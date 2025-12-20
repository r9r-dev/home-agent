/**
 * Update module for Claude Proxy SDK
 * Handles checking for updates and performing updates for both backend and proxy
 */

import { spawn, type ChildProcess } from "child_process";
import { readFileSync } from "fs";
import { join, dirname } from "path";
import { fileURLToPath } from "url";

const __dirname = dirname(fileURLToPath(import.meta.url));

// Configuration
const GITHUB_REPO = "r9r-dev/home-agent";
const GITHUB_API_URL = `https://api.github.com/repos/${GITHUB_REPO}/releases/latest`;
const COMPOSE_PATH = "/home/share/docker/dockge/stacks/home-agent";

export type LogLevel = "info" | "warning" | "error";

export interface LogEntry {
  timestamp: string;
  message: string;
  level: LogLevel;
  source: "backend" | "proxy";
}

export interface VersionInfo {
  current: string;
  latest: string | null;
  updateAvailable: boolean;
}

export interface UpdateStatus {
  backend: VersionInfo;
  proxy: VersionInfo;
}

export type LogCallback = (entry: LogEntry) => void;
export type StatusCallback = (status: "running" | "success" | "error", error?: string) => void;

/**
 * Get the current proxy version from package.json
 */
function getProxyVersion(): string {
  try {
    const packagePath = join(__dirname, "..", "package.json");
    const pkg = JSON.parse(readFileSync(packagePath, "utf-8"));
    return pkg.version || "unknown";
  } catch {
    return "unknown";
  }
}

/**
 * Get the current backend version from the running container
 */
async function getBackendVersion(): Promise<string> {
  return new Promise((resolve) => {
    const proc = spawn("docker", [
      "inspect",
      "--format",
      "{{.Config.Labels.version}}",
      "home-agent",
    ]);

    let output = "";
    proc.stdout.on("data", (data) => {
      output += data.toString();
    });

    proc.on("close", (code) => {
      if (code === 0 && output.trim() && output.trim() !== "<no value>") {
        resolve(output.trim());
      } else {
        // Fallback: try to get image tag
        const imgProc = spawn("docker", [
          "inspect",
          "--format",
          "{{.Config.Image}}",
          "home-agent",
        ]);

        let imgOutput = "";
        imgProc.stdout.on("data", (data) => {
          imgOutput += data.toString();
        });

        imgProc.on("close", () => {
          const match = imgOutput.match(/:([^:]+)$/);
          resolve(match ? match[1].trim() : "latest");
        });
      }
    });

    proc.on("error", () => resolve("unknown"));
  });
}

/**
 * Fetch the latest release version from GitHub
 */
interface GitHubRelease {
  tag_name?: string;
}

async function getLatestVersion(): Promise<string | null> {
  try {
    const response = await fetch(GITHUB_API_URL, {
      headers: {
        Accept: "application/vnd.github.v3+json",
        "User-Agent": "claude-proxy-sdk",
      },
    });

    if (!response.ok) {
      console.error(`GitHub API error: ${response.status}`);
      return null;
    }

    const data = await response.json() as GitHubRelease;
    return data.tag_name || null;
  } catch (error) {
    console.error("Failed to fetch latest version:", error);
    return null;
  }
}

/**
 * Check for available updates
 */
export async function checkForUpdates(): Promise<UpdateStatus> {
  const [proxyVersion, backendVersion, latestVersion] = await Promise.all([
    getProxyVersion(),
    getBackendVersion(),
    getLatestVersion(),
  ]);

  // Normalize versions for comparison (remove 'v' prefix if present)
  const normalize = (v: string) => v.replace(/^v/, "");
  const latest = latestVersion ? normalize(latestVersion) : null;
  const proxyNorm = normalize(proxyVersion);
  const backendNorm = normalize(backendVersion);

  return {
    backend: {
      current: backendVersion,
      latest: latestVersion,
      updateAvailable: latest !== null && backendNorm !== latest,
    },
    proxy: {
      current: proxyVersion,
      latest: latestVersion,
      updateAvailable: latest !== null && proxyNorm !== latest,
    },
  };
}

/**
 * Create a log entry
 */
function createLog(
  message: string,
  source: "backend" | "proxy",
  level: LogLevel = "info"
): LogEntry {
  return {
    timestamp: new Date().toISOString(),
    message,
    level,
    source,
  };
}

/**
 * Execute a command and stream its output
 */
function executeCommand(
  command: string,
  args: string[],
  options: {
    cwd?: string;
    source: "backend" | "proxy";
    onLog: LogCallback;
    sudo?: boolean;
  }
): Promise<{ success: boolean; error?: string }> {
  return new Promise((resolve) => {
    const { cwd, source, onLog, sudo } = options;

    // Prepend sudo if needed
    const finalCommand = sudo ? "sudo" : command;
    const finalArgs = sudo ? [command, ...args] : args;

    onLog(createLog(`Executing: ${finalCommand} ${finalArgs.join(" ")}`, source));

    const proc = spawn(finalCommand, finalArgs, {
      cwd,
      shell: false,
      env: { ...process.env },
    });

    proc.stdout.on("data", (data) => {
      const lines = data.toString().trim().split("\n");
      for (const line of lines) {
        if (line.trim()) {
          onLog(createLog(line, source));
        }
      }
    });

    proc.stderr.on("data", (data) => {
      const lines = data.toString().trim().split("\n");
      for (const line of lines) {
        if (line.trim()) {
          // Detect if it's really an error or just info on stderr
          const level: LogLevel = line.toLowerCase().includes("error") ? "error" : "warning";
          onLog(createLog(line, source, level));
        }
      }
    });

    proc.on("close", (code) => {
      if (code === 0) {
        resolve({ success: true });
      } else {
        resolve({ success: false, error: `Command exited with code ${code}` });
      }
    });

    proc.on("error", (error) => {
      onLog(createLog(`Error: ${error.message}`, source, "error"));
      resolve({ success: false, error: error.message });
    });
  });
}

/**
 * Update the backend (Docker container)
 */
export async function updateBackend(
  onLog: LogCallback,
  onStatus: StatusCallback
): Promise<void> {
  onStatus("running");
  onLog(createLog("Starting backend update...", "backend"));

  try {
    // Step 1: Pull the latest image
    onLog(createLog("Pulling latest image...", "backend"));
    const pullResult = await executeCommand(
      "docker",
      ["compose", "pull"],
      { cwd: COMPOSE_PATH, source: "backend", onLog }
    );

    if (!pullResult.success) {
      onLog(createLog(`Pull failed: ${pullResult.error}`, "backend", "error"));
      onStatus("error", pullResult.error);
      return;
    }

    // Step 2: Recreate container with new image
    onLog(createLog("Recreating container with new image...", "backend"));
    const upResult = await executeCommand(
      "docker",
      ["compose", "up", "-d", "--force-recreate"],
      { cwd: COMPOSE_PATH, source: "backend", onLog }
    );

    if (!upResult.success) {
      onLog(createLog(`Recreate failed: ${upResult.error}`, "backend", "error"));
      onStatus("error", upResult.error);
      return;
    }

    onLog(createLog("Backend update completed successfully", "backend"));
    onStatus("success");
  } catch (error) {
    const message = error instanceof Error ? error.message : "Unknown error";
    onLog(createLog(`Update failed: ${message}`, "backend", "error"));
    onStatus("error", message);
  }
}

/**
 * Update the proxy SDK
 * Note: This will trigger a service restart, so the WebSocket connection will be lost
 */
export async function updateProxy(
  onLog: LogCallback,
  onStatus: StatusCallback
): Promise<void> {
  onStatus("running");
  onLog(createLog("Starting proxy SDK update...", "proxy"));

  try {
    // The install.sh script handles everything:
    // - Stops the service
    // - Downloads new code
    // - Builds
    // - Restarts the service

    // Method 1: Direct execution of install.sh (requires sudo)
    onLog(createLog("Downloading and running install script...", "proxy"));

    // We use bash -c to run the curl | bash pattern
    const result = await executeCommand(
      "bash",
      [
        "-c",
        `curl -fsSL https://raw.githubusercontent.com/${GITHUB_REPO}/main/claude-proxy-sdk/install.sh | sudo bash`,
      ],
      { source: "proxy", onLog }
    );

    if (!result.success) {
      onLog(createLog(`Update failed: ${result.error}`, "proxy", "error"));
      onStatus("error", result.error);
      return;
    }

    onLog(createLog("Proxy SDK update completed successfully", "proxy"));
    onLog(createLog("Service is restarting...", "proxy"));
    onStatus("success");

    // Note: The service will restart, so this process will be terminated
  } catch (error) {
    const message = error instanceof Error ? error.message : "Unknown error";
    onLog(createLog(`Update failed: ${message}`, "proxy", "error"));
    onStatus("error", message);
  }
}

/**
 * Run both updates in sequence (backend first, then proxy)
 */
export async function updateAll(
  onLog: LogCallback,
  onBackendStatus: StatusCallback,
  onProxyStatus: StatusCallback
): Promise<void> {
  // Update backend first
  await updateBackend(onLog, onBackendStatus);

  // Small delay to let backend stabilize
  await new Promise((resolve) => setTimeout(resolve, 3000));

  // Then update proxy (this will restart the service)
  await updateProxy(onLog, onProxyStatus);
}

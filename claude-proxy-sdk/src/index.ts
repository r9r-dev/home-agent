/**
 * Claude Proxy SDK
 * A WebSocket proxy for Claude Agent SDK
 *
 * This service replaces the Go-based claude-proxy with a TypeScript implementation
 * using the official Claude Agent SDK for improved features and maintainability.
 */

import Fastify from "fastify";
import { registerWebSocket } from "./websocket.js";
import { generateTitle } from "./claude.js";
import { checkForUpdates, updateBackend, updateProxy, type LogEntry } from "./update.js";
import type { ProxyConfig } from "./types.js";

// Load configuration from environment
function loadConfig(): ProxyConfig {
  const config: ProxyConfig = {
    port: parseInt(process.env.PROXY_PORT || "9090", 10),
    host: process.env.PROXY_HOST || "0.0.0.0",
    apiKey: process.env.PROXY_API_KEY,
  };

  console.log("Configuration loaded:");
  console.log(`  Host: ${config.host}`);
  console.log(`  Port: ${config.port}`);
  console.log(`  API Key: ${config.apiKey ? "configured" : "not configured"}`);

  return config;
}

// Main entry point
async function main() {
  console.log("Starting Claude Proxy SDK...");

  const config = loadConfig();

  // Create Fastify server
  const app = Fastify({
    logger: {
      level: "info",
      transport: {
        target: "pino-pretty",
        options: {
          colorize: true,
          translateTime: "HH:MM:ss",
          ignore: "pid,hostname",
        },
      },
    },
  });

  // CORS middleware
  app.register(import("@fastify/cors"), {
    origin: "*",
    methods: ["GET", "POST", "OPTIONS"],
    allowedHeaders: ["Origin", "Content-Type", "Accept", "X-API-Key"],
  });

  // Health check endpoint (no auth required)
  app.get("/health", async () => {
    return {
      status: "ok",
      service: "claude-proxy-sdk",
      timestamp: new Date().toISOString(),
    };
  });

  // Title generation endpoint (with auth)
  app.post<{
    Body: { user_message: string; assistant_response: string };
  }>("/api/title", async (request, reply) => {
    // Check API key if configured
    if (config.apiKey) {
      const providedKey = request.headers["x-api-key"];
      if (providedKey !== config.apiKey) {
        return reply.status(401).send({ error: "Unauthorized" });
      }
    }

    const { user_message, assistant_response } = request.body;

    if (!user_message || !assistant_response) {
      return reply.status(400).send({ error: "Missing required fields" });
    }

    try {
      const title = await generateTitle(user_message, assistant_response);
      return { title };
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to generate title";
      console.error(`Title generation error: ${errorMessage}`);
      return reply.status(500).send({ error: errorMessage });
    }
  });

  // Update check endpoint
  app.get("/api/update/check", async (request, reply) => {
    // Check API key if configured
    if (config.apiKey) {
      const providedKey = request.headers["x-api-key"];
      if (providedKey !== config.apiKey) {
        return reply.status(401).send({ error: "Unauthorized" });
      }
    }

    try {
      const status = await checkForUpdates();
      return status;
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to check for updates";
      console.error(`Update check error: ${errorMessage}`);
      return reply.status(500).send({ error: errorMessage });
    }
  });

  // Store for active update WebSocket connections
  const updateClients = new Set<import("@fastify/websocket").WebSocket>();

  // Update backend endpoint
  app.post("/api/update/backend", async (request, reply) => {
    // Check API key if configured
    if (config.apiKey) {
      const providedKey = request.headers["x-api-key"];
      if (providedKey !== config.apiKey) {
        return reply.status(401).send({ error: "Unauthorized" });
      }
    }

    // Start update in background and stream logs to all connected WebSocket clients
    updateBackend(
      (entry: LogEntry) => {
        const message = JSON.stringify({ type: "update_log", ...entry });
        for (const client of updateClients) {
          if (client.readyState === 1) { // OPEN
            client.send(message);
          }
        }
      },
      (status, error) => {
        const message = JSON.stringify({ type: "update_status", target: "backend", status, error });
        for (const client of updateClients) {
          if (client.readyState === 1) {
            client.send(message);
          }
        }
      }
    );

    return { started: true, message: "Backend update started. Watch WebSocket for logs." };
  });

  // Update proxy endpoint
  app.post("/api/update/proxy", async (request, reply) => {
    // Check API key if configured
    if (config.apiKey) {
      const providedKey = request.headers["x-api-key"];
      if (providedKey !== config.apiKey) {
        return reply.status(401).send({ error: "Unauthorized" });
      }
    }

    // Start update in background
    updateProxy(
      (entry: LogEntry) => {
        const message = JSON.stringify({ type: "update_log", ...entry });
        for (const client of updateClients) {
          if (client.readyState === 1) {
            client.send(message);
          }
        }
      },
      (status, error) => {
        const message = JSON.stringify({ type: "update_log", source: "proxy", status, error });
        for (const client of updateClients) {
          if (client.readyState === 1) {
            client.send(message);
          }
        }
      }
    );

    return { started: true, message: "Proxy update started. Service will restart." };
  });

// Register WebSocket handler (includes plugin registration)
  registerWebSocket(app, config.apiKey, updateClients);

  // Graceful shutdown
  const signals: NodeJS.Signals[] = ["SIGINT", "SIGTERM"];
  for (const signal of signals) {
    process.on(signal, async () => {
      console.log(`\nReceived ${signal}, shutting down gracefully...`);
      await app.close();
      process.exit(0);
    });
  }

  // Start server
  try {
    await app.listen({ port: config.port, host: config.host });
    console.log(`Server started on http://${config.host}:${config.port}`);
    console.log(`WebSocket endpoint: ws://${config.host}:${config.port}/ws`);
    console.log(`Health check: http://${config.host}:${config.port}/health`);
  } catch (error) {
    console.error("Failed to start server:", error);
    process.exit(1);
  }
}

main().catch((error) => {
  console.error("Fatal error:", error);
  process.exit(1);
});

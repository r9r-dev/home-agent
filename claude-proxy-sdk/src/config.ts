/**
 * Shared configuration module for Claude Proxy SDK
 */

import type { McpServerConfig } from "./types.js";

// Singleton configuration store
let mcpServers: Record<string, McpServerConfig> | undefined;

/**
 * Set the MCP servers configuration
 */
export function setMcpServers(servers: Record<string, McpServerConfig> | undefined): void {
  mcpServers = servers;
  if (servers) {
    console.log(`[Config] MCP servers configured: ${Object.keys(servers).join(", ")}`);
    for (const [name, config] of Object.entries(servers)) {
      console.log(`[Config]   - ${name}: ${config.url} (${config.transport})`);
    }
  }
}

/**
 * Get the MCP servers configuration
 */
export function getMcpServers(): Record<string, McpServerConfig> | undefined {
  return mcpServers;
}

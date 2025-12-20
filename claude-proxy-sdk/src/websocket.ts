/**
 * WebSocket handler for Claude Proxy
 * Compatible with the existing Go backend protocol
 */

import type { FastifyInstance, FastifyRequest } from "fastify";
import type { WebSocket } from "@fastify/websocket";
import type { ProxyRequest, ProxyResponse } from "./types.js";
import { executePrompt } from "./claude.js";
import { auditLog } from "./hooks/audit.js";

/**
 * Register WebSocket routes
 */
export function registerWebSocket(
  app: FastifyInstance,
  apiKey?: string,
  updateClients?: Set<WebSocket>
): void {
  app.register(import("@fastify/websocket"));

  app.register(async (fastify) => {
    // Main Claude WebSocket endpoint
    fastify.get(
      "/ws",
      { websocket: true },
      (socket: WebSocket, request: FastifyRequest) => {
        // Validate API key if configured
        if (apiKey) {
          const providedKey = request.headers["x-api-key"];
          if (providedKey !== apiKey) {
            console.log(
              `[WS] Unauthorized connection attempt from ${request.ip}`
            );
            socket.close(4001, "Unauthorized");
            return;
          }
        }

        console.log(`[WS] New connection from ${request.ip}`);

        socket.on("message", async (data: Buffer) => {
          try {
            const message = data.toString();
            const request = JSON.parse(message) as ProxyRequest;

            console.log(
              `[WS] Request: type=${request.type}, prompt_len=${request.prompt?.length || 0}, ` +
                `session=${request.session_id || "none"}, model=${request.model || "default"}`
            );

            if (request.type === "execute") {
              await handleExecute(socket, request);
            } else {
              sendError(socket, `Unknown request type: ${request.type}`);
            }
          } catch (error) {
            const errorMessage =
              error instanceof Error ? error.message : "Parse error";
            console.error(`[WS] Error processing message: ${errorMessage}`);
            sendError(socket, errorMessage);
          }
        });

        socket.on("close", () => {
          console.log(`[WS] Connection closed from ${request.ip}`);
        });

        socket.on("error", (error: Error) => {
          console.error(`[WS] Socket error: ${error.message}`);
        });
      }
    );

    // Update WebSocket endpoint for streaming update logs
    if (updateClients) {
      fastify.get(
        "/ws/update",
        { websocket: true },
        (socket: WebSocket, request: FastifyRequest) => {
          // Validate API key via query param
          if (apiKey) {
            const url = new URL(request.url, `http://${request.headers.host}`);
            const providedKey = url.searchParams.get("key");
            if (providedKey !== apiKey) {
              console.log(`[WS/Update] Unauthorized connection attempt`);
              socket.close(4001, "Unauthorized");
              return;
            }
          }

          console.log(`[WS/Update] Client connected from ${request.ip}`);
          updateClients.add(socket);

          socket.on("close", () => {
            console.log(`[WS/Update] Client disconnected from ${request.ip}`);
            updateClients.delete(socket);
          });

          socket.on("error", (error: Error) => {
            console.error(`[WS/Update] Socket error: ${error.message}`);
            updateClients.delete(socket);
          });
        }
      );
    }
  });
}

/**
 * Handle execute request
 */
async function handleExecute(
  socket: WebSocket,
  request: ProxyRequest
): Promise<void> {
  if (!request.prompt) {
    sendError(socket, "Prompt is required");
    return;
  }

  auditLog({
    timestamp: new Date(),
    sessionId: request.session_id,
    event: "SessionStart",
    details: {
      model: request.model,
      thinking: request.thinking,
      is_new_session: request.is_new_session,
    },
  });

  try {
    // Stream responses from Claude Agent SDK
    for await (const response of executePrompt(request)) {
      sendResponse(socket, response);

      // Stop streaming after done or error
      if (response.type === "done" || response.type === "error") {
        break;
      }
    }
  } catch (error) {
    const errorMessage =
      error instanceof Error ? error.message : "Execution failed";
    console.error(`[WS] Execution error: ${errorMessage}`);
    sendError(socket, errorMessage);
  }

  auditLog({
    timestamp: new Date(),
    sessionId: request.session_id,
    event: "SessionEnd",
  });
}

/**
 * Send a response to the client
 */
function sendResponse(socket: WebSocket, response: ProxyResponse): void {
  try {
    socket.send(JSON.stringify(response));
  } catch (error) {
    console.error(`[WS] Failed to send response: ${error}`);
  }
}

/**
 * Send an error response
 */
function sendError(socket: WebSocket, message: string): void {
  sendResponse(socket, {
    type: "error",
    error: message,
  });
}

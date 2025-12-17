# Claude Proxy SDK

Remplacement du proxy Claude Code basé sur Go par une implémentation TypeScript utilisant le Claude Agent SDK officiel.

## Installation rapide (Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/r9r-dev/home-agent/main/claude-proxy-sdk/install.sh | sudo bash
```

Le script:
- Installe Node.js 20+ si necessaire
- Telecharge et compile le projet
- Cree un service systemd
- Genere une cle API
- Demarre le service

### Desinstallation

```bash
curl -fsSL https://raw.githubusercontent.com/r9r-dev/home-agent/main/claude-proxy-sdk/install.sh | sudo bash -s -- --uninstall
```

## Installation manuelle

### Prerequis

- Node.js >= 20.0.0
- Claude Code CLI installe (`claude --version`)
- Cle API Anthropic (`ANTHROPIC_API_KEY`)

### Etapes

```bash
# Cloner le projet
git clone https://github.com/r9r-dev/home-agent.git
cd home-agent/claude-proxy-sdk

# Installer les dependances
npm install

# Compiler
npm run build

# Demarrer en mode developpement
npm run dev

# Ou demarrer en production
npm start
```

## Avantages par rapport au proxy Go

| Aspect | Proxy Go (ancien) | Proxy SDK (nouveau) |
|--------|-------------------|---------------------|
| Exécution | `exec.Command` sur CLI | SDK natif |
| Parsing | Manuel du JSON stream | Géré par le SDK |
| Maintenance | Code custom à maintenir | SDK maintenu par Anthropic |
| Features | Basiques | Hooks, MCP, Subagents |
| Sessions | Gestion manuelle | Intégrée au SDK |

## Prérequis

- Node.js >= 20.0.0
- Claude Code CLI installé (`claude --version`)
- Clé API Anthropic (`ANTHROPIC_API_KEY`)

## Installation

```bash
cd claude-proxy-sdk
npm install
```

## Configuration

Variables d'environnement:

| Variable | Description | Défaut |
|----------|-------------|--------|
| `PROXY_PORT` | Port d'écoute | `9090` |
| `PROXY_HOST` | Adresse d'écoute | `0.0.0.0` |
| `PROXY_API_KEY` | Clé API pour authentification | (vide) |
| `ANTHROPIC_API_KEY` | Clé API Anthropic | (requis) |

## Utilisation

```bash
# Développement
npm run dev

# Production
npm run build
npm start
```

## Endpoints

### WebSocket `/ws`

Protocole identique au proxy Go:

**Request:**
```json
{
  "type": "execute",
  "prompt": "...",
  "session_id": "uuid",
  "is_new_session": true,
  "model": "sonnet",
  "custom_instructions": "...",
  "thinking": false
}
```

**Responses:**
```json
{"type": "chunk", "content": "..."}
{"type": "thinking", "content": "..."}
{"type": "session_id", "session_id": "..."}
{"type": "done", "content": "...", "session_id": "..."}
{"type": "error", "error": "..."}
```

### REST

- `GET /health` - Health check
- `POST /api/title` - Génération de titre (body: `{user_message, assistant_response}`)

## Architecture

```
src/
├── index.ts      # Point d'entrée Fastify
├── websocket.ts  # Handler WebSocket
├── claude.ts     # Wrapper Agent SDK
├── types.ts      # Types TypeScript
└── hooks/
    └── audit.ts  # Logging d'audit
```

## Hooks

Le SDK permet d'exécuter du code à différents moments:

- `PreToolUse` - Avant l'utilisation d'un outil
- `PostToolUse` - Après l'utilisation d'un outil
- `SessionStart` - Début de session
- `SessionEnd` - Fin de session

Exemple actuel: logging des commandes Bash et modifications de fichiers.

## Migration depuis le proxy Go

1. Arrêter le service `claude-proxy` Go
2. Configurer les variables d'environnement
3. Lancer `npm start` dans `claude-proxy-sdk/`
4. Le backend Home Agent se connecte de la même façon

Le protocole WebSocket est 100% compatible.

## Évolutions futures

Avec le Agent SDK, ces fonctionnalités sont maintenant possibles:

### MCP (Model Context Protocol)

```typescript
mcpServers: {
  postgres: { command: "npx", args: ["@modelcontextprotocol/server-postgres"] },
  playwright: { command: "npx", args: ["@playwright/mcp@latest"] }
}
```

### Subagents

```typescript
allowedTools: ["Read", "Glob", "Grep", "Task"]  // Task = subagents
```

### Hooks personnalisés

```typescript
hooks: {
  PreToolUse: [{
    matcher: "Bash",
    hooks: [{ type: "command", command: "validate-command.sh $TOOL_INPUT" }]
  }]
}
```

## Logs d'audit

Les actions sont loggées dans:
- Console (stdout)
- `/tmp/claude-audit.log` (via hooks shell)
- Buffer mémoire (via `getRecentAuditLogs()`)

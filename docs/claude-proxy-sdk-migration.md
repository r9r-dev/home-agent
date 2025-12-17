# Migration vers Claude Proxy SDK

Ce document décrit la migration du proxy Claude Code Go vers le Claude Agent SDK TypeScript.

## Contexte

Le proxy original (`claude-proxy/`) est écrit en Go et exécute le CLI Claude Code via `exec.Command`. Cette approche fonctionne mais présente des limitations:

1. **Parsing manuel** du JSON stream de Claude CLI
2. **Maintenance** du code de parsing à chaque évolution du CLI
3. **Fonctionnalités limitées** aux options du CLI

## Solution: Claude Agent SDK

Le Claude Agent SDK est le SDK officiel d'Anthropic pour construire des agents. Il offre:

- Exécution native (pas de subprocess)
- Streaming natif via async iterators
- Gestion de session intégrée
- Hooks pour personnalisation
- MCP (Model Context Protocol) pour extensions
- Subagents pour tâches complexes

## Comparaison architecturale

### Avant (Go + CLI)

```
[Home Agent Backend] --WebSocket--> [Claude Proxy Go] --exec.Command--> [Claude CLI]
                                           |
                                    Parse JSON stream
```

### Après (TypeScript + SDK)

```
[Home Agent Backend] --WebSocket--> [Claude Proxy SDK] --SDK--> [Claude Agent]
                                           |
                                    Native streaming
```

## Compatibilité

Le nouveau proxy est **100% compatible** avec le backend Home Agent existant:

- Même protocole WebSocket
- Mêmes types de messages (`chunk`, `thinking`, `done`, `error`, `session_id`)
- Mêmes endpoints REST (`/health`, `/api/title`)
- Mêmes variables d'environnement

## Nouvelles fonctionnalités disponibles

### 1. Hooks

Exécuter du code avant/après les actions de l'agent:

```typescript
hooks: {
  PreToolUse: [{
    matcher: "Bash",
    hooks: [{
      type: "command",
      command: "security-check.sh $TOOL_INPUT"
    }]
  }],
  PostToolUse: [{
    matcher: "Edit|Write",
    hooks: [{
      type: "command",
      command: "echo 'File modified' >> audit.log"
    }]
  }]
}
```

Cas d'usage:
- Validation des commandes avant exécution
- Logging des modifications
- Notification (Slack, email) sur certaines actions
- Blocage de commandes dangereuses

### 2. MCP (Model Context Protocol)

Connecter des serveurs externes:

```typescript
mcpServers: {
  // Base de données PostgreSQL
  postgres: {
    command: "npx",
    args: ["@modelcontextprotocol/server-postgres", "postgresql://localhost/mydb"]
  },

  // Automatisation browser
  playwright: {
    command: "npx",
    args: ["@playwright/mcp@latest"]
  },

  // Mémoire persistante
  memory: {
    command: "npx",
    args: ["@modelcontextprotocol/server-memory"]
  }
}
```

Cas d'usage:
- Requêtes SQL directes
- Screenshots et monitoring web
- Mémoire entre sessions
- Intégration API tierces

### 3. Subagents

Déléguer des tâches complexes:

```typescript
allowedTools: ["Read", "Glob", "Grep", "Task"]
```

Claude peut alors spawner des sous-agents pour:
- Analyse de sécurité parallèle
- Review de code sur plusieurs fichiers
- Recherches approfondies

### 4. Permissions granulaires

Contrôle fin des outils:

```typescript
// Agent read-only
allowedTools: ["Read", "Glob", "Grep"]
permissionMode: "bypassPermissions"

// Agent avec édition
allowedTools: ["Read", "Write", "Edit", "Bash"]
permissionMode: "acceptEdits"
```

## Procédure de migration

### Étape 1: Installation

```bash
cd claude-proxy-sdk
npm install
```

### Étape 2: Configuration

```bash
export ANTHROPIC_API_KEY=sk-ant-...
export PROXY_PORT=9090
export PROXY_API_KEY=your-key  # optionnel
```

### Étape 3: Test

```bash
npm run dev
# Vérifier http://localhost:9090/health
```

### Étape 4: Déploiement

Option A: Remplacer le service Go
```bash
# Arrêter l'ancien
systemctl stop claude-proxy

# Créer nouveau service
# /etc/systemd/system/claude-proxy-sdk.service
```

Option B: Exécuter en parallèle sur un autre port
```bash
PROXY_PORT=9091 npm start
# Mettre à jour CLAUDE_PROXY_URL dans Home Agent
```

## Rollback

En cas de problème, revenir à l'ancien proxy:

```bash
cd claude-proxy
go run .
```

Le backend Home Agent se reconnecte automatiquement.

## Évolutions futures prévues

1. **Interface d'administration** des hooks via UI
2. **Marketplace MCP** pour installer des serveurs
3. **Métriques** d'utilisation des tools
4. **Rate limiting** par session/utilisateur
5. **Templates d'agents** préconfigurés

## Références

- [Claude Agent SDK Documentation](https://platform.claude.com/docs/en/agent-sdk/overview)
- [MCP Servers Repository](https://github.com/modelcontextprotocol/servers)
- [Claude Code CLI Documentation](https://docs.anthropic.com/en/docs/claude-code)

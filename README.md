# Home Agent

[![Architecture Refactoring](https://img.shields.io/endpoint?style=for-the-badge&url=https://raw.githubusercontent.com/r9r-dev/home-agent/main/.github/badges/milestone-1.json)](https://github.com/r9r-dev/home-agent/milestone/1)

**Assistant domotique personnel propulse par Claude Code**

Home Agent est une interface web auto-hébergée qui transforme Claude Code en assistant domotique intelligent. Contrairement à un simple chatbot, Home Agent peut exécuter des commandes, contrôler vos appareils, analyser des images et interagir avec votre infrastructure.

## Pourquoi Home Agent ?

- **Exécution locale** : Claude Code s'exécute sur votre machine, avec accès à vos fichiers et outils
- **Contrôle de votre infrastructure** : Gérez vos serveurs, conteneurs et appareils connectés
- **Vision** : Analysez des images, captures d'écran ou flux de caméras
- **Mémoire persistante** : Home Agent se souvient de vos préférences et contexte
- **Extensible** : Connectez vos propres APIs et services
- **Vie privée** : Vos données restent chez vous

## Fonctionnalités

| Disponible | En développement |
|------------|------------------|
| Chat en temps réel (WebSocket) | Recherche dans les conversations |
| Historique des conversations | Instructions personnalisées |
| Génération automatique de titres | Mémoire persistante |
| Mode local et proxy | Gestion de machines SSH |
| Interface responsive | Affichage des outils utilisés |
| | Upload d'images et fichiers |
| | Intégration caméras |
| | Mode projet avec contexte |

Voir les [issues GitHub](https://github.com/r9r-dev/home-agent/issues) pour la roadmap complète.

## Démarrage rapide

### Avec Docker

```bash
docker pull ghcr.io/r9r-dev/home-agent:latest
docker run -d -p 8080:8080 \
  -e CLAUDE_PROXY_URL=http://HOST_IP:9090 \
  -e CLAUDE_PROXY_KEY=your_key \
  -v homeagent-data:/app/data \
  ghcr.io/r9r-dev/home-agent:latest
```

> **Note** : L'image Docker nécessite un [Claude Proxy](docs/claude-proxy.md) sur l'hôte pour exécuter Claude Code.

### Sans Docker

```bash
git clone https://github.com/r9r-dev/home-agent.git
cd home-agent
cp .env.example .env
# Configurer ANTHROPIC_API_KEY dans .env

./start-dev.sh
```

Ouvrir http://localhost:5173

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Home Agent                           │
├─────────────┬─────────────────────┬─────────────────────┤
│  Frontend   │      Backend        │    Claude Proxy     │
│  (Svelte)   │    (Go/Fiber)       │   (optionnel)       │
├─────────────┼─────────────────────┼─────────────────────┤
│ • Chat UI   │ • WebSocket server  │ • Exécute Claude    │
│ • Sidebar   │ • Session manager   │   CLI sur l'hôte    │
│ • Settings  │ • SQLite database   │ • Pour containers   │
└─────────────┴─────────────────────┴─────────────────────┘
                        │
                        ▼
              ┌─────────────────┐
              │   Claude Code   │
              │   CLI + Outils  │
              └─────────────────┘
```

## Documentation

- [Guide d'installation](docs/USER_MANUAL.md)
- [Configuration du proxy](docs/claude-proxy.md)
- [Développement](docs/development.md)

## Licence

MIT

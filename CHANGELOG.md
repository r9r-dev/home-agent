# Changelog

## v1.0.0 (2025-12-21)

- Ajout de la recherche textuelle complète dans les conversations

## v0.21.13 (2025-12-21)

- Ajout de documentation CLAUDE.md pour le Claude Proxy SDK
- Correction : prévention de la boucle infinie de mise à jour après redémarrage du proxy

## v0.21.12 (2025-12-21)

- Amélioration des icônes de sélection de machine

## v0.21.11 (2025-12-21)

Corrections et ameliorations mineures

## v0.21.10 (2025-12-21)

- Correction du versioning des releases pour assurer la cohérence des numéros de version
- Déplacement des dépendances TypeScript de devDependencies vers dependencies pour les builds de production

## v0.21.9 (2025-12-21)

- Déplacement de TypeScript vers les dépendances de production pour les builds du proxy
- Mise à jour des versions vers 0.21.9

## v0.21.8 (2025-12-21)

- Mise à jour des versions vers 0.21.8
- Correction du CHANGELOG.md corrompu

## v0.21.6 (2025-12-21)
- Mise à jour vers la version 0.21.6
- Correction des commandes npm pour les mises à jour
- Inclusion des dépendances de développement lors des mises à jour du système

## v0.21.5 (2025-12-21)
- Mise à jour des versions frontend et claude-proxy-sdk vers 0.21.5
- Résolution définitive des problèmes de permissions lors de la mise à jour du proxy

## v0.21.4 (2025-12-21)
- Mise à jour des versions frontend et claude-proxy-sdk vers 0.21.4

## v0.21.3 (2025-12-21)
- Mise a jour mineure

## v0.21.2 (2025-12-21)
- Ajout du mode Auto pour la sélection des machines SSH
- Claude peut désormais choisir automatiquement la machine appropriée selon le contexte
- Sélection intelligente basée sur les paramètres de la conversation

## v0.21.1 (2025-12-21)
- Correction de la boucle infinie dans la boîte de dialogue de mise à jour
- Suppression de sudo et ajout de délais d'expiration dans la mise à jour du proxy

## v0.21.0 (2025-12-21)
- Interface utilisateur complètement repensée avec support optimisé pour les appareils mobiles
- Amélioration générale de la mise en page pour une meilleure expérience utilisateur sur tous les appareils
- Responsive design adapté aux écrans de petite et grande taille
- Optimisation de la navigation et de l'accessibilité

## v0.20.3 (2025-12-20)
- Mise à jour de la version vers 0.20.3
- Correction : utilisation de sudo pour les opérations de fichiers lors de la mise à jour du proxy

## v0.20.2 (2025-12-20)
- Correction : Réinitialisation du contexte d'exécution lors d'un nouveau tour de conversation
- Documentation : Consolidation de CLAUDE.md - suppression de l'historique des versions et réorganisation de la structure

## v0.20.1 (2025-12-20)
- Correction du défilement automatique dans les dialogues de mise à jour
- Correction de la mise à jour du proxy sans nécessiter les droits sudo
- Mise à jour de la version en 0.20.1

## v0.20.0 (2025-12-20)
- Ajout d'une interface complète de gestion des machines SSH dans les paramètres
- Permet de configurer plusieurs connexions SSH distantes (hôte, port, identifiant, authentification)
- Support de l'authentification par mot de passe ou clé SSH
- Test de connexion avec affichage de la latence
- Sélecteur de machine dans la zone de saisie pour cibler l'exécution à distance
- Chiffrement des identifiants SSH avec AES-256-GCM
- Injection automatique du contexte SSH dans les messages envoyés à Claude
- Stockage des machines dans la base de données SQLite
- Import/export des configurations de machines
- Résolution de l'issue #7

## v0.19.5 (2025-12-20)
- Correction de la reconnexion WebSocket pendant les mises à jour du backend
- Ajout d'un label de version dans le Dockerfile pour l'affichage correct de la version
- Implémentation du sondage de santé du backend pour attendre la disponibilité après redémarrage
- Déclenchement automatique de la mise à jour du proxy après reconnexion du backend
- Ajout d'un indicateur de reconnexion dans l'interface de la boîte de dialogue de mise à jour

## v0.19.4 (2025-12-20)
- Correction de la reconnexion WebSocket et affichage de la version
- Actualisation automatique de la barre latérale lors de la mise à jour des titres
- Séparation des blocs de réflexion pour une meilleure lisibilité
- Ajout de la fonctionnalité de mise à jour système avec vérification des versions et journaux en direct
- Simplification de l'en-tête avec suppression du badge de connexion et notifications par toast
- Indicateur de journaux en temps réel avec correction de la persistance des blocs de réflexion
- Affichage chronologique des appels d'outils avec amélioration de l'UX
- Correction de l'analyse des résultats d'outils provenant des messages utilisateur du SDK
- Correction de l'ordre et de l'initialisation de l'état des appels d'outils
- Amélioration de l'affichage des entrées d'appels d'outils et de l'ordre chronologique

## v0.19.3 (2025-12-20)
- Correction de la reconnexion WebSocket pour une meilleure fiabilité
- Mise à jour de l'affichage de la version

## v0.19.2 (2025-12-20)
- Mise à jour de la version vers 0.19.2
- Mise à jour des dépendances du SDK proxy et du frontend

## v0.19.1 (2025-12-20)
- Mise a jour mineure

## v0.19.0 (2025-12-20)
- Correction automatique de la barre latérale lors de la mise à jour des titres de conversations
- Séparation des blocs de réflexion en blocs individuels et affichage chronologique
- Amélioration de l'interface utilisateur pour les réponses avec mode de réflexion étendu

## v0.18.0 (2025-12-20)
- Ajout d'une fonctionnalité de mise à jour système avec vérification des versions et affichage des logs en direct
- Simplification de l'en-tête en supprimant le badge de connexion et ajout de notifications toast
- Ajout d'un indicateur de logs en temps réel et correction de la persistance des blocs de réflexion

## v0.17.4 (2025-12-19)
- Affichage des appels d'outils dans l'ordre chronologique
- Amélioration de l'expérience utilisateur pour la visualisation des appels d'outils

## v0.17.3 (2025-12-19)
- Correction de l'analyse des résultats d'outils provenant des messages utilisateur du SDK
- Les résultats d'outils sont maintenant correctement extraits et traités des réponses du SDK

## v0.17.2 (2025-12-18)
- Correction de l'ordre d'affichage des appels d'outils
- Correction de l'initialisation de l'état des appels d'outils

## v0.17.1 (2025-12-18)
- Correction de l'affichage des entrées des appels d'outils et de leur ordre chronologique
- Documentation : Ajout des options d'authentification pour Claude Agent SDK
- Documentation : Ajout du lien vers la documentation de Claude Agent SDK

## v0.17.0 (2025-12-18)
- Affichage intégré des appels d'outils dans le flux des messages
- Visualisation des outils utilisés par Claude avec leur statut (en cours, succès, erreur)
- Temps d'exécution visible pour chaque appel d'outil
- Blocs d'outils repliables pour les entrées/sorties détaillées
- Code couleur : bleu (en cours), vert (succès), rouge (erreur)
- Icônes spécifiques par type d'outil (terminal, fichier, recherche, tâches, web)
- Chargement différé des détails des outils pour meilleure performance

## v0.16.4 (2025-12-18)
- Migration de la gestion des sessions : les identifiants de session sont désormais générés par Claude Agent SDK au lieu d'être créés localement
- Simplification de la logique de session dans le backend
- Meilleure cohérence avec l'architecture du proxy Claude

## v0.16.3 (2025-12-18)
- Correction du bloc de réflexion (thinking) pour améliorer l'affichage et la stabilité
- Suppression de la dépendance à hold proxy, simplifiant l'architecture

## v0.16.2 (2025-12-18)
- Affichage du bloc de réflexion au-dessus du message de Claude pour une meilleure lisibilité
- Configuration de SQLite pour supporter l'accès concurrent avec WAL (Write-Ahead Logging) et timeout d'attente
- Ajout de log pour les erreurs de l'endpoint /api/sessions afin de faciliter le débogage

## v0.16.1 (2025-12-18)
- Refactorisation des alias de modèles pour utiliser les noms simples (haiku, sonnet, opus)
- Mise à jour vers les derniers modèles Claude (Haiku/Sonnet/Opus 4.5)
- Mise à jour de Node.js vers la version 24+ (LTS actuelle)
- Augmentation de la version minimale requise de Node.js à 22 (LTS actuelle)

## v0.16.0 (2025-12-17)
- Migration complète du service proxy de Go vers TypeScript utilisant le Claude Agent SDK
- Remplacement de l'implémentation personnalisée par l'SDK officiel d'Anthropic pour une meilleure maintenance et stabilité
- Amélioration de la compatibilité avec les fonctionnalités actuelles de Claude
- Simplification de l'architecture en utilisant des dépendances officielles éprouvées

## v0.15.1 (2025-12-17)
- Mode "thinking" désormais fonctionnel pour afficher le processus de raisonnement de Claude
- Nettoyage des logs pour améliorer la lisibilité et les performances

## v0.15.0 (2025-12-17)
- Ajout de l'affichage des pensées de Claude dans l'interface
- Activation/désactivation via le menu Claude
- Bloc de pensées collapsible avec style distinctif (fond ambre)
- Support de plusieurs blocs de pensées affichés chronologiquement
- Persistance des pensées en base de données avec rôle "thinking"

## v0.14.0 (2025-12-17)
- Architecture unifiée basée sur le proxy uniquement
- Suppression du mode local et de tous les chemins d'exécution alternatifs
- Simplification du code backend en supprimant les dépendances CLI locales
- Réduction de la complexité de configuration et de déploiement
- Meilleure maintenabilité et cohérence de l'architecture

## v0.13.3 (2025-12-17)
- Correction de la création de session : l'ID de session est maintenant créé par le SDK et fourni par le frontend, plutôt que généré par le backend
- Amélioration de la gestion des sessions pour une meilleure synchronisation entre frontend et backend

## v0.13.2 (2025-12-17)
- Correction de la génération de sessionId côté frontend et support du flag `--session-id` pour Claude
- Correction des titres de conversation générés en français dans claude-proxy
- Correction de la séparation des réponses Claude en paragraphes distincts

## v0.13.1 (2025-12-17)
- Refonte complète du menu avec une nouvelle structure organisée en trois sections : Claude (modèle, thinking, instructions), Host (machine SSH, logs), et Paramètres (mémoire, paramètres généraux)
- Correction et refactorisation des chemins d'upload pour utiliser le répertoire /workspace et les GUID complets des fichiers

## v0.13.0 (2025-12-17)
- Ajout d'une barre de navigation Menubar pour améliorer la navigation dans l'interface
- Implémentation de la persistance de la mémoire utilisateur
- Sauvegarde automatique des données de mémoire
- Restauration de la mémoire lors du redémarrage de l'application

## v0.12.3 (2025-12-17)
- Correction du mapping WORKSPACE_PATH pour créer automatiquement le sous-dossier uploads
- Résout les problèmes de chemin lors du téléchargement de fichiers dans l'environnement conteneurisé

## v0.12.2 (2025-12-17)
- Ajout de la variable d'environnement `WORKSPACE_PATH` pour mapper les chemins des fichiers accessibles à Claude
- Permet une meilleure intégration entre le conteneur et l'hôte lors de l'utilisation de Claude CLI
- Facilite l'accès aux fichiers uploadés et aux ressources du workspace depuis Claude

## v0.12.1 (2025-12-17)
- Corrections de l'interface utilisateur pour une meilleure cohérence visuelle
- Améliorations des fonctionnalités existantes
- Corrections de bugs mineurs affectant l'expérience utilisateur

## v0.12.0 (2025-12-17)
- Ajout d'un menu de configuration permettant l'édition des instructions personnalisées
- Résolution de l'issue #2 avec mise en place de la fonctionnalité
- Documentation mise à jour avec les détails de la v0.11.0 concernant les uploads de fichiers

## v0.11.0 (2025-12-17)
- Ajout de la fonctionnalité d'upload de fichiers et d'images dans les conversations
- Transmission du contenu des fichiers à Claude CLI
- Refonte du README pour mieux positionner Home Agent comme assistant domotique
- Documentation des modifications de style de scrollbar dans CLAUDE.md

## v0.10.5 (2025-12-16)
- Correction de la visibilité de la barre de défilement en utilisant les attributs de données bits-ui
- Amélioration de l'affichage du composant ScrollArea pour une meilleure cohérence visuelle
- Application des styles personnalisés de barre de défilement via les sélecteurs bits-ui

## v0.10.4 (2025-12-16)
- Correction de la barre de défilement via des modifications de composants
- Ajout de la documentation des modifications personnalisées dans CLAUDE.md
- Les composants `scroll-area.svelte` et `scroll-area-scrollbar.svelte` doivent être reconfiguré après les mises à jour de shadcn-svelte pour maintenir la barre de défilement visible

## v0.10.3 (2025-12-16)
- Déplacement des styles de scrollbar vers app.css pour une meilleure pérennité et faciliter les mises à jour futures de shadcn-svelte

## v0.10.2 (2025-12-16)
- Correction de la visibilité de la scrollbar avec bits-ui type="always"
- Mise à jour de la documentation du projet (CLAUDE.md)

## v0.10.1 (2025-12-16)
- Correction du poids de police de la barre latérale
- Amélioration de la visibilité de la barre de défilement
- Correction du dialogue de suppression

## v0.10.0 (2025-12-16)
- Amélioration de l'interface de la barre latérale avec bouton de basculement
- État de la barre latérale persistant dans localStorage
- Séparation de la barre latérale en sections actions et historique de conversation
- Migration complète des icônes de Lucide vers MynaUI via @iconify/svelte
- Mise à jour du style du badge de connexion (noir/blanc avec point vert)
- Génération des titres de conversation en français
- Ajout de la police Cal Sans pour les titres de conversation

## v0.9.1 (2025-12-16)
- Correction de la mise en page des messages et de la visibilité de la barre de défilement
- Mise à jour de la documentation CLAUDE.md avec les détails de la pile frontend (Svelte 5, Tailwind CSS v4, shadcn-svelte)

## v0.9.0 (2025-12-16)
- Migration complète du frontend vers shadcn-svelte pour une meilleure cohérence avec l'écosystème Svelte
- Modernisation des composants UI avec les primitives bits-ui
- Amélioration de la maintenabilité et de la flexibilité des composants
- Intégration des icônes MynaUI via @iconify/svelte
- Support complet de Tailwind CSS v4 via le plugin @tailwindcss/vite
- Remplacement des composants personnalisés par des versions shadcn-svelte maintenues
- Meilleure accessibilité et expérience utilisateur grâce aux composants éprouvés de shadcn-svelte

## v0.8.0 (2025-12-16)
- Ajout d'une liste déroulante de sélection de modèle pour choisir entre Haiku, Sonnet et Opus
- Améliorations de l'interface utilisateur et de l'expérience utilisateur

## v0.7.8 (2025-12-16)
- Mise à jour de Go vers la version 1.24 pour assurer la compatibilité avec modernc.org/sqlite

## v0.7.7 (2025-12-16)
- Ajout de Git dans le conteneur pour permettre le téléchargement des dépendances Go
- Correction du processus de construction en incluant Git, nécessaire pour les modules Go distants

## v0.7.6 (2025-12-16)
- Passage à une implémentation SQLite 100% Go (sans CGO) pour accélérer les builds et simplifier le déploiement
- Suppression des dépendances de compilation C, réduisant les temps de build et les problèmes de compatibilité cross-plateforme
- Utilisation de `modernc.org/sqlite` à la place de `github.com/mattn/go-sqlite3` pour une meilleure portabilité

## v0.7.5 (2025-12-16)
- Parallélisation des étapes de build Docker pour accélérer le processus de CI
- Amélioration du temps de compilation et de déploiement via des stages d'optimisation

## v0.7.4 (2025-12-16)
- Correction du rendu markdown avec normalisation des titres et réparation des sauts de ligne

## v0.7.3 (2025-12-16)
- Correction du rendu des sauts de ligne dans les messages de chat
- Correction du script d'installation : arrêt du service avant remplacement du binaire
- Ajout de montages de cache BuildKit pour des compilations plus rapides

## v0.7.2 (2025-12-15)
- Amélioration de l'installateur Claude Proxy avec auto-configuration : génération automatique de clés API, détection de l'adresse IP, démarrage automatique du service et documentation simplifiée
- Support des runners auto-hébergés pour CI/CD : migration vers runners locaux avec cache Docker optimisé
- Simplification de CI pour x86_64 uniquement : suppression des builds arm64 et de QEMU
- Corrections des avertissements CI : configuration du code-splitting Vite et correction du chemin de cache Go
- Injection dynamique de la version depuis le tag git en CI : passage de la version au build Docker et Vite
- Injection de la version depuis package.json au moment du build avec Vite

## v0.7.1 (2025-12-15)
- Correction des avertissements CI relatifs à la taille des chunks et au chemin du cache Go

## v0.7.0 (2025-12-15)
- Amélioration de l'installeur Claude Proxy avec auto-configuration
- Suppression de fichiers inutiles
- Ajout de Claude Proxy pour l'exécution de la CLI sur l'hôte (v0.6.0)
- Ajout de la persistance des sessions Claude et génération dynamique des titres (v0.5.0)
- Ajout de l'historique des conversations avec barre latérale (v0.4.0)
- Ajout du mode skip-permissions, format d'heure 24h et pied de page avec version
- Correction de coquilles
- Amélioration de l'expérience utilisateur et du prompt système
- Correction : détection automatique de l'URL WebSocket depuis la localisation actuelle
- Suppression de fichiers inutiles
- Correction : activation de CGO pour le support SQLite
- Simplification du README et mise à jour du titre frontend
- Nettoyage de la structure du dépôt
- Correction Dockerfile : utilisation de l'utilisateur node existant de node:alpine
- Ajout des variables d'environnement DISABLE_AUTOUPDATER et DISABLE_TELEMETRY
- Ajout de l'intégration continue/déploiement continu GitHub Actions pour le déploiement Docker
- Commit initial : Home Agent - Assistant de gestion d'infrastructure

## v0.6.0 (2025-12-15)
- Architecture entièrement repensée : Claude Proxy SDK externe permet l'exécution de la CLI Claude sur la machine hôte sans l'inclure dans l'image du conteneur
- WebSocket bidirectionnel entre le backend et Claude Proxy pour la communication en temps réel
- Support de l'authentification Claude via clé API ou token OAuth
- Gestion des sessions améliorée avec IDs générés par le SDK
- Streaming des réponses Claude avec support des blocs de réflexion (thinking)
- Affichage des appels d'outils (tool calls) avec métadonnées complètes
- Injection automatique du contexte mémoire et des instructions personnalisées
- Support des machines SSH distantes pour l'exécution de commandes via Claude
- Chiffrement AES-256-GCM des identifiants SSH
- Système de logs en temps réel avec indicateurs de statut (OK/avertissements/erreurs)
- Téléversement de fichiers avec aperçu et validation MIME
- Import/export de mémoire persistante en JSON
- Vérification et mise à jour automatique des versions via GitHub Releases
- Architecture modulaire avec interface ClaudeExecutor pour une flexibilité future

## v0.5.0 (2025-12-15)
- Remplacement du bouton "+ Nouvelle conversation" par un bouton compact "+"
- Ajout de la colonne `claude_session_id` pour supporter correctement l'option `--resume`
- Stockage de l'ID de session Claude CLI séparé de l'ID de session interne
- Génération automatique de titres résumés utilisant Claude (haiku) après la première réponse
- Correction de la persistance des messages pour utiliser les bons ID de session

## v0.4.0 (2025-12-15)
- Historique des conversations avec barre latérale
- Sauvegarde automatique du titre de chaque conversation
- Génération de titres à partir du premier message
- API REST pour gérer les sessions (lister, récupérer, supprimer)
- Chargement des conversations existantes au clic
- Suppression de conversations avec confirmation
- Migration automatique de la base de données

## v0.3.1 (2025-12-15)
- Ajout du mode skip-permissions pour contourner les vérifications de permissions
- Ajout du format d'heure 24h
- Ajout du numéro de version dans le pied de page
- Correction de fautes de frappe

## v0.3.0 (2025-12-15)
- Améliorations de l'expérience utilisateur (UX)
- Mise à jour du système de prompts

## v0.2.2 (2025-12-15)
- Correction de la détection automatique de l'URL WebSocket basée sur l'emplacement courant
- Suppression de fichiers inutiles

## v0.2.1 (2025-12-15)
- Activation du support CGO pour SQLite
- Résout les problèmes de compilation lors de la construction du binaire backend

## v0.2.0 (2025-12-15)
- Simplification de la documentation README pour une meilleure clarté
- Mise à jour du titre de l'application frontend
- Nettoyage de la structure générale du dépôt

## v0.1.2 (2025-12-15)
- Correction du Dockerfile pour utiliser l'utilisateur node existant fourni par l'image node:alpine, éliminant la nécessité de le créer manuellement

## v0.1.1 (2025-12-15)
- Ajout des variables d'environnement `DISABLE_AUTOUPDATER` et `DISABLE_TELEMETRY` pour désactiver respectivement la mise à jour automatique et la télémétrie

## v0.1.0 (2025-12-15)
- Architecture conteneurisée avec Docker pour le déploiement
- Configuration GitHub Actions pour build et push d'images Docker
- Backend Go avec framework Fiber et WebSocket
- Frontend Svelte 5 avec TypeScript et Tailwind CSS
- Intégration Claude Agent SDK via proxy service
- Gestion des sessions et historique des conversations
- Système de mémoire persistant
- Support du Mode Thinking étendu avec affichage du raisonnement Claude
- Visualisation des appels d'outils (Bash, Read, Write, etc.)
- Gestion des machines SSH distantes avec credentials chiffrés
- Upload de fichiers (images et documents) avec preview
- Instructions personnalisées et paramètres de système
- Indicateur de statut en temps réel avec logs
- Vérification et mise à jour des versions système
- API REST complète pour sessions, mémoire et paramètres
- Base de données SQLite pour persistance
- Authentification Claude par clé API ou token OAuth


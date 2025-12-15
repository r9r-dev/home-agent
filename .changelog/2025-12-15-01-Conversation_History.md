# Conversation History Feature (v0.4.0)

## Summary
Implemented conversation history with sidebar navigation, auto-generated titles, and SQLite persistence.

## Backend Changes

### Database (models/database.go)
- Added `title` field to sessions table with migration for existing databases
- Added `ListSessions()` method to retrieve all sessions ordered by last activity
- Added `UpdateSessionTitle()` method to update session titles
- Added `DeleteSession()` method with cascade delete for messages

### Services (services/session.go)
- Added `GenerateTitle()` function to auto-generate titles from first message (max 50 chars)
- Exposed `ListSessions`, `UpdateSessionTitle`, `DeleteSession` through SessionManager

### REST API (main.go)
- `GET /api/sessions` - List all sessions
- `GET /api/sessions/:id` - Get session details
- `GET /api/sessions/:id/messages` - Get messages for a session
- `DELETE /api/sessions/:id` - Delete a session

### Chat Handler (handlers/chat.go)
- Auto-generates title for new sessions from first message content

## Frontend Changes

### API Service (services/api.ts)
- New service with `fetchSessions()`, `fetchMessages()`, `deleteSession()` functions
- TypeScript interfaces for Session and Message types

### Sidebar Component (components/Sidebar.svelte)
- New conversation button
- Session list with title and relative date
- Delete button with confirmation
- Active session highlighting
- Auto-refresh capability

### Chat Store (stores/chatStore.ts)
- Added `loadMessages()` method to load messages from existing sessions

### ChatWindow (components/ChatWindow.svelte)
- Integrated Sidebar component with two-column layout
- Added `handleSelectSession()` to load existing conversations
- Added `handleNewConversation()` to start fresh conversations
- Sidebar auto-refreshes when new sessions are created
- Responsive: sidebar hidden on mobile (< 768px)
- Version bumped to 0.4.0

## Files Modified
- backend/models/database.go
- backend/services/session.go
- backend/handlers/chat.go
- backend/main.go
- frontend/src/services/api.ts (new)
- frontend/src/components/Sidebar.svelte (new)
- frontend/src/stores/chatStore.ts
- frontend/src/components/ChatWindow.svelte

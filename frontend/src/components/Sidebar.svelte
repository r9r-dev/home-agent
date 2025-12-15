<script lang="ts">
  import { onMount } from 'svelte';
  import { fetchSessions, deleteSession, type Session } from '../services/api';

  export let currentSessionId: string | null = null;
  export let onSelectSession: (sessionId: string) => void;
  export let onNewConversation: () => void;

  let sessions: Session[] = [];
  let loading = true;

  async function loadSessions() {
    try {
      sessions = await fetchSessions();
    } catch (error) {
      console.error('Failed to load sessions:', error);
    } finally {
      loading = false;
    }
  }

  async function handleDelete(sessionId: string, event: Event) {
    event.stopPropagation();
    if (!confirm('Supprimer cette conversation ?')) return;

    try {
      await deleteSession(sessionId);
      sessions = sessions.filter(s => s.session_id !== sessionId);
      if (currentSessionId === sessionId) {
        onNewConversation();
      }
    } catch (error) {
      console.error('Failed to delete session:', error);
    }
  }

  function formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (days === 0) {
      return date.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' });
    } else if (days === 1) {
      return 'Hier';
    } else if (days < 7) {
      return date.toLocaleDateString('fr-FR', { weekday: 'short' });
    } else {
      return date.toLocaleDateString('fr-FR', { day: '2-digit', month: '2-digit' });
    }
  }

  // Refresh sessions when a new one might be created
  export function refresh() {
    loadSessions();
  }

  onMount(() => {
    loadSessions();
  });
</script>

<aside class="sidebar">
  <div class="sidebar-header">
    <button class="new-chat-btn" on:click={onNewConversation} title="Nouvelle conversation">
      +
    </button>
  </div>

  <div class="sessions-list">
    {#if loading}
      <div class="loading">Chargement...</div>
    {:else if sessions.length === 0}
      <div class="empty">Aucune conversation</div>
    {:else}
      {#each sessions as session (session.session_id)}
        <button
          class="session-item"
          class:active={currentSessionId === session.session_id}
          on:click={() => onSelectSession(session.session_id)}
        >
          <div class="session-content">
            <span class="session-title">{session.title || 'Sans titre'}</span>
            <span class="session-date">{formatDate(session.last_activity)}</span>
          </div>
          <button
            class="delete-btn"
            on:click={(e) => handleDelete(session.session_id, e)}
            title="Supprimer"
          >
            &times;
          </button>
        </button>
      {/each}
    {/if}
  </div>
</aside>

<style>
  .sidebar {
    width: 260px;
    background: var(--color-bg-secondary);
    border-right: 1px solid var(--color-border);
    display: flex;
    flex-direction: column;
    height: 100vh;
    flex-shrink: 0;
  }

  .sidebar-header {
    padding: 1rem;
    border-bottom: 1px solid var(--color-border);
  }

  .new-chat-btn {
    width: 36px;
    height: 36px;
    padding: 0;
    background: var(--color-bg-tertiary);
    border: 1px solid var(--color-border);
    border-radius: 6px;
    color: var(--color-text-primary);
    font-size: 1.25rem;
    cursor: pointer;
    transition: all 0.15s ease;
    font-family: var(--font-family-mono);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .new-chat-btn:hover {
    background: var(--color-bg-primary);
    border-color: var(--color-border-light);
  }

  .sessions-list {
    flex: 1;
    overflow-y: auto;
    padding: 0.5rem;
  }

  .loading, .empty {
    padding: 1rem;
    text-align: center;
    color: var(--color-text-tertiary);
    font-size: 0.875rem;
  }

  .session-item {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem;
    margin-bottom: 0.25rem;
    background: transparent;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    text-align: left;
    transition: background 0.15s ease;
  }

  .session-item:hover {
    background: var(--color-bg-tertiary);
  }

  .session-item.active {
    background: var(--color-bg-tertiary);
    border: 1px solid var(--color-border);
  }

  .session-content {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .session-title {
    font-size: 0.8125rem;
    color: var(--color-text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-family: var(--font-family-mono);
  }

  .session-date {
    font-size: 0.625rem;
    color: var(--color-text-tertiary);
    font-family: var(--font-family-mono);
  }

  .delete-btn {
    opacity: 0;
    background: none;
    border: none;
    color: var(--color-text-tertiary);
    font-size: 1.25rem;
    cursor: pointer;
    padding: 0.25rem 0.5rem;
    line-height: 1;
    transition: all 0.15s ease;
  }

  .session-item:hover .delete-btn {
    opacity: 1;
  }

  .delete-btn:hover {
    color: var(--color-error);
  }

  /* Scrollbar */
  .sessions-list::-webkit-scrollbar {
    width: 4px;
  }

  .sessions-list::-webkit-scrollbar-track {
    background: transparent;
  }

  .sessions-list::-webkit-scrollbar-thumb {
    background: var(--color-border);
    border-radius: 2px;
  }
</style>

<script lang="ts">
  import { onMount } from 'svelte';
  import { fetchSessions, deleteSession, type Session } from '../services/api';
  import { Button } from "$lib/components/ui/button";
  import { ScrollArea } from "$lib/components/ui/scroll-area";
  import { Separator } from "$lib/components/ui/separator";
  import PlusIcon from "@lucide/svelte/icons/plus";
  import XIcon from "@lucide/svelte/icons/x";

  interface Props {
    currentSessionId?: string | null;
    onSelectSession: (sessionId: string) => void;
    onNewConversation: () => void;
  }

  let { currentSessionId = null, onSelectSession, onNewConversation }: Props = $props();

  let sessions = $state<Session[]>([]);
  let loading = $state(true);

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

<aside class="w-[260px] bg-sidebar border-r border-sidebar-border flex flex-col h-screen shrink-0">
  <div class="p-4">
    <Button
      variant="outline"
      size="icon"
      onclick={onNewConversation}
      title="Nouvelle conversation"
      class="bg-sidebar hover:bg-sidebar-accent"
    >
      <PlusIcon class="size-5" />
    </Button>
  </div>

  <Separator />

  <ScrollArea class="flex-1 px-2 py-2">
    {#if loading}
      <p class="p-4 text-center text-muted-foreground text-sm">
        Chargement...
      </p>
    {:else if sessions.length === 0}
      <p class="p-4 text-center text-muted-foreground text-sm">
        Aucune conversation
      </p>
    {:else}
      <div class="space-y-1">
        {#each sessions as session (session.session_id)}
          <button
            class="group w-full flex items-center justify-between px-3 py-2.5 rounded-md cursor-pointer text-left transition-colors hover:bg-sidebar-accent {currentSessionId === session.session_id ? 'bg-sidebar-accent border border-sidebar-border' : ''}"
            onclick={() => onSelectSession(session.session_id)}
          >
            <div class="flex-1 min-w-0 flex flex-col gap-1">
              <span class="text-[0.8125rem] text-sidebar-foreground truncate font-mono">
                {session.title || 'Sans titre'}
              </span>
              <span class="text-[0.625rem] text-muted-foreground font-mono">
                {formatDate(session.last_activity)}
              </span>
            </div>
            <Button
              variant="ghost"
              size="icon-sm"
              onclick={(e: Event) => handleDelete(session.session_id, e)}
              title="Supprimer"
              class="opacity-0 group-hover:opacity-100 hover:text-destructive h-7 w-7"
            >
              <XIcon class="size-4" />
            </Button>
          </button>
        {/each}
      </div>
    {/if}
  </ScrollArea>
</aside>

<script lang="ts">
  import { onMount } from 'svelte';
  import { fetchSessions, deleteSession, type Session } from '../services/api';
  import { Button } from "$lib/components/ui/button";
  import { ScrollArea } from "$lib/components/ui/scroll-area";
  import { Separator } from "$lib/components/ui/separator";
  import { sidebarStore } from '../stores/sidebarStore';
  import Icon from "@iconify/svelte";

  interface Props {
    currentSessionId?: string | null;
    onSelectSession: (sessionId: string) => void;
    onNewConversation: () => void;
  }

  let { currentSessionId = null, onSelectSession, onNewConversation }: Props = $props();

  let sessions = $state<Session[]>([]);
  let loading = $state(true);
  let isCollapsed = $derived($sidebarStore);

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

  export function refresh() {
    loadSessions();
  }

  onMount(() => {
    loadSessions();
  });
</script>

<aside class="bg-sidebar border-r border-sidebar-border flex flex-col h-screen shrink-0 transition-[width] duration-200 ease-in-out {isCollapsed ? 'w-16' : 'w-[260px]'}">
  <!-- Toggle button -->
  <div class="p-2 flex {isCollapsed ? 'justify-center' : 'justify-end'}">
    <Button
      variant="ghost"
      size="icon"
      onclick={() => sidebarStore.toggle()}
      title={isCollapsed ? "Ouvrir le menu" : "Fermer le menu"}
      class="h-8 w-8 hover:bg-sidebar-accent"
    >
      <Icon icon={isCollapsed ? "mynaui:chevron-right" : "mynaui:chevron-left"} class="size-4" />
    </Button>
  </div>

  <!-- Top section: Actions -->
  <div class="px-2 pb-3 space-y-2">
    <!-- Nouveau chat -->
    <Button
      variant="outline"
      onclick={onNewConversation}
      title="Nouvelle conversation"
      class="bg-sidebar hover:bg-sidebar-accent {isCollapsed ? 'w-full justify-center px-0' : 'w-full justify-start gap-3'}"
      size={isCollapsed ? 'icon' : 'default'}
    >
      <Icon icon="mynaui:edit-one" class="size-5 shrink-0" />
      {#if !isCollapsed}
        <span class="text-sm">Nouveau chat</span>
      {/if}
    </Button>

    <!-- Rechercher (non-functional) -->
    <Button
      variant="ghost"
      disabled
      title="Rechercher (bientot disponible)"
      class="text-muted-foreground {isCollapsed ? 'w-full justify-center px-0' : 'w-full justify-start gap-3'}"
      size={isCollapsed ? 'icon' : 'default'}
    >
      <Icon icon="mynaui:search" class="size-5 shrink-0" />
      {#if !isCollapsed}
        <span class="text-sm">Rechercher</span>
      {/if}
    </Button>
  </div>

  <Separator />

  <!-- Bottom section: Chat History -->
  <div class="flex-1 min-h-0 flex flex-col">
    {#if !isCollapsed}
      <div class="px-3 py-2">
        <span class="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          Vos chats
        </span>
      </div>
    {/if}

    <ScrollArea class="flex-1 min-h-0 px-2 py-2">
      {#if loading}
        <p class="p-4 text-center text-muted-foreground text-sm">
          {isCollapsed ? '...' : 'Chargement...'}
        </p>
      {:else if sessions.length === 0}
        <p class="p-4 text-center text-muted-foreground text-sm">
          {isCollapsed ? '-' : 'Aucune conversation'}
        </p>
      {:else}
        <div class="space-y-1">
          {#each sessions as session (session.session_id)}
            <button
              class="group w-full flex items-center rounded-md cursor-pointer text-left transition-colors hover:bg-sidebar-accent {currentSessionId === session.session_id ? 'bg-sidebar-accent border border-sidebar-border' : ''} {isCollapsed ? 'justify-center px-2 py-2.5' : 'justify-between px-3 py-2.5'}"
              onclick={() => onSelectSession(session.session_id)}
              title={isCollapsed ? (session.title || 'Sans titre') : undefined}
            >
              {#if isCollapsed}
                <!-- Show first letter when collapsed -->
                <span class="text-sm font-medium text-sidebar-foreground">
                  {(session.title || 'S')[0].toUpperCase()}
                </span>
              {:else}
                <div class="flex-1 min-w-0 flex flex-col gap-1">
                  <span class="text-[0.8125rem] text-sidebar-foreground leading-snug" style="font-family: 'Cal Sans', sans-serif;">
                    {session.title || 'Sans titre'}
                  </span>
                  <span class="text-[0.625rem] text-muted-foreground">
                    {formatDate(session.last_activity)}
                  </span>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  onclick={(e: Event) => handleDelete(session.session_id, e)}
                  title="Supprimer"
                  class="opacity-0 group-hover:opacity-100 hover:text-destructive h-7 w-7"
                >
                  <Icon icon="mynaui:trash" class="size-4" />
                </Button>
              {/if}
            </button>
          {/each}
        </div>
      {/if}
    </ScrollArea>
  </div>
</aside>

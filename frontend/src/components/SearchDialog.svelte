<script lang="ts">
  import * as Dialog from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { ScrollArea } from "$lib/components/ui/scroll-area";
  import { Separator } from "$lib/components/ui/separator";
  import Icon from "@iconify/svelte";
  import {
    searchStore,
    searchQuery,
    isSearching,
    searchError,
    groupedResults,
    searchTotal,
  } from '../stores/searchStore';

  interface Props {
    open?: boolean;
    onSelectResult: (sessionId: string, messageId: number) => void;
  }

  let { open = $bindable(false), onSelectResult }: Props = $props();

  // Track input element for autofocus
  let searchInput: HTMLInputElement | null = $state(null);

  // Focus input when dialog opens
  $effect(() => {
    if (open && searchInput) {
      setTimeout(() => searchInput?.focus(), 50);
    }
  });

  // Reset store when dialog closes
  $effect(() => {
    if (!open) {
      searchStore.reset();
    }
  });

  function handleInput(event: Event) {
    const target = event.target as HTMLInputElement;
    searchStore.setQuery(target.value);
  }

  function handleResultClick(sessionId: string, messageId: number) {
    onSelectResult(sessionId, messageId);
    open = false;
  }

  function formatTime(dateStr: string): string {
    const date = new Date(dateStr);
    return date.toLocaleDateString('fr-FR', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  // Derived values
  let groups = $derived($groupedResults);
  let query = $derived($searchQuery);
  let loading = $derived($isSearching);
  let error = $derived($searchError);
  let total = $derived($searchTotal);
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="sm:max-w-[600px] max-h-[80vh] flex flex-col">
    <Dialog.Header>
      <Dialog.Title class="flex items-center gap-2">
        <Icon icon="mynaui:search" class="size-5" />
        Rechercher
      </Dialog.Title>
      <Dialog.Description>
        Rechercher dans toutes les conversations
      </Dialog.Description>
    </Dialog.Header>

    <!-- Search Input -->
    <div class="py-4">
      <div class="relative">
        <Icon icon="mynaui:search" class="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
        <Input
          bind:ref={searchInput}
          value={query}
          oninput={handleInput}
          placeholder="Rechercher..."
          class="pl-9 font-mono"
        />
        {#if loading}
          <div class="absolute right-3 top-1/2 -translate-y-1/2">
            <Icon icon="mynaui:spinner" class="size-4 text-muted-foreground animate-spin" />
          </div>
        {/if}
      </div>
      {#if total > 0}
        <p class="text-xs text-muted-foreground mt-2">
          {total} resultat{total > 1 ? 's' : ''} trouve{total > 1 ? 's' : ''}
        </p>
      {/if}
    </div>

    <!-- Results -->
    <ScrollArea class="flex-1 min-h-0 -mx-6 px-6">
      {#if error}
        <div class="text-sm text-destructive bg-destructive/10 rounded-md p-3">
          {error}
        </div>
      {:else if query.length < 2}
        <div class="text-center text-muted-foreground text-sm py-8">
          Entrez au moins 2 caracteres pour rechercher
        </div>
      {:else if groups.length === 0 && !loading}
        <div class="text-center text-muted-foreground text-sm py-8">
          Aucun resultat pour "{query}"
        </div>
      {:else}
        <div class="space-y-4 pb-4">
          {#each groups as group (group.sessionId)}
            <div>
              <div class="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-2 flex items-center gap-2">
                <Icon icon="mynaui:chat" class="size-3.5" />
                {group.title}
              </div>
              <div class="space-y-2">
                {#each group.results as result (result.message_id)}
                  <button
                    class="w-full text-left p-3 rounded-lg border border-border hover:bg-muted/50 transition-colors"
                    onclick={() => handleResultClick(result.session_id, result.message_id)}
                  >
                    <div class="flex items-center gap-2 mb-1">
                      <Icon
                        icon={result.role === 'user' ? 'mynaui:user' : result.role === 'thinking' ? 'mynaui:lightbulb' : 'mynaui:sparkles'}
                        class="size-3 text-muted-foreground"
                      />
                      <span class="text-xs text-muted-foreground">
                        {formatTime(result.timestamp)}
                      </span>
                    </div>
                    <div class="text-sm font-mono search-snippet line-clamp-3">
                      {@html result.snippet}
                    </div>
                  </button>
                {/each}
              </div>
            </div>
            {#if groups.indexOf(group) < groups.length - 1}
              <Separator />
            {/if}
          {/each}
        </div>
      {/if}
    </ScrollArea>

    <Dialog.Footer class="border-t border-border pt-4">
      <div class="flex items-center gap-2 text-xs text-muted-foreground">
        <kbd class="px-1.5 py-0.5 rounded bg-muted border border-border font-mono text-[10px]">Cmd+K</kbd>
        pour rechercher
      </div>
      <Button variant="outline" onclick={() => open = false}>
        Fermer
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<style>
  /* Highlight search matches */
  :global(.search-snippet mark) {
    background-color: hsl(var(--primary) / 0.2);
    color: hsl(var(--primary));
    padding: 0 2px;
    border-radius: 2px;
    font-weight: 500;
  }
</style>

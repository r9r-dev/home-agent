<script lang="ts">
  import { logStore, logEntries, type LogEntry } from '../stores/logStore';
  import * as Dialog from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import { ScrollArea } from "$lib/components/ui/scroll-area";
  import Icon from "@iconify/svelte";

  interface Props {
    open?: boolean;
  }

  let { open = $bindable(false) }: Props = $props();

  let entries = $derived($logEntries);

  // Mark as read when opening
  $effect(() => {
    if (open) {
      logStore.markAsRead();
    }
  });

  function formatTime(timestamp: string): string {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('fr-FR', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  }

  function getLevelIcon(level: string): string {
    switch (level) {
      case 'error':
        return 'mynaui:danger-circle';
      case 'warning':
        return 'mynaui:warning-circle';
      default:
        return 'mynaui:info-circle';
    }
  }

  function getLevelColor(level: string): string {
    switch (level) {
      case 'error':
        return 'text-red-500';
      case 'warning':
        return 'text-orange-500';
      default:
        return 'text-muted-foreground';
    }
  }

  function handleClearStatus() {
    logStore.clearStatus();
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-2xl max-h-[80vh] flex flex-col">
    <Dialog.Header>
      <Dialog.Title class="flex items-center gap-2">
        <Icon icon="mynaui:terminal" class="size-5" />
        Logs
      </Dialog.Title>
      <Dialog.Description>
        Logs de l'application (derniers 100 messages)
      </Dialog.Description>
    </Dialog.Header>

    <div class="flex-1 min-h-0 mt-4">
      {#if entries.length === 0}
        <div class="flex items-center justify-center h-32 text-muted-foreground">
          Aucun log
        </div>
      {:else}
        <ScrollArea class="h-[400px] rounded border border-border">
          <div class="p-2 space-y-1">
            {#each entries as entry (entry.id)}
              <div class="flex items-start gap-2 px-2 py-1.5 rounded hover:bg-muted/50 font-mono text-xs">
                <Icon
                  icon={getLevelIcon(entry.level)}
                  class="size-4 mt-0.5 shrink-0 {getLevelColor(entry.level)}"
                />
                <span class="text-muted-foreground shrink-0">
                  {formatTime(entry.timestamp)}
                </span>
                <span class="flex-1 break-all {getLevelColor(entry.level)}">
                  {entry.message}
                </span>
              </div>
            {/each}
          </div>
        </ScrollArea>
      {/if}
    </div>

    <Dialog.Footer class="mt-4">
      <Button variant="outline" onclick={handleClearStatus}>
        Effacer les indicateurs
      </Button>
      <Button variant="default" onclick={() => open = false}>
        Fermer
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

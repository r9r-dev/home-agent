<script lang="ts">
  import * as Dialog from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import { Badge } from "$lib/components/ui/badge";
  import { ScrollArea } from "$lib/components/ui/scroll-area";
  import Icon from "@iconify/svelte";
  import {
    updateStore,
    updateAvailable,
    isUpdating,
    isReconnecting,
    isChecking,
    updateError,
    backendVersion,
    proxyVersion,
    backendLogs,
    proxyLogs,
    updateStatus,
    type LogEntry,
  } from '../stores/updateStore';

  interface Props {
    open?: boolean;
  }

  let { open = $bindable(false) }: Props = $props();

  function handleStartUpdate() {
    updateStore.startUpdate();
  }

  function handleCheckUpdates() {
    updateStore.checkForUpdates();
  }

  function formatTime(timestamp: string): string {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('fr-FR', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  }

  function getLogLevelColor(level: LogEntry['level']): string {
    switch (level) {
      case 'error': return 'text-red-400';
      case 'warning': return 'text-yellow-400';
      default: return 'text-green-400';
    }
  }

  // Determine if we should show logs section
  let showLogs = $derived($isUpdating || $backendLogs.length > 0 || $proxyLogs.length > 0);
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="sm:max-w-[900px] max-h-[85vh] flex flex-col">
    <Dialog.Header>
      <Dialog.Title class="flex items-center gap-2">
        <Icon icon="mynaui:refresh" class="size-5" />
        Mises a jour
      </Dialog.Title>
      <Dialog.Description>
        Gerez les mises a jour du systeme
      </Dialog.Description>
    </Dialog.Header>

    <!-- Version Info Section -->
    <div class="py-4 border-b border-border">
      {#if $isChecking}
        <div class="flex items-center gap-2 text-muted-foreground">
          <Icon icon="mynaui:refresh" class="size-4 animate-spin" />
          Verification des mises a jour...
        </div>
      {:else}
        <div class="flex items-center justify-between">
          <div class="space-y-3">
            <!-- Backend version -->
            <div class="flex items-center gap-3">
              <Icon icon="mynaui:box" class="size-4 text-muted-foreground" />
              <span class="text-sm text-muted-foreground w-16">Backend:</span>
              <Badge variant="secondary">{$backendVersion.current}</Badge>
              {#if $backendVersion.updateAvailable && $backendVersion.latest}
                <Icon icon="mynaui:arrow-right" class="size-3 text-muted-foreground" />
                <Badge class="bg-green-600 hover:bg-green-600 text-white">{$backendVersion.latest}</Badge>
              {/if}
            </div>
            <!-- Proxy version -->
            <div class="flex items-center gap-3">
              <Icon icon="mynaui:terminal" class="size-4 text-muted-foreground" />
              <span class="text-sm text-muted-foreground w-16">Proxy:</span>
              <Badge variant="secondary">{$proxyVersion.current}</Badge>
              {#if $proxyVersion.updateAvailable && $proxyVersion.latest}
                <Icon icon="mynaui:arrow-right" class="size-3 text-muted-foreground" />
                <Badge class="bg-green-600 hover:bg-green-600 text-white">{$proxyVersion.latest}</Badge>
              {/if}
            </div>
          </div>

          <div class="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onclick={handleCheckUpdates}
              disabled={$isChecking || $isUpdating}
            >
              <Icon icon="mynaui:refresh" class="size-4 mr-1.5" />
              Verifier
            </Button>
            {#if $updateAvailable}
              <Button
                size="sm"
                onclick={handleStartUpdate}
                disabled={$isUpdating}
              >
                {#if $isUpdating}
                  <Icon icon="mynaui:refresh" class="size-4 mr-1.5 animate-spin" />
                  Mise a jour...
                {:else}
                  <Icon icon="mynaui:download" class="size-4 mr-1.5" />
                  Lancer la mise a jour
                {/if}
              </Button>
            {/if}
          </div>
        </div>

        {#if !$updateAvailable && !$isChecking && $updateStatus !== 'success'}
          <div class="mt-3 text-sm text-green-500 flex items-center gap-1.5">
            <Icon icon="mynaui:check-circle" class="size-4" />
            Votre systeme est a jour
          </div>
        {/if}
      {/if}
    </div>

    <!-- Log Panels Section -->
    {#if showLogs}
      <div class="flex-1 min-h-0 py-4">
        <div class="grid grid-cols-2 gap-4 h-[350px]">
          <!-- Backend (Docker) Logs -->
          <div class="flex flex-col border border-border rounded-lg overflow-hidden">
            <div class="px-3 py-2 bg-muted/50 border-b border-border flex items-center gap-2">
              <Icon icon="mynaui:box" class="size-4 text-muted-foreground" />
              <span class="text-sm font-medium">Backend (Docker)</span>
              {#if $isReconnecting}
                <span class="ml-auto flex items-center gap-1.5 text-xs text-amber-500">
                  <Icon icon="mynaui:refresh" class="size-3 animate-spin" />
                  Reconnexion...
                </span>
              {:else if $isUpdating}
                <span class="ml-auto w-2 h-2 rounded-full bg-blue-500 animate-pulse"></span>
              {/if}
            </div>
            <ScrollArea class="flex-1 min-h-0">
              <div class="p-3 font-mono text-xs bg-[#1a1a1a] min-h-full">
                {#if $backendLogs.length === 0}
                  <span class="text-muted-foreground">En attente...</span>
                {:else}
                  {#each $backendLogs as entry (entry.id)}
                    <div class="flex gap-2 py-0.5">
                      <span class="text-muted-foreground shrink-0">[{formatTime(entry.timestamp)}]</span>
                      <span class={getLogLevelColor(entry.level)}>{entry.message}</span>
                    </div>
                  {/each}
                {/if}
              </div>
            </ScrollArea>
          </div>

          <!-- Proxy SDK Logs -->
          <div class="flex flex-col border border-border rounded-lg overflow-hidden">
            <div class="px-3 py-2 bg-muted/50 border-b border-border flex items-center gap-2">
              <Icon icon="mynaui:terminal" class="size-4 text-muted-foreground" />
              <span class="text-sm font-medium">Claude Proxy SDK</span>
              {#if $isUpdating}
                <span class="ml-auto w-2 h-2 rounded-full bg-blue-500 animate-pulse"></span>
              {/if}
            </div>
            <ScrollArea class="flex-1 min-h-0">
              <div class="p-3 font-mono text-xs bg-[#1a1a1a] min-h-full">
                {#if $proxyLogs.length === 0}
                  <span class="text-muted-foreground">En attente...</span>
                {:else}
                  {#each $proxyLogs as entry (entry.id)}
                    <div class="flex gap-2 py-0.5">
                      <span class="text-muted-foreground shrink-0">[{formatTime(entry.timestamp)}]</span>
                      <span class={getLogLevelColor(entry.level)}>{entry.message}</span>
                    </div>
                  {/each}
                {/if}
              </div>
            </ScrollArea>
          </div>
        </div>
      </div>
    {/if}

    <!-- Error Display -->
    {#if $updateError}
      <div class="text-sm text-destructive bg-destructive/10 rounded-md p-3 mt-2">
        {$updateError}
      </div>
    {/if}

    <!-- Success Message -->
    {#if $updateStatus === 'success'}
      <div class="text-sm text-green-500 bg-green-500/10 rounded-md p-3 mt-2 flex items-center gap-2">
        <Icon icon="mynaui:check-circle" class="size-4" />
        Mise a jour terminee avec succes
      </div>
    {/if}

    <Dialog.Footer class="border-t border-border pt-4">
      <Button variant="outline" onclick={() => open = false}>
        Fermer
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

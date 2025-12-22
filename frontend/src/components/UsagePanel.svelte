<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import { ScrollArea } from "$lib/components/ui/scroll-area";
  import Icon from "@iconify/svelte";
  import {
    currentUsage,
    usageHistory,
    isPanelOpen,
    usageStore,
    formatTokenCount,
    formatCost,
    getUsageColorClass,
    MODEL_CONTEXT_LIMITS,
    type UsageInfo,
  } from '../stores/usageStore';
  import { selectedModel } from '../stores/chatStore';

  let isOpen = $derived($isPanelOpen);
  let usage = $derived($currentUsage);
  let history = $derived($usageHistory);
  let model = $derived($selectedModel);
  let contextLimit = $derived(MODEL_CONTEXT_LIMITS[model] || MODEL_CONTEXT_LIMITS.sonnet);

  function close() {
    usageStore.closePanel();
  }

  function formatTime(date: Date): string {
    return date.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  }

  function calculatePercentage(u: UsageInfo): number {
    const total = u.inputTokens + u.outputTokens;
    return Math.min((total / contextLimit) * 100, 100);
  }
</script>

<!-- Backdrop -->
{#if isOpen}
  <button
    type="button"
    class="fixed inset-0 bg-black/20 z-40 cursor-default"
    onclick={close}
    aria-label="Fermer le panel"
  ></button>
{/if}

<!-- Panel -->
<div
  class="fixed top-0 right-0 h-full w-80 bg-background border-l border-border shadow-lg z-50 transform transition-transform duration-200 ease-out {isOpen ? 'translate-x-0' : 'translate-x-full'}"
>
  <div class="flex flex-col h-full">
    <!-- Header -->
    <div class="flex items-center justify-between p-4 border-b border-border">
      <h2 class="text-sm font-medium">Usage du contexte</h2>
      <Button variant="ghost" size="icon-sm" onclick={close}>
        <Icon icon="mynaui:x" class="size-4" />
      </Button>
    </div>

    <!-- Content -->
    <ScrollArea class="flex-1">
      <div class="p-4 space-y-6">
        <!-- Current Usage Summary -->
        {#if usage}
          <div class="space-y-3">
            <h3 class="text-xs font-medium text-muted-foreground uppercase tracking-wide">Resume</h3>

            <!-- Total tokens with progress -->
            <div class="space-y-2">
              <div class="flex justify-between text-sm">
                <span class="text-muted-foreground">Tokens totaux</span>
                <span class="font-mono {getUsageColorClass(calculatePercentage(usage))}">
                  {formatTokenCount(usage.inputTokens + usage.outputTokens)}
                </span>
              </div>
              <div class="w-full h-2 bg-muted rounded-full overflow-hidden">
                <div
                  class="h-full transition-all duration-300 {calculatePercentage(usage) >= 90 ? 'bg-red-500' : calculatePercentage(usage) >= 70 ? 'bg-orange-500' : calculatePercentage(usage) >= 50 ? 'bg-yellow-500' : 'bg-green-500'}"
                  style="width: {calculatePercentage(usage)}%"
                ></div>
              </div>
              <div class="flex justify-between text-xs text-muted-foreground">
                <span>0</span>
                <span>{formatTokenCount(contextLimit)} ({model})</span>
              </div>
            </div>

            <!-- Details grid -->
            <div class="grid grid-cols-2 gap-3 pt-2">
              <div class="bg-muted/50 rounded-lg p-3">
                <div class="text-xs text-muted-foreground">Input</div>
                <div class="text-sm font-mono">{formatTokenCount(usage.inputTokens)}</div>
              </div>
              <div class="bg-muted/50 rounded-lg p-3">
                <div class="text-xs text-muted-foreground">Output</div>
                <div class="text-sm font-mono">{formatTokenCount(usage.outputTokens)}</div>
              </div>
              {#if usage.cacheReadInputTokens > 0}
                <div class="bg-muted/50 rounded-lg p-3">
                  <div class="text-xs text-muted-foreground">Cache lu</div>
                  <div class="text-sm font-mono">{formatTokenCount(usage.cacheReadInputTokens)}</div>
                </div>
              {/if}
              {#if usage.cacheCreationInputTokens > 0}
                <div class="bg-muted/50 rounded-lg p-3">
                  <div class="text-xs text-muted-foreground">Cache cree</div>
                  <div class="text-sm font-mono">{formatTokenCount(usage.cacheCreationInputTokens)}</div>
                </div>
              {/if}
            </div>

            <!-- Cost -->
            {#if usage.totalCostUSD > 0}
              <div class="flex justify-between items-center pt-2 border-t border-border">
                <span class="text-sm text-muted-foreground">Cout estime</span>
                <span class="text-sm font-mono text-green-600">{formatCost(usage.totalCostUSD)}</span>
              </div>
            {/if}
          </div>
        {:else}
          <div class="text-center text-muted-foreground text-sm py-8">
            Aucune donnee d'usage disponible
          </div>
        {/if}

        <!-- Usage History -->
        {#if history.length > 1}
          <div class="space-y-3">
            <h3 class="text-xs font-medium text-muted-foreground uppercase tracking-wide">Historique</h3>
            <div class="space-y-2">
              {#each history.slice().reverse() as entry, i (entry.timestamp.getTime())}
                <div class="flex items-center justify-between text-xs py-2 border-b border-border/50 last:border-0">
                  <span class="text-muted-foreground">{formatTime(entry.timestamp)}</span>
                  <div class="flex items-center gap-2">
                    <span class="font-mono">{formatTokenCount(entry.inputTokens + entry.outputTokens)}</span>
                    <span class="text-muted-foreground">({calculatePercentage(entry).toFixed(1)}%)</span>
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    </ScrollArea>
  </div>
</div>

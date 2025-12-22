<script lang="ts">
  import {
    currentUsage,
    totalTokens,
    usagePercentage,
    formatTokenCount,
    getUsageBgClass,
    usageStore,
  } from '../stores/usageStore';

  // Only show if we have usage data
  let hasUsage = $derived($currentUsage !== null);
  let tokens = $derived($totalTokens);
  let percentage = $derived($usagePercentage);
  let bgClass = $derived(getUsageBgClass(percentage));

  function handleClick() {
    usageStore.togglePanel();
  }
</script>

{#if hasUsage}
  <button
    type="button"
    onclick={handleClick}
    class="flex items-center gap-2 h-7 px-3 rounded-full bg-muted/50 border border-border hover:bg-muted transition-colors cursor-pointer"
    title="Cliquez pour voir les details d'usage"
  >
    <!-- Token count and percentage -->
    <span class="text-xs font-mono text-muted-foreground whitespace-nowrap">
      {formatTokenCount(tokens)} ({percentage.toFixed(1)}%)
    </span>

    <!-- Progress bar -->
    <div class="w-16 h-1.5 bg-muted rounded-full overflow-hidden">
      <div
        class="h-full transition-all duration-300 {bgClass}"
        style="width: {Math.min(percentage, 100)}%"
      ></div>
    </div>
  </button>
{/if}

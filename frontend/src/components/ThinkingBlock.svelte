<script lang="ts">
  import Icon from '@iconify/svelte';

  interface Props {
    content: string;
    isStreaming?: boolean;
  }

  let { content, isStreaming = false }: Props = $props();

  let expanded = $state(false);

  // Preview shows first 100 characters when collapsed
  const previewLength = 100;

  let preview = $derived(
    content.length > previewLength
      ? content.slice(0, previewLength) + '...'
      : content
  );

  let canExpand = $derived(content.length > previewLength);

  function toggleExpand() {
    if (canExpand) {
      expanded = !expanded;
    }
  }
</script>

<div
  class="thinking-block bg-primary/5 rounded-lg border-l-2 border-primary/50 overflow-hidden mb-3"
>
  <!-- Header -->
  <button
    class="w-full flex items-center justify-between px-3 py-2 text-left hover:bg-primary/10 transition-colors"
    onclick={toggleExpand}
    disabled={!canExpand}
  >
    <div class="flex items-center gap-2 text-primary">
      <Icon icon="mynaui:lightbulb" class="w-4 h-4" />
      <span class="text-sm font-medium">Reflexion</span>
      {#if isStreaming}
        <span class="inline-flex items-center gap-1 text-xs text-muted-foreground">
          <span class="animate-pulse">...</span>
        </span>
      {/if}
    </div>
    {#if canExpand}
      <Icon
        icon={expanded ? "mynaui:chevron-up" : "mynaui:chevron-down"}
        class="w-4 h-4 text-muted-foreground"
      />
    {/if}
  </button>

  <!-- Content -->
  <div class="px-3 pb-3">
    <div class="font-mono text-sm text-muted-foreground whitespace-pre-wrap">
      {#if expanded || !canExpand}
        {content}
      {:else}
        {preview}
      {/if}
    </div>
  </div>
</div>

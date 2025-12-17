<script lang="ts">
  import Icon from '@iconify/svelte';

  interface Props {
    content: string;
    isStreaming?: boolean;
    defaultExpanded?: boolean;
  }

  let { content, isStreaming = false, defaultExpanded = false }: Props = $props();

  let expanded = $state(defaultExpanded);

  function toggleExpand() {
    expanded = !expanded;
  }
</script>

<div
  class="thinking-block bg-primary/5 rounded-lg border-l-2 border-primary/50 overflow-hidden mb-3"
>
  <!-- Header -->
  <button
    class="w-full flex items-center justify-between px-3 py-2 text-left hover:bg-primary/10 transition-colors"
    onclick={toggleExpand}
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
    <Icon
      icon={expanded ? "mynaui:chevron-up" : "mynaui:chevron-down"}
      class="w-4 h-4 text-muted-foreground"
    />
  </button>

  <!-- Content (only visible when expanded) -->
  {#if expanded}
    <div class="px-3 pb-3">
      <div class="font-mono text-sm text-muted-foreground whitespace-pre-wrap">
        {content}
      </div>
    </div>
  {/if}
</div>

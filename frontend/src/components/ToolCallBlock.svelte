<script lang="ts">
  import Icon from '@iconify/svelte';
  import type { ToolCall } from '../stores/chatStore';
  import { fetchToolCallDetail } from '../services/api';

  interface Props {
    toolCall: ToolCall;
    defaultExpanded?: boolean;
  }

  let { toolCall, defaultExpanded = false }: Props = $props();

  let expanded = $state(defaultExpanded);
  let isLoading = $state(false);
  let loadedDetail = $state<{ input: string; output: string } | null>(null);

  // Map tool names to icons
  const toolIcons: Record<string, string> = {
    Bash: 'mynaui:terminal',
    Read: 'mynaui:file',
    Write: 'mynaui:file',
    Edit: 'mynaui:file',
    Glob: 'mynaui:search',
    Grep: 'mynaui:search',
    Task: 'mynaui:layers',
    Agent: 'mynaui:layers',
    WebFetch: 'mynaui:globe',
    WebSearch: 'mynaui:globe',
  };

  // Status colors
  const statusClasses: Record<string, string> = {
    running: 'border-blue-500/50 bg-blue-500/5',
    success: 'border-green-500/50 bg-green-500/5',
    error: 'border-red-500/50 bg-red-500/5',
  };

  // Status icons
  const statusIcons: Record<string, string> = {
    running: 'mynaui:spinner',
    success: 'mynaui:check-circle',
    error: 'mynaui:x-circle',
  };

  // Status text colors
  const statusTextColors: Record<string, string> = {
    running: 'text-blue-600',
    success: 'text-green-600',
    error: 'text-red-600',
  };

  function getIcon(): string {
    return toolIcons[toolCall.toolName] || 'mynaui:code';
  }

  function getStatusClass(): string {
    return statusClasses[toolCall.status] || statusClasses.running;
  }

  function getStatusIcon(): string {
    return statusIcons[toolCall.status] || statusIcons.running;
  }

  function getStatusTextColor(): string {
    return statusTextColors[toolCall.status] || statusTextColors.running;
  }

  function formatDuration(seconds?: number): string {
    if (!seconds) return '';
    if (seconds < 1) return `${Math.round(seconds * 1000)}ms`;
    return `${seconds.toFixed(1)}s`;
  }

  async function toggleExpand() {
    if (!expanded && !loadedDetail && toolCall.status !== 'running') {
      // Lazy load details when expanding
      isLoading = true;
      try {
        const detail = await fetchToolCallDetail(toolCall.toolUseId);
        if (detail) {
          loadedDetail = {
            input: detail.input,
            output: detail.output,
          };
        }
      } catch (err) {
        console.error('Failed to load tool call details:', err);
      } finally {
        isLoading = false;
      }
    }
    expanded = !expanded;
  }

  // Use loaded detail or fall back to toolCall data
  function getInput(): string {
    if (loadedDetail) return loadedDetail.input;
    if (toolCall.input) return JSON.stringify(toolCall.input, null, 2);
    return '{}';
  }

  function getOutput(): string {
    if (loadedDetail) return loadedDetail.output;
    return toolCall.output || '';
  }
</script>

<div class="tool-call-block rounded-lg border-l-2 overflow-hidden mb-2 {getStatusClass()}">
  <!-- Header -->
  <button
    class="w-full flex items-center justify-between px-3 py-2 text-left hover:bg-primary/10 transition-colors"
    onclick={toggleExpand}
  >
    <div class="flex items-center gap-2">
      <Icon icon={getIcon()} class="w-4 h-4 text-muted-foreground" />
      <span class="text-sm font-medium">{toolCall.toolName}</span>
      {#if toolCall.status === 'running' && toolCall.elapsedTimeSeconds}
        <span class="text-xs text-muted-foreground">
          {formatDuration(toolCall.elapsedTimeSeconds)}
        </span>
      {/if}
      <Icon
        icon={getStatusIcon()}
        class="w-4 h-4 {getStatusTextColor()} {toolCall.status === 'running' ? 'animate-spin' : ''}"
      />
    </div>
    <Icon
      icon={expanded ? 'mynaui:chevron-up' : 'mynaui:chevron-down'}
      class="w-4 h-4 text-muted-foreground"
    />
  </button>

  <!-- Content (only visible when expanded) -->
  {#if expanded}
    <div class="px-3 pb-3 space-y-2">
      {#if isLoading}
        <div class="flex items-center gap-2 text-xs text-muted-foreground">
          <Icon icon="mynaui:spinner" class="w-3 h-3 animate-spin" />
          <span>Chargement...</span>
        </div>
      {:else}
        <!-- Input -->
        <div>
          <span class="text-xs text-muted-foreground font-semibold">Input</span>
          <pre
            class="text-xs bg-muted/50 rounded p-2 overflow-x-auto mt-1 max-h-[150px] overflow-y-auto">{getInput()}</pre>
        </div>

        <!-- Output (if completed) -->
        {#if getOutput()}
          <div>
            <span class="text-xs text-muted-foreground font-semibold">Output</span>
            <pre
              class="text-xs bg-muted/50 rounded p-2 overflow-x-auto mt-1 max-h-[200px] overflow-y-auto">{getOutput()}</pre>
          </div>
        {/if}
      {/if}
    </div>
  {/if}
</div>

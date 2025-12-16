<script lang="ts">
  import { onMount } from 'svelte';
  import { Textarea } from "$lib/components/ui/textarea";
  import { Button } from "$lib/components/ui/button";
  import Icon from "@iconify/svelte";

  // Props
  interface Props {
    disabled?: boolean;
    onSend: (message: string) => void;
  }

  let { disabled = false, onSend }: Props = $props();

  // Local state
  let textareaRef = $state<HTMLTextAreaElement | null>(null);
  let message = $state('');

  /**
   * Auto-resize textarea based on content
   */
  function autoResize() {
    if (!textareaRef) return;
    textareaRef.style.height = 'auto';
    textareaRef.style.height = Math.min(textareaRef.scrollHeight, 200) + 'px';
  }

  /**
   * Handle textarea input
   */
  function handleInput() {
    autoResize();
  }

  /**
   * Handle send button click
   */
  function handleSend() {
    if (!message.trim() || disabled) return;

    onSend(message.trim());
    message = '';

    // Reset textarea height
    if (textareaRef) {
      textareaRef.style.height = 'auto';
    }
  }

  /**
   * Handle keyboard events
   */
  function handleKeyDown(event: KeyboardEvent) {
    // Enter without Shift sends the message
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      handleSend();
    }
  }

  /**
   * Focus the textarea (exported for parent components)
   */
  export function focus() {
    if (textareaRef) {
      textareaRef.focus();
    }
  }

  onMount(() => {
    autoResize();
    focus();
  });
</script>

<div class="p-6 pb-0 bg-background border-t border-border">
  <div class="flex gap-3 items-center bg-muted border border-border rounded-lg px-4 py-3 max-w-[900px] mx-auto min-h-12 transition-colors focus-within:border-ring">
    <Textarea
      bind:ref={textareaRef}
      bind:value={message}
      oninput={handleInput}
      onkeydown={handleKeyDown}
      placeholder="Ecrivez votre message..."
      rows={1}
      {disabled}
      aria-label="Message input"
      class="flex-1 bg-transparent border-none shadow-none p-0 min-h-5 max-h-[200px] resize-none focus-visible:ring-0 focus-visible:border-none text-sm font-mono"
    />
    <Button
      variant="outline"
      size="icon-sm"
      onclick={handleSend}
      disabled={disabled || !message.trim()}
      aria-label="Send message"
      type="button"
    >
      <Icon icon="mynaui:send" class="size-4" />
    </Button>
  </div>
  {#if disabled}
    <p class="mt-3 text-xs text-muted-foreground text-center font-mono">
      En attente de réponse...
    </p>
  {/if}
</div>

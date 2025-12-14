<script lang="ts">
  import { onMount } from 'svelte';

  // Props
  export let disabled = false;
  export let onSend: (message: string) => void;

  // Local state
  let textarea: HTMLTextAreaElement;
  let message = '';

  /**
   * Auto-resize textarea based on content
   */
  function autoResize() {
    if (!textarea) return;

    textarea.style.height = 'auto';
    textarea.style.height = Math.min(textarea.scrollHeight, 200) + 'px';
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
    if (textarea) {
      textarea.style.height = 'auto';
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

  onMount(() => {
    autoResize();
  });
</script>

<div class="input-box">
  <div class="input-container">
    <textarea
      bind:this={textarea}
      bind:value={message}
      on:input={handleInput}
      on:keydown={handleKeyDown}
      placeholder="Ecrivez votre message..."
      rows="1"
      {disabled}
      aria-label="Message input"
    />
    <button
      class="send-button"
      on:click={handleSend}
      disabled={disabled || !message.trim()}
      aria-label="Send message"
      type="button"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        fill="currentColor"
        width="20"
        height="20"
      >
        <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z" />
      </svg>
    </button>
  </div>
  {#if disabled}
    <div class="hint">En attente de reponse...</div>
  {:else}
    <div class="hint">Entree pour envoyer, Maj+Entree pour nouvelle ligne</div>
  {/if}
</div>

<style>
  .input-box {
    padding: 1.5rem 2rem;
    background: var(--color-bg-primary);
    border-top: 1px solid var(--color-border);
  }

  .input-container {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    background: var(--color-bg-tertiary);
    border: 1px solid var(--color-border);
    border-radius: 8px;
    padding: 0.75rem 1rem;
    max-width: 900px;
    margin: 0 auto;
    min-height: 48px;
    transition: border-color 0.15s ease;
  }

  .input-container:focus-within {
    border-color: var(--color-border-light);
  }

  textarea {
    flex: 1;
    background: transparent;
    border: none;
    color: var(--color-text-primary);
    font-family: var(--font-family-mono);
    font-size: 0.875rem;
    line-height: 1.4;
    resize: none;
    outline: none;
    max-height: 200px;
    overflow-y: auto;
    padding: 0;
    margin: 0;
    min-height: 20px;
  }

  textarea::placeholder {
    color: var(--color-text-tertiary);
  }

  textarea:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .send-button {
    flex-shrink: 0;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: 1px solid var(--color-border);
    border-radius: 6px;
    color: var(--color-text-secondary);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .send-button:hover:not(:disabled) {
    border-color: var(--color-text-secondary);
    color: var(--color-text-primary);
  }

  .send-button:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  .send-button svg {
    width: 16px;
    height: 16px;
  }

  .hint {
    margin-top: 0.75rem;
    font-size: 0.75rem;
    color: var(--color-text-tertiary);
    text-align: center;
    font-family: var(--font-family-mono);
  }

  /* Scrollbar styles */
  textarea::-webkit-scrollbar {
    width: 4px;
  }

  textarea::-webkit-scrollbar-track {
    background: transparent;
  }

  textarea::-webkit-scrollbar-thumb {
    background: var(--color-border);
    border-radius: 2px;
  }
</style>

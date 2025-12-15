<script lang="ts">
  import { onMount, afterUpdate, tick } from 'svelte';
  import { marked } from 'marked';
  import hljs from 'highlight.js';
  import type { Message } from '../stores/chatStore';

  // Props
  export let messages: Message[] = [];
  export let isTyping = false;

  let messagesContainer: HTMLDivElement;
  let shouldAutoScroll = true;

  // Configure marked
  marked.setOptions({
    gfm: true,
    breaks: true,
  });

  /**
   * Render markdown to HTML
   */
  function renderMarkdown(content: string): string {
    try {
      return marked.parse(content) as string;
    } catch (error) {
      console.error('Markdown parse error:', error);
      return content;
    }
  }

  /**
   * Format timestamp
   */
  function formatTime(date: Date): string {
    return new Intl.DateTimeFormat('en-US', {
      hour: '2-digit',
      minute: '2-digit',
    }).format(date);
  }

  /**
   * Check if user is near bottom of scroll
   */
  function isNearBottom(): boolean {
    if (!messagesContainer) return true;

    const threshold = 150;
    const position = messagesContainer.scrollTop + messagesContainer.clientHeight;
    const height = messagesContainer.scrollHeight;

    return position >= height - threshold;
  }

  /**
   * Scroll to bottom
   */
  function scrollToBottom() {
    if (!messagesContainer) return;

    messagesContainer.scrollTo({
      top: messagesContainer.scrollHeight,
      behavior: 'smooth',
    });
  }

  /**
   * Handle scroll events
   */
  function handleScroll() {
    shouldAutoScroll = isNearBottom();
  }

  /**
   * Copy code to clipboard
   */
  async function copyCode(code: string) {
    try {
      await navigator.clipboard.writeText(code);
      // Could add a toast notification here
    } catch (err) {
      console.error('Failed to copy code:', err);
    }
  }

  /**
   * Apply syntax highlighting to code blocks
   */
  function highlightCodeBlocks() {
    if (!messagesContainer) return;

    const codeBlocks = messagesContainer.querySelectorAll('pre code');
    codeBlocks.forEach((block) => {
      if (block instanceof HTMLElement) {
        hljs.highlightElement(block);
      }
    });
  }

  /**
   * Auto-scroll after updates if user is near bottom
   */
  afterUpdate(async () => {
    if (shouldAutoScroll) {
      await tick();
      scrollToBottom();
    }
    highlightCodeBlocks();
  });

  onMount(() => {
    scrollToBottom();
    highlightCodeBlocks();
  });
</script>

<div
  class="message-list"
  bind:this={messagesContainer}
  on:scroll={handleScroll}
  role="log"
  aria-live="polite"
  aria-label="Chat messages"
>
  {#if messages.length === 0}
    <div class="empty-state">
      <h1 class="hero-title">Le majordome de votre infrastructure</h1>
      <p class="hero-subtitle">Gérez vos machines, serveurs et containers par la conversation.</p>
    </div>
  {:else}
    {#each messages as message (message.id)}
      <div class="message {message.role}" data-role={message.role}>
        <div class="message-content">
          {#if message.role === 'user'}
            <div class="message-text">{message.content}</div>
          {:else}
            <div class="message-text markdown-body">
              {@html renderMarkdown(message.content)}
            </div>
          {/if}
        </div>
        <div class="message-time">{formatTime(message.timestamp)}</div>
      </div>
    {/each}

    {#if isTyping}
      <div class="message assistant typing-indicator">
        <div class="message-content">
          <div class="typing-dots">
            <span></span>
            <span></span>
            <span></span>
          </div>
        </div>
      </div>
    {/if}
  {/if}
</div>

<style>
  .message-list {
    flex: 1;
    overflow-y: auto;
    padding: 2rem;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    max-width: 900px;
    margin: 0 auto;
    width: 100%;
  }

  .empty-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 4rem 2rem;
  }

  .hero-title {
    font-size: 2.5rem;
    font-weight: 500;
    color: var(--color-text-primary);
    margin-bottom: 1rem;
    letter-spacing: -0.02em;
  }

  .hero-subtitle {
    font-size: 1rem;
    color: var(--color-text-secondary);
    max-width: 500px;
  }

  .message {
    display: flex;
    flex-direction: column;
    max-width: 100%;
  }

  .message.user {
    align-self: flex-end;
    align-items: flex-end;
    max-width: 80%;
  }

  .message.assistant {
    align-self: flex-start;
    align-items: flex-start;
  }

  .message-content {
    padding: 1rem 1.25rem;
    border-radius: 8px;
    word-wrap: break-word;
    overflow-wrap: break-word;
  }

  .message.user .message-content {
    background: var(--color-bg-tertiary);
    color: var(--color-text-primary);
    border: 1px solid var(--color-border);
  }

  .message.assistant .message-content {
    background: transparent;
    color: var(--color-text-primary);
    padding: 0;
  }

  .message-text {
    font-size: 0.875rem;
    line-height: 1.7;
    font-family: var(--font-family-mono);
  }

  .message-time {
    margin-top: 0.5rem;
    font-size: 0.625rem;
    color: var(--color-text-tertiary);
    font-family: var(--font-family-mono);
  }

  /* Typing indicator */
  .typing-indicator .message-content {
    padding: 0.75rem 0;
  }

  .typing-dots {
    display: flex;
    gap: 0.3rem;
  }

  .typing-dots span {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--color-text-tertiary);
    animation: typing 1.4s infinite;
  }

  .typing-dots span:nth-child(2) {
    animation-delay: 0.2s;
  }

  .typing-dots span:nth-child(3) {
    animation-delay: 0.4s;
  }

  @keyframes typing {
    0%,
    60%,
    100% {
      opacity: 0.3;
      transform: translateY(0);
    }
    30% {
      opacity: 1;
      transform: translateY(-4px);
    }
  }

  /* Scrollbar styles */
  .message-list::-webkit-scrollbar {
    width: 6px;
  }

  .message-list::-webkit-scrollbar-track {
    background: transparent;
  }

  .message-list::-webkit-scrollbar-thumb {
    background: var(--color-border);
    border-radius: 3px;
  }

  /* Markdown styles */
  .markdown-body :global(pre) {
    background: var(--color-bg-tertiary);
    border: 1px solid var(--color-border);
    border-radius: 6px;
    padding: 1rem;
    overflow-x: auto;
    margin: 1rem 0;
  }

  .markdown-body :global(code) {
    font-family: var(--font-family-mono);
    font-size: 0.8125rem;
    line-height: 1.6;
  }

  .markdown-body :global(p code) {
    background: var(--color-bg-tertiary);
    padding: 0.2em 0.4em;
    border-radius: 4px;
    font-size: 0.85em;
    border: 1px solid var(--color-border);
  }

  .markdown-body :global(p) {
    margin: 0.75rem 0;
  }

  .markdown-body :global(p:first-child) {
    margin-top: 0;
  }

  .markdown-body :global(p:last-child) {
    margin-bottom: 0;
  }

  .markdown-body :global(ul),
  .markdown-body :global(ol) {
    margin: 0.75rem 0;
    padding-left: 1.5rem;
  }

  .markdown-body :global(li) {
    margin: 0.375rem 0;
  }

  .markdown-body :global(h1),
  .markdown-body :global(h2),
  .markdown-body :global(h3),
  .markdown-body :global(h4) {
    margin: 1.5rem 0 0.75rem 0;
    font-weight: 500;
    color: var(--color-text-primary);
  }

  .markdown-body :global(h1) {
    font-size: 1.5rem;
  }

  .markdown-body :global(h2) {
    font-size: 1.25rem;
  }

  .markdown-body :global(h3) {
    font-size: 1.125rem;
  }

  .markdown-body :global(blockquote) {
    border-left: 2px solid var(--color-border);
    padding-left: 1rem;
    margin: 1rem 0;
    color: var(--color-text-secondary);
  }

  .markdown-body :global(a) {
    color: var(--color-text-primary);
    text-decoration: underline;
    text-underline-offset: 2px;
  }

  .markdown-body :global(a:hover) {
    color: var(--color-text-secondary);
  }

  /* Responsive */
  @media (max-width: 768px) {
    .message.user {
      max-width: 90%;
    }

    .message-list {
      padding: 1rem;
    }

    .hero-title {
      font-size: 1.75rem;
    }

    .hero-subtitle {
      font-size: 0.875rem;
    }
  }
</style>

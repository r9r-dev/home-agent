<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { marked } from 'marked';
  import hljs from 'highlight.js';
  import type { Message, MessageAttachment } from '../stores/chatStore';
  import { currentThinking } from '../stores/chatStore';
  import { ScrollArea } from "$lib/components/ui/scroll-area";
  import Icon from "@iconify/svelte";
  import ThinkingBlock from './ThinkingBlock.svelte';

  interface Props {
    messages?: Message[];
    isTyping?: boolean;
  }

  let { messages = [], isTyping = false }: Props = $props();

  let scrollAreaViewport = $state<HTMLElement | null>(null);
  let shouldAutoScroll = $state(true);

  // Configure marked
  marked.setOptions({
    gfm: true,
    breaks: true,
  });

  /**
   * Normalize markdown content to ensure proper parsing
   */
  function normalizeMarkdown(content: string): string {
    let normalized = content;

    // Ensure headers have blank lines before them
    normalized = normalized.replace(/([^\n])(\n?)(#{1,6}\s)/g, '$1\n\n$3');

    // Ensure proper paragraph breaks (double newlines become proper breaks)
    normalized = normalized.replace(/\n{2,}/g, '\n\n');

    return normalized;
  }

  /**
   * Render markdown to HTML
   */
  function renderMarkdown(content: string): string {
    try {
      const normalized = normalizeMarkdown(content);
      return marked.parse(normalized) as string;
    } catch (error) {
      console.error('Markdown parse error:', error);
      return content;
    }
  }

  /**
   * Parse attachments from message content (for historical messages)
   * Format: <!-- attachments:id|filename|path|type,id|filename|path|type,... -->
   */
  function parseAttachmentsFromContent(content: string): { cleanContent: string; attachments: MessageAttachment[] } {
    const attachmentRegex = /<!-- attachments:(.*?) -->/;
    const match = content.match(attachmentRegex);

    if (!match) {
      return { cleanContent: content, attachments: [] };
    }

    const attachments: MessageAttachment[] = [];
    const attachmentStr = match[1];

    if (attachmentStr) {
      const parts = attachmentStr.split(',');
      for (const part of parts) {
        const [id, filename, path, type] = part.split('|');
        if (id && filename && path && type) {
          attachments.push({ id, filename, path, type: type as 'image' | 'file' });
        }
      }
    }

    // Remove the attachment comment from content
    const cleanContent = content.replace(attachmentRegex, '').trim();

    return { cleanContent, attachments };
  }

  /**
   * Get attachments for a message (from message.attachments or parsed from content)
   */
  function getMessageAttachments(message: Message): MessageAttachment[] {
    if (message.attachments && message.attachments.length > 0) {
      return message.attachments;
    }
    const { attachments } = parseAttachmentsFromContent(message.content);
    return attachments;
  }

  /**
   * Get clean content for a message (without attachment comments)
   */
  function getCleanContent(message: Message): string {
    if (message.attachments && message.attachments.length > 0) {
      return message.content;
    }
    const { cleanContent } = parseAttachmentsFromContent(message.content);
    return cleanContent;
  }

  /**
   * Format timestamp (24h format)
   */
  function formatTime(date: Date): string {
    return new Intl.DateTimeFormat('fr-FR', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    }).format(date);
  }

  /**
   * Check if user is near bottom of scroll
   */
  function isNearBottom(): boolean {
    if (!scrollAreaViewport) return true;

    const threshold = 150;
    const position = scrollAreaViewport.scrollTop + scrollAreaViewport.clientHeight;
    const height = scrollAreaViewport.scrollHeight;

    return position >= height - threshold;
  }

  /**
   * Scroll to bottom
   */
  function scrollToBottom() {
    if (!scrollAreaViewport) return;

    scrollAreaViewport.scrollTo({
      top: scrollAreaViewport.scrollHeight,
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
   * Apply syntax highlighting to code blocks
   */
  function highlightCodeBlocks() {
    if (!scrollAreaViewport) return;

    const codeBlocks = scrollAreaViewport.querySelectorAll('pre code');
    codeBlocks.forEach((block) => {
      if (block instanceof HTMLElement) {
        hljs.highlightElement(block);
      }
    });
  }

  /**
   * Auto-scroll after updates if user is near bottom (replaces afterUpdate)
   */
  $effect(() => {
    // Track dependencies
    messages;
    isTyping;
    $currentThinking;

    // Run after render
    tick().then(() => {
      if (shouldAutoScroll) {
        scrollToBottom();
      }
      highlightCodeBlocks();
    });
  });

  onMount(() => {
    // Add scroll listener to viewport
    if (scrollAreaViewport) {
      scrollAreaViewport.addEventListener('scroll', handleScroll);
    }
    scrollToBottom();
    highlightCodeBlocks();

    return () => {
      if (scrollAreaViewport) {
        scrollAreaViewport.removeEventListener('scroll', handleScroll);
      }
    };
  });
</script>

<ScrollArea class="flex-1 min-h-0" bind:viewportRef={scrollAreaViewport}>
  <div
    class="flex flex-col gap-6 p-8 max-w-[900px] mx-auto w-full"
    role="log"
    aria-live="polite"
    aria-label="Chat messages"
  >
    {#if messages.length === 0}
      <div class="flex-1 flex flex-col items-center justify-center text-center py-16 px-8">
        <h1 class="text-4xl font-medium text-foreground mb-4 tracking-tight">
          Bienvenue, Ronan.
        </h1>
        <p class="text-base text-muted-foreground max-w-[500px]">
          Comment puis-je vous aider ?
        </p>
      </div>
    {:else}
      {#each messages as message, index (message.id)}
        <!-- Add separator between consecutive assistant messages -->
        {#if index > 0 && message.role === 'assistant' && messages[index - 1].role === 'assistant'}
          <hr class="border-t border-border my-2 w-full" />
        {/if}

        <!-- Thinking Block: show before the last assistant message when streaming -->
        {#if $currentThinking && message.role === 'assistant' && index === messages.length - 1}
          <div class="self-start w-full max-w-[80%]">
            <ThinkingBlock content={$currentThinking} isStreaming={isTyping} />
          </div>
        {/if}

        <!-- Historical thinking message -->
        {#if message.role === 'thinking'}
          <div class="self-start w-full max-w-[80%]">
            <ThinkingBlock content={message.content} />
          </div>
        {:else}
        <div
          class="flex flex-col max-w-full {message.role === 'user' ? 'self-end items-end max-w-[80%]' : 'self-start items-start'}"
          data-role={message.role}
        >
          <div class="{message.role === 'user' ? 'bg-muted border border-border rounded-lg px-5 py-4' : ''}">
            {#if message.role === 'user'}
              <!-- User message with potential attachments -->
              {@const attachments = getMessageAttachments(message)}
              {@const cleanContent = getCleanContent(message)}
              {#if attachments.length > 0}
                <div class="flex flex-wrap gap-2 mb-3">
                  {#each attachments as attachment (attachment.id)}
                    {#if attachment.type === 'image'}
                      <a href={attachment.path} target="_blank" rel="noopener noreferrer" class="block">
                        <img
                          src={attachment.path}
                          alt={attachment.filename}
                          class="max-w-[200px] max-h-[150px] rounded border border-border object-cover hover:opacity-90 transition-opacity"
                        />
                      </a>
                    {:else}
                      <a
                        href={attachment.path}
                        target="_blank"
                        rel="noopener noreferrer"
                        class="flex items-center gap-2 bg-background border border-border rounded px-3 py-2 hover:bg-muted transition-colors"
                      >
                        <Icon icon="mynaui:file" class="size-4 text-muted-foreground" />
                        <span class="text-xs font-mono truncate max-w-[150px]">{attachment.filename}</span>
                      </a>
                    {/if}
                  {/each}
                </div>
              {/if}
              {#if cleanContent}
                <div class="text-sm leading-relaxed font-mono whitespace-pre-wrap text-foreground">
                  {cleanContent}
                </div>
              {/if}
            {:else}
              <div class="text-sm leading-relaxed font-mono markdown-body text-foreground">
                {@html renderMarkdown(message.content)}
              </div>
            {/if}
          </div>
          <span class="mt-2 text-[0.625rem] text-muted-foreground font-mono">
            {formatTime(message.timestamp)}
          </span>
        </div>
        {/if}
      {/each}

      <!-- Thinking Block: show before typing indicator if no assistant message yet -->
      {#if $currentThinking && (messages.length === 0 || messages[messages.length - 1].role !== 'assistant')}
        <div class="self-start w-full max-w-[80%]">
          <ThinkingBlock content={$currentThinking} isStreaming={isTyping} />
        </div>
      {/if}

      {#if isTyping}
        <div class="flex flex-col self-start items-start">
          <div class="py-3">
            <div class="flex gap-1.5">
              <span class="w-1.5 h-1.5 rounded-full bg-muted-foreground animate-bounce [animation-delay:0ms]"></span>
              <span class="w-1.5 h-1.5 rounded-full bg-muted-foreground animate-bounce [animation-delay:200ms]"></span>
              <span class="w-1.5 h-1.5 rounded-full bg-muted-foreground animate-bounce [animation-delay:400ms]"></span>
            </div>
          </div>
        </div>
      {/if}
    {/if}
  </div>
</ScrollArea>

<style>
  /* Markdown styles - keeping these as they style dynamically rendered HTML */
  .markdown-body {
    word-wrap: break-word;
    overflow-wrap: break-word;
  }

  .markdown-body :global(pre) {
    background: hsl(var(--muted));
    border: 1px solid hsl(var(--border));
    border-radius: 0.375rem;
    padding: 1rem;
    overflow-x: auto;
    margin: 1rem 0;
    white-space: pre-wrap;
  }

  .markdown-body :global(code) {
    font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace;
    font-size: 0.8125rem;
    line-height: 1.6;
  }

  .markdown-body :global(p code) {
    background: hsl(var(--muted));
    padding: 0.2em 0.4em;
    border-radius: 0.25rem;
    font-size: 0.85em;
    border: 1px solid hsl(var(--border));
  }

  .markdown-body :global(p) {
    margin: 0.75rem 0;
    white-space: pre-wrap;
  }

  .markdown-body :global(p:first-child) {
    margin-top: 0;
  }

  .markdown-body :global(p:last-child) {
    margin-bottom: 0;
  }

  /* Ensure line breaks are visible */
  .markdown-body :global(br) {
    display: block;
    content: "";
    margin: 0.25rem 0;
  }

  .markdown-body :global(ul),
  .markdown-body :global(ol) {
    margin: 0.75rem 0;
    padding-left: 1.5rem;
  }

  .markdown-body :global(li) {
    margin: 0.375rem 0;
  }

  .markdown-body :global(li p) {
    margin: 0.25rem 0;
  }

  .markdown-body :global(h1),
  .markdown-body :global(h2),
  .markdown-body :global(h3),
  .markdown-body :global(h4) {
    margin: 1.5rem 0 0.75rem 0;
    font-weight: 500;
    color: hsl(var(--foreground));
  }

  .markdown-body :global(h1:first-child),
  .markdown-body :global(h2:first-child),
  .markdown-body :global(h3:first-child),
  .markdown-body :global(h4:first-child) {
    margin-top: 0;
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
    border-left: 2px solid hsl(var(--border));
    padding-left: 1rem;
    margin: 1rem 0;
    color: hsl(var(--muted-foreground));
  }

  .markdown-body :global(a) {
    color: hsl(var(--primary));
    text-decoration: underline;
    text-underline-offset: 2px;
  }

  .markdown-body :global(a:hover) {
    color: hsl(var(--primary) / 0.8);
  }

  /* Horizontal rules */
  .markdown-body :global(hr) {
    border: none;
    border-top: 1px solid hsl(var(--border));
    margin: 1.5rem 0;
  }

  /* Responsive */
  @media (max-width: 768px) {
    .markdown-body :global(pre) {
      margin: 0.5rem 0;
      padding: 0.75rem;
    }
  }
</style>

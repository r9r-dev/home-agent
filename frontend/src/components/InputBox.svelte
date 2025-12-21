<script lang="ts">
  import { onMount } from 'svelte';
  import { Textarea } from "$lib/components/ui/textarea";
  import { Button } from "$lib/components/ui/button";
  import * as Select from "$lib/components/ui/select";
  import Icon from "@iconify/svelte";
  import { uploadFile, type UploadedFile } from '../services/api';
  import { machinesStore, machines, selectedMachineId, selectedMachine } from '../stores/machinesStore';

  // Props
  interface Props {
    disabled?: boolean;
    onSend: (message: string, attachments?: UploadedFile[]) => void;
    sessionId?: string | null;
  }

  let { disabled = false, onSend, sessionId = null }: Props = $props();

  // Local state
  let textareaRef = $state<HTMLTextAreaElement | null>(null);
  let fileInputRef = $state<HTMLInputElement | null>(null);
  let message = $state('');
  let attachments = $state<UploadedFile[]>([]);
  let isUploading = $state(false);
  let uploadError = $state<string | null>(null);

  // Allowed file types
  const ALLOWED_TYPES = [
    'image/png', 'image/jpeg', 'image/jpg', 'image/gif', 'image/webp',
    'application/pdf', 'text/plain', 'text/markdown', 'application/json',
    'text/csv', 'text/html', 'text/css', 'text/javascript', 'application/javascript'
  ];

  const ALLOWED_EXTENSIONS = [
    '.png', '.jpg', '.jpeg', '.gif', '.webp',
    '.pdf', '.txt', '.md', '.json', '.csv',
    '.html', '.css', '.js', '.ts', '.go', '.py', '.rs', '.java',
    '.c', '.cpp', '.h', '.sh', '.sql', '.log', '.xml', '.yaml', '.yml'
  ];

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
    if ((!message.trim() && attachments.length === 0) || disabled || isUploading) return;

    onSend(message.trim(), attachments.length > 0 ? attachments : undefined);
    message = '';
    attachments = [];
    uploadError = null;

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
   * Open file picker
   */
  function openFilePicker() {
    fileInputRef?.click();
  }

  /**
   * Validate file type
   */
  function isValidFile(file: File): boolean {
    // Check MIME type
    if (ALLOWED_TYPES.includes(file.type)) {
      return true;
    }
    // Check extension as fallback
    const ext = '.' + file.name.split('.').pop()?.toLowerCase();
    return ALLOWED_EXTENSIONS.includes(ext);
  }

  /**
   * Handle file selection
   */
  async function handleFileSelect(event: Event) {
    const input = event.target as HTMLInputElement;
    const files = input.files;
    if (!files || files.length === 0) return;

    uploadError = null;
    isUploading = true;

    try {
      for (const file of files) {
        // Validate file type
        if (!isValidFile(file)) {
          uploadError = `Type de fichier non supporté: ${file.name}`;
          continue;
        }

        // Check file size (10MB max)
        if (file.size > 10 * 1024 * 1024) {
          uploadError = `Fichier trop volumineux: ${file.name} (max 10MB)`;
          continue;
        }

        // Upload file
        const uploaded = await uploadFile(file);
        attachments = [...attachments, uploaded];
      }
    } catch (error) {
      console.error('Upload error:', error);
      uploadError = error instanceof Error ? error.message : 'Erreur lors de l\'upload';
    } finally {
      isUploading = false;
      // Reset file input
      if (fileInputRef) {
        fileInputRef.value = '';
      }
    }
  }

  /**
   * Remove an attachment
   */
  function removeAttachment(index: number) {
    attachments = attachments.filter((_, i) => i !== index);
  }

  /**
   * Format file size
   */
  function formatSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
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
    // Load machines for selector
    machinesStore.loadMachines();
  });

  function handleMachineChange(value: string | undefined) {
    machinesStore.selectMachine(value || null);
  }
</script>

<div class="p-6 pb-0 bg-background border-t border-border">
  <!-- Upload error -->
  {#if uploadError}
    <div class="flex items-center gap-2 text-destructive text-sm mb-3 max-w-[900px] mx-auto">
      <Icon icon="mynaui:danger-circle" class="size-4" />
      <span>{uploadError}</span>
    </div>
  {/if}

  <!-- Input area -->
  <div class="flex gap-3 items-center bg-muted border border-border rounded-lg px-4 py-3 max-w-[900px] mx-auto min-h-12 transition-colors focus-within:border-ring">
    <!-- Hidden file input -->
    <input
      type="file"
      bind:this={fileInputRef}
      onchange={handleFileSelect}
      multiple
      accept={[...ALLOWED_TYPES, ...ALLOWED_EXTENSIONS].join(',')}
      class="hidden"
      aria-label="Selectionner des fichiers"
    />

    <!-- Attachment button -->
    <Button
      variant="ghost"
      size="icon-sm"
      onclick={openFilePicker}
      disabled={disabled || isUploading}
      aria-label="Joindre un fichier"
      type="button"
      class="text-muted-foreground hover:text-foreground"
    >
      {#if isUploading}
        <Icon icon="mynaui:spinner" class="size-5 animate-spin" />
      {:else}
        <Icon icon="mynaui:paperclip" class="size-5" />
      {/if}
    </Button>

    <Textarea
      bind:ref={textareaRef}
      bind:value={message}
      oninput={handleInput}
      onkeydown={handleKeyDown}
      placeholder="Ecrivez votre message..."
      rows={1}
      disabled={disabled || isUploading}
      aria-label="Message input"
      class="flex-1 bg-transparent border-none shadow-none p-0 min-h-5 max-h-[200px] resize-none focus-visible:ring-0 focus-visible:border-none text-sm font-mono"
    />

    <Button
      variant="outline"
      size="icon-sm"
      onclick={handleSend}
      disabled={disabled || isUploading || (!message.trim() && attachments.length === 0)}
      aria-label="Send message"
      type="button"
    >
      <Icon icon="mynaui:send" class="size-4" />
    </Button>
  </div>

  <!-- Attachments and Machine selector (below input) -->
  {#if attachments.length > 0 || $machines.length > 0}
    <div class="flex flex-wrap items-center gap-2 max-w-[900px] mx-auto mt-3 px-1">
      <!-- Attachment chips -->
      {#each attachments as attachment, index (attachment.id)}
        <div class="flex items-center gap-1.5 bg-muted/50 border border-border rounded-full px-3 py-1 text-xs">
          {#if attachment.type === 'image'}
            <img
              src={attachment.path}
              alt={attachment.filename}
              class="w-4 h-4 object-cover rounded-full"
            />
          {:else}
            <Icon icon="mynaui:file" class="size-3.5 text-muted-foreground" />
          {/if}
          <span class="truncate max-w-[100px] font-mono">{attachment.filename}</span>
          <button
            type="button"
            onclick={() => removeAttachment(index)}
            class="p-0.5 hover:bg-destructive/20 rounded-full transition-colors ml-0.5"
            aria-label="Retirer le fichier"
          >
            <Icon icon="mynaui:x" class="size-3 text-muted-foreground hover:text-destructive" />
          </button>
        </div>
      {/each}

      <!-- Machine selector as badge -->
      {#if $machines.length > 0}
        <Select.Root type="single" value={$selectedMachineId ?? ''} onValueChange={handleMachineChange}>
          <Select.Trigger class="h-7 text-xs border-border bg-muted/50 rounded-full px-3 w-auto min-w-[100px]">
            <div class="flex items-center gap-1.5 truncate">
              <Icon icon="mynaui:server" class="size-3.5 shrink-0" />
              <span class="truncate">{$selectedMachine?.name || 'Local'}</span>
            </div>
          </Select.Trigger>
          <Select.Content>
            <Select.Item value="">
              <div class="flex items-center gap-2">
                <Icon icon="mynaui:home" class="size-4" />
                Local
              </div>
            </Select.Item>
            {#each $machines as machine (machine.id)}
              <Select.Item value={machine.id} disabled={machine.status === 'offline'}>
                <div class="flex items-center gap-2">
                  <span class={machine.status === 'online' ? 'text-green-500' : machine.status === 'offline' ? 'text-red-500' : 'text-gray-500'}>
                    <Icon icon="mynaui:server" class="size-4" />
                  </span>
                  <span class="truncate">{machine.name}</span>
                </div>
              </Select.Item>
            {/each}
          </Select.Content>
        </Select.Root>
      {/if}
    </div>
  {/if}

  {#if disabled}
    <p class="mt-3 text-xs text-muted-foreground text-center font-mono">
      En attente de reponse...
    </p>
  {/if}
</div>

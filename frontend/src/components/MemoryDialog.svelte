<script lang="ts">
  import * as Dialog from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Textarea } from "$lib/components/ui/textarea";
  import { Separator } from "$lib/components/ui/separator";
  import { ScrollArea } from "$lib/components/ui/scroll-area";
  import Icon from "@iconify/svelte";
  import {
    memoryStore,
    isMemoryLoading,
    isMemorySaving,
    memoryError,
    formatMemoryPreview,
  } from '../stores/memoryStore';
  import type { MemoryEntry } from '../services/api';

  interface Props {
    open?: boolean;
  }

  let { open = $bindable(false) }: Props = $props();

  // Form state
  let editingEntry = $state<MemoryEntry | null>(null);
  let formTitle = $state('');
  let formContent = $state('');
  let showForm = $state(false);
  let activeTab = $state<'list' | 'preview'>('list');

  // Delete confirmation
  let deleteDialogOpen = $state(false);
  let entryToDelete = $state<string | null>(null);

  // Import file input ref
  let importInput: HTMLInputElement;

  // Load entries when dialog opens
  $effect(() => {
    if (open) {
      memoryStore.loadEntries();
      resetForm();
    }
  });

  function resetForm() {
    editingEntry = null;
    formTitle = '';
    formContent = '';
    showForm = false;
  }

  function startNewEntry() {
    resetForm();
    showForm = true;
  }

  function startEditEntry(entry: MemoryEntry) {
    editingEntry = entry;
    formTitle = entry.title;
    formContent = entry.content;
    showForm = true;
  }

  function cancelForm() {
    resetForm();
  }

  async function saveEntry() {
    if (!formTitle.trim() || !formContent.trim()) return;

    let success: boolean;
    if (editingEntry) {
      success = await memoryStore.updateEntry(editingEntry.id, {
        title: formTitle.trim(),
        content: formContent.trim(),
      });
    } else {
      success = await memoryStore.createEntry(formTitle.trim(), formContent.trim());
    }

    if (success) {
      resetForm();
    }
  }

  async function toggleEntry(entry: MemoryEntry) {
    await memoryStore.toggleEntry(entry.id, !entry.enabled);
  }

  function confirmDelete(id: string) {
    entryToDelete = id;
    deleteDialogOpen = true;
  }

  async function executeDelete() {
    if (entryToDelete) {
      await memoryStore.deleteEntry(entryToDelete);
    }
    deleteDialogOpen = false;
    entryToDelete = null;
  }

  function cancelDelete() {
    deleteDialogOpen = false;
    entryToDelete = null;
  }

  async function handleExport() {
    await memoryStore.exportEntries();
  }

  function triggerImport() {
    importInput?.click();
  }

  async function handleImport(event: Event) {
    const target = event.target as HTMLInputElement;
    const file = target.files?.[0];
    if (file) {
      const result = await memoryStore.importEntries(file);
      if (result) {
        // Reset file input
        target.value = '';
      }
    }
  }

  // Derived values
  let entries = $derived($memoryStore.entries);
  let enabledCount = $derived(entries.filter(e => e.enabled).length);
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="sm:max-w-[800px] lg:max-w-[900px] max-h-[85vh] flex flex-col">
    <Dialog.Header>
      <Dialog.Title>Memoire</Dialog.Title>
      <Dialog.Description>
        Informations injectees dans chaque conversation ({enabledCount} actives)
      </Dialog.Description>
    </Dialog.Header>

    <!-- Tab Navigation -->
    <div class="flex items-center justify-between border-b border-border pb-2 mt-2">
      <div class="flex gap-2">
        <Button
          variant={activeTab === 'list' ? 'default' : 'ghost'}
          size="sm"
          onclick={() => activeTab = 'list'}
        >
          Liste
        </Button>
        <Button
          variant={activeTab === 'preview' ? 'default' : 'ghost'}
          size="sm"
          onclick={() => activeTab = 'preview'}
        >
          Apercu
        </Button>
      </div>
      <div class="flex gap-2">
        <Button variant="outline" size="sm" onclick={handleExport} disabled={entries.length === 0}>
          <Icon icon="mynaui:download" class="size-4 mr-1" />
          Exporter
        </Button>
        <Button variant="outline" size="sm" onclick={triggerImport}>
          <Icon icon="mynaui:upload" class="size-4 mr-1" />
          Importer
        </Button>
        <input
          type="file"
          accept=".json"
          class="hidden"
          bind:this={importInput}
          onchange={handleImport}
        />
      </div>
    </div>

    <!-- Tab Content -->
    <div class="flex-1 min-h-0 overflow-hidden py-4">
      {#if $isMemoryLoading}
        <div class="flex items-center justify-center h-32">
          <span class="text-muted-foreground">Chargement...</span>
        </div>
      {:else if activeTab === 'list'}
        <!-- List Tab -->
        <div class="flex flex-col h-full">
          {#if showForm}
            <!-- Add/Edit Form -->
            <div class="space-y-4 p-4 border border-border rounded-lg bg-muted/30">
              <div class="space-y-2">
                <label for="entry-title" class="text-sm font-medium">Titre</label>
                <Input
                  id="entry-title"
                  bind:value={formTitle}
                  placeholder="Ex: Informations personnelles"
                />
              </div>
              <div class="space-y-2">
                <label for="entry-content" class="text-sm font-medium">Contenu</label>
                <Textarea
                  id="entry-content"
                  bind:value={formContent}
                  placeholder="Ex: Mon nom est Jean Dupont, je suis developpeur..."
                  class="min-h-[100px] resize-none"
                />
              </div>
              <div class="flex justify-end gap-2">
                <Button variant="outline" size="sm" onclick={cancelForm}>
                  Annuler
                </Button>
                <Button
                  size="sm"
                  onclick={saveEntry}
                  disabled={$isMemorySaving || !formTitle.trim() || !formContent.trim()}
                >
                  {$isMemorySaving ? 'Enregistrement...' : editingEntry ? 'Modifier' : 'Ajouter'}
                </Button>
              </div>
            </div>
          {:else}
            <!-- Add Button -->
            <Button
              variant="outline"
              class="w-full mb-4 border-dashed"
              onclick={startNewEntry}
            >
              <Icon icon="mynaui:plus" class="size-4 mr-2" />
              Ajouter une entree
            </Button>
          {/if}

          <!-- Entries List -->
          <ScrollArea class="flex-1 min-h-0">
            {#if entries.length === 0 && !showForm}
              <div class="flex flex-col items-center justify-center h-32 text-muted-foreground">
                <Icon icon="mynaui:brain" class="size-8 mb-2 opacity-50" />
                <p class="text-sm">Aucune entree memoire</p>
                <p class="text-xs">Ajoutez des informations pour personnaliser les conversations</p>
              </div>
            {:else}
              <div class="space-y-2">
                {#each entries as entry (entry.id)}
                  <div class="flex items-start gap-3 p-3 border border-border rounded-lg bg-background {!entry.enabled ? 'opacity-60' : ''}">
                    <!-- Toggle -->
                    <button
                      class="mt-0.5 shrink-0"
                      onclick={() => toggleEntry(entry)}
                      title={entry.enabled ? 'Desactiver' : 'Activer'}
                    >
                      <Icon
                        icon={entry.enabled ? 'mynaui:checkbox' : 'mynaui:square'}
                        class="size-5 {entry.enabled ? 'text-primary' : 'text-muted-foreground'}"
                      />
                    </button>

                    <!-- Content -->
                    <div class="flex-1 min-w-0">
                      <div class="font-medium text-sm truncate">{entry.title}</div>
                      <div class="text-xs text-muted-foreground line-clamp-2 mt-0.5">
                        {entry.content}
                      </div>
                    </div>

                    <!-- Actions -->
                    <div class="flex gap-1 shrink-0">
                      <Button
                        variant="ghost"
                        size="icon"
                        class="h-7 w-7"
                        onclick={() => startEditEntry(entry)}
                        title="Modifier"
                      >
                        <Icon icon="mynaui:edit-one" class="size-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        class="h-7 w-7 hover:text-destructive"
                        onclick={() => confirmDelete(entry.id)}
                        title="Supprimer"
                      >
                        <Icon icon="mynaui:trash" class="size-4" />
                      </Button>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </ScrollArea>
        </div>
      {:else}
        <!-- Preview Tab -->
        <div class="space-y-4">
          <div class="space-y-2">
            <span class="text-sm font-medium">Apercu du contexte memoire</span>
            {#if enabledCount === 0}
              <div class="bg-muted/50 rounded-md p-4 text-center text-muted-foreground text-sm">
                Aucune entree active
              </div>
            {:else}
              <div class="bg-muted/50 rounded-md p-3 font-mono text-xs whitespace-pre-wrap max-h-[300px] overflow-y-auto">
                {formatMemoryPreview(entries)}
              </div>
            {/if}
          </div>
          <div class="text-xs text-muted-foreground">
            Ce contexte sera ajoute au debut de chaque conversation.
          </div>
        </div>
      {/if}

      {#if $memoryError}
        <div class="mt-4 text-sm text-destructive bg-destructive/10 rounded-md p-3">
          {$memoryError}
        </div>
      {/if}
    </div>

    <Dialog.Footer class="border-t border-border pt-4">
      <Button variant="outline" onclick={() => open = false}>
        Fermer
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<!-- Delete Confirmation Dialog -->
<Dialog.Root bind:open={deleteDialogOpen}>
  <Dialog.Content class="sm:max-w-[425px]">
    <Dialog.Header>
      <Dialog.Title>Supprimer l'entree</Dialog.Title>
      <Dialog.Description>
        Etes-vous sur de vouloir supprimer cette entree memoire ? Cette action est irreversible.
      </Dialog.Description>
    </Dialog.Header>
    <Dialog.Footer>
      <Button variant="outline" onclick={cancelDelete}>
        Annuler
      </Button>
      <Button variant="destructive" onclick={executeDelete}>
        Supprimer
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

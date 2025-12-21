<script lang="ts">
  import * as Dialog from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import { Textarea } from "$lib/components/ui/textarea";
  import Icon from "@iconify/svelte";
  import { settingsStore, isSettingsLoading, isSettingsSaving, settingsError } from '../stores/settingsStore';
  import MachinesList from './MachinesList.svelte';

  interface Props {
    open?: boolean;
  }

  let { open = $bindable(false) }: Props = $props();

  const MAX_CHARS = 2000;

  let activeTab = $state<'personnalisation' | 'ssh'>('personnalisation');
  let localInstructions = $state('');
  let showPreview = $state(false);

  // Derived values
  let charCount = $derived(localInstructions.length);
  let isOverLimit = $derived(charCount > MAX_CHARS);
  let hasChanges = $derived(localInstructions !== $settingsStore.customInstructions);

  // Load settings when dialog opens
  $effect(() => {
    if (open) {
      settingsStore.loadSettings();
    }
  });

  // Sync local state with store when settings load
  $effect(() => {
    if (!$isSettingsLoading) {
      localInstructions = $settingsStore.customInstructions;
    }
  });

  async function handleSave() {
    if (isOverLimit) return;
    const success = await settingsStore.saveCustomInstructions(localInstructions);
    if (success) {
      open = false;
    }
  }

  function handleCancel() {
    localInstructions = $settingsStore.customInstructions;
    settingsStore.clearError();
    open = false;
  }

  function buildFinalPrompt(): string {
    if (!localInstructions.trim()) {
      return $settingsStore.systemPrompt;
    }
    return $settingsStore.systemPrompt + '\n\n## Instructions personnalisees\n' + localInstructions;
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="sm:max-w-[800px] lg:max-w-[900px] max-h-[85vh] flex flex-col">
    <Dialog.Header>
      <Dialog.Title>Parametres</Dialog.Title>
      <Dialog.Description>
        Personnalisez le comportement de l'assistant
      </Dialog.Description>
    </Dialog.Header>

    <!-- Tab Navigation -->
    <div class="flex gap-2 border-b border-border pb-2 mt-2">
      <Button
        variant={activeTab === 'personnalisation' ? 'default' : 'ghost'}
        size="sm"
        onclick={() => activeTab = 'personnalisation'}
      >
        Personnalisation
      </Button>
      <Button
        variant={activeTab === 'ssh' ? 'default' : 'ghost'}
        size="sm"
        onclick={() => activeTab = 'ssh'}
      >
        Connexions SSH
      </Button>
    </div>

    <!-- Tab Content -->
    <div class="flex-1 min-h-0 overflow-y-auto py-4">
      {#if $isSettingsLoading && activeTab === 'personnalisation'}
        <div class="flex items-center justify-center h-32">
          <span class="text-muted-foreground">Chargement...</span>
        </div>
      {:else if activeTab === 'personnalisation'}
        <!-- Personnalisation Tab -->
        <div class="space-y-4">
          <div class="space-y-2">
            <label for="custom-instructions" class="text-sm font-medium">
              Instructions personnalisees
            </label>
            <Textarea
              id="custom-instructions"
              bind:value={localInstructions}
              placeholder="Exemples d'instructions :
- Reponds toujours en francais
- Utilise un ton formel
- Privilegie les solutions simples
- Explique tes commandes avant de les executer"
              class="min-h-[150px] resize-none {isOverLimit ? 'border-destructive focus-visible:border-destructive' : ''}"
            />
            <div class="flex justify-between items-center text-xs">
              <button
                type="button"
                onclick={() => showPreview = !showPreview}
                class="flex items-center gap-1 text-muted-foreground hover:text-foreground transition-colors"
              >
                <Icon icon={showPreview ? 'mynaui:chevron-down' : 'mynaui:chevron-right'} class="size-4" />
                Apercu du prompt
              </button>
              <span class={isOverLimit ? 'text-destructive font-medium' : 'text-muted-foreground'}>
                {charCount}/{MAX_CHARS}
              </span>
            </div>
          </div>

          <!-- Preview Section (collapsible) -->
          {#if showPreview}
            <div class="space-y-3 pt-2 border-t border-border">
              <div class="space-y-2">
                <span class="text-xs font-medium text-muted-foreground">Prompt systeme de base</span>
                <div class="bg-muted/50 rounded-md p-3 font-mono text-xs whitespace-pre-wrap max-h-[120px] overflow-y-auto">
                  {$settingsStore.systemPrompt}
                </div>
              </div>

              {#if localInstructions.trim()}
                <div class="space-y-2">
                  <span class="text-xs font-medium text-muted-foreground">Vos instructions</span>
                  <div class="bg-primary/10 rounded-md p-3 font-mono text-xs whitespace-pre-wrap max-h-[80px] overflow-y-auto">
                    {localInstructions}
                  </div>
                </div>
              {/if}

              <div class="text-xs text-muted-foreground">
                Longueur totale : {buildFinalPrompt().length} caracteres
              </div>
            </div>
          {/if}

          {#if $settingsError}
            <div class="text-sm text-destructive bg-destructive/10 rounded-md p-3">
              {$settingsError}
            </div>
          {/if}
        </div>
      {:else}
        <!-- SSH Connections Tab -->
        <MachinesList />
      {/if}
    </div>

    {#if activeTab === 'personnalisation'}
      <Dialog.Footer class="border-t border-border pt-4">
        <Button variant="outline" onclick={handleCancel} disabled={$isSettingsSaving}>
          Annuler
        </Button>
        <Button
          onclick={handleSave}
          disabled={$isSettingsSaving || isOverLimit || !hasChanges}
        >
          {$isSettingsSaving ? 'Enregistrement...' : 'Enregistrer'}
        </Button>
      </Dialog.Footer>
    {/if}
  </Dialog.Content>
</Dialog.Root>

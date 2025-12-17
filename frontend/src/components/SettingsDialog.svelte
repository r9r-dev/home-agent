<script lang="ts">
  import * as Dialog from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import { Textarea } from "$lib/components/ui/textarea";
  import { settingsStore, isSettingsLoading, isSettingsSaving, settingsError } from '../stores/settingsStore';

  interface Props {
    open?: boolean;
  }

  let { open = $bindable(false) }: Props = $props();

  const MAX_CHARS = 2000;

  let activeTab = $state<'personnalisation' | 'apercu'>('personnalisation');
  let localInstructions = $state('');
  let showSystemPrompt = $state(true);
  let showCustomInstructions = $state(true);

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
  <Dialog.Content class="sm:max-w-[600px] max-h-[85vh] flex flex-col">
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
        variant={activeTab === 'apercu' ? 'default' : 'ghost'}
        size="sm"
        onclick={() => activeTab = 'apercu'}
      >
        Apercu du prompt
      </Button>
    </div>

    <!-- Tab Content -->
    <div class="flex-1 min-h-0 overflow-y-auto py-4">
      {#if $isSettingsLoading}
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
              class="min-h-[200px] resize-none {isOverLimit ? 'border-destructive focus-visible:border-destructive' : ''}"
            />
            <div class="flex justify-between text-xs">
              <span class="text-muted-foreground">
                Ces instructions seront ajoutees au prompt systeme
              </span>
              <span class={isOverLimit ? 'text-destructive font-medium' : 'text-muted-foreground'}>
                {charCount}/{MAX_CHARS}
              </span>
            </div>
          </div>

          {#if $settingsError}
            <div class="text-sm text-destructive bg-destructive/10 rounded-md p-3">
              {$settingsError}
            </div>
          {/if}
        </div>
      {:else}
        <!-- Apercu Tab -->
        <div class="space-y-4">
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium">Prompt systeme de base</span>
              <Button
                variant="ghost"
                size="sm"
                onclick={() => showSystemPrompt = !showSystemPrompt}
              >
                {showSystemPrompt ? 'Masquer' : 'Afficher'}
              </Button>
            </div>
            {#if showSystemPrompt}
              <div class="bg-muted/50 rounded-md p-3 font-mono text-xs whitespace-pre-wrap max-h-[200px] overflow-y-auto">
                {$settingsStore.systemPrompt}
              </div>
            {/if}
          </div>

          {#if localInstructions.trim()}
            <div class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium">Vos instructions</span>
                <Button
                  variant="ghost"
                  size="sm"
                  onclick={() => showCustomInstructions = !showCustomInstructions}
                >
                  {showCustomInstructions ? 'Masquer' : 'Afficher'}
                </Button>
              </div>
              {#if showCustomInstructions}
                <div class="bg-primary/10 rounded-md p-3 font-mono text-xs whitespace-pre-wrap max-h-[150px] overflow-y-auto">
                  {localInstructions}
                </div>
              {/if}
            </div>
          {/if}

          <div class="pt-2 border-t border-border">
            <span class="text-xs text-muted-foreground">
              Longueur totale du prompt : {buildFinalPrompt().length} caracteres
            </span>
          </div>
        </div>
      {/if}
    </div>

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
  </Dialog.Content>
</Dialog.Root>

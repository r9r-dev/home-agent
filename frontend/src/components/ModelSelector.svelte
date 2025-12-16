<script lang="ts">
  import * as Select from "$lib/components/ui/select";
  import { chatStore, type ClaudeModel } from '../stores/chatStore';
  import { get } from 'svelte/store';

  const models: { value: ClaudeModel; label: string }[] = [
    { value: 'haiku', label: 'Haiku 4.5' },
    { value: 'sonnet', label: 'Sonnet 4.5' },
    { value: 'opus', label: 'Opus 4.5' },
  ];

  let currentModel = $state<ClaudeModel>(get(chatStore).selectedModel);

  // Sync with store
  chatStore.subscribe((state) => {
    currentModel = state.selectedModel;
  });

  function handleValueChange(value: string | undefined) {
    if (value) {
      chatStore.setModel(value as ClaudeModel);
    }
  }

  function getLabel(value: ClaudeModel): string {
    return models.find(m => m.value === value)?.label ?? value;
  }
</script>

<Select.Root type="single" value={currentModel} onValueChange={handleValueChange}>
  <Select.Trigger size="sm" class="w-[120px] text-xs">
    {getLabel(currentModel)}
  </Select.Trigger>
  <Select.Content>
    {#each models as model}
      <Select.Item value={model.value}>{model.label}</Select.Item>
    {/each}
  </Select.Content>
</Select.Root>

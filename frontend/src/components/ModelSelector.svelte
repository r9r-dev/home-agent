<script lang="ts">
  import { chatStore, selectedModel, type ClaudeModel } from '../stores/chatStore';

  const models: { value: ClaudeModel; label: string }[] = [
    { value: 'haiku', label: 'Haiku 4.5' },
    { value: 'sonnet', label: 'Sonnet 4.5' },
    { value: 'opus', label: 'Opus 4.5' },
  ];

  function handleChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    chatStore.setModel(target.value as ClaudeModel);
  }
</script>

<div class="model-selector">
  <select value={$selectedModel} on:change={handleChange}>
    {#each models as model}
      <option value={model.value}>{model.label}</option>
    {/each}
  </select>
</div>

<style>
  .model-selector {
    display: flex;
    align-items: center;
  }

  select {
    background: var(--color-bg-tertiary);
    color: var(--color-text-secondary);
    border: 1px solid var(--color-border);
    border-radius: 4px;
    padding: 0.375rem 0.5rem;
    font-size: 0.75rem;
    font-family: var(--font-family-sans);
    cursor: pointer;
    outline: none;
    transition: border-color var(--transition-fast), color var(--transition-fast);
  }

  select:hover {
    border-color: var(--color-border-hover, var(--color-border));
    color: var(--color-text-primary);
  }

  select:focus {
    border-color: var(--color-primary, #3b82f6);
  }

  @media (max-width: 768px) {
    select {
      padding: 0.25rem 0.375rem;
      font-size: 0.625rem;
    }
  }
</style>

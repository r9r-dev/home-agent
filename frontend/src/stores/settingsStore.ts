import { writable, derived } from 'svelte/store';
import { fetchSettings, updateSetting, fetchSystemPrompt } from '../services/api';

interface SettingsState {
  customInstructions: string;
  systemPrompt: string;
  isLoading: boolean;
  isSaving: boolean;
  error: string | null;
}

function createSettingsStore() {
  const { subscribe, update, set } = writable<SettingsState>({
    customInstructions: '',
    systemPrompt: '',
    isLoading: false,
    isSaving: false,
    error: null,
  });

  return {
    subscribe,

    loadSettings: async () => {
      update(s => ({ ...s, isLoading: true, error: null }));
      try {
        const [settings, systemPrompt] = await Promise.all([
          fetchSettings(),
          fetchSystemPrompt(),
        ]);
        update(s => ({
          ...s,
          customInstructions: settings['custom_instructions'] || '',
          systemPrompt,
          isLoading: false,
        }));
      } catch (error) {
        update(s => ({
          ...s,
          isLoading: false,
          error: error instanceof Error ? error.message : 'Failed to load settings',
        }));
      }
    },

    saveCustomInstructions: async (value: string) => {
      update(s => ({ ...s, isSaving: true, error: null }));
      try {
        await updateSetting('custom_instructions', value);
        update(s => ({
          ...s,
          customInstructions: value,
          isSaving: false,
        }));
        return true;
      } catch (error) {
        update(s => ({
          ...s,
          isSaving: false,
          error: error instanceof Error ? error.message : 'Failed to save settings',
        }));
        return false;
      }
    },

    setCustomInstructions: (value: string) => {
      update(s => ({ ...s, customInstructions: value }));
    },

    clearError: () => {
      update(s => ({ ...s, error: null }));
    },

    reset: () => {
      set({
        customInstructions: '',
        systemPrompt: '',
        isLoading: false,
        isSaving: false,
        error: null,
      });
    },
  };
}

export const settingsStore = createSettingsStore();

// Derived stores for convenience
export const customInstructions = derived(settingsStore, $s => $s.customInstructions);
export const systemPrompt = derived(settingsStore, $s => $s.systemPrompt);
export const isSettingsLoading = derived(settingsStore, $s => $s.isLoading);
export const isSettingsSaving = derived(settingsStore, $s => $s.isSaving);
export const settingsError = derived(settingsStore, $s => $s.error);

import { writable, derived } from 'svelte/store';
import {
  fetchMemoryEntries,
  createMemoryEntry,
  updateMemoryEntry,
  deleteMemoryEntry,
  exportMemory,
  importMemory,
  type MemoryEntry,
} from '../services/api';

interface MemoryState {
  entries: MemoryEntry[];
  isLoading: boolean;
  isSaving: boolean;
  error: string | null;
}

function createMemoryStore() {
  const { subscribe, update, set } = writable<MemoryState>({
    entries: [],
    isLoading: false,
    isSaving: false,
    error: null,
  });

  return {
    subscribe,

    loadEntries: async () => {
      update(s => ({ ...s, isLoading: true, error: null }));
      try {
        const entries = await fetchMemoryEntries();
        update(s => ({
          ...s,
          entries,
          isLoading: false,
        }));
      } catch (error) {
        update(s => ({
          ...s,
          isLoading: false,
          error: error instanceof Error ? error.message : 'Failed to load memory entries',
        }));
      }
    },

    createEntry: async (title: string, content: string) => {
      update(s => ({ ...s, isSaving: true, error: null }));
      try {
        const entry = await createMemoryEntry(title, content);
        update(s => ({
          ...s,
          entries: [entry, ...s.entries],
          isSaving: false,
        }));
        return true;
      } catch (error) {
        update(s => ({
          ...s,
          isSaving: false,
          error: error instanceof Error ? error.message : 'Failed to create entry',
        }));
        return false;
      }
    },

    updateEntry: async (id: string, data: { title?: string; content?: string; enabled?: boolean }) => {
      update(s => ({ ...s, isSaving: true, error: null }));
      try {
        const updated = await updateMemoryEntry(id, data);
        update(s => ({
          ...s,
          entries: s.entries.map(e => (e.id === id ? updated : e)),
          isSaving: false,
        }));
        return true;
      } catch (error) {
        update(s => ({
          ...s,
          isSaving: false,
          error: error instanceof Error ? error.message : 'Failed to update entry',
        }));
        return false;
      }
    },

    toggleEntry: async (id: string, enabled: boolean) => {
      update(s => ({ ...s, error: null }));
      try {
        const updated = await updateMemoryEntry(id, { enabled });
        update(s => ({
          ...s,
          entries: s.entries.map(e => (e.id === id ? updated : e)),
        }));
        return true;
      } catch (error) {
        update(s => ({
          ...s,
          error: error instanceof Error ? error.message : 'Failed to toggle entry',
        }));
        return false;
      }
    },

    deleteEntry: async (id: string) => {
      update(s => ({ ...s, isSaving: true, error: null }));
      try {
        await deleteMemoryEntry(id);
        update(s => ({
          ...s,
          entries: s.entries.filter(e => e.id !== id),
          isSaving: false,
        }));
        return true;
      } catch (error) {
        update(s => ({
          ...s,
          isSaving: false,
          error: error instanceof Error ? error.message : 'Failed to delete entry',
        }));
        return false;
      }
    },

    exportEntries: async () => {
      try {
        const data = await exportMemory();
        // Create and download JSON file
        const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `memory-export-${new Date().toISOString().split('T')[0]}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        return true;
      } catch (error) {
        update(s => ({
          ...s,
          error: error instanceof Error ? error.message : 'Failed to export memory',
        }));
        return false;
      }
    },

    importEntries: async (file: File) => {
      update(s => ({ ...s, isSaving: true, error: null }));
      try {
        const text = await file.text();
        const data = JSON.parse(text);

        // Handle both direct array and {entries: [...]} format
        const entries = Array.isArray(data) ? data : data.entries || [];

        if (!Array.isArray(entries)) {
          throw new Error('Invalid format: expected array of entries');
        }

        const result = await importMemory(
          entries.map((e: { title: string; content: string; enabled?: boolean }) => ({
            title: e.title,
            content: e.content,
            enabled: e.enabled ?? true,
          }))
        );

        // Reload entries after import
        const updatedEntries = await fetchMemoryEntries();
        update(s => ({
          ...s,
          entries: updatedEntries,
          isSaving: false,
        }));

        return result;
      } catch (error) {
        update(s => ({
          ...s,
          isSaving: false,
          error: error instanceof Error ? error.message : 'Failed to import memory',
        }));
        return null;
      }
    },

    clearError: () => {
      update(s => ({ ...s, error: null }));
    },

    reset: () => {
      set({
        entries: [],
        isLoading: false,
        isSaving: false,
        error: null,
      });
    },
  };
}

export const memoryStore = createMemoryStore();

// Derived stores for convenience
export const memoryEntries = derived(memoryStore, $s => $s.entries);
export const enabledMemoryEntries = derived(memoryStore, $s => $s.entries.filter(e => e.enabled));
export const isMemoryLoading = derived(memoryStore, $s => $s.isLoading);
export const isMemorySaving = derived(memoryStore, $s => $s.isSaving);
export const memoryError = derived(memoryStore, $s => $s.error);

// Helper to format memory for preview
export function formatMemoryPreview(entries: MemoryEntry[]): string {
  if (entries.length === 0) return '';

  const lines = ['<user_memory>'];
  for (const entry of entries) {
    if (entry.enabled) {
      lines.push(`- ${entry.title}: ${entry.content}`);
    }
  }
  lines.push('</user_memory>');
  return lines.join('\n');
}

import { writable, derived } from 'svelte/store';
import { searchMessages, type SearchResult } from '../services/api';

interface SearchState {
  query: string;
  results: SearchResult[];
  total: number;
  isLoading: boolean;
  error: string | null;
  isOpen: boolean;
}

const initialState: SearchState = {
  query: '',
  results: [],
  total: 0,
  isLoading: false,
  error: null,
  isOpen: false,
};

function createSearchStore() {
  const { subscribe, update, set } = writable<SearchState>(initialState);

  // Debounce timer
  let debounceTimer: ReturnType<typeof setTimeout> | null = null;

  return {
    subscribe,

    open: () => {
      update(s => ({ ...s, isOpen: true }));
    },

    close: () => {
      // Clear debounce timer
      if (debounceTimer) {
        clearTimeout(debounceTimer);
        debounceTimer = null;
      }
      set(initialState);
    },

    setQuery: (query: string) => {
      update(s => ({ ...s, query }));

      // Clear previous timer
      if (debounceTimer) {
        clearTimeout(debounceTimer);
      }

      // Debounce search by 300ms
      if (query.trim().length >= 2) {
        update(s => ({ ...s, isLoading: true, error: null }));
        debounceTimer = setTimeout(async () => {
          try {
            const response = await searchMessages(query);
            update(s => ({
              ...s,
              results: response.results,
              total: response.total,
              isLoading: false,
            }));
          } catch (error) {
            update(s => ({
              ...s,
              isLoading: false,
              error: error instanceof Error ? error.message : 'Recherche echouee',
            }));
          }
        }, 300);
      } else {
        update(s => ({ ...s, results: [], total: 0, isLoading: false }));
      }
    },

    reset: () => {
      if (debounceTimer) {
        clearTimeout(debounceTimer);
        debounceTimer = null;
      }
      set(initialState);
    },
  };
}

export const searchStore = createSearchStore();

// Derived stores for convenience
export const searchResults = derived(searchStore, $s => $s.results);
export const searchQuery = derived(searchStore, $s => $s.query);
export const isSearching = derived(searchStore, $s => $s.isLoading);
export const searchError = derived(searchStore, $s => $s.error);
export const isSearchOpen = derived(searchStore, $s => $s.isOpen);
export const searchTotal = derived(searchStore, $s => $s.total);

// Group results by session for display
interface GroupedResults {
  title: string;
  sessionId: string;
  results: SearchResult[];
}

export const groupedResults = derived(searchStore, $s => {
  const groups = new Map<string, GroupedResults>();

  for (const result of $s.results) {
    if (!groups.has(result.session_id)) {
      groups.set(result.session_id, {
        title: result.session_title || 'Sans titre',
        sessionId: result.session_id,
        results: [],
      });
    }
    groups.get(result.session_id)!.results.push(result);
  }

  return Array.from(groups.values());
});

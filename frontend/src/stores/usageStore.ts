/**
 * Svelte store for tracking token usage and context consumption
 */

import { writable, derived } from 'svelte/store';

// Context window limits per model (in tokens)
export const MODEL_CONTEXT_LIMITS: Record<string, number> = {
  haiku: 200000,
  sonnet: 200000,
  opus: 200000,
};

export interface UsageInfo {
  inputTokens: number;
  outputTokens: number;
  cacheCreationInputTokens: number;
  cacheReadInputTokens: number;
  totalCostUSD: number;
  timestamp: Date;
}

export interface UsageState {
  // Current session usage (cumulative)
  currentUsage: UsageInfo | null;
  // History of usage updates for the panel
  usageHistory: UsageInfo[];
  // Is the usage panel open?
  isPanelOpen: boolean;
}

const initialState: UsageState = {
  currentUsage: null,
  usageHistory: [],
  isPanelOpen: false,
};

function createUsageStore() {
  const { subscribe, set, update } = writable<UsageState>(initialState);

  return {
    subscribe,

    /**
     * Update usage with new data from the server
     */
    updateUsage: (usage: {
      input_tokens: number;
      output_tokens: number;
      cache_creation_input_tokens?: number;
      cache_read_input_tokens?: number;
      total_cost_usd?: number;
    }) => {
      update((state) => {
        const newUsage: UsageInfo = {
          inputTokens: usage.input_tokens,
          outputTokens: usage.output_tokens,
          cacheCreationInputTokens: usage.cache_creation_input_tokens || 0,
          cacheReadInputTokens: usage.cache_read_input_tokens || 0,
          totalCostUSD: usage.total_cost_usd || 0,
          timestamp: new Date(),
        };

        return {
          ...state,
          currentUsage: newUsage,
          usageHistory: [...state.usageHistory, newUsage],
        };
      });
    },

    /**
     * Toggle the usage panel visibility
     */
    togglePanel: () => {
      update((state) => ({
        ...state,
        isPanelOpen: !state.isPanelOpen,
      }));
    },

    /**
     * Open the usage panel
     */
    openPanel: () => {
      update((state) => ({
        ...state,
        isPanelOpen: true,
      }));
    },

    /**
     * Close the usage panel
     */
    closePanel: () => {
      update((state) => ({
        ...state,
        isPanelOpen: false,
      }));
    },

    /**
     * Clear usage data (e.g., when starting a new session)
     */
    clearUsage: () => {
      update((state) => ({
        ...state,
        currentUsage: null,
        usageHistory: [],
      }));
    },

    /**
     * Initialize usage from session data (when loading a session from API)
     */
    initFromSession: (inputTokens: number, outputTokens: number, totalCostUSD: number) => {
      // Only initialize if there's actual usage data
      if (inputTokens === 0 && outputTokens === 0) {
        return;
      }

      update((state) => {
        const sessionUsage: UsageInfo = {
          inputTokens,
          outputTokens,
          cacheCreationInputTokens: 0,
          cacheReadInputTokens: 0,
          totalCostUSD,
          timestamp: new Date(),
        };

        return {
          ...state,
          currentUsage: sessionUsage,
          usageHistory: [sessionUsage], // Start fresh history with session usage
        };
      });
    },

    /**
     * Reset the store to initial state
     */
    reset: () => {
      set(initialState);
    },
  };
}

export const usageStore = createUsageStore();

// Derived stores for convenience
export const currentUsage = derived(usageStore, ($store) => $store.currentUsage);
export const usageHistory = derived(usageStore, ($store) => $store.usageHistory);
export const isPanelOpen = derived(usageStore, ($store) => $store.isPanelOpen);

// Derived store for total tokens used
export const totalTokens = derived(usageStore, ($store) => {
  if (!$store.currentUsage) return 0;
  return $store.currentUsage.inputTokens + $store.currentUsage.outputTokens;
});

// Derived store for usage percentage based on selected model
export const usagePercentage = derived(
  [usageStore],
  ([$store]) => {
    if (!$store.currentUsage) return 0;
    // Use sonnet limit as default (most common)
    const limit = MODEL_CONTEXT_LIMITS.sonnet;
    const total = $store.currentUsage.inputTokens + $store.currentUsage.outputTokens;
    return Math.min((total / limit) * 100, 100);
  }
);

/**
 * Format token count for display (e.g., 103.2k)
 */
export function formatTokenCount(tokens: number): string {
  if (tokens >= 1000000) {
    return `${(tokens / 1000000).toFixed(1)}M`;
  }
  if (tokens >= 1000) {
    return `${(tokens / 1000).toFixed(1)}k`;
  }
  return tokens.toString();
}

/**
 * Format cost for display
 */
export function formatCost(costUSD: number): string {
  if (costUSD < 0.01) {
    return `$${costUSD.toFixed(4)}`;
  }
  return `$${costUSD.toFixed(2)}`;
}

/**
 * Get color class based on usage percentage
 */
export function getUsageColorClass(percentage: number): string {
  if (percentage >= 90) return 'text-red-500';
  if (percentage >= 70) return 'text-orange-500';
  if (percentage >= 50) return 'text-yellow-500';
  return 'text-green-500';
}

/**
 * Get background color class based on usage percentage
 */
export function getUsageBgClass(percentage: number): string {
  if (percentage >= 90) return 'bg-red-500';
  if (percentage >= 70) return 'bg-orange-500';
  if (percentage >= 50) return 'bg-yellow-500';
  return 'bg-green-500';
}

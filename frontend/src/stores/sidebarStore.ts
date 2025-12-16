import { writable } from 'svelte/store';

const STORAGE_KEY = 'sidebar-collapsed';

function createSidebarStore() {
  const stored = typeof window !== 'undefined'
    ? localStorage.getItem(STORAGE_KEY) === 'true'
    : false;

  const { subscribe, update } = writable<boolean>(stored);

  return {
    subscribe,
    toggle: () => update(v => {
      const newVal = !v;
      if (typeof window !== 'undefined') {
        localStorage.setItem(STORAGE_KEY, String(newVal));
      }
      return newVal;
    }),
  };
}

export const sidebarStore = createSidebarStore();

import { writable, derived } from 'svelte/store';
import {
  fetchMachines,
  createMachine,
  updateMachine,
  deleteMachine,
  testMachineConnection,
  type Machine,
  type CreateMachineData,
  type TestConnectionResult,
} from '../services/api';

interface MachinesState {
  machines: Machine[];
  selectedMachineId: string | null;
  isLoading: boolean;
  isSaving: boolean;
  testingMachineIds: Set<string>;
  error: string | null;
}

function createMachinesStore() {
  const { subscribe, update, set } = writable<MachinesState>({
    machines: [],
    selectedMachineId: 'auto', // Default to auto mode
    isLoading: false,
    isSaving: false,
    testingMachineIds: new Set(),
    error: null,
  });

  return {
    subscribe,

    loadMachines: async () => {
      update(s => ({ ...s, isLoading: true, error: null }));
      try {
        const machines = await fetchMachines();
        update(s => ({
          ...s,
          machines,
          isLoading: false,
        }));
      } catch (error) {
        update(s => ({
          ...s,
          isLoading: false,
          error: error instanceof Error ? error.message : 'Echec du chargement des machines',
        }));
      }
    },

    createMachine: async (data: CreateMachineData): Promise<Machine | null> => {
      update(s => ({ ...s, isSaving: true, error: null }));
      try {
        const machine = await createMachine(data);
        update(s => ({
          ...s,
          machines: [...s.machines, machine].sort((a, b) => a.name.localeCompare(b.name)),
          isSaving: false,
        }));
        return machine;
      } catch (error) {
        update(s => ({
          ...s,
          isSaving: false,
          error: error instanceof Error ? error.message : 'Echec de la creation',
        }));
        return null;
      }
    },

    updateMachine: async (id: string, data: CreateMachineData): Promise<Machine | null> => {
      update(s => ({ ...s, isSaving: true, error: null }));
      try {
        const machine = await updateMachine(id, data);
        update(s => ({
          ...s,
          machines: s.machines.map(m => (m.id === id ? machine : m)).sort((a, b) => a.name.localeCompare(b.name)),
          isSaving: false,
        }));
        return machine;
      } catch (error) {
        update(s => ({
          ...s,
          isSaving: false,
          error: error instanceof Error ? error.message : 'Echec de la mise a jour',
        }));
        return null;
      }
    },

    deleteMachine: async (id: string): Promise<boolean> => {
      update(s => ({ ...s, isSaving: true, error: null }));
      try {
        await deleteMachine(id);
        update(s => ({
          ...s,
          machines: s.machines.filter(m => m.id !== id),
          selectedMachineId: s.selectedMachineId === id ? 'auto' : s.selectedMachineId,
          isSaving: false,
        }));
        return true;
      } catch (error) {
        update(s => ({
          ...s,
          isSaving: false,
          error: error instanceof Error ? error.message : 'Echec de la suppression',
        }));
        return false;
      }
    },

    testConnection: async (id: string): Promise<TestConnectionResult | null> => {
      update(s => ({
        ...s,
        testingMachineIds: new Set([...s.testingMachineIds, id]),
        error: null,
      }));
      try {
        const result = await testMachineConnection(id);
        // Update machine status based on result
        update(s => ({
          ...s,
          machines: s.machines.map(m =>
            m.id === id ? { ...m, status: result.success ? 'online' : 'offline' } : m
          ),
          testingMachineIds: new Set([...s.testingMachineIds].filter(mid => mid !== id)),
        }));
        return result;
      } catch (error) {
        update(s => ({
          ...s,
          testingMachineIds: new Set([...s.testingMachineIds].filter(mid => mid !== id)),
          error: error instanceof Error ? error.message : 'Echec du test de connexion',
        }));
        return null;
      }
    },

    selectMachine: (id: string | null) => {
      update(s => ({ ...s, selectedMachineId: id }));
    },

    clearError: () => {
      update(s => ({ ...s, error: null }));
    },

    reset: () => {
      set({
        machines: [],
        selectedMachineId: 'auto',
        isLoading: false,
        isSaving: false,
        testingMachineIds: new Set(),
        error: null,
      });
    },
  };
}

export const machinesStore = createMachinesStore();

// Derived stores for convenience
export const machines = derived(machinesStore, $s => $s.machines);
export const selectedMachineId = derived(machinesStore, $s => $s.selectedMachineId);
export const selectedMachine = derived(machinesStore, $s =>
  $s.machines.find(m => m.id === $s.selectedMachineId)
);
export const isMachinesLoading = derived(machinesStore, $s => $s.isLoading);
export const isMachinesSaving = derived(machinesStore, $s => $s.isSaving);
export const machinesError = derived(machinesStore, $s => $s.error);

// Helper to check if a machine is being tested
export function isMachineTesting(state: MachinesState, id: string): boolean {
  return state.testingMachineIds.has(id);
}

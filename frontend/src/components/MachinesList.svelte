<script lang="ts">
  import { machinesStore, isMachinesLoading, machinesError, isMachineTesting } from '../stores/machinesStore';
  import { Button } from "$lib/components/ui/button";
  import { Badge } from "$lib/components/ui/badge";
  import { ScrollArea } from "$lib/components/ui/scroll-area";
  import * as Dialog from "$lib/components/ui/dialog";
  import Icon from "@iconify/svelte";
  import MachineForm from './MachineForm.svelte';
  import type { Machine } from '../services/api';

  // Local state
  let showForm = $state(false);
  let editingMachine = $state<Machine | null>(null);
  let deleteDialogOpen = $state(false);
  let machineToDelete = $state<Machine | null>(null);
  let lastTestResult = $state<{ id: string; success: boolean; message: string } | null>(null);

  // Load machines on mount
  $effect(() => {
    machinesStore.loadMachines();
  });

  // Status badge styling
  function getStatusBadge(status: string): { class: string; label: string } {
    switch (status) {
      case 'online':
        return { class: 'bg-green-500 hover:bg-green-500', label: 'En ligne' };
      case 'offline':
        return { class: 'bg-red-500 hover:bg-red-500', label: 'Hors ligne' };
      default:
        return { class: 'bg-gray-500 hover:bg-gray-500', label: 'Non teste' };
    }
  }

  function handleAddMachine() {
    editingMachine = null;
    showForm = true;
  }

  function handleEditMachine(machine: Machine) {
    editingMachine = machine;
    showForm = true;
  }

  function handleFormSave() {
    showForm = false;
    editingMachine = null;
  }

  function handleFormCancel() {
    showForm = false;
    editingMachine = null;
  }

  function handleDeleteClick(machine: Machine) {
    machineToDelete = machine;
    deleteDialogOpen = true;
  }

  async function confirmDelete() {
    if (machineToDelete) {
      await machinesStore.deleteMachine(machineToDelete.id);
      deleteDialogOpen = false;
      machineToDelete = null;
    }
  }

  async function handleTestConnection(machine: Machine) {
    const result = await machinesStore.testConnection(machine.id);
    if (result) {
      lastTestResult = {
        id: machine.id,
        success: result.success,
        message: result.message + (result.latency_ms ? ` (${result.latency_ms}ms)` : ''),
      };
      // Clear after 5 seconds
      setTimeout(() => {
        if (lastTestResult?.id === machine.id) {
          lastTestResult = null;
        }
      }, 5000);
    }
  }
</script>

<div class="flex flex-col h-full">
  {#if showForm}
    <MachineForm
      machine={editingMachine}
      onSave={handleFormSave}
      onCancel={handleFormCancel}
    />
  {:else}
    <!-- Add button -->
    <Button
      variant="outline"
      class="w-full mb-4 border-dashed"
      onclick={handleAddMachine}
    >
      <Icon icon="mynaui:plus" class="size-4 mr-2" />
      Ajouter une machine
    </Button>

    <!-- Loading state -->
    {#if $isMachinesLoading}
      <div class="flex items-center justify-center h-32">
        <Icon icon="mynaui:spinner" class="size-6 animate-spin text-muted-foreground" />
      </div>
    {:else if $machinesStore.machines.length === 0}
      <!-- Empty state -->
      <div class="flex flex-col items-center justify-center h-32 text-muted-foreground">
        <Icon icon="mynaui:server" class="size-10 mb-2 opacity-50" />
        <span class="text-sm">Aucune machine configuree</span>
        <span class="text-xs mt-1">Ajoutez votre premiere machine SSH</span>
      </div>
    {:else}
      <!-- Machine list -->
      <ScrollArea class="flex-1 min-h-0">
        <div class="space-y-2 pr-3">
          {#each $machinesStore.machines as machine (machine.id)}
            {@const isTesting = isMachineTesting($machinesStore, machine.id)}
            {@const statusBadge = getStatusBadge(machine.status)}
            <div class="flex items-center gap-3 p-3 border border-border rounded-lg hover:bg-muted/30 transition-colors">
              <!-- Machine info -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <span class="font-medium text-sm truncate">{machine.name}</span>
                  <Badge class={statusBadge.class + " text-white text-xs px-1.5 py-0"}>
                    {statusBadge.label}
                  </Badge>
                </div>
                <div class="text-xs text-muted-foreground font-mono">
                  {machine.username}@{machine.host}:{machine.port}
                </div>
                {#if machine.description}
                  <div class="text-xs text-muted-foreground mt-0.5 truncate">
                    {machine.description}
                  </div>
                {/if}
                {#if lastTestResult?.id === machine.id}
                  <div class="text-xs mt-1 {lastTestResult.success ? 'text-green-600' : 'text-red-600'}">
                    {lastTestResult.message}
                  </div>
                {/if}
              </div>

              <!-- Actions -->
              <div class="flex gap-1">
                <Button
                  variant="ghost"
                  size="icon"
                  onclick={() => handleTestConnection(machine)}
                  disabled={isTesting}
                  title="Tester la connexion"
                >
                  <Icon icon="mynaui:refresh" class="size-4 {isTesting ? 'animate-spin' : ''}" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  onclick={() => handleEditMachine(machine)}
                  title="Modifier"
                >
                  <Icon icon="mynaui:edit-one" class="size-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  onclick={() => handleDeleteClick(machine)}
                  title="Supprimer"
                >
                  <Icon icon="mynaui:trash" class="size-4" />
                </Button>
              </div>
            </div>
          {/each}
        </div>
      </ScrollArea>
    {/if}

    <!-- Error display -->
    {#if $machinesError}
      <div class="mt-4 text-sm text-destructive bg-destructive/10 rounded-md p-3">
        {$machinesError}
      </div>
    {/if}
  {/if}
</div>

<!-- Delete confirmation dialog -->
<Dialog.Root bind:open={deleteDialogOpen}>
  <Dialog.Content class="sm:max-w-[400px]">
    <Dialog.Header>
      <Dialog.Title>Supprimer la machine</Dialog.Title>
      <Dialog.Description>
        Etes-vous sur de vouloir supprimer la machine "{machineToDelete?.name}" ?
        Cette action est irreversible.
      </Dialog.Description>
    </Dialog.Header>
    <Dialog.Footer>
      <Button variant="outline" onclick={() => deleteDialogOpen = false}>
        Annuler
      </Button>
      <Button variant="destructive" onclick={confirmDelete}>
        Supprimer
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

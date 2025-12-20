<script lang="ts">
  import { Input } from "$lib/components/ui/input";
  import { Textarea } from "$lib/components/ui/textarea";
  import { Button } from "$lib/components/ui/button";
  import Icon from "@iconify/svelte";
  import { machinesStore, isMachinesSaving } from '../stores/machinesStore';
  import type { Machine } from '../services/api';

  interface Props {
    machine?: Machine | null;
    onSave: () => void;
    onCancel: () => void;
  }

  let { machine = null, onSave, onCancel }: Props = $props();

  // Form state
  let name = $state(machine?.name || '');
  let description = $state(machine?.description || '');
  let host = $state(machine?.host || '');
  let port = $state(machine?.port || 22);
  let username = $state(machine?.username || '');
  let authType = $state<'password' | 'key'>(machine?.auth_type || 'password');
  let authValue = $state(''); // Always empty, user must re-enter

  // File upload for SSH key
  let fileInputRef = $state<HTMLInputElement | null>(null);
  let keyLoaded = $state(false);

  // Validation
  let isValid = $derived(
    name.trim() !== '' &&
    host.trim() !== '' &&
    username.trim() !== '' &&
    authValue.trim() !== '' &&
    port > 0 && port <= 65535
  );

  async function handleKeyUpload(event: Event) {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    if (file) {
      try {
        authValue = await file.text();
        keyLoaded = true;
      } catch (error) {
        console.error('Failed to read key file:', error);
      }
    }
  }

  async function handleSubmit() {
    if (!isValid) return;

    const data = {
      name: name.trim(),
      description: description.trim(),
      host: host.trim(),
      port,
      username: username.trim(),
      auth_type: authType,
      auth_value: authValue,
    };

    let result;
    if (machine) {
      result = await machinesStore.updateMachine(machine.id, data);
    } else {
      result = await machinesStore.createMachine(data);
    }

    if (result) {
      onSave();
    }
  }
</script>

<div class="space-y-4 p-4 border border-border rounded-lg bg-muted/30">
  <div class="flex items-center justify-between mb-2">
    <span class="text-sm font-medium">
      {machine ? 'Modifier la machine' : 'Nouvelle machine'}
    </span>
    <Button variant="ghost" size="icon" onclick={onCancel}>
      <Icon icon="mynaui:x" class="size-4" />
    </Button>
  </div>

  <div class="grid grid-cols-2 gap-4">
    <div class="space-y-2">
      <label for="name" class="text-xs font-medium text-muted-foreground">Nom *</label>
      <Input
        id="name"
        bind:value={name}
        placeholder="Mon serveur"
        class="h-8 text-sm"
      />
    </div>
    <div class="space-y-2">
      <label for="description" class="text-xs font-medium text-muted-foreground">Description</label>
      <Input
        id="description"
        bind:value={description}
        placeholder="Serveur de production"
        class="h-8 text-sm"
      />
    </div>
  </div>

  <div class="grid grid-cols-4 gap-4">
    <div class="col-span-3 space-y-2">
      <label for="host" class="text-xs font-medium text-muted-foreground">Hote *</label>
      <Input
        id="host"
        bind:value={host}
        placeholder="192.168.1.100 ou server.example.com"
        class="h-8 text-sm font-mono"
      />
    </div>
    <div class="space-y-2">
      <label for="port" class="text-xs font-medium text-muted-foreground">Port *</label>
      <Input
        id="port"
        type="number"
        bind:value={port}
        min="1"
        max="65535"
        class="h-8 text-sm font-mono"
      />
    </div>
  </div>

  <div class="space-y-2">
    <label for="username" class="text-xs font-medium text-muted-foreground">Utilisateur *</label>
    <Input
      id="username"
      bind:value={username}
      placeholder="root"
      class="h-8 text-sm font-mono"
    />
  </div>

  <div class="space-y-2">
    <label class="text-xs font-medium text-muted-foreground">Authentification *</label>
    <div class="flex gap-2">
      <Button
        type="button"
        variant={authType === 'password' ? 'default' : 'outline'}
        size="sm"
        onclick={() => { authType = 'password'; authValue = ''; keyLoaded = false; }}
      >
        <Icon icon="mynaui:lock-password" class="size-4 mr-1" />
        Mot de passe
      </Button>
      <Button
        type="button"
        variant={authType === 'key' ? 'default' : 'outline'}
        size="sm"
        onclick={() => { authType = 'key'; authValue = ''; keyLoaded = false; }}
      >
        <Icon icon="mynaui:key" class="size-4 mr-1" />
        Cle SSH
      </Button>
    </div>
  </div>

  {#if authType === 'password'}
    <div class="space-y-2">
      <label for="password" class="text-xs font-medium text-muted-foreground">
        Mot de passe * {#if machine}<span class="text-yellow-600">(re-saisie requise)</span>{/if}
      </label>
      <Input
        id="password"
        type="password"
        bind:value={authValue}
        placeholder="********"
        class="h-8 text-sm"
      />
    </div>
  {:else}
    <div class="space-y-2">
      <label class="text-xs font-medium text-muted-foreground">
        Cle SSH privee * {#if machine}<span class="text-yellow-600">(re-saisie requise)</span>{/if}
      </label>
      <div class="flex gap-2 items-center">
        <Button variant="outline" size="sm" onclick={() => fileInputRef?.click()}>
          <Icon icon="mynaui:upload" class="size-4 mr-1" />
          Charger un fichier
        </Button>
        <input
          type="file"
          class="hidden"
          bind:this={fileInputRef}
          onchange={handleKeyUpload}
          accept=".pem,.key,.pub,id_rsa,id_ed25519,id_ecdsa"
        />
        {#if keyLoaded}
          <span class="text-xs text-green-600 flex items-center gap-1">
            <Icon icon="mynaui:check" class="size-3" />
            Cle chargee
          </span>
        {/if}
      </div>
      <Textarea
        bind:value={authValue}
        placeholder="-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----"
        class="font-mono text-xs min-h-[100px] resize-none"
      />
    </div>
  {/if}

  <div class="flex justify-end gap-2 pt-2 border-t border-border">
    <Button variant="outline" size="sm" onclick={onCancel} disabled={$isMachinesSaving}>
      Annuler
    </Button>
    <Button
      size="sm"
      onclick={handleSubmit}
      disabled={!isValid || $isMachinesSaving}
    >
      {#if $isMachinesSaving}
        <Icon icon="mynaui:spinner" class="size-4 mr-1 animate-spin" />
      {/if}
      {machine ? 'Modifier' : 'Ajouter'}
    </Button>
  </div>
</div>

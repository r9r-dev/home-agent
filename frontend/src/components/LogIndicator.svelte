<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { logStore, logStatus, unreadLogCount } from '../stores/logStore';
  import { Badge } from "$lib/components/ui/badge";
  import Icon from "@iconify/svelte";

  interface Props {
    onclick?: () => void;
  }

  let { onclick }: Props = $props();

  let status = $derived($logStatus);
  let unreadCount = $derived($unreadLogCount);

  // Color based on status
  let indicatorColor = $derived.by(() => {
    switch (status) {
      case 'error':
        return 'bg-red-500';
      case 'warning':
        return 'bg-orange-500';
      default:
        return 'bg-green-500';
    }
  });

  let badgeVariant = $derived.by(() => {
    switch (status) {
      case 'error':
        return 'destructive' as const;
      case 'warning':
        return 'outline' as const;
      default:
        return 'secondary' as const;
    }
  });

  onMount(() => {
    logStore.connect();
  });

  onDestroy(() => {
    logStore.disconnect();
  });
</script>

<button
  class="relative flex items-center gap-2 px-3 py-1.5 rounded-md border border-border bg-background hover:bg-muted transition-colors"
  onclick={onclick}
  title="Voir les logs"
>
  <Icon icon="mynaui:terminal" class="size-4 text-muted-foreground" />
  <span class="w-2 h-2 rounded-full {indicatorColor}"></span>
  {#if unreadCount > 0}
    <span class="absolute -top-1 -right-1 flex items-center justify-center min-w-[18px] h-[18px] px-1 text-[10px] font-medium rounded-full bg-primary text-primary-foreground">
      {unreadCount > 99 ? '99+' : unreadCount}
    </span>
  {/if}
</button>

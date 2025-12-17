<script lang="ts">
	import { Menubar as MenubarPrimitive } from "bits-ui";
	import { cn, type WithoutChild } from "$lib/utils.js";
	import Icon from "@iconify/svelte";

	let {
		ref = $bindable(null),
		class: className,
		checked = $bindable(false),
		children,
		...restProps
	}: WithoutChild<MenubarPrimitive.CheckboxItemProps> = $props();
</script>

<MenubarPrimitive.CheckboxItem
	bind:ref
	bind:checked
	data-slot="menubar-checkbox-item"
	class={cn(
		"focus:bg-accent focus:text-accent-foreground relative flex cursor-default items-center gap-2 rounded-sm py-1.5 pe-2 ps-8 text-sm outline-hidden select-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
		className
	)}
	{...restProps}
>
	{#snippet children({ checked, indeterminate })}
		<span class="pointer-events-none absolute start-2 flex size-3.5 items-center justify-center">
			{#if checked}
				<Icon icon="mynaui:check" class="size-4" />
			{/if}
		</span>
		{@render children?.({ checked, indeterminate })}
	{/snippet}
</MenubarPrimitive.CheckboxItem>

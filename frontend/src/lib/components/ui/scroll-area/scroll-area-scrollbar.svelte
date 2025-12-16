<script lang="ts">
	import { ScrollArea as ScrollAreaPrimitive } from "bits-ui";
	import { cn, type WithoutChild } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		class: className,
		orientation = "vertical",
		children,
		...restProps
	}: WithoutChild<ScrollAreaPrimitive.ScrollbarProps> = $props();
</script>

<ScrollAreaPrimitive.Scrollbar
	bind:ref
	data-slot="scroll-area-scrollbar"
	{orientation}
	class={cn(
		"flex touch-none p-0.5 transition-colors select-none bg-muted/50",
		orientation === "vertical" && "h-full w-3",
		orientation === "horizontal" && "h-3 flex-col",
		className
	)}
	{...restProps}
>
	{@render children?.()}
	<ScrollAreaPrimitive.Thumb
		data-slot="scroll-area-thumb"
		class="bg-muted-foreground/40 hover:bg-muted-foreground/60 relative flex-1 rounded-full transition-colors"
	/>
</ScrollAreaPrimitive.Scrollbar>

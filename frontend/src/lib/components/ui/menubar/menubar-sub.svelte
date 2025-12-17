<script lang="ts">
	import { Menubar as MenubarPrimitive } from "bits-ui";

	let { ...restProps }: MenubarPrimitive.SubProps = $props();
</script>

<MenubarPrimitive.Sub {...restProps} />

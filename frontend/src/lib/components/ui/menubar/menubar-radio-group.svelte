<script lang="ts">
	import { Menubar as MenubarPrimitive } from "bits-ui";

	let { ...restProps }: MenubarPrimitive.RadioGroupProps = $props();
</script>

<MenubarPrimitive.RadioGroup data-slot="menubar-radio-group" {...restProps} />

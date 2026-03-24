<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	onMount(() => {
		goto('/');
	});
</script>

<div class="container mx-auto p-8 max-w-4xl text-center">
	<p class="text-gray-600">Redirecionando...</p>
</div>

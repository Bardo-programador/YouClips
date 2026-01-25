<script lang="ts">
	import { onMount } from 'svelte';

	interface Clip {
		id: number;
		status: string;
		title: string;
		duration: number;
		size: number;
		download_url?: string;
	}

	let clips: Clip[] = [];
	let loading = true;
	let error = '';
	let page = 1;
	let total = 0;
	let deletingId: number | null = null;

	function formatTime(seconds: number): string {
		const hours = Math.floor(seconds / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);
		const secs = seconds % 60;

		if (hours > 0) {
			return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
		}
		return `${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
	}

	function formatSize(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'completed':
				return 'bg-green-100 text-green-800';
			case 'processing':
				return 'bg-yellow-100 text-yellow-800';
			case 'failed':
				return 'bg-red-100 text-red-800';
			default:
				return 'bg-gray-100 text-gray-800';
		}
	}

	function getStatusText(status: string): string {
		switch (status) {
			case 'completed':
				return 'Concluído';
			case 'processing':
				return 'Processando';
			case 'failed':
				return 'Falhou';
			default:
				return status;
		}
	}

	async function fetchClips() {
		loading = true;
		error = '';

		try {
			const response = await fetch(`/api/clips?page=${page}&limit=20`);
			if (!response.ok) throw new Error('Falha ao buscar clips');

			const data = await response.json();
			clips = data.clips || [];
			total = data.total || 0;
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	async function deleteClip(id: number) {
		if (!confirm('Tem certeza que deseja apagar este clip?')) return;

		deletingId = id;
		try {
			const response = await fetch(`/api/clips/${id}`, {
				method: 'DELETE'
			});

			if (!response.ok) throw new Error('Falha ao apagar clip');

			await fetchClips(); // Recarrega a lista
		} catch (e: any) {
			error = e.message;
		} finally {
			deletingId = null;
		}
	}

	onMount(() => {
		fetchClips();
		// Auto-refresh a cada 5 segundos se houver clips processando
		const interval = setInterval(() => {
			if (clips.some((c) => c.status === 'processing')) {
				fetchClips();
			}
		}, 5000);

		return () => clearInterval(interval);
	});
</script>

<div class="container mx-auto p-8 max-w-4xl">
	<div class="flex justify-between items-center mb-6">
		<div>
			<h1 class="text-3xl font-bold">Meus Clips</h1>
			<p class="text-gray-600 mt-1">Total: {total} clips</p>
		</div>
		<a href="/" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
			+ Novo Clip
		</a>
	</div>

	{#if loading && clips.length === 0}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
			<p class="mt-4 text-gray-600">Carregando clips...</p>
		</div>
	{:else if error}
		<div class="p-4 bg-red-100 text-red-700 rounded-lg">{error}</div>
	{:else if clips.length === 0}
		<div class="text-center py-12">
			<svg
				class="mx-auto h-12 w-12 text-gray-400"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"
				/>
			</svg>
			<h3 class="mt-4 text-lg font-medium text-gray-900">Nenhum clip ainda</h3>
			<p class="mt-2 text-gray-600">Comece criando seu primeiro clip!</p>
			<a
				href="/"
				class="mt-4 inline-block px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
			>
				Criar Clip
			</a>
		</div>
	{:else}
		<div class="space-y-4">
			{#each clips as clip (clip.id)}
				<div class="border rounded-lg p-4 hover:shadow-md transition-shadow">
					<div class="flex justify-between items-start">
						<div class="flex-1">
							<div class="flex items-center gap-2 mb-2">
								<h3 class="font-semibold text-lg">
									{clip.title || `Clip #${clip.id}`}
								</h3>
								<span class={`px-2 py-1 text-xs rounded-full ${getStatusColor(clip.status)}`}>
									{getStatusText(clip.status)}
								</span>
							</div>
							<div class="flex gap-4 text-sm text-gray-600">
								<span>Duração: {formatTime(clip.duration)}</span>
								{#if clip.size > 0}
									<span>Tamanho: {formatSize(clip.size)}</span>
								{/if}
							</div>
						</div>
						<div class="flex gap-2">
							{#if clip.status === 'completed' && clip.download_url}
								<a
									href={clip.download_url}
									download
									class="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 flex items-center gap-2"
								>
									<svg
										class="w-4 h-4"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
										/>
									</svg>
									Download
								</a>
							{/if}
							<a
								href={`/clips/${clip.id}`}
								class="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300"
							>
								Detalhes
							</a>
							<button
								type="button"
								on:click={() => deleteClip(clip.id)}
								disabled={deletingId === clip.id}
								class="px-3 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50 flex items-center gap-2"
								title="Apagar clip"
							>
								{#if deletingId === clip.id}
									<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
								{:else}
									<svg
										class="w-4 h-4"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
										/>
									</svg>
								{/if}
							</button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

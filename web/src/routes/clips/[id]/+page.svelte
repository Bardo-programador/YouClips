<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	interface Clip {
		id: number;
		status: string;
		title: string;
		duration: number;
		size: number;
		download_url?: string;
	}

	interface Progress {
		id: number;
		status: string;
		progress: number;
	}

	let clip: Clip | null = null;
	let progress: Progress | null = null;
	let loading = true;
	let error = '';
	let deleting = false;

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
		return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
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

	async function fetchClip() {
		loading = true;
		error = '';

		try {
			const id = $page.params.id;
			const response = await fetch(`/clips/${id}`);
			if (!response.ok) throw new Error('Clip não encontrado');

			const data = await response.json();
			clip = { ...data }; // Force reactivity update
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	async function fetchProgress() {
		try {
			const id = $page.params.id;
			const response = await fetch(`/clips/${id}/progress`);
			if (!response.ok) {
				progress = null;
				return;
			}

			const data = await response.json();
			progress = { ...data }; // Force reactivity update
		} catch (e: any) {
			console.error('Failed to fetch progress:', e);
			progress = null;
		}
	}

	async function deleteClip() {
		if (!clip || !confirm('Tem certeza que deseja apagar este clip?')) return;

		deleting = true;
		try {
			const response = await fetch(`/clips/${clip.id}`, {
				method: 'DELETE'
			});

			if (!response.ok) throw new Error('Falha ao apagar clip');

			goto('/');
		} catch (e: any) {
			error = e.message;
			deleting = false;
		}
	}

	onMount(() => {
		fetchClip();
		fetchProgress();

		const interval = setInterval(async () => {
			await fetchClip();
			
			// Only fetch progress if still processing
			if (clip?.status === 'processing') {
				await fetchProgress();
			} else {
				clearInterval(interval);
			}
		}, 1500);

		return () => {
			clearInterval(interval);
		};
	});
</script>

<div class="container mx-auto p-8 max-w-2xl">
	<div class="mb-6">
		<a href="/" class="text-blue-600 hover:text-blue-800 flex items-center gap-2">
			<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
			</svg>
			Voltar
		</a>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
			<p class="mt-4 text-gray-600">Carregando clip...</p>
		</div>
	{:else if error}
		<div class="p-4 bg-red-100 text-red-700 rounded-lg">{error}</div>
	{:else if clip}
		<div class="bg-white border rounded-lg p-6 shadow-sm">
			<div class="flex justify-between items-start mb-6">
				<div>
					<h1 class="text-2xl font-bold mb-2">{clip.title || `Clip #${clip.id}`}</h1>
					<span class={`inline-block px-3 py-1 text-sm rounded-full ${getStatusColor(clip.status)}`}>
						{getStatusText(clip.status)}
					</span>
				</div>
				<button
					type="button"
					on:click={deleteClip}
					disabled={deleting}
					class="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50 flex items-center gap-2"
					title="Apagar clip"
				>
					{#if deleting}
						<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
						Apagando...
					{:else}
						<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
							/>
						</svg>
						Apagar
					{/if}
				</button>
			</div>

			<div class="space-y-4">
				<!-- Progress Bar (only for processing) -->
				{#if clip.status === 'processing' && progress}
					<div class="space-y-2">
						<div class="flex justify-between items-center">
							<span class="text-sm font-medium text-gray-700">Progresso do Processamento</span>
							<span class="text-sm font-bold text-blue-600">{progress.progress}%</span>
						</div>
						<div class="w-full bg-gray-200 rounded-full h-3 overflow-hidden">
							<div
								class="bg-blue-600 h-full transition-all duration-500"
								style="width: {progress.progress}%"
							></div>
						</div>
					</div>
				{/if}

				<div class="grid grid-cols-2 gap-4">
					<div class="p-4 bg-gray-50 rounded-lg">
						<p class="text-sm text-gray-600 mb-1">Duração</p>
						<p class="text-lg font-semibold">{formatTime(clip.duration)}</p>
					</div>

					{#if clip.size > 0}
						<div class="p-4 bg-gray-50 rounded-lg">
							<p class="text-sm text-gray-600 mb-1">Tamanho</p>
							<p class="text-lg font-semibold">{formatSize(clip.size)}</p>
						</div>
					{/if}
				</div>

				{#if clip.status === 'completed' && clip.download_url}
					<a
						href={clip.download_url}
						download
						class="w-full block text-center px-6 py-3 bg-green-600 text-white rounded-lg hover:bg-green-700 font-medium flex items-center justify-center gap-2"
					>
						<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
							/>
						</svg>
						Fazer Download
					</a>
				{:else if clip.status === 'processing'}
					<div class="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
						<div class="flex items-center gap-3">
							<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-yellow-600"></div>
							<div>
								<p class="font-medium text-yellow-900">Processando clip...</p>
								<p class="text-sm text-yellow-700">
									Isso pode levar alguns minutos. A página atualizará automaticamente.
								</p>
							</div>
						</div>
					</div>
				{:else if clip.status === 'failed'}
					<div class="p-4 bg-red-50 border border-red-200 rounded-lg">
						<p class="font-medium text-red-900">Falha no processamento</p>
						<p class="text-sm text-red-700 mt-1">
							Ocorreu um erro ao processar este clip. Tente criar um novo.
						</p>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

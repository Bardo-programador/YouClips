<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';

	interface Clip {
		id: number;
		status: string;
		title: string;
		duration: number;
		size: number;
		download_url?: string;
	}

	let clip: Clip | null = null;
	let loading = true;
	let error = '';

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

			clip = await response.json();
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		fetchClip();

		// Auto-refresh se estiver processando
		const interval = setInterval(() => {
			if (clip?.status === 'processing') {
				fetchClip();
			}
		}, 3000);

		return () => clearInterval(interval);
	});
</script>

<div class="container mx-auto p-8 max-w-2xl">
	<div class="mb-6">
		<a href="/clips" class="text-blue-600 hover:text-blue-800 flex items-center gap-2">
			<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
			</svg>
			Voltar para lista
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
			</div>

			<div class="space-y-4">
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

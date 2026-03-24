<script lang="ts">
	import { goto } from '$app/navigation';

	let url = '';
	let startTime = 0;
	let endTime = 30;
	let format: 'video' | 'audio' = 'video';
	let quality = '720p'; // Default quality
	let loading = false;
	let loadingMetadata = false;
	let error = '';
	let metadata: { title: string; duration: number } | null = null;

	// Quality options for video
	const videoQualities = [
		{ value: '360p', label: '360p (Low)' },
		{ value: '480p', label: '480p (Medium)' },
		{ value: '720p', label: '720p (HD)' },
		{ value: '1080p', label: '1080p (Full HD)' }
	];

	function formatTime(seconds: number): string {
		const hours = Math.floor(seconds / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);
		const secs = seconds % 60;

		if (hours > 0) {
			return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
		}
		return `${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
	}

	async function fetchMetadata() {
		if (!url) return;

		loadingMetadata = true;
		error = '';
		metadata = null;

		try {
			const response = await fetch('/api/metadata', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ url })
			});

			if (!response.ok) throw new Error('Falha ao buscar metadados');

			metadata = await response.json();
			// Reset sliders to video range
			startTime = 0;
			endTime = Math.min(30, metadata.duration);
		} catch (e: any) {
			error = e.message;
		} finally {
			loadingMetadata = false;
		}
	}

	async function createClip() {
		if (loading) return;
		
		loading = true;
		error = '';

		try {
			const response = await fetch('/clips', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ 
					url, 
					start_time: startTime, 
					end_time: endTime, 
					format,
					quality: format === 'video' ? quality : '240p' // Use 240p for audio
				})
			});

			if (!response.ok) throw new Error('Falha ao criar clip');

			const result = await response.json();
			await goto(`/clips/${result.id}`);
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}
</script>

<div class="container mx-auto p-8 max-w-2xl">
	<div class="mb-6">
		<h1 class="text-3xl font-bold">YouClips</h1>
		<p class="text-gray-600">Crie clips de vídeos do YouTube</p>
	</div>

	<div class="space-y-6 mt-8">
		<!-- URL Input -->
		<div>
			<label for="url" class="block mb-2 font-medium">URL do YouTube</label>
			<div class="flex gap-2">
				<input
					id="url"
					type="url"
					bind:value={url}
					required
					placeholder="https://youtube.com/watch?v=..."
					class="flex-1 px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
				/>
				<button
					type="button"
					on:click={fetchMetadata}
					disabled={!url || loadingMetadata}
					class="px-6 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 disabled:opacity-50"
				>
					{loadingMetadata ? 'Carregando...' : 'Buscar'}
				</button>
			</div>
		</div>

		{#if metadata}
			<!-- Video Info -->
			<div class="p-4 bg-blue-50 rounded-lg">
				<h3 class="font-semibold text-blue-900 mb-1">{metadata.title}</h3>
				<p class="text-sm text-blue-700">Duração: {formatTime(metadata.duration)}</p>
			</div>

			<!-- Time Range Slider -->
			<div>
				<label class="block mb-2 font-medium">Selecione o intervalo</label>
				<div class="space-y-4">
					<div>
						<div class="flex justify-between text-sm text-gray-600 mb-2">
							<span>Início: {formatTime(startTime)}</span>
							<span>Fim: {formatTime(endTime)}</span>
						</div>
						<div class="relative h-2 bg-gray-200 rounded-full mt-6 mb-6">
							<!-- Track highlight -->
							<div 
								class="absolute h-full bg-blue-600 rounded-full"
								style="left: {(startTime / metadata.duration) * 100}%; right: {100 - (endTime / metadata.duration) * 100}%;"
							></div>
							
							<!-- Start thumb -->
							<input
								type="range"
								min="0"
								max={metadata.duration}
								bind:value={startTime}
								on:input={() => {
									if (startTime >= endTime) {
										endTime = Math.min(startTime + 1, metadata.duration);
									}
								}}
								class="absolute w-full h-2 appearance-none bg-transparent pointer-events-none [&::-webkit-slider-thumb]:pointer-events-auto [&::-moz-range-thumb]:pointer-events-auto [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-4 [&::-webkit-slider-thumb]:h-4 [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-blue-600 [&::-webkit-slider-thumb]:cursor-pointer [&::-moz-range-thumb]:appearance-none [&::-moz-range-thumb]:w-4 [&::-moz-range-thumb]:h-4 [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:bg-blue-600 [&::-moz-range-thumb]:cursor-pointer [&::-moz-range-thumb]:border-0"
							/>
							
							<!-- End thumb -->
							<input
								type="range"
								min="0"
								max={metadata.duration}
								bind:value={endTime}
								on:input={() => {
									if (endTime <= startTime) {
										startTime = Math.max(endTime - 1, 0);
									}
								}}
								class="absolute w-full h-2 appearance-none bg-transparent pointer-events-none [&::-webkit-slider-thumb]:pointer-events-auto [&::-moz-range-thumb]:pointer-events-auto [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-4 [&::-webkit-slider-thumb]:h-4 [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-blue-600 [&::-webkit-slider-thumb]:cursor-pointer [&::-moz-range-thumb]:appearance-none [&::-moz-range-thumb]:w-4 [&::-moz-range-thumb]:h-4 [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:bg-blue-600 [&::-moz-range-thumb]:cursor-pointer [&::-moz-range-thumb]:border-0"
							/>
						</div>
						<div class="text-sm text-gray-600 text-center mt-2">
							Duração do clip: {formatTime(endTime - startTime)}
						</div>
					</div>
				</div>
			</div>

			<!-- Format Selection -->
			<div>
				<label class="block mb-2 font-medium">Formato</label>
				<div class="flex gap-4">
					<label class="flex items-center">
						<input type="radio" bind:group={format} value="video" class="mr-2" />
						Vídeo
					</label>
					<label class="flex items-center">
						<input type="radio" bind:group={format} value="audio" class="mr-2" />
						Áudio
					</label>
				</div>
			</div>

			<!-- Quality Selection (only for video) -->
			{#if format === 'video'}
				<div>
					<label for="quality" class="block mb-2 font-medium">Qualidade do Vídeo</label>
					<select
						id="quality"
						bind:value={quality}
						class="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
					>
						{#each videoQualities as q}
							<option value={q.value}>{q.label}</option>
						{/each}
					</select>
					<p class="mt-1 text-sm text-gray-500">
						Maior qualidade = arquivo maior e processamento mais lento
					</p>
				</div>
			{:else}
				<div class="p-3 bg-blue-50 border border-blue-200 rounded-lg">
					<p class="text-sm text-blue-800">
						<strong>💡 Dica:</strong> Para áudio, usamos automaticamente vídeo de baixa qualidade (240p) antes da extração. Isso torna o download mais rápido sem afetar a qualidade do áudio.
					</p>
				</div>
			{/if}

			<!-- Submit Button -->
			<button
				type="button"
				on:click={createClip}
				disabled={loading}
				class="w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 disabled:opacity-50"
			>
				{loading ? 'Processando...' : 'Criar Clip'}
			</button>
		{/if}
	</div>

	{#if error}
		<div class="mt-4 p-4 bg-red-100 text-red-700 rounded-lg">{error}</div>
	{/if}
</div>

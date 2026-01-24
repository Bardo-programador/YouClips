<script lang="ts">
	let url = '';
	let startTime = 0;
	let endTime = 30;
	let format: 'video' | 'audio' = 'video';
	let loading = false;
	let loadingMetadata = false;
	let result: any = null;
	let error = '';
	let metadata: { title: string; duration: number } | null = null;

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
			const response = await fetch('/metadata', {
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
		loading = true;
		error = '';
		result = null;

		try {
			const response = await fetch('/clips', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ url, start_time: startTime, end_time: endTime, format })
			});

			if (!response.ok) throw new Error('Falha ao criar clip');

			result = await response.json();
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}
</script>

<div class="container mx-auto p-8 max-w-2xl">
	<div class="flex justify-between items-center mb-6">
		<div>
			<h1 class="text-3xl font-bold">YouClips</h1>
			<p class="text-gray-600">Crie clips de vídeos do YouTube</p>
		</div>
		<a href="/clips" class="px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700">
			Ver Clips
		</a>
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
						<div class="space-y-2">
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
								class="w-full"
							/>
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
								class="w-full"
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

	{#if result}
		<div class="mt-4 p-4 bg-green-100 text-green-700 rounded-lg">
			<p class="font-medium">Clip criado com sucesso!</p>
			<p class="text-sm mt-2">ID: {result.id} - Status: {result.status}</p>
		</div>
	{/if}
</div>

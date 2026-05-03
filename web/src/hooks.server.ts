import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	// Get API_URL at runtime (not build time) to support environment variable changes
	const API_URL = process.env.API_URL || 'http://api:8080';
	// Proxy /api/* and /clips* requests to the backend
	if (event.url.pathname.startsWith('/api/') || event.url.pathname.startsWith('/clips')) {
		let path = event.url.pathname;
		
		// Remove /api prefix if present
		if (path.startsWith('/api/')) {
			path = path.slice(4);
		}
		
		const queryString = event.url.search;
		const url = `${API_URL}${path}${queryString}`;

		try {
			const response = await fetch(url, {
				method: event.request.method,
				headers: {
					'Content-Type': event.request.headers.get('Content-Type') || 'application/json',
					...Object.fromEntries(
						Array.from(event.request.headers.entries()).filter(
							([key]) => !['host', 'connection'].includes(key.toLowerCase())
						)
					)
				},
				body: ['GET', 'HEAD'].includes(event.request.method)
					? undefined
					: await event.request.text()
			});

			// Read entire response and return with proper headers
			const buffer = await response.arrayBuffer();
			const headers = new Headers(response.headers);
			headers.set('Content-Length', buffer.byteLength.toString());
			
			return new Response(buffer, {
				status: response.status,
				headers
			});
		} catch (error) {
			console.error('API proxy error:', error);
			return new Response(JSON.stringify({ error: 'API proxy failed' }), {
				status: 502,
				headers: { 'Content-Type': 'application/json' }
			});
		}
	}

	return resolve(event);
};

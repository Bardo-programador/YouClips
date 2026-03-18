import type { Handle } from '@sveltejs/kit';

const API_URL = process.env.API_URL || 'http://api:8080';

export const handle: Handle = async ({ event, resolve }) => {
	// Proxy /api/* requests to the backend
	if (event.url.pathname.startsWith('/api/')) {
		const path = event.url.pathname.slice(4); // Remove /api prefix
		const queryString = event.url.search;
		const url = `${API_URL}/${path}${queryString}`;

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

			return new Response(response.body, {
				status: response.status,
				headers: response.headers
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

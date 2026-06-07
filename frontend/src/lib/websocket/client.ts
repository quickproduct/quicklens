import { wsConnected } from '$stores/ui';

export interface WebSocketClient {
	send: (data: unknown) => void;
	close: () => void;
}

export function createWebSocket(
	channel: string,
	onMessage: (data: unknown) => void,
	onOpen?: () => void,
	onClose?: () => void
): WebSocketClient {
	let ws: WebSocket | null = null;
	let reconnectAttempt = 0;
	let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	let pingTimer: ReturnType<typeof setInterval> | null = null;
	let destroyed = false;

	const maxReconnectDelay = 30000;
	const baseDelay = 1000;

	function getWsUrl(): string {
		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		const host = window.location.host;
		return `${protocol}//${host}/ws/${channel}`;
	}

	function connect(): void {
		if (destroyed) return;

		try {
			const token = localStorage.getItem('ql_access_token');
			const url = token
				? `${getWsUrl()}?token=${encodeURIComponent(token)}`
				: getWsUrl();

			ws = new WebSocket(url);

			ws.onopen = () => {
				reconnectAttempt = 0;
				wsConnected.set(true);
				onOpen?.();

				// Start ping
				pingTimer = setInterval(() => {
					if (ws && ws.readyState === WebSocket.OPEN) {
						ws.send(JSON.stringify({ type: 'ping' }));
					}
				}, 30000);
			};

			ws.onmessage = (event) => {
				try {
					const data = JSON.parse(event.data);
					if (data.type === 'pong') return;
					onMessage(data);
				} catch {
					onMessage(event.data);
				}
			};

			ws.onclose = () => {
				wsConnected.set(false);
				cleanup();
				onClose?.();
				scheduleReconnect();
			};

			ws.onerror = () => {
				wsConnected.set(false);
			};
		} catch {
			scheduleReconnect();
		}
	}

	function cleanup(): void {
		if (pingTimer) {
			clearInterval(pingTimer);
			pingTimer = null;
		}
	}

	function scheduleReconnect(): void {
		if (destroyed) return;

		const delay = Math.min(baseDelay * Math.pow(2, reconnectAttempt), maxReconnectDelay);
		reconnectAttempt++;

		reconnectTimer = setTimeout(() => {
			connect();
		}, delay);
	}

	function send(data: unknown): void {
		if (ws && ws.readyState === WebSocket.OPEN) {
			ws.send(typeof data === 'string' ? data : JSON.stringify(data));
		}
	}

	function close(): void {
		destroyed = true;
		cleanup();
		if (reconnectTimer) {
			clearTimeout(reconnectTimer);
		}
		if (ws) {
			ws.onclose = null;
			ws.onerror = null;
			ws.close();
			ws = null;
		}
		wsConnected.set(false);
	}

	// Auto-connect
	connect();

	return { send, close };
}

type WSCallback = (data: unknown) => void;

class WSClient {
  private ws: WebSocket | null = null;
  private listeners: Map<string, Set<WSCallback>> = new Map();
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private baseURL: string;
  private connected = false;

  constructor(baseURL?: string) {
    this.baseURL =
      baseURL ||
      (typeof window !== 'undefined' ? window.location.host : 'localhost:8080');
  }

  get isConnected(): boolean {
    return this.connected;
  }

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const token =
      typeof window !== 'undefined'
        ? localStorage.getItem('token')
        : '';
    const tokenParam = token ? `?token=${encodeURIComponent(token)}` : '';
    const url = `${protocol}//${this.baseURL}/api/ws${tokenParam}`;

    try {
      this.ws = new WebSocket(url);
    } catch {
      this.scheduleReconnect();
      return;
    }

    this.ws.onopen = () => {
      this.connected = true;
    };

    this.ws.onmessage = (event: MessageEvent) => {
      try {
        const msg = JSON.parse(event.data);
        const type = msg.type as string;
        const payload = msg.payload;

        const typeListeners = this.listeners.get(type);
        if (typeListeners) {
          typeListeners.forEach((cb) => cb(payload));
        }
      } catch {
        // Ignore malformed messages
      }
    };

    this.ws.onclose = () => {
      this.connected = false;
      this.ws = null;
      this.scheduleReconnect();
    };

    this.ws.onerror = () => {
      this.ws?.close();
    };
  }

  on(type: string, callback: WSCallback): () => void {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, new Set());
    }
    this.listeners.get(type)!.add(callback);

    // Return unsubscribe function
    return () => {
      this.listeners.get(type)?.delete(callback);
    };
  }

  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.ws?.close();
    this.ws = null;
    this.connected = false;
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) return;
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, 5000);
  }
}

export const wsClient = new WSClient();

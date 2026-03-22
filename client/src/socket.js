// WebSocket connection for Go backend
class WebSocketClient {
  constructor() {
    this.ws = null;
    this.listeners = {};
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 3000;
  }

  connect() {
    const username = localStorage.getItem('username') || 'guest';
    const token = localStorage.getItem('token');

    if (!token) {
      console.error('No token found, cannot connect to WebSocket');
      return;
    }

    const wsUrl = `${process.env.REACT_APP_SOCKET_ENDPOINT.replace('http', 'ws')}/ws?username=${username}&token=${token}`;

    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
      this.emit('connect');
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        const { type, payload } = data;

        if (this.listeners[type]) {
          this.listeners[type].forEach(callback => callback(payload));
        }

        // Also emit to 'any' listeners
        if (this.listeners['*']) {
          this.listeners['*'].forEach(callback => callback(type, payload));
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.emit('disconnect');
      this.attemptReconnect();
    };
  }

  attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
      setTimeout(() => this.connect(), this.reconnectDelay);
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  on(event, callback) {
    if (!this.listeners[event]) {
      this.listeners[event] = [];
    }
    this.listeners[event].push(callback);
  }

  onAny(callback) {
    this.on('*', callback);
  }

  off(event, callback) {
    if (this.listeners[event]) {
      this.listeners[event] = this.listeners[event].filter(cb => cb !== callback);
    }
  }

  emit(event, data) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const message = JSON.stringify({ type: event, payload: data });
      this.ws.send(message);
    } else {
      console.error('WebSocket is not connected');
    }
  }
}

const socket = new WebSocketClient();

export default socket;
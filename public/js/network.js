export class NetworkManager {
    constructor(game) {
        this.game = game;
        this.menu = document.getElementById('upgrade-menu');
        this.skillPointsElement = document.getElementById('skill-points');
        this.upgradeButtons = document.querySelectorAll('.upgrade-btn');
        this.socket = new WebSocket('ws://localhost:8080/ws');
        this.setupEventListeners();
    }

    setupEventListeners() {
        this.socket.onmessage = (event) => this.handleMessage(event);
        this.socket.onclose = () => {
            this.game.statusDiv.textContent = "Disconnected. Refresh page to reconnect.";
        };
    }

    handleMessage(event) {
        try {
            const data = JSON.parse(event.data);
            // Обработка сообщений...
        } catch (e) {
            console.error('Error parsing message:', e);
        }
    }

    send(data) {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify(data));
        } else {
            console.warn('WebSocket is not connected');
        }
    }
}
import { InputHandler } from './input.js';
import { NetworkManager } from './network.js';
import { Renderer } from './render.js';
import { UpgradeSystem } from './upgrades.js';

export class Game {
    constructor() {
        this.canvas = document.getElementById('gameCanvas');
        this.ctx = this.canvas.getContext('2d');
        this.statusDiv = document.getElementById('status');

        this.network = new NetworkManager(this);
        this.renderer = new Renderer(this);
        this.upgrades = new UpgradeSystem(this);
        this.input = new InputHandler(this);

        this.players = {};
        this.bullets = [];
        this.myPlayerId = null;
        this.isAuthenticated = false;
    }

    init() {
        this.resizeCanvas();
        window.addEventListener('resize', () => this.resizeCanvas());
        this.renderer.drawGrid();

        // Инициализация других систем
        this.upgrades.init();
        this.input.init();
    }

    resizeCanvas() {
        this.canvas.width = window.innerWidth;
        this.canvas.height = window.innerHeight;
        this.renderer.drawGrid();
    }
}

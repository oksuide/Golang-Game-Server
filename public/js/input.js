export class InputHandler {
    constructor(game) {
        this.game = game;
        this.canvas = game.canvas;
        this.inputState = {
            up: false, down: false, left: false, right: false,
            angle: 0, shoot: false
        };
    }

    init() {
        this.setupMouseEvents();
        this.setupKeyboardEvents();
    }

    setupMouseEvents() {
        this.canvas.addEventListener('mousemove', (e) => {
            if (!this.game.isAuthenticated || !this.game.myPlayerId || !this.game.players[this.game.myPlayerId]) {
                console.warn("Not authenticated or player not initialized");
                return;
            }

            const rect = this.canvas.getBoundingClientRect();
            const mouseX = e.clientX - rect.left;
            const mouseY = e.clientY - rect.top;

            const gameMouseX = mouseX * (1920 / this.canvas.width);
            const gameMouseY = mouseY * (1080 / this.canvas.height);

            const player = this.game.players[this.game.myPlayerId];
            const dx = gameMouseX - player.X;
            const dy = gameMouseY - player.Y;
            this.inputState.angle = Math.atan2(dy, dx);

            this.sendInput();
        });

        // Обработчик клика для стрельбы
        this.canvas.addEventListener('click', (e) => {
            this.inputState.shoot = true;
            this.sendInput();
            setTimeout(() => {
                this.inputState.shoot = false;
                this.sendInput();
            }, 100);
        });

        this.canvas.addEventListener('mousedown', (e) => {
            if (e.button === 0) {
                this.inputState.shoot = true;
                this.sendInput();
            }
        });

        this.canvas.addEventListener('mouseup', (e) => {
            if (e.button === 0) {
                this.inputState.shoot = false;
                this.sendInput();
            }
        });
    }

    setupKeyboardEvents() {
        document.addEventListener('keydown', (e) => {
            if (e.key === 'w' || e.key === 'ArrowUp') this.inputState.up = true;
            if (e.key === 's' || e.key === 'ArrowDown') this.inputState.down = true;
            if (e.key === 'a' || e.key === 'ArrowLeft') this.inputState.left = true;
            if (e.key === 'd' || e.key === 'ArrowRight') this.inputState.right = true;
            this.sendInput();
        });

        document.addEventListener('keyup', (e) => {
            if (e.key === 'w' || e.key === 'ArrowUp') this.inputState.up = false;
            if (e.key === 's' || e.key === 'ArrowDown') this.inputState.down = false;
            if (e.key === 'a' || e.key === 'ArrowLeft') this.inputState.left = false;
            if (e.key === 'd' || e.key === 'ArrowRight') this.inputState.right = false;
            this.sendInput();
        });
    }

    sendInput() {
        this.game.network.send(this.inputState);
    }
}
export class Renderer {
    constructor(game) {
        this.game = game;
        this.canvas = game.canvas;
        this.ctx = game.ctx;
    }

    toCanvasX(x) {
        return x * (this.canvas.width / 1920);
    }

    toCanvasY(y) {
        return y * (this.canvas.height / 1080);
    }

    drawGrid() {
        this.ctx.strokeStyle = 'rgba(255, 255, 255, 0.1)';
        this.ctx.lineWidth = 1;

        const gridSize = 50;
        const cols = Math.ceil(this.canvas.width / gridSize);
        const rows = Math.ceil(this.canvas.height / gridSize);

        for (let i = 0; i < cols; i++) {
            this.ctx.beginPath();
            this.ctx.moveTo(i * gridSize, 0);
            this.ctx.lineTo(i * gridSize, this.canvas.height);
            this.ctx.stroke();
        }

        for (let i = 0; i < rows; i++) {
            this.ctx.beginPath();
            this.ctx.moveTo(0, i * gridSize);
            this.ctx.lineTo(this.canvas.width, i * gridSize);
            this.ctx.stroke();
        }
    }

    drawPlayer(player) {
        const x = this.toCanvasX(player.X);
        const y = this.toCanvasY(player.Y);
        const isCurrentPlayer = player.ID === this.game.myPlayerId;

        // Player body
        this.ctx.beginPath();
        this.ctx.arc(x, y, 12, 0, Math.PI * 2);
        this.ctx.fillStyle = isCurrentPlayer ? '#00FF00' : '#FF0000';
        this.ctx.fill();

        // Player border
        this.ctx.lineWidth = 2;
        this.ctx.strokeStyle = isCurrentPlayer ? '#FFFFFF' : '#000000';
        this.ctx.stroke();

        // Direction indicator
        this.ctx.beginPath();
        this.ctx.moveTo(x, y);
        this.ctx.lineTo(
            x + 25 * Math.cos(player.Angle),
            y + 25 * Math.sin(player.Angle)
        );
        this.ctx.lineWidth = 3;
        this.ctx.strokeStyle = '#FFFFFF';
        this.ctx.stroke();

        // Player ID
        this.ctx.fillStyle = '#FFFFFF';
        this.ctx.font = 'bold 12px Arial';
        this.ctx.textAlign = 'center';
        this.ctx.textBaseline = 'middle';
        this.ctx.fillText(player.ID, x, y);
    }

    drawBullet(bullet) {
        this.ctx.beginPath();
        this.ctx.arc(
            this.toCanvasX(bullet.X),
            this.toCanvasY(bullet.Y),
            5, 0, Math.PI * 2
        );
        this.ctx.fillStyle = '#FFFF00';
        this.ctx.fill();

        // Bullet trail
        this.ctx.beginPath();
        this.ctx.moveTo(
            this.toCanvasX(bullet.X - 10 * Math.cos(bullet.Angle)),
            this.toCanvasY(bullet.Y - 10 * Math.sin(bullet.Angle))
        );
        this.ctx.lineTo(this.toCanvasX(bullet.X), this.toCanvasY(bullet.Y));
        this.ctx.strokeStyle = 'rgba(255, 255, 0, 0.5)';
        this.ctx.lineWidth = 2;
        this.ctx.stroke();
    }

    updateScene() {
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
        this.drawGrid();

        Object.values(this.game.players).forEach(player => this.drawPlayer(player));
        this.game.bullets.forEach(bullet => this.drawBullet(bullet));

        // Debug info
        if (this.game.myPlayerId && this.game.players[this.game.myPlayerId]) {
            const player = this.game.players[this.game.myPlayerId];
            this.ctx.fillStyle = 'white';
            this.ctx.font = '12px Arial';
            this.ctx.textAlign = 'left';
            this.ctx.fillText(
                `Position: (${Math.round(player.X)}, ${Math.round(player.Y)})`,
                10, 20
            );
            this.ctx.fillText(
                `Angle: ${Math.round(player.Angle * 180 / Math.PI)}°`,
                10, 40
            );
        }

        if (!this.game.isAuthenticated) {
            this.ctx.fillStyle = 'rgba(0, 0, 0, 0.7)';
            this.ctx.fillRect(
                this.canvas.width / 2 - 150,
                this.canvas.height / 2 - 25,
                300,
                50
            );
            this.ctx.fillStyle = 'white';
            this.ctx.font = '20px Arial';
            this.ctx.textAlign = 'center';
            this.ctx.fillText(
                'Connecting to server...',
                this.canvas.width / 2,
                this.canvas.height / 2
            );
        }
    }
}
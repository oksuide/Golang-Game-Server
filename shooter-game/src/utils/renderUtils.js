export function drawGrid(ctx, canvas) {
    ctx.strokeStyle = 'rgba(255, 255, 255, 0.1)';
    ctx.lineWidth = 1;

    const gridSize = 50;
    const cols = Math.ceil(canvas.width / gridSize);
    const rows = Math.ceil(canvas.height / gridSize);

    for (let i = 0; i < cols; i++) {
        ctx.beginPath();
        ctx.moveTo(i * gridSize, 0);
        ctx.lineTo(i * gridSize, canvas.height);
        ctx.stroke();
    }

    for (let i = 0; i < rows; i++) {
        ctx.beginPath();
        ctx.moveTo(0, i * gridSize);
        ctx.lineTo(canvas.width, i * gridSize);
        ctx.stroke();
    }
}

export function drawPlayer(ctx, player, canvas, isCurrentPlayer) {
    const x = player.X * (canvas.width / 1920);
    const y = player.Y * (canvas.height / 1080);

    // Player body
    ctx.beginPath();
    ctx.arc(x, y, 12, 0, Math.PI * 2);
    ctx.fillStyle = isCurrentPlayer ? '#00FF00' : '#FF0000';
    ctx.fill();

    // Player border
    ctx.lineWidth = 2;
    ctx.strokeStyle = isCurrentPlayer ? '#FFFFFF' : '#000000';
    ctx.stroke();

    // Direction indicator
    ctx.beginPath();
    ctx.moveTo(x, y);
    ctx.lineTo(
        x + 25 * Math.cos(player.Angle),
        y + 25 * Math.sin(player.Angle)
    );
    ctx.lineWidth = 3;
    ctx.strokeStyle = '#FFFFFF';
    ctx.stroke();

    // Player ID
    ctx.fillStyle = '#FFFFFF';
    ctx.font = 'bold 12px Arial';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(player.ID, x, y);
}

export function drawBullet(ctx, bullet, canvas) {
    const x = bullet.X * (canvas.width / 1920);
    const y = bullet.Y * (canvas.height / 1080);

    ctx.beginPath();
    ctx.arc(x, y, 5, 0, Math.PI * 2);
    ctx.fillStyle = '#FFFF00';
    ctx.fill();

    // Bullet trail
    ctx.beginPath();
    ctx.moveTo(
        x - 10 * Math.cos(bullet.Angle),
        y - 10 * Math.sin(bullet.Angle)
    );
    ctx.lineTo(x, y);
    ctx.strokeStyle = 'rgba(255, 255, 0, 0.5)';
    ctx.lineWidth = 2;
    ctx.stroke();
}

export function renderScene(ctx, canvas, gameState) {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    drawGrid(ctx, canvas);

    // Рендерим игроков
    Object.values(gameState.players).forEach(player => {
        drawPlayer(
            ctx,
            player,
            canvas,
            player.ID === gameState.myPlayerId
        );
    });

    // Рендерим пули
    gameState.bullets.forEach(bullet => {
        drawBullet(ctx, bullet, canvas);
    });

    // Debug информация
    if (gameState.myPlayerId && gameState.players[gameState.myPlayerId]) {
        const player = gameState.players[gameState.myPlayerId];
        ctx.fillStyle = 'white';
        ctx.font = '12px Arial';
        ctx.textAlign = 'left';
        ctx.fillText(
            `Position: (${Math.round(player.X)}, ${Math.round(player.Y)})`,
            10, 20
        );
        ctx.fillText(
            `Angle: ${Math.round(player.Angle * 180 / Math.PI)}°`,
            10, 40
        );
    }

    // Сообщение о подключении
    if (!gameState.isAuthenticated) {
        ctx.fillStyle = 'rgba(0, 0, 0, 0.7)';
        ctx.fillRect(
            canvas.width / 2 - 150,
            canvas.height / 2 - 25,
            300,
            50
        );
        ctx.fillStyle = 'white';
        ctx.font = '20px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(
            'Connecting to server...',
            canvas.width / 2,
            canvas.height / 2
        );
    }
}
import { useEffect, useRef } from 'react';

export default function useInputHandler({ canvasRef, gameState, sendInput }) {
    const inputState = useRef({
        up: false,
        down: false,
        left: false,
        right: false,
        angle: 0,
        shoot: false
    });

    // Общая функция отправки состояния ввода
    const sendCurrentInput = () => {
        sendInput(inputState.current);
    };

    useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas || !gameState.isAuthenticated) return;

        // Обработчики мыши
        const handleMouseMove = (e) => {
            const rect = canvas.getBoundingClientRect();
            const mouseX = e.clientX - rect.left;
            const mouseY = e.clientY - rect.top;

            const gameMouseX = mouseX * (1920 / canvas.width);
            const gameMouseY = mouseY * (1080 / canvas.height);

            const player = gameState.players[gameState.myPlayerId];
            if (!player) return;

            const dx = gameMouseX - player.X;
            const dy = gameMouseY - player.Y;
            inputState.current.angle = Math.atan2(dy, dx);
            sendCurrentInput();
        };

        const handleMouseClick = (e) => {
            inputState.current.shoot = true;
            sendCurrentInput();
            setTimeout(() => {
                inputState.current.shoot = false;
                sendCurrentInput();
            }, 100);
        };

        const handleMouseDown = (e) => {
            if (e.button === 0) {
                inputState.current.shoot = true;
                sendCurrentInput();
            }
        };

        const handleMouseUp = (e) => {
            if (e.button === 0) {
                inputState.current.shoot = false;
                sendCurrentInput();
            }
        };

        // Обработчики клавиатуры
        const handleKeyDown = (e) => {
            if (e.key === 'w' || e.key === 'ArrowUp') inputState.current.up = true;
            if (e.key === 's' || e.key === 'ArrowDown') inputState.current.down = true;
            if (e.key === 'a' || e.key === 'ArrowLeft') inputState.current.left = true;
            if (e.key === 'd' || e.key === 'ArrowRight') inputState.current.right = true;
            sendCurrentInput();
        };

        const handleKeyUp = (e) => {
            if (e.key === 'w' || e.key === 'ArrowUp') inputState.current.up = false;
            if (e.key === 's' || e.key === 'ArrowDown') inputState.current.down = false;
            if (e.key === 'a' || e.key === 'ArrowLeft') inputState.current.left = false;
            if (e.key === 'd' || e.key === 'ArrowRight') inputState.current.right = false;
            sendCurrentInput();
        };

        // Подписываемся на события
        canvas.addEventListener('mousemove', handleMouseMove);
        canvas.addEventListener('click', handleMouseClick);
        canvas.addEventListener('mousedown', handleMouseDown);
        canvas.addEventListener('mouseup', handleMouseUp);
        document.addEventListener('keydown', handleKeyDown);
        document.addEventListener('keyup', handleKeyUp);

        // Отписываемся при размонтировании
        return () => {
            canvas.removeEventListener('mousemove', handleMouseMove);
            canvas.removeEventListener('click', handleMouseClick);
            canvas.removeEventListener('mousedown', handleMouseDown);
            canvas.removeEventListener('mouseup', handleMouseUp);
            document.removeEventListener('keydown', handleKeyDown);
            document.removeEventListener('keyup', handleKeyUp);
        };
    }, [
        canvasRef,
        gameState.isAuthenticated,
        gameState.myPlayerId,
        gameState.players,
        sendInput
    ]);

    // Возвращаем текущее состояние для отладки
    return inputState.current;
}
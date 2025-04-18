import { useCallback, useEffect, useRef, useState } from 'react';
import { renderScene } from '../utils/renderUtils';
import useInputHandler from './useInputHandler';
import useNetworkManager from './useNetworkManager';
import useUpgradeSystem from './useUpgradeSystem';

export default function useGameEngine() {
    const canvasRef = useRef(null);
    const animationFrameId = useRef(null);
    const [gameState, setGameState] = useState({
        players: {},
        bullets: [],
        myPlayerId: null,
        isAuthenticated: false
    });

    const networkManager = useNetworkManager({ gameState, setGameState });
    const upgradeSystem = useUpgradeSystem({ gameState, setGameState, networkManager });
    const inputHandler = useInputHandler({
        canvasRef,
        gameState,
        sendInput: networkManager.send
    });

    // Функция игрового цикла
    const gameLoop = useCallback(() => {
        const canvas = canvasRef.current;
        if (!canvas) return;

        const ctx = canvas.getContext('2d');
        if (!ctx) return;

        // Отрисовка текущего состояния игры
        renderScene(ctx, canvas, gameState);

        // Продолжаем цикл
        animationFrameId.current = requestAnimationFrame(gameLoop);
    }, [gameState]);

    // Запуск и остановка игрового цикла
    useEffect(() => {
        animationFrameId.current = requestAnimationFrame(gameLoop);
        return () => {
            if (animationFrameId.current) {
                cancelAnimationFrame(animationFrameId.current);
            }
        };
    }, [gameLoop]);

    // Обработка изменения размера окна
    useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas) return;

        const resizeCanvas = () => {
            canvas.width = window.innerWidth;
            canvas.height = window.innerHeight;

            // Принудительная отрисовка при изменении размера
            const ctx = canvas.getContext('2d');
            if (ctx) {
                renderScene(ctx, canvas, gameState);
            }
        };

        resizeCanvas();
        window.addEventListener('resize', resizeCanvas);

        return () => {
            window.removeEventListener('resize', resizeCanvas);
        };
    }, [gameState]);

    return {
        canvasRef,
        gameState,
        setGameState,
        inputHandler,
        networkManager,
        upgradeSystem
    };
}
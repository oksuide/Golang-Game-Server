import { useEffect, useRef } from 'react';
import { renderScene } from '../../utils/renderUtils';
import styles from './GameCanvas.module.css';

export default function GameCanvas({ gameState }) {  // gameState получаем из пропсов
    const canvasRef = useRef(null);
    const animationFrameId = useRef(null);

    useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas || !gameState) return;

        const ctx = canvas.getContext('2d');
        if (!ctx) return;

        // Игровой цикл
        const gameLoop = () => {
            renderScene(ctx, canvas, gameState);
            animationFrameId.current = requestAnimationFrame(gameLoop);
        };

        animationFrameId.current = requestAnimationFrame(gameLoop);

        // Очистка
        return () => {
            if (animationFrameId.current) {
                cancelAnimationFrame(animationFrameId.current);
            }
        };
    }, [gameState]); // Зависимость от gameState

    return <canvas ref={canvasRef} className={styles.canvas} />;
}
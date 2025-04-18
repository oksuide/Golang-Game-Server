import { useCallback, useRef } from 'react';

export default function useNetworkManager({ setGameState }) {
    const socketRef = useRef(null);

    const connect = useCallback((token) => {
        if (!token) {
            console.error('WebSocket: Token is required');
            return;
        }

        if (socketRef.current) {
            socketRef.current.close();
        }

        const socket = new WebSocket(`ws://localhost:8080/api/ws?token=${encodeURIComponent(token)}`);
        socketRef.current = socket;

        socket.onopen = () => {
            console.log('WebSocket connected');
            setGameState(prev => ({ ...prev, isAuthenticated: true }));
        };

        socket.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                setGameState(prev => ({ ...prev, ...data }));
            } catch (e) {
                console.error('WebSocket message error:', e);
            }
        };

        socket.onclose = () => {
            setGameState(prev => ({ ...prev, isAuthenticated: false }));
        };

        socket.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        return socket;
    }, [setGameState]);

    const send = useCallback((data) => {
        if (socketRef.current?.readyState === WebSocket.OPEN) {
            socketRef.current.send(JSON.stringify(data));
        }
    }, []);

    return { connect, send };
}
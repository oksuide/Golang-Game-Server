import { useState } from 'react';
import AuthPage from './components/Auth/AuthPage';
import GameCanvas from './components/GameCanvas/GameCanvas';
import useNetworkManager from './hooks/useNetworkManager';

export default function App() {
  const [gameState, setGameState] = useState({
    isAuthenticated: false,
    players: {},
    bullets: []
  });

  const { connect } = useNetworkManager({ setGameState });

  const handleAuthSuccess = (token) => {
    connect(token);
    setGameState(prev => ({ ...prev, isAuthenticated: true }));
  };

  return (
    <div className="app">
      {gameState.isAuthenticated ? (
        <GameCanvas gameState={gameState} />
      ) : (
        <AuthPage onAuthSuccess={handleAuthSuccess} />
      )}
    </div>
  );
}
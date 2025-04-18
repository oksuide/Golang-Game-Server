import { useState } from 'react';
import { login, register } from '../../api/auth';
import AuthForm from './AuthForm';

export default function AuthPage({ onAuthSuccess }) {
    const [isLogin, setIsLogin] = useState(true);

    const handleAuth = async (credentials) => {
        try {
            const response = isLogin
                ? await login(credentials)
                : await register(credentials);

            if (!response.token) {
                throw new Error('Authentication failed: No token received');
            }

            localStorage.setItem('authToken', response.token);
            onAuthSuccess(response.token); // Передаём токен
        } catch (error) {
            console.error('Auth error:', error);
            throw error;
        }
    };

    return (
        <div className="auth-page">
            <AuthForm
                type={isLogin ? 'login' : 'register'}
                onSubmit={handleAuth}
            />
            <button onClick={() => setIsLogin(!isLogin)}>
                {isLogin ? 'Need an account? Register' : 'Have an account? Login'}
            </button>
        </div>
    );
}
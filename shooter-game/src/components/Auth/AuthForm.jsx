import { useState } from 'react';
import styles from './AuthForm.module.css';

export default function AuthForm({ type, onSubmit }) {
    const [username, setUsername] = useState('');
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const handleSubmit = async (e) => {
        e.preventDefault();
        setLoading(true);
        setError('');

        try {
            await onSubmit({ username, email, password });
        } catch (err) {
            setError(err.message || 'Ошибка регистрации');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className={styles.authContainer}>
            <h2>{type === 'login' ? 'Вход' : 'Регистрация'}</h2>
            {error && <div className={styles.error}>{error}</div>}

            <form onSubmit={handleSubmit}>
                {type === 'register' && (
                    <input
                        type="text"
                        placeholder="Имя пользователя"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        required
                        minLength="3"
                    />
                )}

                <input
                    type="email"
                    placeholder="Email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                />

                <input
                    type="password"
                    placeholder="Пароль"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                    minLength="6"
                />

                <button type="submit" disabled={loading}>
                    {loading ? 'Загрузка...' : type === 'login' ? 'Войти' : 'Зарегистрироваться'}
                </button>
            </form>
        </div>
    );
}
const API_URL = 'http://localhost:8080/api';

export const register = async ({ username, email, password }) => {
    const response = await fetch('http://localhost:8080/api/register', {
        method: 'POST', // Явно указываем метод
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            username,
            email,
            password
        }),
        credentials: 'include' // Если используете куки
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message);
    }

    return response.json();
};

export const login = async ({ email, password }) => {
    const response = await fetch(`${API_URL}/login`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            email,
            password
        }),
    });

    if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Ошибка входа');
    }

    return response.json();
};
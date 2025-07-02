document.getElementById('loginForm').addEventListener('submit', function(event) {
    event.preventDefault();

    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    fetch('http://127.0.0.1:8080/login', {
        method: 'POST',
        credentials: 'include', 

        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            username: username,
            password: password
        })
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Ошибка входа');
        }   
        return response.json();
    })
    .then(data => {
        localStorage.setItem('token', data.token); 
        window.location.href = "http://127.0.0.1";
  
    })
    .catch((error) => {
        console.error('Ошибка:', error);
        alert('Ошибка входа. Пожалуйста, проверьте ваши учетные данные.');
    });
});
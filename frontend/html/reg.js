document.getElementById('sendCodeButton').addEventListener('click', function(event) {
    event.preventDefault();

    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const email = document.getElementById('email').value;

    fetch('http://127.0.0.1:8080/sendmail', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            username: username,
            password: password,
            email: email
        })
    })
    .then(response => {
        if (!response.ok) {
            return response.json().then(errData => {
                throw new Error(errData.error || 'Ошибка регистрации');
            });
        }
        return response.json();
    })
    .then(data => {
        console.log('Успех:', data);
        alert(data.message);
    })
    .catch(error => {
        console.error('Ошибка:', error);
        alert(error.message);
    });
})

document.getElementById('registerForm').addEventListener('submit', function(event) {
    event.preventDefault(); 
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const email = document.getElementById('email').value;
    const code = document.getElementById('code').value; 
    const userData = {
        username: username,
        password: password,
        email: email,
        code: code
    };
    fetch('http://127.0.0.1:8080/reg', {
        method: 'POST',
        credentials: 'include',

        headers: {
            'Content-Type': 'application/json'
            
        },
        body: JSON.stringify(userData)
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Ошибка при регистрации: ' + response.message);
        }
        return response.json();
    })
    .then(data => {
        window.location.href = "http://127.0.0.1/login";

    })
    .catch(error => {
        console.error('Ошибка:', error);
        alert('Произошла ошибка: ' + error.message);
    });

})
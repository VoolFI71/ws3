

const token = localStorage.getItem('token');

function checkEnter(event) {
    if (event.key === 'Enter') {
        event.preventDefault(); 
        createMessage();
    }
}


function logout() {
    localStorage.removeItem('token');
    window.location.reload();
}

if (token) {
    document.getElementById('auth-buttons').style.display = 'none';
    document.getElementById('user-info').style.display = 'block';
    getUserInfo();
    getMessages();
} else {
    document.getElementById('auth-buttons').style.display = 'block';
    document.getElementById('user-info').style.display = 'none';
}

function redirectToRegister() {
    window.location.href = 'http://127.0.0.1/reg'; 
}

function redirectToLogin() {
    window.location.href = 'http://127.0.0.1/login'; 
}

function getUserInfo() {
    fetch('http://127.0.0.1:8080/userinfo', {
        method: 'GET',
        credentials: 'include',

        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        return response.json();
    })
    .then(data => {
        document.getElementById('username-display').textContent = data.username;
    })
    .catch(error => {
        console.error('Error fetching user info:', error);
    });
}

function getMessages() {
    fetch('http://127.0.0.1:8080/getmsg', {
        method: 'GET', 
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        return response.json();
    })
    .then(data => {
        const messagesList = document.getElementById('messages'); 
        messagesList.innerHTML = ''; 
        const fragment = document.createDocumentFragment(); 

        data.forEach(msg => {
            const li = document.createElement('li');
            if (msg.message) {
                li.textContent += `${msg.username}: ${msg.message} `;
            } else {
                li.textContent += `${msg.username}: `;
            }
        
            if (msg.image) {
                const img = document.createElement('img');
                img.src = msg.image; // Устанавливаем src на строку Base64
                img.style.maxWidth = '200px';
                img.style.display = 'block'; // Отображаем изображение как блок
                li.appendChild(img);
            }
        
            if (msg.audio) {
                const audio = document.createElement('audio');
                audio.controls = true; 
                audio.src = msg.audio;
                audio.load();
                li.appendChild(audio);
            }
            fragment.appendChild(li);
        });
        messagesList.append(fragment); 

    })
    .catch(error => {
        console.error('Error fetching messages:', error);
    });
}

window.onload = function() {
    getMessages(); 
};


const conn = new WebSocket(`ws://127.0.0.1:8080/ws`);

const messagesList = document.getElementById('messages');

conn.onmessage = function(event) {
    const data = JSON.parse(event.data);
    const li = document.createElement('li');
    if (data.message) {
        li.textContent = `${data.username}: ${data.message}`; 
    }

    if (data.image) {
        const img = document.createElement('img');
        img.src = data.image;
        img.style.maxWidth = '200px';
        img.style.display = 'block'; 
        li.textContent = `${data.username}:`
        li.appendChild(img);
    }
    if (data.audio) {
        const audio = document.createElement('audio');
        audio.controls = true; 
        audio.src = `data:audio/wav;base64,${data.audio}`;
        li.textContent = `${data.username}:`;
        li.appendChild(audio);
    }

    document.getElementById('messages').prepend(li);
};


function createMessage() {
    const messageInput = document.getElementById('message');
    const message = messageInput.value.trim();
    
    if (message) {
        const messageData = { message: message }; // Создаем объект с полем Message

        fetch('http://127.0.0.1:8080/savemsg', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json' // Устанавливаем заголовок для JSON
            },
            body: JSON.stringify(messageData) // Преобразуем объект в строку JSON
        })
        .then(response => {
            if (response.ok) {
                return response.json();
            } else if (response.status === 401) { 
                alert('Необходимо авторизоваться');
                throw new Error('Необходимо авторизоваться');
            } else if (response.status === 400) {
                alert('Некорректный запрос');
                throw new Error('Некорректный запрос');
            } else {
                alert('Произошла ошибка: ' + response.status);
                throw new Error('Произошла ошибка: ' + response.status);
            }
        })
        .then(data => {
            const msg = { username: data.username, message: message };
            conn.send(JSON.stringify(msg)); 
            document.getElementById('message').value = ''; 
            })
            .catch(error => {
                console.error('Ошибка:', error);
        });
    }
}
    
function createImage() {
    const imageInput = document.getElementById('image')
    const image = imageInput.files.length > 0 ? imageInput.files[0] : null;
    const formData = new FormData(); 

    formData.append('image', image);
    if (image) {
        fetch('http://127.0.0.1:8080/saveimage', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            },
            body: formData
        })
        .then(response => {
            if (response.ok) {
                return response.json();
            } else if (response.status === 401) { 
                alert('Необходимо авторизоваться');
                throw new Error('Необходимо авторизоваться');
            } else if (response.status === 400) {
                alert('Некорректный запрос');
                throw new Error('Некорректный запрос');
            } else {
                alert('Произошла ошибка: ' + response.status);
                throw new Error('Произошла ошибка: ' + response.status);
            }
        })
        .then(data => {
            const msg = { username: data.username, image: image };

            const reader = new FileReader();
            reader.onload = function(event) {
                const base64Image = event.target.result; // Получаем Base64 строку
                msg.image = base64Image; // Добавляем изображение в сообщение
                conn.send(JSON.stringify(msg)); // Отправляем сообщение по WebSocket
            };
            reader.readAsDataURL(image);
            imageInput.value = '';
            })
            .catch(error => {
                console.error('Ошибка:', error);
        })
    }
}

function checktype() {
    const imageInput = document.getElementById('image')
    if (imageInput){
        const image = imageInput.files.length > 0 ? imageInput.files[0] : null;
    }
    const message = document.getElementById('message').value.trim();
    if (message){
        createMessage()
    }
    if (image){
        createImage()
    }
}


let mediaRecorder;
let audioChunks = [];
let isRecording = false;

function toggleRecording() {
    if (isRecording) {
        mediaRecorder.stop();
        isRecording = false;
        document.getElementById('recordButton').innerText = 'Записать аудио';
    } else {
        startRecording();
        isRecording = true;
        document.getElementById('recordButton').innerText = 'Остановить запись';
    }
}

async function startRecording() {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    mediaRecorder = new MediaRecorder(stream);
    let audioChunks = []; // Объявляем переменную для хранения аудиочастей

    mediaRecorder.ondataavailable = event => {
        audioChunks.push(event.data);
    };

    mediaRecorder.onstop = async () => {
        const audioBlob = new Blob(audioChunks, { type: 'audio/wav' });
        audioChunks = [];
        const audioUrl = URL.createObjectURL(audioBlob);

        const formData = new FormData();
        formData.append('audio', audioBlob, 'audio.wav'); // Добавляем аудиофайл в FormData

        fetch('http://127.0.0.1:8080/saveaudio', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            },  
            body: formData
        })
        .then(response => {
            if (response.ok) {
                return response.json();
            } else if (response.status === 401) { 
                alert('Необходимо авторизоваться');
                throw new Error('Необходимо авторизоваться');
            } else if (response.status === 400) {
                alert('Некорректный запрос');
                throw new Error('Некорректный запрос');
            } else {
                alert('Произошла ошибка: ' + response.status);
                throw new Error('Произошла ошибка: ' + response.status);
            }
        })
        .then(data => {
            const msg = { username: data.username, audio: 1 }; 
            
            const reader = new FileReader();
            reader.onload = function(event) {
                const base64Audio = event.target.result.split(',')[1]; // Получаем только Base64 часть
                msg.audio = base64Audio; // Добавляем аудиоданные в сообщение
                conn.send(JSON.stringify(msg)); // Отправляем сообщение по WebSocket
            };
            reader.readAsDataURL(audioBlob); 
        })
        .catch(error => {
            console.error('Ошибка сети:', error);
        });
    };

    mediaRecorder.start();
}


document.addEventListener("DOMContentLoaded", function() {
    const chatListElement = document.getElementById("chatList");

    fetch("http://127.0.0.1:8080/chats", { // Замените на ваш URL API
        method: "GET",
        headers: {
            "Authorization": `Bearer ${token}`,
            "Content-Type": "application/json"
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error("Ошибка при получении чатов: " + response.statusText);
        }
        return response.json();
    })
    .then(data => {
        // Обработка полученных данных
        if (data.length === 0) {
            chatListElement.innerHTML = "<p>Нет доступных чатов.</p>";
            return;
        }

        data.forEach(chat => {
            const chatButton = document.createElement("button");
            chatButton.id = "chatbutton"; // Устанавливаем id для кнопки
            chatButton.innerText = `Чат ID: ${chat.chat_id}, Название: ${chat.name}`;
            
            // Добавляем обработчик события для кнопки
            chatButton.addEventListener("click", function() {
                // Здесь можно добавить логику для обработки нажатия на кнопку
                console.log(`Кнопка чата ${chat.chat_id} нажата`);
                // Например, можно открыть чат или выполнить другой запрос
            });

            chatListElement.appendChild(chatButton);
        });
    })
    .catch(error => {
        console.error("Ошибка:", error);
        chatListElement.innerHTML = "<p>Произошла ошибка при загрузке чатов.</p>";
    });
});
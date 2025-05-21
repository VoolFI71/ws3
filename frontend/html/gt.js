const token = localStorage.getItem('token');

        if (!token) {
            alert('Пожалуйста, войдите в систему для доступа к этому ресурсу.');
        } else {
            fetch('http://glebase.ru:8080/gt', {
                method: 'GET',
                credentials: 'include',
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            })
            .then(response => {
                if (!response.ok) {
                    throw new Error('Ошибка доступа к защищенному ресурсу');
                }
                return response.json();
            })
            .then(data => {
                console.log('Данные защищенного ресурса:', data);
            })
            .catch((error) => {
                console.error('Ошибка:', error);
            });
        }
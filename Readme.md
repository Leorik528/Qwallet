# Используемые технологии:
golang
postgres
docker / docker-compose

# Библиотеки:
github.com/gorilla/mux - Настройка роутера

github.com/lib/pq - Драйвер для postgres

github.com/joho/godotenv - Загрузка переменных окружения из файла .env


# Сборка образа
docker build -t qwallet-app .

# Запуск контейнера
docker run -p 8080:8080 --env-file .env qwallet-app

# Или с Docker Compose
docker-compose up --build

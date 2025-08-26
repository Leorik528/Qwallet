# Сборка образа
docker build -t qwallet-app .

# Запуск контейнера
docker run -p 8080:8080 --env-file .env qwallet-app

# Или с Docker Compose
docker-compose up --build

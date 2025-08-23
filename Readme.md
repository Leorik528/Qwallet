


docker run --name qwallets \
  -e POSTGRES_PASSWORD=12345 \
  -e POSTGRES_DB=qwallets \
  -p 5432:5432 \
  -d postgres





psql -h localhost -p 5432 -U postgres -d qwallets

migrate -path migrations -database "postgresql://postgres:12345@localhost:5432/qwallets?sslmode=disable" -verbose up

migrate create -ext sql -dir migrations -seq init

migrate -database "postgresql://root:12345@localhost:5432/qwallets?sslmode=disable" -path migrations up

migrate create -ext sql -dir migrations -seq transactions

git clone <твой репозиторий>
cd проект
docker-compose up -d



version: "3.9"

services:
  db:
    image: postgres:15
    container_name: qwallets
    environment:
      POSTGRES_PASSWORD: 12345
      POSTGRES_DB: qwallets
    ports:
      - "5432:5432"
    volumes:
      - db_data:/var/lib/postgresql/data

volumes:
  db_data:

DATABASE_DSN=postgres://myuser:mypass@localhost:5432/mydb?sslmode=disable

run:
	DATABASE_DSN="$(DATABASE_DSN)" go run cmd/shortener/main.go

run-db: 
	docker-compose up -d

down-db:
	docker-compose down

clean: 
	docker-compose down -v	
	
FROM golang:1.23.6-alpine
WORKDIR /app
RUN apk update && apk add --no-cache ffmpeg
COPY go.mod go.sum ./              
RUN go mod download                # Загружаем зависимости
COPY . .                           
RUN go build -o /app/music-service ./cmd/music-service/main.go
RUN go build -o /app/user-service ./cmd/user-service/main.go

EXPOSE 11000 12000
CMD ["./music-service", "&", "./user-service"]



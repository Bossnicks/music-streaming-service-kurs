package auth

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

func LoadEnv() {
	if err := godotenv.Load("pkg/auth/.env"); err != nil {
		log.Println("Предупреждение: .env файл не найден, используются системные переменные окружения")
	}
}

func SendResetEmail(to string, resetLink string) error {
	// Получаем параметры из .env
	LoadEnv()
	from := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))

	if err != nil {
		return fmt.Errorf("Неверный порт: %v", err)
	}

	// Формируем ссылку для сброса пароля
	message := fmt.Sprintf("<html><body><p>Для сброса пароля перейдите по <a href='%s'>ссылке</a>.</p></body></html>", resetLink)

	// Подключаемся к SMTP-серверу
	dialer := gomail.NewDialer(host, port, from, password)

	// Создаём письмо
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", "Восстановление пароля")
	msg.SetBody("text/html", message) // Указываем, что письмо в формате HTML

	// Отправляем письмо
	if err := dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("ошибка отправки письма: %v", err)
	}
	return nil
}

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/siavoid/task-manager/repo/dbsqlite"
	"github.com/siavoid/task-manager/tests"
	"github.com/siavoid/task-manager/usecase/httpserver"
)

func main() {

	db, err := dbsqlite.New()
	if err != nil {
		fmt.Print(err)
		return
	}

	db.GetTask(1)

	// Установка порта по умолчанию
	port := strconv.Itoa(tests.Port)
	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		port = envPort
	}

	// Директория, из которой будут раздаваться файлы
	webDir := "./web"
	server := httpserver.New(webDir, db)

	// Запуск сервера
	log.Printf("Сервер запущен на порту %s", port)
	server.Run("127.0.0.1:" + port)
}

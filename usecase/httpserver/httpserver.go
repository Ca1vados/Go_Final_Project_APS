package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/siavoid/task-manager/entity"
	sqliterepo "github.com/siavoid/task-manager/repo/sqlite_repo"
	"github.com/siavoid/task-manager/usecase"
)

// Server структура для HTTP-сервера
type Server struct {
	Router *mux.Router
	u      *usecase.Usecase
}

// New создает новый экземпляр сервера
func New(webDir string, db *sqliterepo.SqliteRepo) *Server {
	u := usecase.New(db)
	s := &Server{
		Router: mux.NewRouter(),
		u:      u,
	}

	// Определение обработчиков
	s.Router.HandleFunc("/api/nextdate", s.NextDateAPI).Methods("GET")
	s.Router.HandleFunc("/api/tasks", s.GetAllTasksAPI).Methods("GET")
	s.Router.HandleFunc("/api/task", s.CreateTaskAPI).Methods("POST")
	s.Router.HandleFunc("/api/task", s.GetTaskAPI).Methods("GET")
	s.Router.HandleFunc("/api/task", s.UpdateTaskAPI).Methods("PUT")
	s.Router.HandleFunc("/api/task/done", s.MarkTaskDoneAPI).Methods("POST")
	s.Router.HandleFunc("/api/task", s.DeleteTaskAPI).Methods("DELETE")

	s.Router.PathPrefix("/").Handler(http.FileServer(http.Dir(webDir)))
	return s
}

// Run запускает HTTP-сервер
func (s *Server) Run(addr string) {
	fmt.Println("Сервер запущен на", addr)
	http.ListenAndServe(addr, s.Router)
}

// Заглушки для обработчиков (реализация функций не требуется)
func (s *Server) NextDateAPI(w http.ResponseWriter, r *http.Request) {
	// Извлечение параметров запроса
	nowStr := r.URL.Query().Get("now")
	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	// Проверка обязательных параметров
	if nowStr == "" || date == "" || repeat == "" {
		http.Error(w, "все параметры (now, date, repeat) обязательны.", http.StatusBadRequest)
		return
	}

	// Разбор дат
	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "неверный формат даты для параметра 'now'.", http.StatusBadRequest)
		return
	}

	// Вызов функции NextDate для вычисления следующей даты
	nextDate, err := usecase.NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Форматирование и отправка результата
	fmt.Fprintf(w, nextDate)
}

// CreateTaskAPI обработчик для создания задачи
func (s *Server) CreateTaskAPI(w http.ResponseWriter, r *http.Request) {
	var task entity.Task

	// Декодируем JSON-запрос в структуру Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Сохраняем задачу
	taskId, err := s.u.CreateTask(task)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Успешный ответ
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]int{"id": taskId})
}

func (s *Server) GetAllTasksAPI(w http.ResponseWriter, r *http.Request) {
	// Обработчик для получения всех задач
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	tasks, err := s.u.GetAllTask()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	type TaskResponse struct {
		ID      string `json:"id"`
		Date    string `json:"date"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
		Repeat  string `json:"repeat"`
	}

	var response struct {
		Tasks []TaskResponse `json:"tasks"`
	}
	response.Tasks = make([]TaskResponse, 0)

	for _, task := range tasks {
		response.Tasks = append(response.Tasks, TaskResponse{
			ID:      strconv.Itoa(task.ID), // Конвертация ID в строку
			Date:    task.Date,
			Title:   task.Title,
			Comment: task.Comment,
			Repeat:  task.Repeat,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) GetTaskAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Не указан идентификатор"})
		return
	}

	// Конвертируем id в целое число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат идентификатора"})
		return
	}

	// Получаем задачу из базы данных
	task, err := s.u.GetTask(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Задача не найдена"})
		return
	}

	// Формируем ответ
	response := map[string]string{
		"id":      strconv.Itoa(task.ID),
		"date":    task.Date,
		"title":   task.Title,
		"comment": task.Comment,
		"repeat":  task.Repeat,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) UpdateTaskAPI(w http.ResponseWriter, r *http.Request) {
	type TaskReq struct {
		ID      string `json:"id"`
		Date    string `json:"date"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
		Repeat  string `json:"repeat"`
	}

	var req TaskReq
	// Декодируем JSON-запрос в структуру Task
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный запрос"})
		return
	}
	defer r.Body.Close()
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID"})
		return
	}

	task := entity.Task{
		ID:      id,
		Date:    req.Date,
		Title:   req.Title,
		Comment: req.Comment,
		Repeat:  req.Repeat,
	}

	// Проверяем обязательные поля
	if task.ID == 0 || task.Title == "" {
		json.NewEncoder(w).Encode(map[string]string{"error": "Некорректные данные"})
		return
	}

	// Проверяем формат даты
	if task.Date != "" {
		if _, err := time.Parse("20060102", task.Date); err != nil {
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат даты"})
			return
		}
	}

	// Обновляем задачу
	err = s.u.UpdateTask(task)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Успешный ответ
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{})
}

func (s *Server) MarkTaskDoneAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Не указан идентификатор"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат идентификатора"})
		return
	}

	err = s.u.MarkTaskDone(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Успешный ответ
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

func (s *Server) DeleteTaskAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Не указан идентификатор"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат идентификатора"})
		return
	}

	if err := s.u.DeleteTask(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка удаления задачи"})
		return
	}

	// Успешный ответ
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

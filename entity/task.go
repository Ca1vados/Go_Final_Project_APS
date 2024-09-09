package entity

// Task представляет собой задачу в TODO-листе
type Task struct {
	ID      int    `json:"id"`      // Автоинкрементный идентификатор
	Date    string `json:"date"`    // Дата задачи в формате YYYYMMDD
	Title   string `json:"title"`   // Заголовок задачи
	Comment string `json:"comment"` // Комментарий к задаче
	Repeat  string `json:"repeat"`  // Правила повторения задачи
}

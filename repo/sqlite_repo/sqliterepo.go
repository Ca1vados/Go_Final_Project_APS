package sqliterepo

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/siavoid/task-manager/entity"
)

// SqliteRepo предоставляет методы для взаимодействия с базой данных
type SqliteRepo struct {
	db *sql.DB
}

// New создает новое соединение с базой данных и инициализирует её, если необходимо
func New() (*SqliteRepo, error) {
	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	// Определяем путь к базе данных
	dbPath := filepath.Join(filepath.Dir(appPath), "scheduler.db")
	if envDBPath := os.Getenv("TODO_DBFILE"); envDBPath != "" {
		dbPath = envDBPath
	}

	// Проверяем существование файла базы данных
	_, err = os.Stat(dbPath)
	createDB := os.IsNotExist(err)

	// Открываем соединение с базой данных
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Если база данных не существует, создаем таблицу и индекс
	if createDB {
		createTableQuery := `
	 CREATE TABLE scheduler (
	  id INTEGER PRIMARY KEY AUTOINCREMENT,
	  date TEXT,
	  title TEXT,
	  comment TEXT,
	  repeat TEXT
	 );
	 CREATE INDEX idx_date ON scheduler(date);
	 `

		_, err = db.Exec(createTableQuery)
		if err != nil {
			return nil, err
		}
		log.Println("Таблица scheduler и индекс созданы")
	}

	return &SqliteRepo{db: db}, nil
}

// AddTask добавляет новую задачу в базу данных
func (repo *SqliteRepo) CreateTask(task entity.Task) (int, error) {
	// Начинаем транзакцию
	tx, err := repo.db.Begin()
	if err != nil {
		return 0, err
	}

	// Выполняем вставку задачи
	result, err := tx.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
		task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		tx.Rollback() // Откатываем транзакцию в случае ошибки
		return 0, err
	}

	// Получаем ID последней вставленной записи
	taskId, err := result.LastInsertId()
	if err != nil {
		tx.Rollback() // Откатываем транзакцию в случае ошибки
		return 0, err
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return int(taskId), nil
}

// GetAllTasks возвращает все задачи из базы данных
func (repo *SqliteRepo) GetAllTasks() ([]entity.Task, error) {
	rows, err := repo.db.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []entity.Task
	for rows.Next() {
		var task entity.Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// RemoveTask удаляет задачу по ID
func (repo *SqliteRepo) RemoveTask(id int) error {
	_, err := repo.db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	return err
}

// GetTask возвращает задачу по ID
func (repo *SqliteRepo) GetTask(id int) (entity.Task, error) {
	row := repo.db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id)

	var task entity.Task
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	return task, err
}

// UpdateTask обновляет задачу в базе данных
func (repo *SqliteRepo) UpdateTask(task entity.Task) error {
	// Проверяем существование задачи с данным ID
	var exists bool
	err := repo.db.QueryRow("SELECT EXISTS(SELECT 1 FROM scheduler WHERE id = ?)", task.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка проверки существования задачи: %w", err)
	}

	if !exists {
		return errors.New("задача не найдена")
	}

	_, err = repo.db.Exec("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?",
		task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	return err
}

// dbPath возвращает путь к базе данных
func (repo *SqliteRepo) dbPath() string {
	appPath, _ := os.Executable()
	return filepath.Join(filepath.Dir(appPath), "scheduler.db")
}

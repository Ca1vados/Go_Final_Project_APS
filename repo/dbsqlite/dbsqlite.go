package dbsqlite

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

// DbSqlite предоставляет методы для взаимодействия с базой данных
type DbSqlite struct {
	db *sql.DB
}

// New создает новое соединение с базой данных и инициализирует её, если необходимо
func New() (*DbSqlite, error) {
	appPath, err := os.Executable()
	if err != nil {
		return nil, err
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

	return &DbSqlite{db: db}, nil
}

// AddTask добавляет новую задачу в базу данных
func (repo *DbSqlite) CreateTask(task entity.Task) (int, error) {
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
func (repo *DbSqlite) GetAllTasks() ([]entity.Task, error) {
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// RemoveTask удаляет задачу по ID
func (repo *DbSqlite) RemoveTask(id int) error {
	_, err := repo.db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	return err
}

// GetTask возвращает задачу по ID
func (repo *DbSqlite) GetTask(id int) (entity.Task, error) {
	row := repo.db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id)

	var task entity.Task
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	return task, err
}

// UpdateTask обновляет задачу в базе данных
func (repo *DbSqlite) UpdateTask(task entity.Task) error {
	result, err := repo.db.Exec(
		"UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?",
		task.Date, task.Title, task.Comment, task.Repeat, task.ID,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления задачи: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("задача не найдена, обновление не выполнено")
	}

	return nil
}

// dbPath возвращает путь к базе данных
func (repo *DbSqlite) dbPath() string {
	appPath, _ := os.Executable()
	return filepath.Join(filepath.Dir(appPath), "scheduler.db")
}

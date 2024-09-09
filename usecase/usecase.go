package usecase

import (
	sqliterepo "github.com/siavoid/task-manager/repo/sqlite_repo"
)

type Usecase struct {
	db *sqliterepo.SqliteRepo
}

func New(db *sqliterepo.SqliteRepo) *Usecase {
	return &Usecase{db: db}
}

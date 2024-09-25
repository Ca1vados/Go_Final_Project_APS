package usecase

import (
	"github.com/siavoid/task-manager/repo/dbsqlite"
)

type Usecase struct {
	db *dbsqlite.DbSqlite
}

func New(db *dbsqlite.DbSqlite) *Usecase {
	return &Usecase{db: db}
}

package bookRepository

import (
	"bitbucket.org/graph-ql-schema/sbuilder/v2"
	"database/sql"
)

type BookRepositoryInterface interface {
	HandleError(err error, message string) error
	InitTx(baseTx *sql.Tx) (tx *sql.Tx, err error)
	GetBooks(params sbuilder.Parameters) (interface{}, error)
	AddBook(params sbuilder.Parameters) (interface{}, error)
	UpdateBook(params sbuilder.Parameters) (interface{}, error)
	RemoveBook(params sbuilder.Parameters) (interface{}, error)
}
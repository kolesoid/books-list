package bookRepository

import (
	"bitbucket.org/graph-ql-schema/gql-sql-converter"
	"bitbucket.org/graph-ql-schema/sbuilder/v2"
	"books-list/models"
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"log"
	"strings"
)

type BookRepository struct {
	db *sqlx.DB
}

func NewBookRepository(connection *sqlx.DB) BookRepositoryInterface {
	return &BookRepository{
		db: connection,
	}
}

func (b BookRepository) HandleError(err error, message string) error {
	err = errors.Wrap(err, message)
	log.Fatalln(message)

	return err
}

func (b BookRepository) InitTx(baseTx *sql.Tx) (tx *sql.Tx, err error) {
	tx = baseTx

	if nil == tx {
		tx, err = b.db.Begin()
		if nil != err {
			err = b.HandleError(err, "failed to start transaction")

			return
		}
	}

	return
}

func (b BookRepository) GetBooks(params sbuilder.Parameters) (i interface{}, err error) {
	where := ""
	if nil != params.Arguments.Where {
		where, err = params.Arguments.Where.ToSQL(params.GraphQlObject, map[string]string{})

		if nil != err {
			return nil, err
		}
	}

	q := `SELECT * FROM books`
	if 0 != len(where) {
		q += ` where ` + where
	}

	tx, err := b.InitTx(nil)
	if err != nil {
		log.Fatal(err)
	}

	books := make([]models.Book, 0, 10)
	ctx := context.Background()
	rows, err := tx.QueryContext(ctx, q)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var book models.Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Year)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		books = append(books, book)
		if len(books) == cap(books) {
			booksExpanded := make([]models.Book, len(books), cap(books) + 10)
			copy(booksExpanded, books)
			books = booksExpanded
		}
	}

	return books, nil
}

func (b BookRepository) AddBook(params sbuilder.Parameters) (i interface{}, err error) {
	if 0 == len(params.Arguments.Objects) {
		return nil, fmt.Errorf(`you should pass any entity to store in 'objects'`)
	}

	tx, err := b.InitTx(nil)
	if err != nil {
		log.Fatal(err)
	}

	ids := []string{}
	for _, object := range params.Arguments.Objects {
		id, err := insertNewEntity(object, tx, map[string]string{})

		if nil != err {
			_ = tx.Rollback()
			return nil, err
		}

		if nil == id {
			_ = tx.Rollback()
			return nil, fmt.Errorf(`failed to create entity. returned empty ID`)
		}

		ids = append(ids, fmt.Sprintf(`%v`, *id))
	}

	if 0 == len(ids) {
		_ = tx.Rollback()
		return map[string]interface{}{
			"affected_rows": 0,
			"returning":     []models.Book{},
		}, nil
	}

	var result []models.Book
	q := fmt.Sprintf(
		`select id, title, author, year from books where id in (%v)`,
		strings.Join(ids, `,`),
	)

	ctx := context.Background()
	rows, err := tx.QueryContext(ctx, q)
	if nil != err {
		_ = tx.Rollback()
		return nil, err
	}

	for rows.Next() {
		var id int64
		var title string
		var author string
		var year int64

		err := rows.Scan(&id, &title, &author, &year)
		if nil != err {
			_ = tx.Rollback()
			return nil, err
		}

		data := models.Book{
			ID:         id,
			Title: 		title,
			Author:     author,
			Year:       year,
		}

		result = append(result, data)
	}

	err = tx.Commit()
	if nil != err {
		err = b.HandleError(err, "failed to commit transaction")
	}

	return map[string]interface{}{
		"affected_rows": len(result),
		"returning":     result,
	}, nil
}

func (b BookRepository) UpdateBook(params sbuilder.Parameters) (i interface{}, err error) {
	valueConverter := gql_sql_converter.NewGraphQlSqlConverter()

	if 0 == len(params.Arguments.Set) {
		return nil, fmt.Errorf(`you should pass any field to store in 'set'`)
	}

	tx, err := b.InitTx(nil)
	if err != nil {
		log.Fatal(err)
	}

	mappedFields := map[string]string{}

	fields := []string{}
	arguments := []interface{}{}
	for code, val := range params.Arguments.Set {
		fieldCode := code
		if mappedField, ok := mappedFields[fieldCode]; ok {
			fieldCode = mappedField
		}

		convertedVal, err := valueConverter.ToSQLValue(params.GraphQlObject, code, val)
		if nil != err {
			_ = tx.Rollback()
			return nil, err
		}

		arguments = append(arguments, convertedVal)

		fields = append(
			fields,
			fmt.Sprintf(
				`%v = $%v`,
				fieldCode,
				len(arguments),
			),
		)
	}

	if 0 == len(fields) {
		_ = tx.Rollback()
		return nil, fmt.Errorf(`you should pass any field to store in 'set'`)
	}

	query := `update books set ` + strings.Join(fields, `, `)
	where := ""
	if nil != params.Arguments.Where {
		where, err = params.Arguments.Where.ToSQL(params.GraphQlObject, mappedFields)

		if nil != err {
			_ = tx.Rollback()
			return nil, err
		}
	}

	if 0 != len(where) {
		query += ` where ` + where + ` returning id, title, year, author`
	}

	ctx := context.Background()
	result, err := tx.ExecContext(ctx, query, arguments...)
	if nil != err {
		_ = tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if nil != err {
		err = b.HandleError(err, "failed to commit transaction")
	}

	rowsUpdated, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"affected_rows": rowsUpdated,
	}, nil
}

func (b BookRepository) RemoveBook(params sbuilder.Parameters) (i interface{}, err error) {
	where := ""
	if nil != params.Arguments.Where {
		where, err = params.Arguments.Where.ToSQL(params.GraphQlObject, map[string]string{})

		if nil != err {
			return nil, err
		}
	}

	query := `delete from books`

	if 0 != len(where) {
		query += ` where ` + where
	}

	ctx := context.Background()
	tx, err := b.InitTx(nil)
	if err != nil {
		log.Fatal(err)
	}

	result, err := tx.ExecContext(ctx, query)

	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if nil != err {
		err = b.HandleError(err, "failed to commit transaction")
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"affected_rows": affectedRows,
	}, nil
}

func insertNewEntity(
	object map[string]interface{},
	transaction *sql.Tx,
	fieldsMap map[string]string,
) (*int, error) {
	fields := []string{}
	values := []interface{}{}

	if 0 == len(object) {
		return nil, fmt.Errorf(`passed empty object to insert`)
	}

	for fieldCode, value := range object {
		if mappedField, ok := fieldsMap[fieldCode]; ok {
			fieldCode = mappedField
		}

		fields = append(fields, fieldCode)
		values = append(values, value)
	}

	parsedValues := []string{}
	for _, val := range values {
		switch v := val.(type) {
		case string:
			parsedValues = append(parsedValues, fmt.Sprintf(`'%v'`, v))
			break
		default:
			parsedValues = append(parsedValues, fmt.Sprintf(`%v`, v))
		}
	}

	var lastInsertId int
	query := fmt.Sprintf(
		`insert into books(%v) values (%v) returning id`,
		strings.Join(fields, `,`),
		strings.Join(parsedValues, `,`),
	)

	err := transaction.QueryRow(query).Scan(&lastInsertId)
	if nil != err {
		return nil, err
	}

	return &lastInsertId, nil
}

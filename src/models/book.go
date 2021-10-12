package models

import (
	"github.com/graphql-go/graphql"
)

type Book struct {
	ID       int64     `pg.field:"id" json:"id" pg.type:"int"`
	Title    string    `pg.field:"title" json:"title" pg.type:"string"`
	Author   string    `pg.field:"author" json:"author" pg.type:"string"`
	Year     int64    `pg.field:"year" json:"year" pg.type:"int"`
}

var BookObject = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "book",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.ID,
				Name: "id",
			},
			"title": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Name: "title",
			},
			"author": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Name: "author",
			},
			"year": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
				Name: "year",
			},
		},
		Description: "book entity",
	},
)


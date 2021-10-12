package main

import (
	"bitbucket.org/graph-ql-schema/sbuilder/v2"
	"books-list/connection"
	"books-list/models"
	"books-list/repository/bookRepository"
	"log"
)

func main()  {
	schemaBuilder := sbuilder.NewGraphQlSchemaBuilder()
	BookRepository := bookRepository.NewBookRepository(connection.Connection)

	schemaBuilder.RegisterEntity(models.BookObject, sbuilder.EntityQueries{
		sbuilder.ListQuery: BookRepository.GetBooks,
		sbuilder.InsertMutation: BookRepository.AddBook,
		sbuilder.UpdateMutation: BookRepository.UpdateBook,
		sbuilder.DeleteMutation: BookRepository.RemoveBook,
	})

	schemaBuilder.SetServerConfig(sbuilder.GraphQlServerConfig{
		Host: "0.0.0.0",
		Port: 9000,
		Uri:  "/graphql",
	})

	server, err := schemaBuilder.BuildServer()
	if nil != err {
		log.Fatal(err)
	}

	log.Fatal(server.Run())
}
package connection

import (
	"fmt"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/subosito/gotenv"
	"log"
	"os"
)


var Connection *sqlx.DB

func init()  {
	_ = gotenv.Load()

	log.Println("Init connection")

	driver := &stdlib.DriverConfig{
		ConnConfig: pgx.ConnConfig{
			RuntimeParams: map[string]string{
				"standard_conforming_strings": "on",
			},
			PreferSimpleProtocol: false,
		},
	}
	stdlib.RegisterDriverConfig(driver)

	db, err := sqlx.Connect(
		"pgx",
		driver.ConnectionString(fmt.Sprintf(
			`host=%v port=%v user=%v password=%v dbname=%v sslmode=%v`,
			os.Getenv("PG_HOST"),
			os.Getenv("PG_PORT"),
			os.Getenv("PG_USER"),
			os.Getenv("PG_PASSWORD"),
			os.Getenv("PG_DB"),
			os.Getenv("PG_SSL_MODE"),
		)),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to DB")
	Connection = db
	Connection.SetMaxIdleConns(10)
	Connection.SetMaxOpenConns(20)
}

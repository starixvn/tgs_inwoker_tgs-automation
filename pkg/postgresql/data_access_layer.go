package postgresql

import (
	"database/sql"
	"fmt"
	"tgs-automation/internal/config"
	"tgs-automation/internal/log"
)

func NewDataAccessLayer(config config.PostgresqlConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		config.PgUsername,
		config.PgPassword,
		config.PgHost,
		config.PgDb,
	)

	log.LogInfo("conn:" + connStr)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.LogFatal(fmt.Sprintf("Failed to connect to database: %v", err))
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		log.LogFatal(fmt.Sprintf("Failed to ping database: %v", err))
		panic(err)
	}

	return db, nil
}

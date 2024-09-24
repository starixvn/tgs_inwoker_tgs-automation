package postgresql

import (
	"fmt"
	"tgs-automation/internal/config"
	"tgs-automation/internal/log"

	_ "github.com/lib/pq"
)

func GetBrandId(config config.PostgresqlConfig, brandCode string) (string, error) {
	db, _ := NewDataAccessLayer(config)
	defer db.Close()
	var brandId string
	query := `SELECT id FROM dbo.brands WHERE code = $1`
	err := db.QueryRow(query, brandCode).Scan(&brandId)

	if err != nil {
		log.LogError(fmt.Sprintf("Error executing query: %v", err))
		return "", err
	}

	return brandId, nil
}

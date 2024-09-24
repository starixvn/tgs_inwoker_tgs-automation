package postgresql

import (
	"database/sql"
	"tgs-automation/internal/config"

	_ "github.com/lib/pq"
)

type GetPromotionInterface interface {
	GetPromotionDistributions(brand, startDate, endDate string) ([]PromotionDistribute, error)
	GetPromotionTypes() ([]Category, error)
	Close()
}

type GetPromotionService struct {
	Db     *sql.DB
	Config config.PostgresqlConfig
}

func (s *GetPromotionService) Close() {
	s.Db.Close()
}
func NewPromotionInterface(config config.PostgresqlConfig) GetPromotionInterface {
	db, err := NewDataAccessLayer(config)

	if err != nil {
		panic(err)
	}
	return &GetPromotionService{
		Db:     db,
		Config: config,
	}
}

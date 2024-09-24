package postgresql

import (
	"tgs-automation/internal/config"

	_ "github.com/lib/pq"
)

// 領取紅利人數

type BonusPlayerCountService struct {
	Config config.PostgresqlConfig
}

func (bpc *BonusPlayerCountService) GetPlayerCount(brandCode, startDate, endDate string) (int, error) {
	query := "SELECT report.get_bonus_player_count($1, $2, $3);"
	return getPlayerCount(bpc.Config, query, brandCode, startDate, endDate)
}

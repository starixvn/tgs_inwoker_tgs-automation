package main

// go install github.com/google/wire/cmd/wire@latest
import (
	"net/http"
	"tgs-automation/internal/config"
	tgsHttp "tgs-automation/internal/http"
	"tgs-automation/pkg/googlesheet"
	"tgs-automation/pkg/namecheap"
	"tgs-automation/pkg/postgresql"
	"time"
)

func NewExpiredDomainsService(config config.TgsConfig) *ExpiredDomainsService {
	client := providerHttpClient()
	return &ExpiredDomainsService{
		httpClient:           client,
		googlesheetInterface: providerGoogleSheetInterface(config.GoogleSheet),
		googlesheetSvc:       providerGoogleSheetSvc(config.GoogleSheet),
		postgresqlInterface:  providerPostgresql(config.Postgresql),
		namecheapInterface:   providerNameCheap(config, client),
	}
}

func providerNameCheap(cfg config.TgsConfig, client *tgsHttp.StandardHTTPClient) namecheap.NamecheapApi {
	return namecheap.New(cfg)
}

func providerPostgresql(cfg config.PostgresqlConfig) postgresql.GetAgentServiceInterface {
	return postgresql.New(cfg)
}

func providerGoogleSheetSvc(cfg config.GoogleSheetConfig) *googlesheet.GoogleSheetService {
	itf, svc, err := googlesheet.New(cfg)
	if err != nil {
		return svc
	}

	if itf != nil {
		return svc
	}
	return nil
}

func providerGoogleSheetInterface(cfg config.GoogleSheetConfig) googlesheet.GoogleSheetServiceInterface {
	itf, svc, err := googlesheet.New(cfg)
	if err != nil {
		return itf
	}
	if svc != nil {
		return svc
	}
	return nil
}

func providerHttpClient() *tgsHttp.StandardHTTPClient {
	return &tgsHttp.StandardHTTPClient{
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

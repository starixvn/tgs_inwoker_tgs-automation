package config

type TgsConfig struct {
	CloudflareToken     string              `yaml:"cloudflare_token"`
	ChromeDriverPath    string              `yaml:"chrome_driver_path"`
	NatsUrl             string              `yaml:"nats_url"`
	OpenTelemetryConfig OpenTelemetryConfig `yaml:"opentelemetry"`
	GoogleSheet         GoogleSheetConfig   `yaml:"google_sheet"`
	CdnNetwork          CdnConfig           `yaml:"cdn_network"`
	Namecheap           NamecheapConfig     `yaml:"namecheap"`
	Postgresql          PostgresqlConfig    `yaml:"postgresql"`
	CreateSiteDb        PostgresqlConfig    `yaml:"create_site_db"`
	Telegram            TelegramConfig      `yaml:"telegram"`
	AwsS3               AwsS3Config         `yaml:"aws_s3"`
	MomoTelegram        MomoTelegramConfig  `yaml:"momo_telegram"`
	LetsTalk            LetsTalkConfig      `yaml:"letstalk"`
	Dockerhub           DockerhubConfig     `yaml:"dockerhub"`
	ApiUrl              ApiUrl              `yaml:"api_url"`
}

type ApiUrl struct {
	BrandCert      string `yaml:"brand_cert"`
	TelegramApi    string `yaml:"telegram_api"`
	TelegramApiKey string `yaml:"telegram_apikey"`
}

type OpenTelemetryConfig struct {
	JaegerCollectorUrl string `yaml:"jaeger_collector_url"`
	OtelCollectorUrl   string `yaml:"otel_collector_url"`
}

type DockerhubConfig struct {
	BaseUrl  string `yaml:"base_url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type MomoTelegramConfig struct {
	Token       string `yaml:"token"`
	Movn2ChatId int64  `yaml:"movn2_chat_id"`
	MophChatId  int64  `yaml:"moph_chat_id"`
}

type AwsS3Config struct {
	Region       string `yaml:"region"`
	AccessKey    string `yaml:"access_key"`
	AccessSecret string `yaml:"access_secret"`
	Bucket       string `yaml:"bucket"`
}

type GoogleSheetConfig struct {
	GoogleApiKey string `yaml:"google_api_key"`
	SheetId      string `yaml:"sheet_id"`
}

type CdnConfig struct {
	CdnUserName               string `yaml:"cdn_user_name"`
	CdnApiKey                 string `yaml:"cdn_api_key"`
	CdnEndPoint               string `yaml:"cdn_end_point"`
	CdnLoginUrl               string `yaml:"cdn_login_url"`
	CdnCertificateCreationUrl string `yaml:"cdn_certificate_creation_url"`
	CdnPassword               string `yaml:"cdn_password"`
	DnsContent                string `yaml:"dns_content"`
}

type NamecheapConfig struct {
	NamecheapApiKey      string `yaml:"namecheap_api_key"`
	NamecheapUsername    string `yaml:"namecheap_username"`
	NamecheapPassword    string `yaml:"namecheap_password"`
	NamecheapClientIp    string `yaml:"namecheap_client_ip"`
	NamecheapBaseUrl     string `yaml:"namecheap_baseurl"`
	NamecheapEmail       string `yaml:"namecheap_email"`
	NamecheapAddress     string `yaml:"namecheap_address"`
	NamecheapNameServers string `yaml:"namecheap_nameservers"`
}

type PostgresqlConfig struct {
	PgHost     string `yaml:"pg_host"`
	PgPort     string `yaml:"pg_port"`
	PgDb       string `yaml:"pg_database"`
	PgUsername string `yaml:"pg_username"`
	PgPassword string `yaml:"pg_password"`
}

type TelegramConfig struct {
	TelegramBotToken string `yaml:"telegram_bot_token"`
	TelegramChatId   string `yaml:"telegram_chat_id"`
	TelegramWebhook  string `yaml:"telegram_webhook"`
}

type LetsTalkConfig struct {
	AccountId string `yaml:"account_id"`
	ApiKey    string `yaml:"api_key"`
}

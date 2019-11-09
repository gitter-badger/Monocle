package monocle

type Config struct {
	AppName           string `envconfig:"APP_NAME" required:"true"`
	AppVersion        string `envconfig:"APP_VERSION" required:"true"`
	AppDeveloper      string `envconfig:"APP_DEVELOPER" required:"true"`
	AppDeveloperEmail string `envconfig:"APP_DEVELOPER_EMAIL" required:"true"`
	AppPort           uint   `envconfig:"APP_PORT" required:"true"`
	DBDsn             string `envconfig:"DB_DSN" required:"true"`
	LogLevel          string `envconfig:"LOG_LEVEL" required:"true"`
	EsiHost           string `envconfig:"ESI_HOST" required:"true"`
	ApiUserAgent      string `envconfig:"API_USER_AGENT" required:"true"`
	DiscordToken      string `envconfig:"DISCORD_TOKEN" required:"true"`
}

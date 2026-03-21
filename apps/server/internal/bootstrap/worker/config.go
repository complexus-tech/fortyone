package workerbootstrap

import (
	"maps"
	"time"

	"github.com/josemukorivo/config"
)

var defaultQueues = map[string]int{
	"critical":      6,
	"default":       3,
	"integrations":  2,
	"low":           1,
	"onboarding":    5,
	"cleanup":       2,
	"notifications": 4,
	"automation":    3,
}

type Config struct {
	DB struct {
		Host         string `default:"localhost" env:"APP_DB_HOST"`
		Port         string `default:"5432" env:"APP_DB_PORT"`
		User         string `default:"postgres" env:"APP_DB_USER"`
		Password     string `default:"password" env:"APP_DB_PASSWORD"`
		Name         string `default:"complexus" env:"APP_DB_NAME"`
		MaxIdleConns int    `default:"25" env:"APP_DB_MAX_IDLE_CONNS"`
		MaxOpenConns int    `default:"25" env:"APP_DB_MAX_OPEN_CONNS"`
		DisableTLS   bool   `default:"true" env:"APP_DB_DISABLE_TLS"`
	}
	Redis struct {
		Host               string        `default:"localhost" env:"APP_REDIS_HOST"`
		Port               string        `default:"6379" env:"APP_REDIS_PORT"`
		Password           string        `default:"" env:"APP_REDIS_PASSWORD"`
		Name               int           `default:"0" env:"APP_REDIS_DB"`
		DisableTLS         bool          `default:"false" env:"APP_REDIS_DISABLE_TLS"`
		InsecureSkipVerify bool          `default:"false" env:"APP_REDIS_INSECURE_SKIP_VERIFY"`
		DialTimeout        time.Duration `default:"10s" env:"APP_REDIS_DIAL_TIMEOUT"`
		ReadTimeout        time.Duration `default:"30s" env:"APP_REDIS_READ_TIMEOUT"`
		WriteTimeout       time.Duration `default:"30s" env:"APP_REDIS_WRITE_TIMEOUT"`
		PoolSize           int           `default:"30" env:"APP_REDIS_POOL_SIZE"`
	}
	Email struct {
		Host        string `default:"smtp.gmail.com" env:"APP_EMAIL_HOST"`
		Port        int    `default:"587" env:"APP_EMAIL_PORT"`
		Username    string `env:"APP_EMAIL_USERNAME"`
		Password    string `env:"APP_EMAIL_PASSWORD"`
		FromAddress string `env:"APP_EMAIL_FROM_ADDRESS"`
		FromName    string `default:"FortyOne" env:"APP_EMAIL_FROM_NAME"`
		Environment string `default:"development" env:"APP_EMAIL_ENVIRONMENT"`
		BaseDir     string `default:"." env:"APP_EMAIL_BASE_DIR"`
	}
	Brevo struct {
		APIKey string `env:"APP_BREVO_API_KEY"`
	}
	Auth struct {
		SecretKey string `default:"secret" env:"APP_AUTH_SECRET_KEY"`
	}
	Website struct {
		URL string `default:"http://localhost:3000" env:"APP_WEBSITE_URL"`
	}
	GitHub struct {
		AppID         int64  `env:"APP_GITHUB_APP_ID"`
		AppSlug       string `env:"GITHUB_APP_SLUG"`
		PrivateKey    string `env:"GITHUB_PRIVATE_KEY"`
		RedirectURL   string `env:"GITHUB_REDIRECT_URL"`
		WebhookSecret string `env:"GITHUB_WEBHOOK_SECRET"`
	}
	Queues map[string]int `default:"{\"critical\":6,\"default\":3,\"integrations\":2,\"low\":1,\"onboarding\":5,\"cleanup\":2,\"notifications\":4,\"automation\":3}"`
}

func loadConfig() (Config, error) {
	var cfg Config
	if err := config.Parse("app", &cfg); err != nil {
		return Config{}, err
	}
	if cfg.Queues == nil {
		cfg.Queues = cloneQueueConfig(defaultQueues)
	}
	return cfg, nil
}

func cloneQueueConfig(src map[string]int) map[string]int {
	dst := make(map[string]int, len(src))
	maps.Copy(dst, src)
	return dst
}

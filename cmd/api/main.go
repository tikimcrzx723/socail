package main

import (
	"expvar"
	"runtime"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikimcrzx723/social/internal/auth"
	"github.com/tikimcrzx723/social/internal/db"
	"github.com/tikimcrzx723/social/internal/env"
	"github.com/tikimcrzx723/social/internal/mailer"
	"github.com/tikimcrzx723/social/internal/ratelimiter"
	"github.com/tikimcrzx723/social/internal/store"
	"github.com/tikimcrzx723/social/internal/store/cache"
	"go.uber.org/zap"
)

const version = ""

//	@title			GopherSocial
//	@description	API for GopherSocial, a social network for gophers
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {
	cfg := config{
		addr:       env.GetString("ADDR", ":8080"),
		apiURL:     env.GetString("EXTERNAL_URL", "localhost:8080"),
		frontedURL: env.GetString("FRONTED_URL", "localhost:4000"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		redisCfg: redisConfig{
			addr:    env.GetString("REDIS_ADDR", "localhost:6379"),
			pw:      env.GetString("REDIS_PW", ""),
			db:      env.GetInt("REDIS_DB", 0),
			enabled: env.GetBool("REDIS_ENABLED", true),
		},
		env: env.GetString("ENV", "development"),
		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER", "admin"),
				pass: env.GetString("AUTH_BASIC_PASS", "admin"),
			},
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", "superdupersecretmen"),
				exp:    time.Hour * 48,
				iss:    env.GetString("AUTH_TOKEN_ISS", "gophersocial"),
			},
		},
		rateLimiter: ratelimiter.Config{
			RequestPerTimeFrame: env.GetInt("RATELIMITER_REQUESTS_COUNT", 20),
			TimeFrame:           time.Second * 5,
			Enabled:             env.GetBool("RATE_LIMITER_ENABLED", true),
		},
	}

	smtpHost := env.GetString("SMTP_HOST", "sandbox.smtp.mailtrap.io")
	smtpPort := env.GetInt("SMTP_PORT", 2525)
	smtpUsername := env.GetString("SMTP_USERNAME", "a4fe83d1cfb872")
	smtpPassword := env.GetString("SMTP_PASSWORD", "3ffc3bd12f1e58")
	smtpSender := env.GetString("SMTP_SENDER", "<no-reply@gophersocial>")

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()

	defer logger.Sync()

	// Database
	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("database connection pool stablished")

	// Cache
	var rdb *redis.Client
	if cfg.redisCfg.enabled {
		rdb = cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.pw, cfg.redisCfg.db)
		logger.Info("redis cache connection established")

		defer rdb.Close()
	}
	store := store.NewStorage(db)
	cacheStorage := cache.NewRedisStorage(rdb)

	jwtAuthenticator := auth.NewJWTAuthenticator(
		cfg.auth.token.secret,
		cfg.auth.token.iss,
		cfg.auth.token.iss,
	)

	// Rate limiter
	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiter.RequestPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	app := &application{
		config:        cfg,
		store:         store,
		cacheStorage:  cacheStorage,
		logger:        logger,
		mailer:        mailer.New(smtpHost, smtpPort, smtpUsername, smtpPassword, smtpSender),
		authenticator: jwtAuthenticator,
		rateLimiter:   rateLimiter,
	}

	expvar.NewString("version").Set(version)
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	logger.Fatal(app.run(app.mount()))
}

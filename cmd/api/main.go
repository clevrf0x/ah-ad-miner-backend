package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"

	"com.activehacks.ad-miner-backend/internal/database"
	"com.activehacks.ad-miner-backend/internal/env"
	"com.activehacks.ad-miner-backend/internal/queue"
	"com.activehacks.ad-miner-backend/internal/smtp"
	"com.activehacks.ad-miner-backend/internal/version"

	"github.com/lmittmann/tint"
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))

	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

type config struct {
	baseURL  string
	httpPort int
	db       struct {
		dsn         string
		automigrate bool
	}
	redis struct {
		addr     string
		password string
	}
	jwt struct {
		secretKey string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		from     string
	}
}

type application struct {
	config      config
	db          *database.DB
	logger      *slog.Logger
	mailer      *smtp.Mailer
	queueClient *queue.Client
	wg          sync.WaitGroup
}

func run(logger *slog.Logger) error {
	// Handle loading environment variabled
	envFile := flag.String("env", "", "path to environment file (required)")

	showVersion := flag.Bool("version", false, "display version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	if *envFile == "" {
		return fmt.Errorf("environment file path is required.")
	}
	err := env.LoadEnv(envFile)
	if err != nil {
		return err
	}

	var cfg config

	cfg.baseURL = env.GetString("BASE_URL", "http://localhost:6666")
	cfg.httpPort = env.GetInt("HTTP_PORT", 6666)
	cfg.db.dsn = env.GetString("DB_DSN", "user:pass@localhost:5432/db")
	cfg.db.automigrate = env.GetBool("DB_AUTOMIGRATE", true)
	cfg.redis.addr = env.GetString("REDIS_ADDR", "localhost:6379")
	cfg.redis.password = env.GetString("REDIS_PASSWORD", "SuperSecure@123")
	cfg.jwt.secretKey = env.GetString("JWT_SECRET_KEY", "4jgnm2k4z7cf54gnyyudvqfqsd7cyeh3")
	cfg.smtp.host = env.GetString("SMTP_HOST", "example.smtp.host")
	cfg.smtp.port = env.GetInt("SMTP_PORT", 25)
	cfg.smtp.username = env.GetString("SMTP_USERNAME", "example_username")
	cfg.smtp.password = env.GetString("SMTP_PASSWORD", "pa55word")
	cfg.smtp.from = env.GetString("SMTP_FROM", "Example Name <no_reply@example.org>")

	db, err := database.New(cfg.db.dsn, cfg.db.automigrate)
	if err != nil {
		return err
	}
	defer db.Close()

	queueClient := queue.NewClient(cfg.redis.addr, cfg.redis.password)
	defer queueClient.Close()

	mailer, err := smtp.NewMailer(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.from)
	if err != nil {
		return err
	}

	app := &application{
		config:      cfg,
		db:          db,
		logger:      logger,
		mailer:      mailer,
		queueClient: queueClient,
	}

	// Start worker
	go func() {
		worker := queue.NewWorker(cfg.redis.addr, cfg.redis.password, logger)
		handler := queue.NewTaskHandler(db)

		logger.Info("Starting worker")
		if err := worker.Start(handler, logger); err != nil {
			logger.Error("Worker failed", "Error", err)
			os.Exit(1)
		}
	}()

	return app.serveHTTP()
}

package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"awesome-chances/backend/internal/adapters/demo"
	"awesome-chances/backend/internal/adapters/memory"
	"awesome-chances/backend/internal/adapters/postgres"
	"awesome-chances/backend/internal/config"
	"awesome-chances/backend/internal/handler"
	"awesome-chances/backend/internal/router"
	"awesome-chances/backend/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("startup failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	catalog, err := demo.Load()
	if err != nil {
		return errors.New("加载示例数据失败")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	var auth *service.Auth
	if cfg.Database.Enabled {
		db, err := postgres.Connect(ctx, cfg.Database)
		if err != nil {
			return err
		}
		defer db.Close()
		if cfg.Database.AutoMigrate {
			if err := db.Migrate(ctx); err != nil {
				return err
			}
		}
		if err := db.CheckSchema(ctx); err != nil {
			return err
		}
		auth, err = service.NewAuth(postgres.NewAuthRepository(db), cfg.Auth)
		if err != nil {
			return errors.New("初始化用户认证失败")
		}
		logger.Info("database connected", "orm", "gorm", "dialect", "postgres", "auth", "enabled")
	} else {
		logger.Info("database disabled", "auth", "unavailable")
	}
	addr := config.Address(cfg)
	app := service.New(catalog, memory.NewFeedbackStore())
	server := &http.Server{Addr: addr, Handler: router.New(handler.New(app).WithAuth(auth, cfg.Auth), logger), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		<-ctx.Done()
		stop, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()
		if err := server.Shutdown(stop); err != nil {
			logger.Error("shutdown", "error", err)
		}
	}()
	logger.Info("demo backend started", "address", addr, "mode", "demo")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.New("启动 HTTP 服务失败，请检查后端地址与端口占用")
	}
	return nil
}

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	_ "github.com/NanoBoom/asethub/docs"
	"github.com/NanoBoom/asethub/internal/cache"
	"github.com/NanoBoom/asethub/internal/config"
	"github.com/NanoBoom/asethub/internal/database"
	"github.com/NanoBoom/asethub/internal/handlers"
	"github.com/NanoBoom/asethub/internal/logger"
	"github.com/NanoBoom/asethub/internal/middleware"
	"github.com/NanoBoom/asethub/internal/repositories"
	"github.com/NanoBoom/asethub/internal/services"
	"github.com/NanoBoom/asethub/pkg/storage"
)

// @title           AssetHub API
// @version         1.0
// @description     资产管理系统 API - 支持图片、视频等资产的上传、管理和检索
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@assethub.example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8003
// @BasePath  /

// @schemes   http https

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load("./configs")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	zapLogger, err := logger.New(cfg.App.Env)
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	defer zapLogger.Sync()

	db, err := database.New(&cfg.Database)
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
	}

	redisClient, err := cache.New(&cfg.Redis)
	if err != nil {
		zapLogger.Fatal("Failed to connect to redis", zap.Error(err))
	}
	defer redisClient.Close()

	router := setupRouter(cfg, zapLogger, db, redisClient)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: router,
	}

	go func() {
		zapLogger.Info("Server starting", zap.Int("port", cfg.App.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Fatal("Server forced shutdown", zap.Error(err))
	}

	zapLogger.Info("Server exited")
}

func setupRouter(cfg *config.Config, zapLogger *zap.Logger, db *gorm.DB, redisClient *cache.RedisClient) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(middleware.Recovery(zapLogger))
	router.Use(middleware.Logger(zapLogger))
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CORS())

	// Health check
	healthHandler := handlers.NewHealthHandler(db, redisClient)
	router.GET("/health", healthHandler.Check)

	// File API - 初始化 Storage 和 FileService
	fileRepo := repositories.NewFileRepository(db)

	// 根据配置创建 Storage 实现（工厂函数在 storage 包中）
	storageBackend, err := storage.NewStorage(context.Background(), &cfg.Storage)
	if err != nil {
		zapLogger.Fatal("Failed to initialize storage", zap.Error(err))
	}

	fileService := services.NewFileService(fileRepo, storageBackend, db)
	fileHandler := handlers.NewFileHandler(fileService)

	api := router.Group("/api/v1")
	{
		files := api.Group("/files")
		{
			// 小文件上传
			files.POST("/upload", fileHandler.UploadDirect)
			files.POST("/upload/presigned", fileHandler.InitPresignedUpload)
			files.POST("/upload/confirm", fileHandler.ConfirmUpload)

			// 大文件分片上传
			files.POST("/upload/multipart/init", fileHandler.InitMultipartUpload)
			files.POST("/upload/multipart/part-url", fileHandler.GeneratePartURL)
			files.POST("/upload/multipart/complete", fileHandler.CompleteMultipartUpload)

			// 通用操作
			files.GET("/:id/download-url", fileHandler.GetDownloadURL)
			files.GET("/:id", fileHandler.GetFile)
			files.DELETE("/:id", fileHandler.DeleteFile)
		}
	}

	// Swagger 文档路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}

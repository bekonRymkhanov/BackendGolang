package main

import (
	"auth-service/internal/config"
	"auth-service/internal/domain"
	"auth-service/internal/handler"
	"auth-service/internal/middleware"
	"auth-service/internal/repository/postgres"
	"auth-service/internal/service"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// Здесь переименовали что бы избежать конфликта с уже импортированным package-ом
	pgdriver "gorm.io/driver/postgres"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting auth service in %s mode", cfg.Environment)

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// логи ОРМ
	gormLogLevel := logger.Info
	if cfg.Environment == "production" {
		gormLogLevel = logger.Error
	}

	// Подключение к БДшке
	db, err := gorm.Open(pgdriver.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connected successfully")

	// Конфигурация пула коннекшинов к БД
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to configure database connection pool: %v", err)
	}
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetime) * time.Second)

	// Миграция
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration applied successfully")

	// Репозиторий
	userRepo := postgres.NewUserRepository(db)

	// Сервисы
	jwtService := service.NewJWTService(
		cfg.JWTSecret,
		cfg.JWTRefreshSecret,
		time.Duration(cfg.JWTExpiryMinutes)*time.Minute,
		time.Duration(cfg.JWTRefreshExpiryDays)*24*time.Hour,
	)

	authService := service.NewAuthService(
		userRepo,
		jwtService,
		time.Duration(cfg.JWTExpiryMinutes)*time.Minute,
	)

	// Хендлеры
	authHandler := handler.NewAuthHandler(authService)

	// Мидлвейры
	authMiddleware := middleware.NewAuthMiddleware(jwtService, userRepo)

	// Router
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Конфигурация CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health Check - возможно пригодиться в будущем, посмотрим
	router.GET("/health", func(c *gin.Context) {
		// Check database connection
		if sqlDB, err := db.DB(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "down",
				"error":  "database connection lost",
			})
			return
		} else if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "down",
				"error":  "database ping failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "up",
			"service": "auth-service",
			"version": cfg.Version,
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		v1.POST("/register", authHandler.Register)
		v1.POST("/login", authHandler.Login)
		v1.POST("/refresh", authHandler.RefreshToken)

		// Protected routes
		protected := v1.Group("/")
		protected.Use(authMiddleware.AuthRequired())
		{
			protected.GET("/me", authHandler.Me)
			protected.PATCH("/users", authHandler.UpdateUser)
			protected.POST("/logout", authHandler.Logout)
		}

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(authMiddleware.AuthRequired(), authMiddleware.RoleRequired("admin"))
		{
			admin.GET("/users", authHandler.ListUsers)
			admin.GET("/users/:id", authHandler.GetUser)
			admin.DELETE("/users/:id", authHandler.DeleteUser)
			admin.PATCH("/users/:id/activate", authHandler.ActivateUser)
			admin.PATCH("/users/:id/deactivate", authHandler.DeactivateUser)
			admin.PATCH("/users/:id/role", authHandler.ChangeUserRole)
		}
	}

	// Создаем сервер
	srv := &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Старт сервера в горотине
	go func() {
		log.Printf("Starting auth service on %s", cfg.ServerAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Создание дефолтного админа
	if cfg.CreateDefaultAdmin {
		createDefaultAdmin(authService)
	}

	// Gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

// createDefaultAdmin creates a default admin user if it doesn't exist
func createDefaultAdmin(authService service.AuthService) {
	ctx := context.Background()
	input := domain.UserRegisterInput{
		Username:  "admin",
		Email:     "admin@example.com",
		Password:  "Admin@123",
		FirstName: "Admin",
		LastName:  "User",
	}

	user, err := authService.Register(ctx, input)
	if err != nil {
		if err.Error() == "username already exists" || err.Error() == "email already exists" {
			log.Println("Default admin user already exists")
			return
		}
		log.Printf("Failed to create default admin user: %v", err)
		return
	}

	if err := authService.ChangeUserRole(ctx, user.ID, "admin"); err != nil {
		log.Printf("Failed to set admin role for default user: %v", err)
		return
	}

	log.Println("Default admin user created successfully")
}

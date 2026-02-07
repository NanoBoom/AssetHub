# Golang Web Project Stack Guide
**Research Date:** 2026-01-13
**Stack:** Gin + GORM + Viper + Zerolog/Zap

## Table of Contents
1. [Framework Versions & Installation](#framework-versions--installation)
2. [Project Structure](#project-structure)
3. [Gin Web Framework](#gin-web-framework)
4. [GORM ORM v2](#gorm-orm-v2)
5. [Configuration Management (Viper)](#configuration-management-viper)
6. [Logging (Zerolog vs Zap)](#logging-zerolog-vs-zap)
7. [Error Handling Patterns](#error-handling-patterns)
8. [Middleware Patterns](#middleware-patterns)
9. [Integration Examples](#integration-examples)

---

## Framework Versions & Installation

### Gin Web Framework
- **Latest Version:** v1.11.0 (Released: Sept 20, 2025)
- **Go Version Required:** 1.23+
- **Installation:**
```bash
go get -u github.com/gin-gonic/gin
```
- **Import Path:** `github.com/gin-gonic/gin`
- **Documentation:** https://gin-gonic.com/en/docs/
- **Go Package Docs:** https://pkg.go.dev/github.com/gin-gonic/gin

### GORM ORM
- **Latest Version:** v1.31.1 (Released: Nov 2, 2025)
- **Import Path:** `gorm.io/gorm`
- **Installation:**
```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/sqlite  # or mysql, postgres, etc.
```
- **Documentation:** https://gorm.io/docs/
- **Go Package Docs:** https://pkg.go.dev/gorm.io/gorm

### Viper Configuration
- **Latest Version:** v1.21.0 (Released: Sept 8, 2025)
- **Import Path:** `github.com/spf13/viper`
- **Installation:**
```bash
go get github.com/spf13/viper
```
- **Documentation:** https://github.com/spf13/viper

### Logging Libraries

#### Zerolog
- **Import Path:** `github.com/rs/zerolog`
- **Performance:** 27 ns/op, 0 allocs/op
- **Best For:** Minimal overhead, JSON-only output

#### Zap
- **Import Path:** `go.uber.org/zap`
- **Performance:** 71 ns/op, 0 allocs/op (standard), 87 ns/op, 1 alloc/op (sugared)
- **Best For:** Customization, multiple output formats
- **Documentation:** https://pkg.go.dev/go.uber.org/zap

---

## Project Structure

Based on [golang-standards/project-layout](https://github.com/golang-standards/project-layout) for web applications:

```
myproject/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── handlers/                # HTTP handlers (Gin controllers)
│   ├── models/                  # GORM models
│   ├── services/                # Business logic
│   ├── repositories/            # Data access layer
│   ├── middleware/              # Custom middleware
│   └── config/                  # Configuration structs
├── pkg/
│   └── utils/                   # Reusable utility functions
├── configs/
│   ├── config.yaml             # Default configuration
│   ├── config.dev.yaml         # Development config
│   └── config.prod.yaml        # Production config
├── web/
│   ├── static/                 # Static assets
│   └── templates/              # HTML templates
├── migrations/                  # Database migrations
├── scripts/                     # Build and deployment scripts
├── docs/                        # API documentation
├── .env.example                # Environment variable template
├── go.mod
├── go.sum
└── README.md
```

### Directory Guidelines

**`/cmd`** - Main applications. Keep minimal; delegate to `/internal`.

**`/internal`** - Private application code (Go enforces this). External packages cannot import from here.

**`/pkg`** - Library code safe for external use.

**`/configs`** - Configuration file templates.

**`/web`** - Web application assets, templates, SPAs.

**Key Principles:**
- Start simple (single `main.go` for small projects)
- Only add structure as project grows
- Use `/internal` to prevent external dependencies
- The `/pkg` directory is optional and debated in the community

---

## Gin Web Framework

### Basic Setup

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func main() {
    // Create router with default middleware (Logger, Recovery)
    router := gin.Default()

    // Basic route
    router.GET("/ping", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "pong",
        })
    })

    // Run on default port :8080
    router.Run()

    // Or specify port
    // router.Run(":3000")
}
```

### Recommended Patterns

#### 1. Use HTTP Constants
```go
import "net/http"

c.JSON(http.StatusOK, data)        // Instead of c.JSON(200, data)
c.JSON(http.StatusBadRequest, err) // Instead of c.JSON(400, err)
```

#### 2. Route Grouping
```go
v1 := router.Group("/api/v1")
{
    v1.GET("/users", getUsers)
    v1.POST("/users", createUser)
    v1.GET("/users/:id", getUser)
}

v2 := router.Group("/api/v2")
{
    v2.GET("/users", getUsersV2)
}
```

#### 3. Structured Handlers
```go
type UserHandler struct {
    service UserService
}

func NewUserHandler(service UserService) *UserHandler {
    return &UserHandler{service: service}
}

func (h *UserHandler) GetUser(c *gin.Context) {
    id := c.Param("id")
    user, err := h.service.GetUserByID(id)
    if err != nil {
        c.Error(err) // Attach error to context
        return
    }
    c.JSON(http.StatusOK, user)
}
```

### Key Features in v1.11.0

- **HTTP/3 Support:** Via quic-go integration
- **Custom JSON Codecs:** Runtime codec configuration
- **BindPlain:** New binding method for plain text
- **Enhanced Form Binding:** Default values for collections
- **AbortWithStatusPureJSON():** New response method

---

## GORM ORM v2

### Basic Setup

```go
import (
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
)

func InitDB() (*gorm.DB, error) {
    dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    return db, nil
}
```

### Model Definition

```go
import "gorm.io/gorm"

type User struct {
    gorm.Model              // Includes ID, CreatedAt, UpdatedAt, DeletedAt
    Name   string    `gorm:"size:100;not null"`
    Email  string    `gorm:"uniqueIndex;not null"`
    Age    int       `gorm:"default:18"`
    Active bool      `gorm:"default:true"`
}

// Custom table name
func (User) TableName() string {
    return "users"
}
```

### Context Support (New in v2)

```go
// Always use context for timeouts and cancellation
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

var users []User
result := db.WithContext(ctx).Find(&users)
```

### CRUD Operations

```go
// Create
user := User{Name: "John", Email: "john@example.com"}
result := db.Create(&user)
// user.ID is now populated

// Read
var user User
db.First(&user, 1)                 // Find by primary key
db.First(&user, "email = ?", "john@example.com")

// Update
db.Model(&user).Update("Name", "Jane")
db.Model(&user).Updates(User{Name: "Jane", Age: 30})
db.Model(&user).Updates(map[string]interface{}{"Name": "Jane", "Age": 30})

// Delete (soft delete with gorm.Model)
db.Delete(&user, 1)

// Permanent delete
db.Unscoped().Delete(&user, 1)
```

### GORM v2 Best Practices

#### 1. Always Use Context
```go
func (r *UserRepository) FindAll(ctx context.Context) ([]User, error) {
    var users []User
    err := r.db.WithContext(ctx).Find(&users).Error
    return users, err
}
```

#### 2. Use Select for Performance
```go
// Only select needed fields
db.Select("name", "email").Find(&users)
```

#### 3. Batch Processing
```go
// Process large datasets in batches
result := db.FindInBatches(&results, 100, func(tx *gorm.DB, batch int) error {
    for _, result := range results {
        // process each batch
    }
    return nil
})
```

#### 4. Error Handling
```go
err := db.First(&user, id).Error
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, ErrUserNotFound
    }
    return nil, err
}
```

#### 5. Prevent Global Updates (Enabled by Default in v2)
```go
// This will fail by default
db.Model(&User{}).Update("active", false)

// Must use conditions
db.Model(&User{}).Where("active = ?", true).Update("active", false)

// Or enable global updates explicitly
db.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&User{}).Update("active", false)
```

### Critical v2 Breaking Changes

1. **Import Path:** `gorm.io/gorm` (was `github.com/jinzhu/gorm`)
2. **Drivers:** Separate packages (`gorm.io/driver/postgres`, etc.)
3. **Tags:** Use camelCase (`foreignKey`, not `foreign_key`)
4. **Soft Delete:** Must use `gorm.DeletedAt` type explicitly
5. **Table Operations:** Use migrator
   ```go
   db.Migrator().CreateTable(&User{})
   db.Migrator().DropTable(&User{})
   ```
6. **ErrRecordNotFound:** Only from `First`, `Last`, `Take` methods

---

## Configuration Management (Viper)

### Basic Setup with Environment Support

```go
package config

import (
    "strings"
    "github.com/spf13/viper"
)

type Config struct {
    App struct {
        Name string `mapstructure:"name"`
        Port int    `mapstructure:"port"`
        Env  string `mapstructure:"env"`
    } `mapstructure:"app"`
    Database struct {
        Host     string `mapstructure:"host"`
        Port     int    `mapstructure:"port"`
        User     string `mapstructure:"user"`
        Password string `mapstructure:"password"`
        DBName   string `mapstructure:"dbname"`
    } `mapstructure:"database"`
}

func LoadConfig(path string) (*Config, error) {
    // Set defaults
    viper.SetDefault("app.port", 8080)
    viper.SetDefault("app.env", "development")

    // Config file settings
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(path)
    viper.AddConfigPath("./configs")
    viper.AddConfigPath(".")

    // Environment variable support
    viper.AutomaticEnv()
    viper.SetEnvPrefix("APP")
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    // Read config file
    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    // Unmarshal to struct
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    return &config, nil
}
```

### Environment-Based Configuration

**config.yaml** (default/development):
```yaml
app:
  name: "MyApp"
  port: 8080
  env: "development"

database:
  host: "localhost"
  port: 5432
  user: "dev_user"
  password: "dev_pass"
  dbname: "myapp_dev"
```

**config.prod.yaml** (production):
```yaml
app:
  name: "MyApp"
  port: 8080
  env: "production"

database:
  host: "${DB_HOST}"
  port: 5432
  user: "${DB_USER}"
  password: "${DB_PASSWORD}"
  dbname: "myapp_prod"
```

### Environment Variable Override

Environment variables take precedence:
```bash
# These will override config file values
export APP_APP_PORT=3000
export APP_DATABASE_HOST=prod-db.example.com
export APP_DATABASE_PASSWORD=super_secret
```

### Best Practices

1. **Always Set Defaults:** Ensure app can start without config file
2. **Use Struct Binding:** Better type safety than individual `Get()` calls
3. **Validate After Loading:**
   ```go
   if config.App.Port < 1024 || config.App.Port > 65535 {
       return nil, errors.New("invalid port number")
   }
   ```
4. **Never Log Secrets:** Avoid logging passwords/API keys
5. **Use Instance, Not Global:** For testability
   ```go
   v := viper.New()
   v.SetConfigName("config")
   // ... configure v
   ```
6. **Separate Sensitive Data:** Store secrets in environment variables

### Configuration Priority (Highest to Lowest)

1. Explicit `viper.Set()` calls
2. Command-line flags
3. Environment variables
4. Config file
5. Defaults

---

## Logging (Zerolog vs Zap)

### Performance Comparison

| Library  | Speed       | Allocations  | Use Case |
|----------|-------------|--------------|----------|
| Zerolog  | 27 ns/op    | 0 allocs/op  | JSON-only, minimal overhead |
| Zap      | 71 ns/op    | 0 allocs/op  | Customizable, multiple formats |
| Zap Sugar| 87 ns/op    | 1 alloc/op   | Ergonomic API |

### Zerolog Implementation

```go
package logger

import (
    "os"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func InitLogger(env string) {
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

    if env == "development" {
        // Pretty console output for dev
        log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
    } else {
        // JSON output for production
        zerolog.SetGlobalLevel(zerolog.InfoLevel)
    }
}

// Usage
func Example() {
    log.Info().Msg("Server started")
    log.Error().Err(err).Msg("Failed to connect to database")
    log.Debug().
        Str("user_id", "123").
        Int("status", 200).
        Msg("Request processed")
}
```

### Zap Implementation

```go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func InitLogger(env string) (*zap.Logger, error) {
    var config zap.Config

    if env == "production" {
        config = zap.NewProductionConfig()
    } else {
        config = zap.NewDevelopmentConfig()
        config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    }

    logger, err := config.Build()
    if err != nil {
        return nil, err
    }

    return logger, nil
}

// Usage
func Example(logger *zap.Logger) {
    logger.Info("Server started",
        zap.String("host", "localhost"),
        zap.Int("port", 8080),
    )

    // Or use sugared logger for easier syntax
    sugar := logger.Sugar()
    sugar.Infow("Request processed",
        "user_id", "123",
        "status", 200,
    )
}
```

### Recommendation

**Choose Zerolog if:**
- JSON output meets your needs
- You want minimal overhead
- Simple, intuitive API preferred
- Built-in context support needed

**Choose Zap if:**
- Need multiple output formats
- Require extensive customization
- Want strongly-typed fields
- Team values performance tuning options

---

## Error Handling Patterns

### Custom Error Types

```go
package errors

import "fmt"

type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Err     error  `json:"-"`
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Err)
    }
    return e.Message
}

// Error constructors
func NewNotFoundError(message string) *AppError {
    return &AppError{
        Code:    404,
        Message: message,
    }
}

func NewValidationError(message string, err error) *AppError {
    return &AppError{
        Code:    400,
        Message: message,
        Err:     err,
    }
}

func NewInternalError(err error) *AppError {
    return &AppError{
        Code:    500,
        Message: "Internal server error",
        Err:     err,
    }
}
```

### Idiomatic Error Wrapping (Go 1.13+)

```go
import (
    "fmt"
    "errors"
)

// Wrap errors with context
func (s *UserService) GetUser(id string) (*User, error) {
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %s: %w", id, err)
    }
    return user, nil
}

// Check wrapped errors
var ErrUserNotFound = errors.New("user not found")

if errors.Is(err, ErrUserNotFound) {
    // handle not found
}

// Extract specific error types
var appErr *AppError
if errors.As(err, &appErr) {
    // handle app error
}
```

### Error Handling Best Practices

1. **Wrap errors with context:** Use `fmt.Errorf` with `%w`
2. **Define sentinel errors:** For common error conditions
3. **Never leak internal errors:** Convert to domain errors at boundaries
4. **Log at appropriate level:** Error vs Info vs Debug
5. **Return errors, don't panic:** Except for unrecoverable conditions

---

## Middleware Patterns

### Error Handling Middleware

```go
package middleware

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "myapp/internal/errors"
)

func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next() // Execute handlers

        // Check for errors after handler execution
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err

            // Handle different error types
            switch e := err.(type) {
            case *errors.AppError:
                c.JSON(e.Code, gin.H{
                    "error": e.Message,
                    "code":  e.Code,
                })
            default:
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": "Internal server error",
                    "code":  500,
                })
            }

            c.Abort()
        }
    }
}

// Usage
router := gin.New()
router.Use(ErrorHandler())
```

### Recovery Middleware (Custom)

```go
func Recovery(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                logger.Error("Panic recovered",
                    zap.Any("error", err),
                    zap.String("path", c.Request.URL.Path),
                )

                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": "Internal server error",
                })
                c.Abort()
            }
        }()
        c.Next()
    }
}
```

### Request Logging Middleware

```go
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery

        c.Next()

        latency := time.Since(start)

        logger.Info("Request",
            zap.String("method", c.Request.Method),
            zap.String("path", path),
            zap.String("query", query),
            zap.Int("status", c.Writer.Status()),
            zap.Duration("latency", latency),
            zap.String("ip", c.ClientIP()),
        )
    }
}
```

### CORS Middleware

```go
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }

        c.Next()
    }
}
```

### Authentication Middleware

```go
func AuthRequired(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "No authorization header"})
            c.Abort()
            return
        }

        // Validate JWT token (simplified)
        userID, err := validateToken(token, jwtSecret)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        // Set user context
        c.Set("user_id", userID)
        c.Next()
    }
}
```

### Middleware Best Practices

1. **Single Responsibility:** Each middleware does one thing
2. **Order Matters:** Recovery → Logger → Error Handler → Business Logic
3. **Use c.Next():** Call explicitly to control flow
4. **Use c.Abort():** Stop execution when necessary
5. **Set Context Values:** Share data between middleware and handlers using `c.Set()`

---

## Integration Examples

### Complete Application Structure

```go
// cmd/api/main.go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    "myapp/internal/config"
    "myapp/internal/handlers"
    "myapp/internal/middleware"
    "myapp/internal/repositories"
    "myapp/internal/services"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig("./configs")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize logger
    logger, err := initLogger(cfg.App.Env)
    if err != nil {
        log.Fatalf("Failed to init logger: %v", err)
    }
    defer logger.Sync()

    // Initialize database
    db, err := initDatabase(cfg)
    if err != nil {
        logger.Fatal("Failed to connect to database", zap.Error(err))
    }

    // Initialize layers
    userRepo := repositories.NewUserRepository(db)
    userService := services.NewUserService(userRepo)
    userHandler := handlers.NewUserHandler(userService, logger)

    // Setup router
    router := setupRouter(cfg, logger, userHandler)

    // Start server with graceful shutdown
    srv := &http.Server{
        Addr:    ":8080",
        Handler: router,
    }

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("Server failed", zap.Error(err))
        }
    }()

    logger.Info("Server started", zap.Int("port", cfg.App.Port))

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Info("Shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        logger.Fatal("Server forced shutdown", zap.Error(err))
    }

    logger.Info("Server exited")
}

func setupRouter(cfg *config.Config, logger *zap.Logger, userHandler *handlers.UserHandler) *gin.Engine {
    if cfg.App.Env == "production" {
        gin.SetMode(gin.ReleaseMode)
    }

    router := gin.New()

    // Global middleware
    router.Use(middleware.Recovery(logger))
    router.Use(middleware.RequestLogger(logger))
    router.Use(middleware.ErrorHandler())
    router.Use(middleware.CORS())

    // Health check
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    // API routes
    v1 := router.Group("/api/v1")
    {
        users := v1.Group("/users")
        {
            users.GET("", userHandler.List)
            users.GET("/:id", userHandler.Get)
            users.POST("", userHandler.Create)
            users.PUT("/:id", userHandler.Update)
            users.DELETE("/:id", userHandler.Delete)
        }
    }

    return router
}
```

### Handler Layer

```go
// internal/handlers/user_handler.go
package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    "myapp/internal/models"
    "myapp/internal/services"
)

type UserHandler struct {
    service services.UserService
    logger  *zap.Logger
}

func NewUserHandler(service services.UserService, logger *zap.Logger) *UserHandler {
    return &UserHandler{
        service: service,
        logger:  logger,
    }
}

func (h *UserHandler) Get(c *gin.Context) {
    id := c.Param("id")

    user, err := h.service.GetByID(c.Request.Context(), id)
    if err != nil {
        c.Error(err)
        return
    }

    c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Create(c *gin.Context) {
    var req models.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.Error(NewValidationError("Invalid request body", err))
        return
    }

    user, err := h.service.Create(c.Request.Context(), &req)
    if err != nil {
        c.Error(err)
        return
    }

    c.JSON(http.StatusCreated, user)
}
```

### Service Layer

```go
// internal/services/user_service.go
package services

import (
    "context"
    "fmt"

    "myapp/internal/models"
    "myapp/internal/repositories"
)

type UserService interface {
    GetByID(ctx context.Context, id string) (*models.User, error)
    Create(ctx context.Context, req *models.CreateUserRequest) (*models.User, error)
}

type userService struct {
    repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserService {
    return &userService{repo: repo}
}

func (s *userService) GetByID(ctx context.Context, id string) (*models.User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return user, nil
}

func (s *userService) Create(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
    user := &models.User{
        Name:  req.Name,
        Email: req.Email,
    }

    if err := s.repo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    return user, nil
}
```

### Repository Layer

```go
// internal/repositories/user_repository.go
package repositories

import (
    "context"
    "errors"

    "gorm.io/gorm"

    "myapp/internal/models"
)

type UserRepository interface {
    FindByID(ctx context.Context, id string) (*models.User, error)
    Create(ctx context.Context, user *models.User) error
}

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
    var user models.User
    err := r.db.WithContext(ctx).First(&user, id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}
```

---

## Sources & References

### Official Documentation
- [Gin Web Framework](https://gin-gonic.com/en/docs/)
- [Gin Quickstart](https://gin-gonic.com/en/docs/quickstart/)
- [Gin GitHub](https://github.com/gin-gonic/gin)
- [GORM Documentation](https://gorm.io/docs/)
- [GORM v2 Release Notes](https://gorm.io/docs/v2_release_note.html)
- [Viper GitHub](https://github.com/spf13/viper)
- [Go Official: Organizing Modules](https://go.dev/doc/modules/layout)
- [Go Official: Error Handling](https://go.dev/blog/go1.13-errors)

### Project Structure
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- [Project Layout & Structure - Awesome Go](https://mehdihadeli.github.io/awesome-go-education/project-layout-structure/)

### Logging
- [Better Stack: Go Logging Libraries Comparison](https://betterstack.com/community/guides/logging/best-golang-logging-libraries/)
- [Dash0: Best Go Logging Tools 2025](https://www.dash0.com/faq/best-go-logging-tools-in-2025-a-comprehensive-guide)
- [Go Logging Benchmarks](https://betterstack-community.github.io/go-logging-benchmarks/)

### Error Handling & Middleware
- [Gin Error Handling Best Practices](https://tillitsdone.com/blogs/gin-error-handling-best-practices/)
- [Gin Official: Error Handling Middleware](https://gin-gonic.com/en/docs/examples/error-handling-middleware/)
- [Datadog: Go Error Handling Guide](https://www.datadoghq.com/blog/go-error-handling/)

### Configuration
- [Viper Configuration Guide](https://dev.to/kittipat1413/a-guide-to-configuration-management-in-go-with-viper-5271)
- [How to Manage Configuration in Go with Viper (2026)](https://oneuptime.com/blog/post/2026-01-07-go-viper-configuration/view)

### Integration Examples
- [Gin + GORM Integration](https://www.compilenrun.com/docs/framework/gin/gin-database-integration/gin-gorm-integration/)
- [GitHub: gingorm1 Example](https://github.com/kokizzu/gingorm1)
- [GitHub: gin-rest-api-example](https://github.com/zacscoding/gin-rest-api-example)

---

**Last Updated:** 2026-01-13
**Next Review:** When upgrading major versions of dependencies

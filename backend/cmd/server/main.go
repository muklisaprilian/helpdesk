package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/simplejourney/helpdesk/internal/config"
	"github.com/simplejourney/helpdesk/internal/database"
	"github.com/simplejourney/helpdesk/internal/handlers"
	"github.com/simplejourney/helpdesk/internal/middleware"
	"github.com/simplejourney/helpdesk/internal/models"
	"github.com/simplejourney/helpdesk/internal/repository"
	"github.com/simplejourney/helpdesk/internal/services"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = findMigrationsDir()
	}
	if err := database.Migrate(db, migrationsDir); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := database.Seed(db); err != nil {
		log.Fatalf("seed: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	authSvc := services.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryH)
	ticketSvc := services.NewTicketService(ticketRepo, commentRepo, userRepo)
	userSvc := services.NewUserService(userRepo)
	h := handlers.New(authSvc, ticketSvc, userSvc)

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(corsMiddleware(cfg.CORSOrigin))

	r.GET("/health", h.Health)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", h.Register)
			auth.POST("/login", h.Login)
		}

		secured := api.Group("")
		secured.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			secured.GET("/me", h.Me)
			secured.GET("/tickets", h.ListTickets)
			secured.POST("/tickets", h.CreateTicket)
			secured.GET("/tickets/:id", h.GetTicket)
			secured.PATCH("/tickets/:id", h.UpdateTicket)
			secured.POST("/tickets/:id/comments", h.AddComment)

			admin := secured.Group("")
			admin.Use(middleware.RequireRoles(models.RoleAdmin))
			{
				admin.GET("/users", h.ListUsers)
				admin.PATCH("/users/:id/role", h.UpdateUserRole)
			}
		}
	}

	addr := ":" + cfg.Port
	log.Printf("helpdesk API listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(origin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET,POST,PATCH,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func findMigrationsDir() string {
	candidates := []string{
		"migrations",
		filepath.Join("..", "migrations"),
		filepath.Join("..", "..", "migrations"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return "migrations"
}

package main

import (
	"BlackPawnChess-server/internal/auth"
	"BlackPawnChess-server/internal/game"
	gameservice "BlackPawnChess-server/internal/gameService"
	"BlackPawnChess-server/internal/matchmaking"
	"BlackPawnChess-server/internal/models"
	"BlackPawnChess-server/internal/router"
	"BlackPawnChess-server/internal/sessions"
	"BlackPawnChess-server/internal/ws"
	"BlackPawnChess-server/pkg/helpers"
	"strconv"
	"strings"

	"BlackPawnChess-server/internal/storage/pdb"
	"BlackPawnChess-server/internal/storage/rdb"

	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	dsn := helpers.GetEnv("DB_DSN", "host=localhost user=chess_user password=secret dbname=chess port=5432 sslmode=disable")
	log.Println("--- TRYING TO CONNECT WITH DSN:", dsn)
	db, err := pdb.NewPostgres(dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&gameservice.Game{}); err != nil {
		log.Fatal(err)
	}

	userRepo := pdb.NewRepository(db)

	log.Println("Database connected and migrated successfully:", db != nil)

	redisAddr := helpers.GetEnv("REDIS_ADDR", "localhost:6379")
	redisPass := helpers.GetEnv("REDIS_PASSWORD", "")
	serverPort := helpers.GetEnv("PORT", "8080")

	sessionHours, _ := strconv.Atoi(helpers.GetEnv("SESSION_DURATION_HOURS", "168"))
	sessionDuration := time.Duration(sessionHours) * time.Hour

	redisStorage, err := rdb.NewRedis(redisAddr, redisPass, 0)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Redis connected succesfully:", redisStorage != nil)

	sessionManager := sessions.NewManager(redisStorage, sessionDuration)
	authHandler := auth.NewHandler(userRepo, sessionManager)

	hub := ws.NewHub()
	gameHub := ws.NewGameHub()
	spectratorHub := ws.NewSpectratorHub()
	matchmaker := matchmaking.NewRedisMatchmaker(redisStorage)
	gameServiceRepo := gameservice.NewRepository(db, redisStorage)
	gameService := gameservice.NewService(gameServiceRepo, userRepo)
	wsHandler := ws.NewHandler(hub, gameHub, spectratorHub, matchmaker, gameService)

	gameHandler := game.NewHandler(gameService)

	gin.SetMode(helpers.GetEnv("GIN_MODE", "debug"))
	r := gin.Default()

	allowedOrigins := strings.Split(helpers.GetEnv("ALLOWED_ORIGINS", "http://localhost:5173"), ",")

	r.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowOriginFunc: func(origin string) bool {
			// Разрешаем любой поддомен trycloudflare.com
			return strings.HasSuffix(origin, ".trycloudflare.com")
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Register(r, authHandler, wsHandler, gameHandler, sessionManager)

	srv := &http.Server{
		Addr:    ":" + serverPort,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	log.Printf("Server started on :%s", serverPort)

	// -----------------------------
	// 5. Ожидание сигнала завершения
	// -----------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		println("Server forced to shutdown:", err.Error())
	}
	//TODO: make db.close()
	println("Server exited gracefully")

}

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

	"BlackPawnChess-server/internal/storage/pdb"
	"BlackPawnChess-server/internal/storage/rdb"

	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	dsn := "host=localhost user=chess_user password=secret dbname=chess port=5432 sslmode=disable"
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

	redisStorage, err := rdb.NewRedis("localhost:6379", "", 0)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Redis connected succesfully:", redisStorage != nil)

	sessionManager := sessions.NewManager(redisStorage, 7*24*time.Hour)
	authHandler := auth.NewHandler(userRepo, sessionManager)

	hub := ws.NewHub()
	gameHub := ws.NewGameHub()
	matchmaker := matchmaking.NewRedisMatchmaker(redisStorage)
	gameServiceRepo := gameservice.NewRepository(db, redisStorage)
	gameService := gameservice.NewService(gameServiceRepo, userRepo)
	wsHandler := ws.NewHandler(hub, gameHub, matchmaker, gameService)

	gameHandler := game.NewHandler(gameService)

	r := gin.Default()
	router.Register(r, authHandler, wsHandler, gameHandler, sessionManager)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	println("Server started on :8080")

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

package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Mahider-T/autoSphere/internal/database"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port   int
	models database.Models
	db     database.Service
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	db, dbConn := database.New()
	NewServer := &Server{
		port:   port,
		db:     db,
		models: database.NewModels(dbConn),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

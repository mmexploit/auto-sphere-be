package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Mahider-T/autoSphere/internal/database"
	"github.com/Mahider-T/autoSphere/internal/jsonlog"
	"github.com/Mahider-T/autoSphere/internal/mailer"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port   int
	models database.Models
	db     database.Service
	logger *jsonlog.Logger
	mailer mailer.Mailer
}

func NewServer() *http.Server {

	// logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	smtp_host := os.Getenv("SMTP_HOST")
	smtp_port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	smtp_username := os.Getenv("SMTP_USERNAME")
	smtp_password := os.Getenv("SMTP_PASSWORD")
	smtp_sender := os.Getenv("SMTP_SENDER")
	db, dbConn := database.New()
	NewServer := &Server{
		port:   port,
		db:     db,
		models: database.NewModels(dbConn),
		logger: logger,
		mailer: mailer.New(smtp_host, smtp_port, smtp_username, smtp_password, smtp_sender),
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

package main

import (
	"github.com/alexedwards/scs/v2"
	_ "modernc.org/sqlite"
	"forum/internal/models"
	"html/template"
	"database/sql"
	"crypto/tls"
	"log/slog"
	"net/http"
	"time"
	"flag"
	"fmt"
	"os"
)

// application is the main application struct
type application struct {
	logger 						*slog.Logger
	threads 					*models.ThreadsModel
	messages 					*models.MessagesModel
	users 						*models.UserModel
	templateCache map[string]	*template.Template
	sessionManager 				*scs.SessionManager
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB()
	if err != nil { logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil { logger.Error(err.Error())
		os.Exit(1)
	}

	ThreadsModel, UserModel, MessagesModel, err := models.NewModels(db)
	if err != nil { logger.Error(err.Error())
		os.Exit(1)
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	app := &application{
		logger: 			logger,
		threads: 			ThreadsModel,
		users: 				UserModel,
		messages:			MessagesModel,
		templateCache: 		templateCache,
		sessionManager: 	sessionManager,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	server := &http.Server{
		Addr: *addr,
		MaxHeaderBytes: 524288,
		Handler: app.routes(),
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig: tlsConfig,

		IdleTimeout: time.Minute,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	
	logger.Info("Starting server", slog.Any("addr", server.Addr))
	err = server.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	logger.Error(err.Error())
	os.Exit(1)
}

// openDB opens a database connection
func openDB() (*sql.DB, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" { return nil, fmt.Errorf("DATABASE_URL is not set") }

	db, err := sql.Open("postgres", connStr)
	if err != nil { return nil, err }

	err = db.Ping()
	if err != nil { db.Close(); return nil, err }
	return db, nil
}
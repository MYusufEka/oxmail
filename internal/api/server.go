package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	authmw "github.com/MYusufEka/oxmail/internal/api/middleware"
	"github.com/MYusufEka/oxmail/internal/config"
	"github.com/MYusufEka/oxmail/internal/database"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/MYusufEka/oxmail/internal/health"
	"github.com/MYusufEka/oxmail/internal/logs"
	"github.com/MYusufEka/oxmail/internal/mail"
)

// Server holds the HTTP server and router.
type Server struct {
	router *chi.Mux
	logger *slog.Logger
}

// NewServer creates a new Server with configured routes and middleware.
func NewServer(conn ...*sql.DB) *Server {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))

	// Security headers + CORS
	allowOrigin := os.Getenv("OXMAIL_WEB_URL")
	if allowOrigin == "" || IsDevMode() {
		allowOrigin = "*"
	}
	router.Use(authmw.SecurityHeaders(allowOrigin))

	// JWT secret
	jwtSecret := os.Getenv("OXMAIL_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = generateRandomSecret()
		logger.Warn("OXMAIL_JWT_SECRET not set, generated random secret (tokens won't survive restart)")
	}

	// Admin password
	adminPassword := os.Getenv("OXMAIL_ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "changeme123"
	}

	srv := &Server{
		router: router,
		logger: logger,
	}

	var db *sql.DB
	if len(conn) > 0 {
		db = conn[0]
	}
	srv.registerRoutes(db, jwtSecret, adminPassword)

	return srv
}

// Router returns the chi router for testing.
func (s *Server) Router() *chi.Mux {
	return s.router
}

// registerRoutes sets up all API routes.
func (s *Server) registerRoutes(conn *sql.DB, jwtSecret, adminPassword string) {
	// Render config templates on startup
	mailDomain := os.Getenv("OXMAIL_DOMAIN")
	if mailDomain == "" {
		mailDomain = "local.test"
	}
	hostname := fmt.Sprintf("mail.%s", mailDomain)
	if err := config.RenderAll("/etc/oxmail", config.RenderPayload{
		Hostname:         hostname,
		Domain:           mailDomain,
		MessageSizeLimit: "10240000",
		DevMode:          IsDevMode(),
	}); err != nil {
		s.logger.Error("failed to render config templates", "error", err)
	} else {
		s.logger.Info("config templates rendered", "output", "/etc/oxmail")
	}

	// Public routes (no auth required)
	checkers := map[string]health.ServiceChecker{
		"postfix": &health.PostfixChecker{Address: "postfix:25", Timeout: 3 * time.Second},
		"dovecot": &health.DovecotChecker{Address: "dovecot:143", Timeout: 3 * time.Second},
		"rspamd":  &health.RspamdChecker{URL: "http://rspamd:11333/ping", Timeout: 3 * time.Second},
		"redis":   &health.RedisChecker{Address: "redis:6379", Timeout: 3 * time.Second},
	}
	healthSvc := health.NewService(checkers, "0.1.0")
	s.router.Get("/health", newHealthHandler(healthSvc))
	s.router.Get("/api/health", newHealthHandler(healthSvc)) // frontend calls this

	// Logs handler (public REST + WebSocket)
	logBuffer := logs.NewRingBuffer(1000)
	logParser := logs.NewParser()
	logCollector := logs.NewCollector(nil, logBuffer, logParser)
	logsHandler := NewLogsHandler(logBuffer, logCollector)
	s.router.Get("/api/logs", logsHandler.HandleGetLogs)
	s.router.Get("/api/logs/stream", logsHandler.HandleStream)

	// Auth endpoint (public)
	var userSvc *domain.UserService
	var dovecotMgr *mail.DovecotManager
	var postfixMgr *mail.PostfixManager
	var domainSvc *domain.DomainService

	if conn != nil {
		db := &database.DB{Conn: conn}
		domainSvc = domain.NewDomainService(db)
		userSvc = domain.NewUserService(db, domainSvc)
		postfixMgr = mail.NewPostfixManager(mail.PostfixConfig{
			DomainsPath: "/etc/oxmail/postfix/virtual_domains",
			AliasesPath: "/etc/oxmail/postfix/virtual_aliases",
		}, &mail.DockerExecExecutor{ContainerName: "oxmail-postfix"})
		dovecotMgr = mail.NewDovecotManager(
			"/etc/oxmail/dovecot",
			"/var/mail/vhosts",
			&mail.DockerExecExecutor{ContainerName: "oxmail-dovecot"},
		)
	}

	authHandler := NewAuthHandler(jwtSecret, adminPassword, userSvc, dovecotMgr)
	authHandler.RegisterRoutes(s.router)

	// Protected routes group
	s.router.Group(func(r chi.Router) {
		if !IsDevMode() {
			r.Use(authmw.JWTAuth(jwtSecret))
		}

		if conn != nil {
			aliasSvc := domain.NewAliasService(conn)
			aliasHandler := NewAliasHandler(aliasSvc, postfixMgr)
			aliasHandler.RegisterRoutes(r)

			domainsHandler := NewDomainsHandler(domainSvc, "/etc/oxmail/postfix/virtual_domains", postfixMgr)
			domainsHandler.RegisterRoutes(r)

			usersHandler := NewUsersHandler(userSvc, dovecotMgr).WithDomainResolver(domainSvc)
			usersHandler.RegisterRoutes(r)
			RegisterDomainScopedRoutes(r, usersHandler, aliasHandler)

			contactSvc := domain.NewContactService(conn)
			contactsHandler := NewContactsHandler(contactSvc)
			contactsHandler.RegisterRoutes(r)
		}

		// Mail handler (IMAP bridge) — works without DB
		imapBridge := mail.NewDovecotBridge("dovecot:143")
		mailHandler := NewMailHandler(imapBridge)
		mailHandler.RegisterRoutes(r)

		// Sieve handler — manage Dovecot sieve scripts
		sieveMgr := mail.NewSieveManager(
			"/var/lib/sieve/scripts",
			"/var/lib/sieve/global",
			&mail.DockerExecExecutor{ContainerName: "oxmail-dovecot"},
		)
		if err := sieveMgr.SetGlobalScript("spam-global", mail.SpamGlobalSieveScript); err != nil {
			s.logger.Error("failed to deploy global spam sieve script", "error", err)
		} else {
			s.logger.Info("global spam sieve script deployed")
			// Reload dovecot to pick up sieve_before config and compiled script
			if dovecotMgr != nil {
				if err := dovecotMgr.Reload(); err != nil {
					s.logger.Error("failed to reload dovecot after sieve deploy", "error", err)
				}
			}
		}
		sieveHandler := NewSieveHandler(sieveMgr)
		sieveHandler.RegisterRoutes(r)

		// DNS records and verification
		dkimSvc := domain.NewDKIMService(conn, "/etc/oxmail/dkim")
		RegisterDKIMRoutes(r, dkimSvc)
		dnsHandler := NewDNSHandler(
			os.Getenv("OXMAIL_DOMAIN"),
			os.Getenv("OXMAIL_PUBLIC_IP"),
			dkimSvc,
			&NetDNSResolver{},
		)
		dnsHandler.RegisterRoutes(r)

		smtpSender := mail.NewSMTPSender(mail.SMTPSenderConfig{
			Host: "postfix",
			Port: "25",
		})
		sendHandler := NewSendHandler(smtpSender, logBuffer)
		if IsDevMode() {
			sendHandler.rateLimit = 1000
		}
		sendHandler.RegisterRoutes(r)
	})

	// Dev-only routes: test endpoints available only in dev mode
	if IsDevMode() {
		smtpSender := mail.NewSMTPSender(mail.SMTPSenderConfig{
			Host: "postfix",
			Port: "25",
		})
		devHandler := NewDevHandler(smtpSender)
		devHandler.RegisterRoutes(s.router)
		s.logger.Info("dev mode enabled: registered dev endpoints")
	}
}

// generateRandomSecret creates a random 32-byte hex string for JWT signing.
func generateRandomSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "fallback-insecure-secret-change-me"
	}
	return hex.EncodeToString(b)
}

// ListenAndServe starts the HTTP server with graceful shutdown.
func (s *Server) ListenAndServe(port string) error {
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)

	go func() {
		s.logger.Info("server starting", "port", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case sig := <-shutdownCh:
		s.logger.Info("shutdown signal received", "signal", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}

		s.logger.Info("server stopped gracefully")
		return nil
	}
}

package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"intermark/internal/files"
	"intermark/internal/utils"

	"github.com/Data-Corruption/blog"
	"github.com/go-chi/chi/v5"
)

const CloseTimeout = 3 * time.Second

var (
	ServerInstance Server // Global server instance
	Version        string // Version of the application
)

type Server struct {
	ShutdownRoute  chan bool      // Channel to trigger a shutdown from the route
	ShutdownSignal chan os.Signal // Channel to listen for OS signals
	router         *chi.Mux
	server         *http.Server // The http or https server
	usingTLS       bool
	port           string // e.g. ":80"
	certPath       string
	keyPath        string
}

func (s *Server) Start() {
	if utils.Config.Server.TLSCertPath != "" && utils.Config.Server.TLSKeyPath != "" {
		s.certPath = filepath.Clean(utils.Config.Server.TLSCertPath)
		s.keyPath = filepath.Clean(utils.Config.Server.TLSKeyPath)
		if filesExist, err := files.Exists(s.keyPath, s.certPath); (err == nil) && filesExist {
			s.usingTLS = true
		} else {
			blog.Warnf("TLS files '%s', and '%s' not found, using HTTP", s.keyPath, s.certPath)
			s.usingTLS = false
		}
	} else {
		blog.Warn("TLS files not configured, using HTTP")
		s.usingTLS = false
	}

	// Init port
	if utils.Config.Server.Port == 0 {
		if s.usingTLS {
			s.port = ":443"
		} else {
			s.port = ":80"
		}
	} else {
		s.port = fmt.Sprintf(":%d", utils.Config.Server.Port)
	}

	// Initialize the server
	var err error
	s.ShutdownRoute = make(chan bool, 1)
	s.router = NewRouter(&s.usingTLS)

	// Configure the server
	if s.usingTLS {
		s.server = &http.Server{
			Addr:    s.port,
			Handler: s.router,
			TLSConfig: &tls.Config{
				MinVersion:               tls.VersionTLS12,
				PreferServerCipherSuites: true,
			},
		}
	} else {
		s.server = &http.Server{
			Addr:    s.port,
			Handler: s.router,
		}
	}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for shutdown signals
	s.ShutdownSignal = make(chan os.Signal, 1)
	signal.Notify(s.ShutdownSignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		select {
		case <-s.ShutdownRoute:
			blog.Info("Shutdown initiated via route...")
		case <-s.ShutdownSignal:
			blog.Info("Shutdown initiated via OS signal...")
		}

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, cancel := context.WithTimeout(serverCtx, CloseTimeout)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				panic("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		if err = s.server.Shutdown(shutdownCtx); err != nil {
			panic(err)
		}
		serverStopCtx()
	}()

	waitForServer := true

	var ip string
	ip, err = utils.GetPublicIP()
	if err != nil {
		blog.Errorf("issue getting public IP: %v", err)
		fmt.Println("issue getting public IP:", err)
		waitForServer = false
	}

	lan := fmt.Sprintf("http://localhost%s", s.port)
	wan := fmt.Sprintf("http://%s%s", ip, s.port)

	log.Print("Starting server...")
	log.Printf("LAN: %s %s/edit", lan, lan)
	log.Printf("WAN: %s %s/edit", wan, wan)
	log.Printf("Press Ctrl+C to stop the server...")

	if s.usingTLS {
		err = s.server.ListenAndServeTLS(s.certPath, s.keyPath)
	} else {
		err = s.server.ListenAndServe()
	}

	if err != nil {
		if err != http.ErrServerClosed {
			fmt.Println("server error:", err)
			blog.Errorf("server error: %v", err)
			waitForServer = false
		}
	}

	if waitForServer {
		<-serverCtx.Done()
		blog.Info("Server shutdown gracefully")
	}
}

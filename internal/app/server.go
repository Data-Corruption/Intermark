package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
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

	// Attempt to start the server, if port is in use, prompt for a new port and try again
	for {
		waitForServer = true

		var ip string
		ip, err = utils.GetPublicIP()
		if err != nil {
			blog.Errorf("Error getting public IP: %v", err)
		}

		protocol := utils.Ternary(s.usingTLS, "https://", "http://")
		lan := fmt.Sprintf("%slocalhost%s", protocol, s.port)
		wan := fmt.Sprintf("%s%s%s", protocol, ip, s.port)

		log.Print("Starting server...")
		log.Printf("LAN: %s %s/edit", lan, lan)
		log.Printf("WAN: %s %s/edit", wan, wan)
		log.Printf("Press Ctrl+C to stop the server...")

		var err error = nil
		if s.usingTLS {
			err = s.server.ListenAndServeTLS(s.certPath, s.keyPath)
		} else {
			err = s.server.ListenAndServe()
		}

		if err != nil {
			// Standard exit
			if err == http.ErrServerClosed {
				break
			}

			// if error is not a network error, print and exit
			opErr, ok := err.(*net.OpError)
			if !(ok && opErr.Op == "listen") {
				blog.Errorf("Error starting server: %v", err)
				fmt.Println("Error starting server:", err)
				waitForServer = false
				break
			}

			// if error is not an address in use error, print and exit
			sysErr, ok := opErr.Err.(*os.SyscallError)
			if !(ok && sysErr.Err == syscall.EADDRINUSE) {
				blog.Errorf("Error starting server: %v", err)
				fmt.Println("Error starting server:", err)
				waitForServer = false
				break
			}

			// Get a new port
			newPort := utils.PromptInt("Port in use. Enter a new port or 0 to abort: ")
			if newPort == 0 {
				waitForServer = false
				break
			} else {
				// Update the port and try again
				s.port = fmt.Sprintf(":%d", newPort)
				s.server.Addr = s.port
				utils.Config.Server.Port = newPort
				utils.Config.Save()
			}
		} else {
			break
		}
	}

	if waitForServer {
		<-serverCtx.Done()
		blog.Info("Server shutdown gracefully")
	}
}

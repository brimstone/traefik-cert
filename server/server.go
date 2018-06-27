package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"time"

	"github.com/brimstone/jwt/jwt"
	"github.com/brimstone/traefik-cert/types"
)

type Server struct {
	acmefile string
	address  string
	healthy  *int32
	key      string
	logger   *log.Logger
	router   *http.ServeMux
	server   *http.Server
}

type ServerOptions struct {
	AcmeFile string
	Address  string
	Key      string
}

func NewServer(o ServerOptions) (*Server, error) {
	s := &Server{
		address:  o.Address,
		key:      o.Key,
		acmefile: o.AcmeFile,
	}
	return s, nil
}

func (s *Server) Serve() error {
	s.logger = log.New(os.Stdout, "http: ", log.LstdFlags)
	s.router = http.NewServeMux()
	s.router.Handle("/", index())
	s.router.Handle("/cert/", getCert(s.key, s.acmefile))
	s.router.Handle("/healthz", healthz(s.healthy))
	s.server = &http.Server{
		Addr:         s.address,
		Handler:      (logging(s.logger)(s.router)),
		ErrorLog:     s.logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		s.logger.Println("Server is shutting down...")
		atomic.StoreInt32(s.healthy, 0)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		s.server.SetKeepAlivesEnabled(false)
		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	s.logger.Println("Server is ready to handle requests at", s.address)
	s.healthy = new(int32)
	atomic.StoreInt32(s.healthy, 1)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("could not listen on %s: %v", s.address, err)
	}

	<-done
	s.logger.Println("Server stopped")
	return nil
}

func index() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Hello, World!")
	})
}

func ifExists(needle string, haystack []string) bool {
	for _, object := range haystack {
		if object == needle {
			return true
		}
	}
	return false
}

func getCert(key string, acmefile string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for /domain/
		domain := strings.TrimPrefix(r.URL.Path, "/cert/")
		// Check for /domain/
		if domain == "" {
			http.Error(w, "Expected cert", http.StatusBadRequest)
			return
		}

		// Check for empty JWT
		clientToken := r.Header.Get("Authorization")
		if clientToken == "" {
			http.Error(w, "Expected authorization", http.StatusBadRequest)
			return
		}
		// Check for Bearer
		if !strings.HasPrefix(clientToken, "Bearer ") {
			http.Error(w, "Authorization in the wrong form.", http.StatusBadRequest)
			return
		}
		clientToken = strings.TrimPrefix(clientToken, "Bearer ")

		var clientPayload types.Auth
		err := jwt.Verify(key, clientToken, &clientPayload)
		if err != nil {
			log.Printf("Authorization failed: %s %s\n", clientToken, err)
			http.Error(w, "Authorization failed", http.StatusUnauthorized)
			return
		}

		if !ifExists(domain, clientPayload.Cert.Domains) {
			http.Error(w, "Unauthorized domain", http.StatusUnauthorized)
			return
		}

		acmecertsraw, err := ioutil.ReadFile(acmefile)
		if err != nil {
			http.Error(w, "Unable to read ACME file", http.StatusInternalServerError)
			return
		}
		var acmecerts types.Acme
		err = json.Unmarshal(acmecertsraw, &acmecerts)
		if err != nil {
			http.Error(w, "Unable to parse ACME file", http.StatusInternalServerError)
			return
		}

		var response types.CertResponse
		for _, acmecert := range acmecerts.Certificates {
			if acmecert.Domain.Main == domain {
				response.Cert, err = base64.StdEncoding.DecodeString(acmecert.Certificate)
				response.Key, err = base64.StdEncoding.DecodeString(acmecert.Key)
				break
			}
		}

		if response.Cert == nil {
			http.Error(w, "Domain not found", http.StatusNotFound)
			return
		}

		responsejson, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Unable to marshal response")
			return
		}

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responsejson)
	})
}

func healthz(healthy *int32) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(healthy) == 1 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				logger.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
	"github.com/vektah/gqlparser/gqlerror"
	"golang.org/x/time/rate"
)

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "600")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func NewStructuredLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&StructuredLogger{logger})
}

func (s *Server) RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Real-IP")
		limiter := s.getVisitor(ip)
		if !limiter.Allow() {

			b, _ := json.Marshal(gqlerror.Error{
				Message: "Too Many Requests",
			})

			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write(b)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) addVisitor(ip string) *rate.Limiter {
	limiter := rate.NewLimiter(2, 5)
	mtx.Lock()
	s.visitors[ip] = &visitor{limiter, time.Now()}
	mtx.Unlock()
	return limiter
}

func (s *Server) getVisitor(ip string) *rate.Limiter {
	mtx.Lock()
	v, exists := s.visitors[ip]
	if !exists {
		mtx.Unlock()
		return s.addVisitor(ip)
	}

	v.lastSeen = time.Now()
	mtx.Unlock()
	return v.limiter
}

func (s *Server) cleanUpVisitors() {
	for {
		time.Sleep(time.Minute)
		mtx.Lock()
		for ip, v := range s.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(s.visitors, ip)
			}
		}
		mtx.Unlock()
	}
}

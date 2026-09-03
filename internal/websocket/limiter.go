package websocket

import (
	"net"
	"net/http"
	"sync"
)

type connectionLimiter struct {
	mu      sync.Mutex
	clients map[string]int
	limit   int
}

func newConnectionLimiter(limit int) *connectionLimiter {
	return &connectionLimiter{
		clients: make(map[string]int),
		limit:   limit,
	}
}

func (l *connectionLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.clients[ip] >= l.limit {
		return false
	}

	l.clients[ip]++

	return true
}

func (l *connectionLimiter) Release(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.clients[ip] <= 1 {
		delete(l.clients, ip)
		return
	}

	l.clients[ip]--
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		return r.RemoteAddr
	}

	return host
}

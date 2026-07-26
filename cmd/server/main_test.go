package main

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/config"
)

func Test_newHTTPServer_configures_resource_timeouts(t *testing.T) {
	// Given
	cfg := &config.Config{Server: config.ServerConfig{
		Port:              "8080",
		ReadHeaderTimeout: 2,
		ReadTimeout:       5,
		WriteTimeout:      7,
		IdleTimeout:       30,
	}}

	// When
	server := newHTTPServer(cfg, gin.New())

	// Then
	if server.ReadHeaderTimeout != 2*time.Second {
		t.Fatalf("ReadHeaderTimeout = %s, want 2s", server.ReadHeaderTimeout)
	}
	if server.ReadTimeout != 5*time.Second {
		t.Fatalf("ReadTimeout = %s, want 5s", server.ReadTimeout)
	}
	if server.WriteTimeout != 7*time.Second {
		t.Fatalf("WriteTimeout = %s, want 7s", server.WriteTimeout)
	}
	if server.IdleTimeout != 30*time.Second {
		t.Fatalf("IdleTimeout = %s, want 30s", server.IdleTimeout)
	}
}

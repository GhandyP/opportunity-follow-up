package config

import "testing"

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		appAddr  string
		wantAddr string
		wantErr  bool
	}{
		{
			name:     "defaults to loopback",
			appAddr:  "",
			wantAddr: "127.0.0.1:8080",
		},
		{
			name:     "uses configured loopback address",
			appAddr:  "127.0.0.1:9090",
			wantAddr: "127.0.0.1:9090",
		},
		{
			name:     "allows localhost hostname",
			appAddr:  "localhost:9090",
			wantAddr: "localhost:9090",
		},
		{
			name:    "rejects address without port",
			appAddr: "127.0.0.1",
			wantErr: true,
		},
		{
			name:    "rejects non-loopback bind",
			appAddr: "0.0.0.0:9090",
			wantErr: true,
		},
		{
			name:    "rejects non-numeric port",
			appAddr: "127.0.0.1:not-a-port",
			wantErr: true,
		},
		{
			name:    "rejects an empty host",
			appAddr: ":9090",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_ADDR", tt.appAddr)

			cfg, err := Load()
			if tt.wantErr {
				if err == nil {
					t.Fatal("Load() error = nil, want non-nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.HTTPAddr != tt.wantAddr {
				t.Errorf("HTTPAddr = %q, want %q", cfg.HTTPAddr, tt.wantAddr)
			}
		})
	}
}

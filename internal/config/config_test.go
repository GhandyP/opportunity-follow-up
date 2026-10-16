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

func TestLoadHistoryAndGeminiConfiguration(t *testing.T) {
	tests := []struct {
		name            string
		historyPath     string
		apiKey          string
		model           string
		wantHistoryPath string
		wantAPIKey      string
		wantModel       string
	}{
		{
			name:            "uses ignored local database path by default",
			wantHistoryPath: "data/opportunity-follow-up.sqlite",
		},
		{
			name:            "uses configured local database and server-side Gemini settings",
			historyPath:     " data/custom-history.sqlite ",
			apiKey:          " test-only-api-key ",
			model:           " gemini-test-model ",
			wantHistoryPath: "data/custom-history.sqlite",
			wantAPIKey:      "test-only-api-key",
			wantModel:       "gemini-test-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_ADDR", "127.0.0.1:9090")
			t.Setenv("HISTORY_DB_PATH", tt.historyPath)
			t.Setenv("GOOGLE_API_KEY", tt.apiKey)
			t.Setenv("GEMINI_MODEL", tt.model)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.HTTPAddr != "127.0.0.1:9090" {
				t.Errorf("HTTPAddr = %q, want 127.0.0.1:9090", cfg.HTTPAddr)
			}
			if cfg.HistoryDBPath != tt.wantHistoryPath {
				t.Errorf("HistoryDBPath = %q, want %q", cfg.HistoryDBPath, tt.wantHistoryPath)
			}
			if cfg.GoogleAPIKey != tt.wantAPIKey {
				t.Errorf("GoogleAPIKey = %q, want %q", cfg.GoogleAPIKey, tt.wantAPIKey)
			}
			if cfg.GeminiModel != tt.wantModel {
				t.Errorf("GeminiModel = %q, want %q", cfg.GeminiModel, tt.wantModel)
			}
		})
	}
}

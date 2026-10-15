package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const defaultHTTPAddr = "127.0.0.1:8080"

type Config struct {
	HTTPAddr string
}

func Load() (Config, error) {
	addr := strings.TrimSpace(os.Getenv("APP_ADDR"))
	if addr == "" {
		addr = defaultHTTPAddr
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil || !isLoopbackHost(host) {
		return Config{}, fmt.Errorf("APP_ADDR must use a loopback host and valid port")
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return Config{}, fmt.Errorf("APP_ADDR must use a loopback host and valid port")
	}

	return Config{HTTPAddr: addr}, nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

// checkMetric fetches the specified Prometheus metric and determines the server's health and weight.
func checkMetric(metricsURL string) (bool, int) {
	resp, err := http.Get(metricsURL)
	if err != nil {
		log.Println("Error fetching metrics:", err)
		return false, 0 // Consider the server unhealthy if metrics cannot be fetched.
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return false, 0
	}

	re := regexp.MustCompile(`sync_execution_network_diff (\d+)`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		log.Println("Metric not found")
		return false, 0
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		log.Println("Error parsing metric value:", err)
		return false, 0
	}
	log.Printf("Metric value: %d\n", value)

	// Assuming "healthy" means the metric is 0, setting full weight.
	// Any other value indicates less health, reducing weight to 50%.
	if value == 0 {
		return true, 100 // Healthy, full weight
	}
	return false, 50 // Not fully healthy, reduced weight
}

// startAgentCheckServer starts a TCP server to handle HAProxy's agent checks, using the health determined by checkMetric.
func startAgentCheckServer(metricsURL string) {
	listenPort := os.Getenv("AGENT_LISTEN_PORT")
	if listenPort == "" {
		listenPort = "9876" // Default port for HAProxy agent checks
	}

	listener, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		log.Fatalf("Failed to start agent check server: %v", err)
	}
	defer listener.Close()
	log.Printf("Agent check server listening on :%s\n", listenPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go func(c net.Conn) {
			defer c.Close()
			_, weight := checkMetric(metricsURL) // Only weight is used for HAProxy agent check response.
			response := fmt.Sprintf("up weight=%d\n", weight)
			log.Printf("Responding to HAProxy agent check: %s", response)
			c.Write([]byte(response))
		}(conn)
	}
}

func main() {
	metricsHostname := os.Getenv("METRICS_HOSTNAME")
	if metricsHostname == "" {
		metricsHostname = "localhost"
	}

	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9090" // Default Prometheus port
	}

	metricsURL := fmt.Sprintf("http://%s:%s/metrics", metricsHostname, metricsPort)
	log.Printf("Using metrics URL: %s\n", metricsURL)

	// Start the agent check server
	startAgentCheckServer(metricsURL)
}

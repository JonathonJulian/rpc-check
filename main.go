package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

func checkMetric(metricsURL string) int {
	resp, err := http.Get(metricsURL)
	if err != nil {
		fmt.Println("Error fetching metrics:", err)
		return 500
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return 500
	}

	re := regexp.MustCompile(`sync_execution_network_diff (\d+)`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		fmt.Println("Metric not found")
		return 500
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		fmt.Println("Error parsing metric value:", err)
		return 500
	}
	fmt.Println("Metric value", value)

	if value <= 2 {
		return 200
	}
	return 500
}

func main() {
	listenPort := os.Getenv("APP_LISTEN_PORT")
	if listenPort == "" {
		listenPort = "8282"
	}

	metricsHostname := os.Getenv("METRICS_HOSTNAME")
	if metricsHostname == "" {
		metricsHostname = "localhost"
	}

	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "3737"
	}

	metricsURL := fmt.Sprintf("http://%s:%s/metrics", metricsHostname, metricsPort)

	http.HandleFunc("/rpc-health", func(w http.ResponseWriter, r *http.Request) {
		status := checkMetric(metricsURL)
		w.WriteHeader(status)
	})

	fmt.Printf("Server is running on http://localhost:%s\n", listenPort)
	err := http.ListenAndServe(":"+listenPort, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

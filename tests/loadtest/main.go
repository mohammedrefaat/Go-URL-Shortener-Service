package main

import (
	"fmt"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

func main() {
	rate := vegeta.Rate{Freq: 1000, Per: time.Second}
	duration := 30 * time.Second
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    "http://localhost:8080/RzzYNV",
	})

	attacker := vegeta.NewAttacker()
	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, duration, "Load Test") {
		metrics.Add(res)
	}
	metrics.Close()

	fmt.Printf("99th percentile: %s\n", metrics.Latencies.P99)
	fmt.Printf("Requests: %d\n", metrics.Requests)
	fmt.Printf("Success rate: %.2f%%\n", metrics.Success*100)
	fmt.Printf("Error rate: %.2f%%\n", (1-metrics.Success)*100)
	fmt.Printf("Total errors: %d\n", len(metrics.Errors))
}

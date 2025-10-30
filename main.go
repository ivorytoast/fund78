package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"fund78/internal/queue"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "engine" {
		testEngineQueue()
	} else {
		testReplayQueue()
	}
}

func testReplayQueue() {
	fmt.Println("Testing ReplayQueue - reading latest simulation input.log")
	q := queue.NewReplayQueue()
	defer q.Stop()

	p, err := latestSimulationInput()
	if err != nil {
		fmt.Printf("Error finding latest simulation input: %v\n", err)
		return
	}

	err = q.StartReadingLogFile(p)
	if err != nil {
		fmt.Printf("Error starting log file reader: %v\n", err)
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
	case <-q.GetQuit():
		fmt.Println("File processing complete")
	}
}

func testEngineQueue() {
	fmt.Println("Testing EngineQueue - programmatic enqueue")
	q := queue.NewEngineQueue()
	defer q.Stop()

	testData := []string{
		// Valid enveloped events (payload is JSON string)
		`{"topic":"auth","payload":"{\"name\":\"user1\",\"action\":\"login\",\"timestamp\":\"2024-01-01T10:00:00Z\"}"}`,
		`{"topic":"auth","payload":"{\"name\":\"user2\",\"action\":\"logout\",\"timestamp\":\"2024-01-01T10:05:00Z\"}"}`,
		`{"topic":"auth","payload":"{\"name\":\"user3\",\"action\":\"login\",\"timestamp\":\"2024-01-01T10:10:00Z\"}"}`,
		// Invalid envelope: payload is not JSON
		`{"topic":"auth","payload":"not json at all"}`,
		// Invalid envelope: payload is malformed JSON string
		`{"topic":"auth","payload":"{\"name\":\"broken\",\"action\":\"oops\""}`,
		// Invalid envelope: missing topic
		`{"payload":"{\"name\":\"user5\",\"action\":\"purchase\",\"timestamp\":\"2024-01-01T10:20:00Z\",\"amount\":99.99}"}`,
		// Completely non-conforming (raw JSON, not enveloped)
		`{"name":"raw_no_envelope"}`,
		// Empty and whitespace
		``,
		`   `,
	}

	fmt.Println("Enqueuing test data...")
	for i, data := range testData {
		fmt.Printf("Enqueuing item %d: %s\n", i+1, data)
		err := q.Enqueue(data)
		if err != nil {
			fmt.Printf("Error enqueuing item %d: %v\n", i+1, err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("Waiting for processing to complete...")
	time.Sleep(2 * time.Second)

	fmt.Println("EngineQueue test complete")
}

func latestSimulationInput() (string, error) {
	base := os.Getenv("SIMULATIONS_DIR")
	if base == "" {
		base = "simulations"
	}

	entries, err := os.ReadDir(base)
	if err != nil {
		return "", err
	}
	var years []string
	for _, e := range entries {
		if e.IsDir() {
			years = append(years, e.Name())
		}
	}
	if len(years) == 0 {
		return "", fmt.Errorf("no simulations found")
	}
	sort.Strings(years)
	y := years[len(years)-1]

	monthsEntries, err := os.ReadDir(filepath.Join(base, y))
	if err != nil {
		return "", err
	}
	var months []string
	for _, e := range monthsEntries {
		if e.IsDir() {
			months = append(months, e.Name())
		}
	}
	if len(months) == 0 {
		return "", fmt.Errorf("no months in %s", y)
	}
	sort.Strings(months)
	m := months[len(months)-1]

	daysEntries, err := os.ReadDir(filepath.Join(base, y, m))
	if err != nil {
		return "", err
	}
	var days []string
	for _, e := range daysEntries {
		if e.IsDir() {
			days = append(days, e.Name())
		}
	}
	if len(days) == 0 {
		return "", fmt.Errorf("no days in %s/%s", y, m)
	}
	sort.Strings(days)
	d := days[len(days)-1]

	p := filepath.Join(base, y, m, d, "input.log")
	if _, err := os.Stat(p); err != nil {
		return "", err
	}
	return p, nil
}

package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"fund78/internal/queue"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "engine":
			testEngineQueue()
			return
		case "compare":
			runCompare()
			return
		}
	}
	testReplayQueue()
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
	q.RegisterApplications(appOne{}, appTwo{}, appThree{})
	defer q.Stop()

	testData := []string{
		`{"topic":"topic-1","payload":"{\"name\":\"user1\",\"action\":\"login\",\"timestamp\":\"2024-01-01T10:00:00Z\"}"}`,
		`{"topic":"topic-2","payload":"{\"name\":\"user2\",\"action\":\"logout\",\"timestamp\":\"2024-01-01T10:05:00Z\"}"}`,
		`{"topic":"topic-3","payload":"{\"name\":\"user3\",\"action\":\"login\",\"timestamp\":\"2024-01-01T10:10:00Z\"}"}`,
		`{"topic":"other","payload":"{\"note\":\"no handler\"}"}`,
		`{"topic":"topic-1","payload":"not json at all"}`,
		`{"topic":"topic-2","payload":"{\"name\":\"broken\",\"action\":\"oops\""}`,
		`{"payload":"{\"name\":\"user5\",\"action\":\"purchase\",\"timestamp\":\"2024-01-01T10:20:00Z\",\"amount\":99.99}"}`,
		`{"name":"raw_no_envelope"}`,
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

func runCompare() {
	base := os.Getenv("SIMULATIONS_DIR")
	if base == "" {
		base = "simulations"
	}
	var rows []struct {
		outPath   string
		identical bool
		diffLine  int
		note      string
		outSample string
		inSample  string
	}
	_ = filepath.Walk(base, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info == nil || info.IsDir() {
			return nil
		}
		name := info.Name()
		if strings.HasPrefix(name, "output_debug_") && strings.HasSuffix(name, ".log") {
			ts := strings.TrimSuffix(strings.TrimPrefix(name, "output_debug_"), ".log")
			debugDir := filepath.Dir(p)
			dayDir := filepath.Dir(debugDir)
			inPath := filepath.Join(dayDir, "output_"+ts+".log")
			same := false
			diffLine := 0
			note := ""
			outSample, inSample := "", ""
			if bOut, e1 := os.ReadFile(p); e1 == nil {
				if bIn, e2 := os.ReadFile(inPath); e2 == nil {
					same, diffLine, note, outSample, inSample = compareFiles(string(bOut), string(bIn))
				}
			}
			rows = append(rows, struct {
				outPath   string
				identical bool
				diffLine  int
				note      string
				outSample string
				inSample  string
			}{outPath: p, identical: same, diffLine: diffLine, note: note, outSample: outSample, inSample: inSample})
		}
		return nil
	})

	if len(rows) == 0 {
		fmt.Println("No debug output files found.")
		return
	}
	fmt.Println("Debug comparisons (output_debug vs input_debug):")
	for _, r := range rows {
		status := "different"
		if r.identical {
			status = "identical"
		}
		if r.identical {
			fmt.Printf("- %s => %s\n", r.outPath, status)
		} else {
			if r.diffLine > 0 {
				fmt.Printf("- %s => %s (first difference at line %d; %s)\n", r.outPath, status, r.diffLine, r.note)
				if r.outSample != "" || r.inSample != "" {
					fmt.Printf("  output: %s\n", r.outSample)
					fmt.Printf("  input : %s\n", r.inSample)
				}
			} else {
				fmt.Printf("- %s => %s (unable to locate differing line)\n", r.outPath, status)
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func compareFiles(outContent, inContent string) (bool, int, string, string, string) {
	normalize := func(s string) []string {
		s = strings.ReplaceAll(s, "\r\n", "\n")
		s = strings.ReplaceAll(s, "\r", "\n")
		lines := strings.Split(s, "\n")
		// trim trailing empty lines
		i := len(lines) - 1
		for i >= 0 && strings.TrimSpace(lines[i]) == "" {
			i--
		}
		lines = lines[:i+1]
		for j := range lines {
			lines[j] = strings.TrimRight(lines[j], " \t")
		}
		return lines
	}
	outLines := normalize(outContent)
	inLines := normalize(inContent)
	max := len(outLines)
	if len(inLines) > max {
		max = len(inLines)
	}
	for i := 0; i < max; i++ {
		var ol, il string
		if i < len(outLines) {
			ol = outLines[i]
		}
		if i < len(inLines) {
			il = inLines[i]
		}
		if ol != il {
			return false, i + 1, "output != input", ol, il
		}
	}
	return true, 0, "", "", ""
}

type appOne struct{}

func (appOne) Accept(topic string) bool {
	return strings.Contains(topic, "1") || strings.EqualFold(topic, "one")
}
func (appOne) Handle(env queue.EventEnvelope) (queue.EventEnvelope, error) { return env, nil }

type appTwo struct{}

func (appTwo) Accept(topic string) bool {
	return strings.Contains(topic, "2") || strings.EqualFold(topic, "two")
}
func (appTwo) Handle(env queue.EventEnvelope) (queue.EventEnvelope, error) { return env, nil }

type appThree struct{}

func (appThree) Accept(topic string) bool {
	return strings.Contains(topic, "3") || strings.EqualFold(topic, "three")
}
func (appThree) Handle(env queue.EventEnvelope) (queue.EventEnvelope, error) { return env, nil }

type envelopeAlias = struct{ Topic, Payload string }

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

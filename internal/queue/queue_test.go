package queue

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "simulations_test_*")
	if err != nil {
		os.Exit(1)
	}
	os.Setenv("SIMULATIONS_DIR", dir)
	code := m.Run()
	os.Unsetenv("SIMULATIONS_DIR")
	// Also remove any local debug/ created by older behavior
	_ = os.RemoveAll("debug")
	os.RemoveAll(dir)
	os.Exit(code)
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func ev(topic, payload string) string {
	return fmt.Sprintf(`{"topic":"%s","payload":%q}`, topic, payload)
}

func TestEngineQueue_WritesTimestampedAndLatestFiles(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	err := q.Enqueue(`{"k":"v"}`)
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	base := os.Getenv("SIMULATIONS_DIR")
	now := time.Now()
	y, m, d := now.Format("2006"), now.Format("01"), now.Format("02")
	dayDir := base + "/" + y + "/" + m + "/" + d

	// Latest real files should exist and not be symlinks
	latestIn := dayDir + "/input.log"
	latestOut := dayDir + "/output.log"
	if !pathExists(latestIn) || !pathExists(latestOut) {
		t.Fatalf("latest files not found: %s or %s", latestIn, latestOut)
	}
	if fi, err := os.Lstat(latestIn); err != nil {
		t.Fatalf("lstat latest input: %v", err)
	} else if (fi.Mode() & os.ModeSymlink) != 0 {
		t.Fatalf("latest input is symlink, expected regular file")
	}
	if fo, err := os.Lstat(latestOut); err != nil {
		t.Fatalf("lstat latest output: %v", err)
	} else if (fo.Mode() & os.ModeSymlink) != 0 {
		t.Fatalf("latest output is symlink, expected regular file")
	}

	// Timestamped files should also exist
	// Find any input_*.log created now
	entries, err := os.ReadDir(dayDir)
	if err != nil {
		t.Fatalf("readdir dayDir: %v", err)
	}
	foundTS := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "input_") && strings.HasSuffix(e.Name(), ".log") {
			foundTS = true
			break
		}
	}
	if !foundTS {
		t.Fatalf("no timestamped input_*.log found in %s", dayDir)
	}
}

func TestReplay_CreatesDebugWithMatchingTimestamp(t *testing.T) {
	base := os.Getenv("SIMULATIONS_DIR")
	now := time.Now()
	y, m, d := now.Format("2006"), now.Format("01"), now.Format("02")
	dayDir := base + "/" + y + "/" + m + "/" + d
	os.MkdirAll(dayDir, 0755)

	// Create two timestamped input files, the later should be selected for input.log derivation
	ts1 := "000000"
	ts2 := "235959"
	f1 := dayDir + "/input_" + ts1 + ".log"
	f2 := dayDir + "/input_" + ts2 + ".log"
	os.WriteFile(f1, []byte(`{"a":1}\n`), 0644)
	os.WriteFile(f2, []byte(`{"b":2}\n`), 0644)

	// Create latest as real file mirroring f2
	latest := dayDir + "/input.log"
	os.WriteFile(latest, []byte(`{"b":2}\n`), 0644)

	q := NewReplayQueue()
	defer q.Stop()

	// Start from latest input.log (no symlink), debug ts should match ts2
	if err := q.StartReadingLogFile(latest); err != nil {
		t.Fatalf("start reading: %v", err)
	}

	// Allow processing
	select {
	case <-q.GetQuit():
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout waiting for quit")
	}

	debugDir := dayDir + "/debug"
	expIn := debugDir + "/input_debug_" + ts2 + ".log"
	expOut := debugDir + "/output_debug_" + ts2 + ".log"
	if !pathExists(expIn) || !pathExists(expOut) {
		t.Fatalf("expected debug files %s and %s", expIn, expOut)
	}
}

func TestNewEngineQueue(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	if q.mode != EngineMode {
		t.Errorf("Expected EngineMode, got %v", q.mode)
	}

	if q.input == nil {
		t.Error("Expected input channel to be initialized")
	}

	if q.output == nil {
		t.Error("Expected output channel to be initialized")
	}
}

func TestNewReplayQueue(t *testing.T) {
	q := NewReplayQueue()
	defer q.Stop()

	if q.mode != ReplayMode {
		t.Errorf("Expected ReplayMode, got %v", q.mode)
	}

	if q.input == nil {
		t.Error("Expected input channel to be initialized")
	}

	if q.output == nil {
		t.Error("Expected output channel to be initialized")
	}
}

func TestEngineQueue_Enqueue(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	err := q.Enqueue("test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestEngineQueue_Enqueue_ReplayMode(t *testing.T) {
	q := NewReplayQueue()
	defer q.Stop()

	err := q.Enqueue("test")
	if err == nil {
		t.Error("Expected error for Enqueue in ReplayMode")
	}

	expected := "Enqueue is only available in Engine mode"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected error to contain '%s', got '%s'", expected, err.Error())
	}
}

func TestReplayQueue_StartReadingLogFile(t *testing.T) {
	q := NewReplayQueue()
	defer q.Stop()

	err := q.StartReadingLogFile("nonexistent.log")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestReplayQueue_StartReadingLogFile_EngineMode(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	err := q.StartReadingLogFile("test.log")
	if err == nil {
		t.Error("Expected error for StartReadingLogFile in EngineMode")
	}

	expected := "StartReadingLogFile cannot be called in Engine mode"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected error to contain '%s', got '%s'", expected, err.Error())
	}
}

func TestQueue_ProcessItems(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	testFile := "test_input.log"
	defer os.Remove(testFile)

	err := q.SetInputLogFile(testFile)
	if err != nil {
		t.Fatalf("Failed to set input log file: %v", err)
	}

	outputFile := "test_output.log"
	defer os.Remove(outputFile)

	err = q.SetOutputLogFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to set output log file: %v", err)
	}

	err = q.Enqueue(ev("t", `{"name": "test1"}`))
	if err != nil {
		t.Fatalf("Failed to enqueue: %v", err)
	}

	err = q.Enqueue(ev("t", `{"name": "test2"}`))
	if err != nil {
		t.Fatalf("Failed to enqueue: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Input log file was not created")
	}

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output log file was not created")
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, ev("t", `{"name": "test1"}`)) {
		t.Error("Expected enveloped test1 event in output file")
	}

	if !strings.Contains(output, ev("t", `{"name": "test2"}`)) {
		t.Error("Expected enveloped test2 event in output file")
	}
}

func TestQueue_ReadFromFile(t *testing.T) {
	testFile := "test_input_file.log"
	defer os.Remove(testFile)

	content := ev("t", `{"name":"line1"}`) + "\n" + ev("t", `{"name":"line2"}`) + "\n" + ev("t", `{"name":"line3"}`) + "\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	q := NewReplayQueue()
	defer q.Stop()

	inputFile := "test_input_read.log"
	defer os.Remove(inputFile)

	err = q.SetInputLogFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to set input log file: %v", err)
	}

	outputFile := "test_output_read.log"
	defer os.Remove(outputFile)

	err = q.SetOutputLogFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to set output log file: %v", err)
	}

	err = q.StartReadingLogFile(testFile)
	if err != nil {
		t.Fatalf("Failed to start reading log file: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(outputContent)
	expectedLines := []string{ev("t", `{"name":"line1"}`), ev("t", `{"name":"line2"}`), ev("t", `{"name":"line3"}`)}

	for _, expected := range expectedLines {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected '%s' in output file", expected)
		}
	}
}

func TestQueue_QuitSignal(t *testing.T) {
	testFile := "test_quit.log"
	defer os.Remove(testFile)

	content := "test1\ntest2\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	q := NewReplayQueue()
	defer q.Stop()

	err = q.StartReadingLogFile(testFile)
	if err != nil {
		t.Fatalf("Failed to start reading log file: %v", err)
	}

	select {
	case <-q.GetQuit():
	case <-time.After(1 * time.Second):
		t.Error("Expected quit signal within 1 second")
	}
}

func TestQueue_SetInputLogFile(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	testFile := "test_input_set.log"
	defer os.Remove(testFile)

	err := q.SetInputLogFile(testFile)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if q.inputFile == nil {
		t.Error("Expected input file to be set")
	}
}

func TestQueue_SetOutputLogFile(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	testFile := "test_output_set.log"
	defer os.Remove(testFile)

	err := q.SetOutputLogFile(testFile)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if q.outputFile == nil {
		t.Error("Expected output file to be set")
	}
}

func TestQueue_GetOutput(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	output := q.GetOutput()
	if output == nil {
		t.Error("Expected output channel to be returned")
	}
}

func TestQueue_GetQuit(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	quit := q.GetQuit()
	if quit == nil {
		t.Error("Expected quit channel to be returned")
	}
}

func TestQueue_Stop(t *testing.T) {
	q := NewEngineQueue()

	err := q.Enqueue("test")
	if err != nil {
		t.Fatalf("Failed to enqueue: %v", err)
	}

	q.Stop()

	time.Sleep(100 * time.Millisecond)
}

func TestQueue_ConcurrentEnqueue(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	outputFile := "test_concurrent.log"
	defer os.Remove(outputFile)

	err := q.SetOutputLogFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to set output log file: %v", err)
	}

	for i := 0; i < 10; i++ {
		go func(i int) {
			payload := fmt.Sprintf(`{"id": %d, "name": "test"}`, i)
			jsonStr := ev("t", payload)
			err := q.Enqueue(jsonStr)
			if err != nil {
				t.Errorf("Failed to enqueue: %v", err)
			}
		}(i)
	}

	time.Sleep(200 * time.Millisecond)

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	expectedCount := 10
	nonEmptyLines := 0
	for _, ln := range strings.Split(strings.TrimSpace(output), "\n") {
		if strings.TrimSpace(ln) != "" {
			nonEmptyLines++
		}
	}
	if nonEmptyLines != expectedCount {
		t.Errorf("Expected %d processed items, got %d", expectedCount, nonEmptyLines)
	}
}

func TestQueue_EmptyFile(t *testing.T) {
	testFile := "test_empty.log"
	defer os.Remove(testFile)

	err := os.WriteFile(testFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty test file: %v", err)
	}

	q := NewReplayQueue()
	defer q.Stop()

	outputFile := "test_empty_output.log"
	defer os.Remove(outputFile)

	err = q.SetOutputLogFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to set output log file: %v", err)
	}

	err = q.StartReadingLogFile(testFile)
	if err != nil {
		t.Fatalf("Failed to start reading log file: %v", err)
	}

	select {
	case <-q.GetQuit():
	case <-time.After(1 * time.Second):
		t.Error("Expected quit signal for empty file")
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(strings.TrimSpace(string(content))) > 0 {
		t.Error("Expected empty output file for empty input")
	}
}

func TestQueue_FileWithEmptyLines(t *testing.T) {
	testFile := "test_empty_lines.log"
	defer os.Remove(testFile)

	content := ev("t", `{"name":"line1"}`) + "\n\n" + ev("t", `{"name":"line2"}`) + "\n\n\n\n" + ev("t", `{"name":"line3"}`) + "\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	q := NewReplayQueue()
	defer q.Stop()

	outputFile := "test_empty_lines_output.log"
	defer os.Remove(outputFile)

	err = q.SetOutputLogFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to set output log file: %v", err)
	}

	err = q.StartReadingLogFile(testFile)
	if err != nil {
		t.Fatalf("Failed to start reading log file: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(outputContent)
	expectedLines := []string{ev("t", `{"name":"line1"}`), ev("t", `{"name":"line2"}`), ev("t", `{"name":"line3"}`)}

	for _, expected := range expectedLines {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected '%s' in output file", expected)
		}
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines (empty/whitespace lines filtered out), got %d", len(lines))
	}
}

func TestQueue_JSONValidation_ValidJSON(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	outputFile := "test_valid_json.log"
	defer os.Remove(outputFile)

	err := q.SetOutputLogFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to set output log file: %v", err)
	}

	validJSON := ev("t", `{"name": "test", "value": 123}`)
	err = q.Enqueue(validJSON)
	if err != nil {
		t.Fatalf("Failed to enqueue valid JSON: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, validJSON) {
		t.Errorf("Expected '%s' in output file, got '%s'", validJSON, output)
	}
}

func TestQueue_JSONValidation_InvalidJSON(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	outputFile := "test_invalid_json.log"
	defer os.Remove(outputFile)

	err := q.SetOutputLogFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to set output log file: %v", err)
	}

	invalidJSON := ev("t", `{"name": "test", "value": 123`) // missing closing brace
	err = q.Enqueue(invalidJSON)
	if err != nil {
		t.Fatalf("Failed to enqueue invalid JSON: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	if !(strings.Contains(output, `"topic":"error"`) && strings.Contains(output, `\"reason\":\"invalid_event\"`)) {
		t.Errorf("Expected error envelope with invalid_event, got '%s'", output)
	}
}

func TestQueue_JSONValidation_MixedValidInvalid(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	outputFile := "test_mixed_json.log"
	defer os.Remove(outputFile)

	err := q.SetOutputLogFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to set output log file: %v", err)
	}

	validJSON := ev("t", `{"name": "test", "value": 123}`)
	invalidJSON := ev("t", `{"name": "test", "value": 123`)

	err = q.Enqueue(validJSON)
	if err != nil {
		t.Fatalf("Failed to enqueue valid JSON: %v", err)
	}

	err = q.Enqueue(invalidJSON)
	if err != nil {
		t.Fatalf("Failed to enqueue invalid JSON: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)

	if !strings.Contains(output, validJSON) {
		t.Errorf("Expected '%s' in output file", validJSON)
	}
	if !(strings.Contains(output, `"topic":"error"`) && strings.Contains(output, `\"reason\":\"invalid_event\"`)) {
		t.Errorf("Expected error envelope with invalid_event in output")
	}
}

func TestQueue_JSONValidation_FromFile(t *testing.T) {
	testFile := "test_json_file.log"
	defer os.Remove(testFile)

	content := ev("t", `{"name": "test1", "value": 123}`) + "\n" + ev("t", `{"name": "test2", "value": 456`) + "\n" + ev("t", `{"name": "test3", "value": 789}`)

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	q := NewReplayQueue()
	defer q.Stop()

	outputFile := "test_json_file_output.log"
	defer os.Remove(outputFile)

	err = q.SetOutputLogFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to set output log file: %v", err)
	}

	err = q.StartReadingLogFile(testFile)
	if err != nil {
		t.Fatalf("Failed to start reading log file: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(outputContent)

	expectedLines := []string{
		ev("t", `{"name": "test1", "value": 123}`),
		ev("t", `{"name": "test3", "value": 789}`),
	}

	for _, expected := range expectedLines {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected '%s' in output file", expected)
		}
	}
	if !(strings.Contains(output, `"topic":"error"`) && strings.Contains(output, `\"reason\":\"invalid_event\"`)) {
		t.Errorf("Expected error envelope for invalid line")
	}
}

func TestQueue_JSONValidation_EmptyString(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	outputFile := "test_empty_json.log"
	defer os.Remove(outputFile)

	err := q.SetOutputLogFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to set output log file: %v", err)
	}

	err = q.Enqueue("")
	if err != nil {
		t.Fatalf("Failed to enqueue empty string: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	if output != "" {
		t.Errorf("Expected empty output file (empty strings filtered out), got '%s'", output)
	}
}

func TestQueue_JSONValidation_NonJSONString(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	outputFile := "test_non_json.log"
	defer os.Remove(outputFile)

	err := q.SetOutputLogFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to set output log file: %v", err)
	}

	nonJSON := ev("t", "this is not json at all")
	err = q.Enqueue(nonJSON)
	if err != nil {
		t.Fatalf("Failed to enqueue non-JSON string: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	if !(strings.Contains(output, `"topic":"error"`) && strings.Contains(output, `\"reason\":\"invalid_event\"`)) {
		t.Errorf("Expected error envelope with invalid_event, got '%s'", output)
	}
}

func TestQueue_EmptyEvents_Filtered(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	outputFile := "test_empty_events.log"
	defer os.Remove(outputFile)

	err := q.SetOutputLogFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to set output log file: %v", err)
	}

	err = q.Enqueue("")
	if err != nil {
		t.Fatalf("Failed to enqueue empty string: %v", err)
	}

	err = q.Enqueue("   ")
	if err != nil {
		t.Fatalf("Failed to enqueue whitespace string: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	if output != "" {
		t.Errorf("Expected empty output file (empty/whitespace strings filtered out), got '%s'", output)
	}
}

func TestQueue_EmptyEvents_FromFile(t *testing.T) {
	testFile := "test_empty_events_file.log"
	defer os.Remove(testFile)

	content := ev("t", `{"name":"valid1"}`) + "\n\n" + ev("t", `{"name":"valid2"}`) + "\n\n   "

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	q := NewReplayQueue()
	defer q.Stop()

	outputFile := "test_empty_events_output.log"
	defer os.Remove(outputFile)

	err = q.SetOutputLogFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to set output log file: %v", err)
	}

	err = q.StartReadingLogFile(testFile)
	if err != nil {
		t.Fatalf("Failed to start reading log file: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(outputContent)

	expectedLines := []string{
		ev("t", `{"name":"valid1"}`),
		ev("t", `{"name":"valid2"}`),
	}

	for _, expected := range expectedLines {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected '%s' in output file", expected)
		}
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines (empty/whitespace lines filtered out), got %d", len(lines))
	}
}

func TestQueue_EmptyStrings_NotLoggedToInput(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	inputFile := "test_empty_input.log"
	defer os.Remove(inputFile)

	err := q.SetInputLogFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to set input log file: %v", err)
	}

	err = q.Enqueue("")
	if err != nil {
		t.Fatalf("Failed to enqueue empty string: %v", err)
	}

	err = q.Enqueue("   ")
	if err != nil {
		t.Fatalf("Failed to enqueue whitespace string: %v", err)
	}

	err = q.Enqueue("\t\n")
	if err != nil {
		t.Fatalf("Failed to enqueue tab/newline string: %v", err)
	}

	err = q.Enqueue(`{"name": "valid"}`)
	if err != nil {
		t.Fatalf("Failed to enqueue valid JSON: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to read input file: %v", err)
	}

	input := string(content)

	lines := strings.Split(strings.TrimSpace(input), "\n")

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			t.Errorf("Line %d is empty or whitespace-only: '%s'", i+1, line)
		}
	}

	if !strings.Contains(input, `{"name": "valid"}`) {
		t.Error("Input file should contain valid JSON")
	}

	if len(lines) != 1 {
		t.Errorf("Expected 1 line in input file (only valid JSON), got %d", len(lines))
	}
}

func TestQueue_EmptyStrings_NotLoggedToInput_FromFile(t *testing.T) {
	testFile := "test_empty_input_file.log"
	defer os.Remove(testFile)

	content := `{"name": "valid1"}

{"name": "valid2"}

   `

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	q := NewReplayQueue()
	defer q.Stop()

	inputFile := "test_empty_input_replay.log"
	defer os.Remove(inputFile)

	err = q.SetInputLogFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to set input log file: %v", err)
	}

	err = q.StartReadingLogFile(testFile)
	if err != nil {
		t.Fatalf("Failed to start reading log file: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	inputContent, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to read input file: %v", err)
	}

	input := string(inputContent)

	lines := strings.Split(strings.TrimSpace(input), "\n")

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			t.Errorf("Line %d is empty or whitespace-only: '%s'", i+1, line)
		}
	}

	expectedLines := []string{
		`{"name": "valid1"}`,
		`{"name": "valid2"}`,
	}

	for _, expected := range expectedLines {
		if !strings.Contains(input, expected) {
			t.Errorf("Expected '%s' in input file", expected)
		}
	}

	if len(lines) != 2 {
		t.Errorf("Expected 2 lines in input file (empty lines filtered out), got %d", len(lines))
	}
}

func TestRouting_ApplicationsAndInvalidCases(t *testing.T) {
	q := NewEngineQueue()
	// Register simple pass-through apps like in main
	q.RegisterApplications(
		appMock{accept: func(topic string) bool { return strings.Contains(topic, "1") || strings.EqualFold(topic, "one") }},
		appMock{accept: func(topic string) bool { return strings.Contains(topic, "2") || strings.EqualFold(topic, "two") }},
		appMock{accept: func(topic string) bool { return strings.Contains(topic, "3") || strings.EqualFold(topic, "three") }},
	)
	defer q.Stop()

	outFile := "test_routing_output.log"
	defer os.Remove(outFile)
	if err := q.SetOutputLogFile(outFile); err != nil {
		t.Fatalf("set output: %v", err)
	}

	cases := []struct {
		name          string
		event         string
		wantProcessed bool
	}{
		{"appOne_valid", `{"topic":"topic-1","payload":"{\"name\":\"user1\",\"action\":\"login\",\"timestamp\":\"2024-01-01T10:00:00Z\"}"}`, true},
		{"appTwo_valid", `{"topic":"topic-2","payload":"{\"name\":\"user2\",\"action\":\"logout\",\"timestamp\":\"2024-01-01T10:05:00Z\"}"}`, true},
		{"appThree_valid", `{"topic":"topic-3","payload":"{\"name\":\"user3\",\"action\":\"login\",\"timestamp\":\"2024-01-01T10:10:00Z\"}"}`, true},
		{"no_handler_valid", `{"topic":"other","payload":"{\"note\":\"no handler\"}"}`, true},
		{"payload_not_json", `{"topic":"topic-1","payload":"not json at all"}`, false},
		{"payload_malformed", `{"topic":"topic-2","payload":"{\"name\":\"broken\",\"action\":\"oops\""}`, false},
		{"missing_topic", `{"payload":"{\"name\":\"user5\",\"action\":\"purchase\",\"timestamp\":\"2024-01-01T10:20:00Z\",\"amount\":99.99}"}`, false},
		{"raw_no_envelope", `{"name":"raw_no_envelope"}`, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := q.Enqueue(tc.event); err != nil {
				t.Fatalf("enqueue: %v", err)
			}
			time.Sleep(50 * time.Millisecond)
		})
	}

	b, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read out: %v", err)
	}
	out := string(b)

	// Check expectations
	// processed cases should have the full event
	wantProcessed := []string{
		`{"topic":"topic-1","payload":"{\"name\":\"user1\",\"action\":\"login\",\"timestamp\":\"2024-01-01T10:00:00Z\"}"}`,
		`{"topic":"topic-2","payload":"{\"name\":\"user2\",\"action\":\"logout\",\"timestamp\":\"2024-01-01T10:05:00Z\"}"}`,
		`{"topic":"topic-3","payload":"{\"name\":\"user3\",\"action\":\"login\",\"timestamp\":\"2024-01-01T10:10:00Z\"}"}`,
		`{"topic":"other","payload":"{\"note\":\"no handler\"}"}`,
	}
	for _, s := range wantProcessed {
		if !strings.Contains(out, s) {
			t.Errorf("expected processed line: %s", s)
		}
	}
	// invalid cases should include error envelope lines
	if !(strings.Contains(out, `"topic":"error"`) && strings.Contains(out, `\"reason\":\"invalid_event\"`)) {
		t.Errorf("expected error envelope with invalid_event in output")
	}
}

type appMock struct{ accept func(topic string) bool }

func (a appMock) Accept(topic string) bool                        { return a.accept(topic) }
func (a appMock) Handle(env EventEnvelope) (EventEnvelope, error) { return env, nil }

func TestOneToOne_Engine_InputToOutput(t *testing.T) {
	q := NewEngineQueue()
	defer q.Stop()

	inputFile := "one_to_one_engine_input.log"
	outputFile := "one_to_one_engine_output.log"
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	if err := q.SetInputLogFile(inputFile); err != nil {
		t.Fatalf("set input: %v", err)
	}
	if err := q.SetOutputLogFile(outputFile); err != nil {
		t.Fatalf("set output: %v", err)
	}

	inputs := []string{
		ev("t", `{"id":1}`),
		ev("t", `{"id":2}`),
		ev("t", `{"id":3}`),
		ev("t", `not json`), // invalid
		ev("", `{"id":4}`),  // invalid (empty topic)
		"",                  // filtered blank
		"   ",               // filtered blank
	}

	for _, s := range inputs {
		if err := q.Enqueue(s); err != nil {
			t.Fatalf("enqueue: %v", err)
		}
	}

	time.Sleep(300 * time.Millisecond)

	inB, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("read in: %v", err)
	}
	outB, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("read out: %v", err)
	}

	countNonEmpty := func(s string) int {
		n := 0
		for _, ln := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
			if strings.TrimSpace(ln) != "" {
				n++
			}
		}
		return n
	}

	inCount := countNonEmpty(string(inB))
	outCount := countNonEmpty(string(outB))

	if inCount != outCount {
		t.Fatalf("expected 1:1 lines, got input=%d output=%d", inCount, outCount)
	}
}

func TestOneToOne_Replay_InputToOutput(t *testing.T) {
	// Prepare a source file with mixed valid/invalid/blank lines
	src := "one_to_one_replay_source.log"
	defer os.Remove(src)
	srcContent := strings.Join([]string{
		ev("t", `{"id":1}`),
		"",
		ev("t", `{"id":2}`),
		"   ",
		ev("t", `not json`), // invalid
		ev("", `{"id":4}`),  // invalid (empty topic)
		ev("t", `{"id":5}`),
		"\n",
	}, "\n")
	if err := os.WriteFile(src, []byte(srcContent), 0644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	q := NewReplayQueue()
	defer q.Stop()

	inputFile := "one_to_one_replay_input.log"
	outputFile := "one_to_one_replay_output.log"
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	if err := q.SetInputLogFile(inputFile); err != nil {
		t.Fatalf("set input: %v", err)
	}
	if err := q.SetOutputLogFile(outputFile); err != nil {
		t.Fatalf("set output: %v", err)
	}

	if err := q.StartReadingLogFile(src); err != nil {
		t.Fatalf("start reading: %v", err)
	}

	select {
	case <-q.GetQuit():
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for quit")
	}

	inB, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("read in: %v", err)
	}
	outB, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("read out: %v", err)
	}

	countNonEmpty := func(s string) int {
		n := 0
		for _, ln := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
			if strings.TrimSpace(ln) != "" {
				n++
			}
		}
		return n
	}

	inCount := countNonEmpty(string(inB))
	outCount := countNonEmpty(string(outB))

	if inCount != outCount {
		t.Fatalf("expected 1:1 lines, got input=%d output=%d", inCount, outCount)
	}
}

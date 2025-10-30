package queue

import (
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
    os.RemoveAll(dir)
    os.Exit(code)
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

	err = q.Enqueue(`{"name": "test1"}`)
	if err != nil {
		t.Fatalf("Failed to enqueue: %v", err)
	}

	err = q.Enqueue(`{"name": "test2"}`)
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
	if !strings.Contains(output, `{"name": "test1"} (processed)`) {
		t.Error("Expected '{\"name\": \"test1\"} (processed)' in output file")
	}

	if !strings.Contains(output, `{"name": "test2"} (processed)`) {
		t.Error("Expected '{\"name\": \"test2\"} (processed)' in output file")
	}
}

func TestQueue_ReadFromFile(t *testing.T) {
	testFile := "test_input_file.log"
	defer os.Remove(testFile)

	content := `{"name": "line1"}
{"name": "line2"}
{"name": "line3"}
`
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
	expectedLines := []string{`{"name": "line1"} (processed)`, `{"name": "line2"} (processed)`, `{"name": "line3"} (processed)`}

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
			jsonStr := `{"id": ` + string(rune('0' + i)) + `, "name": "test"}`
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
	actualCount := strings.Count(output, " (processed)")
	if actualCount != expectedCount {
		t.Errorf("Expected %d processed items, got %d", expectedCount, actualCount)
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

	content := `{"name": "line1"}

{"name": "line2"}



{"name": "line3"}
`
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
	expectedLines := []string{`{"name": "line1"} (processed)`, `{"name": "line2"} (processed)`, `{"name": "line3"} (processed)`}

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

	validJSON := `{"name": "test", "value": 123}`
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
	expected := validJSON + " (processed)"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected '%s' in output file, got '%s'", expected, output)
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

	invalidJSON := `{"name": "test", "value": 123` // missing closing brace
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
	expected := invalidJSON + " (rejected - invalid JSON)"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected '%s' in output file, got '%s'", expected, output)
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

	validJSON := `{"name": "test", "value": 123}`
	invalidJSON := `{"name": "test", "value": 123`

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
	
	expectedValid := validJSON + " (processed)"
	expectedInvalid := invalidJSON + " (rejected - invalid JSON)"
	
	if !strings.Contains(output, expectedValid) {
		t.Errorf("Expected '%s' in output file", expectedValid)
	}
	
	if !strings.Contains(output, expectedInvalid) {
		t.Errorf("Expected '%s' in output file", expectedInvalid)
	}
}

func TestQueue_JSONValidation_FromFile(t *testing.T) {
	testFile := "test_json_file.log"
	defer os.Remove(testFile)

	content := `{"name": "test1", "value": 123}
{"name": "test2", "value": 456
{"name": "test3", "value": 789}`
	
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
		`{"name": "test1", "value": 123} (processed)`,
		`{"name": "test2", "value": 456 (rejected - invalid JSON)`,
		`{"name": "test3", "value": 789} (processed)`,
	}

	for _, expected := range expectedLines {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected '%s' in output file", expected)
		}
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

	nonJSON := "this is not json at all"
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
	expected := nonJSON + " (rejected - invalid JSON)"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected '%s' in output file, got '%s'", expected, output)
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

	content := `{"name": "valid1"}

{"name": "valid2"}

   `

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
		`{"name": "valid1"} (processed)`,
		`{"name": "valid2"} (processed)`,
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

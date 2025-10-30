package queue

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Mode int

const (
	EngineMode Mode = iota
	ReplayMode
)

type Queue struct {
	mode             Mode
	input            chan string
	output           chan string
	done             chan struct{}
	quit             chan struct{}
	inputFile        *os.File
	outputFile       *os.File
	inputFileLatest  *os.File
	outputFileLatest *os.File
}

func NewEngineQueue() *Queue {
	q := &Queue{
		mode:   EngineMode,
		input:  make(chan string),
		output: make(chan string, 100),
		done:   make(chan struct{}),
		quit:   make(chan struct{}),
	}

	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	timestamp := now.Format("150405")

	baseDir := os.Getenv("SIMULATIONS_DIR")
	if baseDir == "" {
		baseDir = "simulations"
	}
	simDir := filepath.Join(baseDir, year, month, day)
	os.MkdirAll(simDir, 0755)

	inputFile := filepath.Join(simDir, fmt.Sprintf("input_%s.log", timestamp))
	outputFile := filepath.Join(simDir, fmt.Sprintf("output_%s.log", timestamp))

	q.SetInputLogFile(inputFile)
	q.SetOutputLogFile(outputFile)

	latestInput := filepath.Join(simDir, "input.log")
	latestOutput := filepath.Join(simDir, "output.log")
	os.Remove(latestInput)
	os.Remove(latestOutput)
	lfIn, _ := os.OpenFile(latestInput, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	lfOut, _ := os.OpenFile(latestOutput, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	q.inputFileLatest = lfIn
	q.outputFileLatest = lfOut

	go q.process()

	return q
}

func NewReplayQueue() *Queue {
	q := &Queue{
		mode:   ReplayMode,
		input:  make(chan string),
		output: make(chan string, 100),
		done:   make(chan struct{}),
		quit:   make(chan struct{}),
	}

	go q.process()

	return q
}

func (q *Queue) SetInputLogFile(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	if q.inputFile != nil {
		q.inputFile.Close()
	}
	q.inputFile = file

	return nil
}

func (q *Queue) SetOutputLogFile(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	if q.outputFile != nil {
		q.outputFile.Close()
	}
	q.outputFile = file

	return nil
}

func (q *Queue) process() {
	defer close(q.output)
	defer func() {
		if q.inputFile != nil {
			q.inputFile.Close()
		}
		if q.outputFile != nil {
			q.outputFile.Close()
		}
		if q.inputFileLatest != nil {
			q.inputFileLatest.Close()
		}
		if q.outputFileLatest != nil {
			q.outputFileLatest.Close()
		}
	}()

	for {
		select {
		case item := <-q.input:
			if strings.TrimSpace(item) == "" {
				continue
			}

			if q.isValidJSON(item) {
				processedItem := item + " (processed)"
				q.logToOutputFile(processedItem)
				select {
				case q.output <- item:
				case <-q.done:
					return
				}
			} else {
				rejectedItem := item + " (rejected - invalid JSON)"
				q.logToOutputFile(rejectedItem)
			}
		case <-q.done:
			return
		}
	}
}

func (q *Queue) logToInputFile(item string) {
	if q.inputFile != nil {
		fmt.Fprintln(q.inputFile, item)
	}
	if q.inputFileLatest != nil {
		fmt.Fprintln(q.inputFileLatest, item)
	}
}

func (q *Queue) logToOutputFile(item string) {
	if q.outputFile != nil {
		fmt.Fprintln(q.outputFile, item)
	}
	if q.outputFileLatest != nil {
		fmt.Fprintln(q.outputFileLatest, item)
	}
}

func (q *Queue) isValidJSON(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

func (q *Queue) Enqueue(item string) error {
	if q.mode != EngineMode {
		return errors.New("Enqueue is only available in Engine mode")
	}
	if strings.TrimSpace(item) != "" {
		q.logToInputFile(item)
		q.input <- item
	}
	return nil
}

func (q *Queue) StartReadingLogFile(filePath string) error {
	if q.mode != ReplayMode {
		return errors.New("StartReadingLogFile cannot be called in Engine mode")
	}

    resolved := filePath
    if fi, err := os.Lstat(filePath); err == nil && (fi.Mode()&os.ModeSymlink) != 0 {
        if p, e := filepath.EvalSymlinks(filePath); e == nil {
            resolved = p
        }
    }

    simDir := filepath.Dir(resolved)
    base := filepath.Base(resolved)
    ts := ""
    if strings.HasPrefix(base, "input_") && strings.HasSuffix(base, ".log") {
        ts = strings.TrimSuffix(strings.TrimPrefix(base, "input_"), ".log")
    } else if base == "input.log" {
        // Derive timestamp from the latest input_*.log in the same directory
        entries, err := os.ReadDir(simDir)
        if err == nil {
            latest := ""
            for _, e := range entries {
                name := e.Name()
                if strings.HasPrefix(name, "input_") && strings.HasSuffix(name, ".log") {
                    if name > latest { // lexicographic compare works for HHMMSS
                        latest = name
                    }
                }
            }
            if latest != "" {
                ts = strings.TrimSuffix(strings.TrimPrefix(latest, "input_"), ".log")
            }
        }
        if ts == "" {
            ts = time.Now().Format("150405")
        }
    } else {
        ts = time.Now().Format("150405")
    }

	debugDir := filepath.Join(simDir, "debug")
	os.MkdirAll(debugDir, 0755)
	q.SetInputLogFile(filepath.Join(debugDir, "input_debug_"+ts+".log"))
	q.SetOutputLogFile(filepath.Join(debugDir, "output_debug_"+ts+".log"))

	file, err := os.Open(resolved)
	if err != nil {
		return err
	}

	go q.readLogFile(file)

	return nil
}

func (q *Queue) readLogFile(file *os.File) {
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for {
		select {
		case <-q.done:
			return
		default:
			if scanner.Scan() {
				line := scanner.Text()
				if strings.TrimSpace(line) != "" {
					q.logToInputFile(line)
					q.input <- line
				}
			} else {
				q.waitForQueueEmpty()
				close(q.quit)
				return
			}
		}
	}
}

func (q *Queue) waitForQueueEmpty() {
	for {
		select {
		case <-q.done:
			return
		default:
			if len(q.input) == 0 {
				return
			}
		}
	}
}

func (q *Queue) GetOutput() <-chan string {
	return q.output
}

func (q *Queue) Stop() {
	close(q.done)
}

func (q *Queue) GetQuit() <-chan struct{} {
	return q.quit
}

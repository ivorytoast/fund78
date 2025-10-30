# Instructions for AI Coding Agents

This document provides specific guidelines for AI coding assistants (such as Cursor, GitHub Copilot, etc.) working on the `fund78` project.

## Project Setup

- **Language:** Go (Golang)
- **Module:** fund78
- **Go Version:** 1.21+
- **Entry Point:** `main.go` (currently)

## Code Generation Guidelines

### When Creating New Code

1. **Follow Go Idioms:**
   - Use interfaces for abstractions
   - Prefer composition over inheritance
   - Use channels and goroutines for concurrency
   - Return errors explicitly (no exceptions)

2. **Project Structure:**
   - If adding new functionality, consider where it fits:
     - `cmd/` - Application entry points
     - `internal/` - Private application code
     - `pkg/` - Public, reusable packages
     - Test files alongside source files

3. **Error Handling Pattern:**
   ```go
   result, err := doSomething()
   if err != nil {
       return fmt.Errorf("failed to do something: %w", err)
   }
   ```

4. **Testing Pattern:**
   - Create test files with `_test.go` suffix
   - Use table-driven tests for multiple scenarios
   - Example:
   ```go
   func TestFunction(t *testing.T) {
       tests := []struct {
           name     string
           input    string
           expected string
           wantErr  bool
       }{
       }
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
           })
       }
   }
   ```

## Code Quality Checklist

Before completing any task, ensure:
- [ ] Code is formatted with `gofmt`
- [ ] No linter errors (`go vet`, `golangci-lint`)
- [ ] Tests are written for new functionality
- [ ] **NO comments added to code - code only, no comments**
- [ ] Error handling is implemented
- [ ] No hardcoded secrets or sensitive data
- [ ] Context is used for cancellation/timeouts (if applicable)

## Common Patterns to Use

### Configuration
- Use environment variables or config files
- Avoid hardcoded values
- Example:
  ```go
  port := os.Getenv("PORT")
  if port == "" {
      port = "8080"
  }
  ```

### Logging
- Use structured logging when adding logging
- Example:
  ```go
  log.Printf("Processing request: %s", requestID)
  ```

### Context Usage
```go
func processWithContext(ctx context.Context, data string) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
}
```

## What NOT to Do

- ❌ **NEVER add comments to code - write code only, no comments**
- ❌ Don't ignore errors (no `_` for error returns unless intentional)
- ❌ Don't use `panic()` for normal error conditions
- ❌ Don't commit secrets, API keys, or sensitive data
- ❌ Don't create overly complex abstractions
- ❌ Don't add dependencies without justification
- ❌ Don't break existing functionality without updating tests

## Communication Style

When suggesting changes or implementations:
- Explain why the approach was chosen
- Mention alternatives considered
- Point out potential issues or limitations
- Suggest improvements if applicable

## Questions to Consider

Before implementing a feature:
1. Does similar functionality already exist?
2. What's the simplest solution that meets requirements?
3. Are there edge cases to handle?
4. Should this be configurable or hardcoded?
5. Does this need to be tested? (Answer: almost always yes)

## Quick Reference

- **Format:** `go fmt ./...`
- **Test:** `go test ./...`
- **Build:** `go build`
- **Dependencies:** `go mod tidy`
- **Lint:** `go vet ./...`

## Context-Aware Decisions

- If the project is in early stages: prioritize simplicity and clarity
- If adding to existing code: match existing patterns and style
- If fixing bugs: add tests to prevent regression
- If refactoring: ensure all tests pass before and after

## Updates to This Document

Update this document when:
- Project structure changes significantly
- New coding standards are established
- Common patterns emerge that should be documented
- Technology stack changes


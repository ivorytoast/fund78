# Project Requirements and Future Plan

## Project Overview
**Project Name:** fund78  
**Language:** Go (Golang)  
**Status:** Core Queue System Implemented

## Current State
The project implements a string queue system with two modes:
- **Engine Mode**: Processes strings via `Enqueue()` method
- **Replay Mode**: Processes strings from log files

The queue logs incoming events to `input.log` and processed events to `output.log`.

Logging/storage behavior:
- Engine Mode
  - Writes timestamped files per run under `simulations/YYYY/MM/DD/`:
    - `input_HHMMSS.log`, `output_HHMMSS.log`
  - Also writes latest real files (not symlinks) in the same folder:
    - `input.log`, `output.log` (truncated and rewritten each run)
  - Base folder can be overridden with env var `SIMULATIONS_DIR` (defaults to `simulations`).
- Replay Mode
  - Automatically finds the latest simulation folder (`YYYY/MM/DD`) and uses its `input.log`.
  - Creates per-run debug logs in `simulations/YYYY/MM/DD/debug/` with matching timestamp to the source input file:
    - `input_debug_HHMMSS.log`, `output_debug_HHMMSS.log`
  - If starting from `input.log`, the timestamp is derived from the latest `input_*.log` in the same directory.

## Application Requirements

### Phase 1: Core Foundation
- [x] Basic Go project setup
- [x] Project structure definition
- [x] Queue system implementation
- [x] Logging setup (input/output files)
- [x] Error handling patterns
- [x] Mode-based queue operation (Engine/Replay)

### Phase 2: Development Environment
- [x] Development dependencies setup
- [ ] Testing framework integration
- [ ] Code quality tools (golangci-lint, gofmt)
- [ ] Build and deployment scripts
- [x] Documentation standards

### Phase 3: Core Features
- [x] String Queue System
  - Engine Mode: Programmatic input via `Enqueue()`
  - Replay Mode: File-based input via `StartReadingLogFile()`
  - Engine logging to timestamped files + latest real files
  - Replay debug logging per-run with matching timestamps
  - Graceful shutdown when processing complete
- [ ] Queue persistence
- [ ] Queue monitoring/metrics
- [ ] Multiple queue support

## Technical Stack

### Required
- **Go Version:** 1.21 or later
- **Module:** fund78

### Optional (To be determined)
- Database (if needed)
- Web framework (if needed)
- External APIs (if needed)
- Message queue (if needed)

## Project Structure (Current)

```
fund78/
├── main.go
├── internal/
│   └── queue/
│       └── queue.go
├── simulations/
│   └── YYYY/MM/DD/
│       ├── input_HHMMSS.log
│       ├── output_HHMMSS.log
│       ├── input.log
│       ├── output.log
│       └── debug/
│           ├── input_debug_HHMMSS.log
│           └── output_debug_HHMMSS.log
├── docs/
│   ├── REQUIREMENTS.md
│   ├── AGENT_INSTRUCTIONS.md
│   └── .cursorrules
├── go.mod
└── go.sum
```

## Development Guidelines

### Code Standards
- Follow Go best practices and idioms
- Use `gofmt` for code formatting
- Maintain test coverage above 80%
- Write clear, self-documenting code
- Add comments for exported functions and types

### Git Workflow
- Use meaningful commit messages
- Create feature branches for new functionality
- Review code before merging to main

## Future Roadmap

### Short Term (Next 1-2 months)
- [x] Define specific application requirements
- [x] Set up project structure
- [x] Implement core queue functionality
- [x] Add comprehensive test suite
- [ ] Add configuration management
- [ ] Add CLI interface

### Medium Term (3-6 months)
- Add advanced features
- Performance optimization
- Integration with external services

### Long Term (6+ months)
- Scalability improvements
- Monitoring and observability
- Production deployment strategy

## Notes
- This document should be updated as requirements are clarified
- Regular review and updates are recommended
- Keep the document aligned with actual project progress
 - Tests create logs in a temporary `SIMULATIONS_DIR` and clean them up after running


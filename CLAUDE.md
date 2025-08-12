# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a native Elgato StreamDeck plugin for controlling BlackmagicDesign ATEM switchers (primarily ATEM Mini series). The plugin is built in Go and provides direct ATEM control without requiring additional software like Companion.

## Development Commands

The project uses Task for build automation. Key commands:

- `task setup` - Download Go modules and prepare development environment
- `task test` - Run all tests with verbose output (`go test -v ./...`)
- `task vet` - Run Go static analysis (`go vet ./...`)
- `task build` - Build the complete plugin for all platforms
- `task build-server` - Build only the Go backend for Windows and macOS
- `task clean` - Remove build artifacts and prepare fresh build directories
- `task release` - Create release distribution

All Go development happens in the `Source/code/` directory.

## Architecture

### Core Components

- **main.go** - Entry point that initializes StreamDeck client, logger, and runs the main application
- **stdatem package** - Main application logic containing the App struct that manages ATEM connections and StreamDeck interactions
- **connectionmanager package** - Manages multiple ATEM device connections with automatic reconnection logic
- **di package** - Dependency injection setup for StreamDeck client and logger initialization
- **logger package** - Multi-logger implementation supporting both file and StreamDeck logging
- **setting package** - Generic setting store for managing action configurations

### ATEM Integration

The application connects to ATEM switchers using the `github.com/FlowingSPDG/go-atem` library. Key features:

- **Multi-host support** - Can connect to multiple ATEM devices simultaneously
- **Automatic reconnection** - Handles connection drops with exponential backoff
- **Real-time tally feedback** - Updates StreamDeck button states based on ATEM preview/program changes
- **Event-driven architecture** - Uses ATEM event callbacks for state synchronization

### StreamDeck Actions

Four main actions are supported:
1. **Preview** (`dev.flowingspdg.atem.preview`) - Set preview input
2. **Program** (`dev.flowingspdg.atem.program`) - Set program input  
3. **Cut** (`dev.flowingspdg.atem.cut`) - Perform cut transition
4. **Auto** (`dev.flowingspdg.atem.auto`) - Perform auto transition

Each action has corresponding property inspector HTML files in `Source/pi/`.

### Connection Management

The `ConnectionManager` maintains several concurrent-safe maps:
- `atemByIP` - Maps IP addresses to ATEM instances
- `atemByContext` - Maps StreamDeck contexts to ATEM instances
- `contextsByIP` - Maps IP addresses to associated StreamDeck contexts
- `usageCounts` - Tracks reference counts for connection cleanup

### Plugin Structure

This is an Elgato StreamDeck plugin with the standard structure:
- `manifest.json` - Plugin metadata and action definitions
- `images/` - Plugin icons and assets
- `inspector/` - HTML property inspector files for configuration
- Binary executables (`atem_go.exe` for Windows, `atem_go` for macOS)

## Key Patterns

- **Context-based operations** - All ATEM operations are tied to specific StreamDeck button contexts
- **Concurrent connection handling** - Each ATEM host gets its own reconnection goroutine
- **Event-driven state management** - ATEM state changes trigger StreamDeck UI updates
- **Reference counting** - Connections are shared across multiple contexts and cleaned up when no longer needed

## Testing and Code Quality

Run `task test` to execute the test suite and `task vet` for static analysis before committing changes.
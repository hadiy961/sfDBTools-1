# sfDBTools - AI Coding Agent Instructions

## Project Overview
sfDBTools is a MariaDB/MySQL database backup and management utility written in Go. It follows a modular architecture with dependency injection and configuration-driven design.

## Architecture & Key Patterns

### Dependency Injection System
- **Global Dependencies**: All services use `pkg/globals/global_deps.go` for dependency injection
- **Bootstrap Flow**: `main.go` → loads config → creates logger → injects into `globals.Deps` → passes to commands
- **Access Pattern**: Use `globals.GetLogger()` and `globals.GetConfig()` throughout the codebase
- **Service Pattern**: Each module has a `Service` struct that receives injected dependencies (see `internal/backup/backup_main.go`)

### Command Structure (Cobra CLI)
- **Root Command**: `cmd/cmd_root.go` with `PersistentPreRunE` that validates dependencies
- **Sub-commands**: Organized in `cmd/*/` directories (e.g., `backup_cmd/`, `dbconfig_cmd/`)
- **Registration**: Each command registers itself in `init()` functions
- **Validation**: Commands validate dependencies before execution in `PersistentPreRunE`

### Configuration System
- **YAML-based**: Primary config in `config/config.yaml` with extensive nesting
- **Environment Variables**: Support via struct tags (e.g., `env:"SFDB_DB_HOST"`)
- **Default Values**: Handled via struct tags and `internal/default_value/` packages
- **Loading**: Configuration loaded once at startup via `appconfig.LoadConfigFromEnv()`

### Flag Management Pattern
- **Dynamic Flags**: Use reflection-based system in `pkg/flag/dynamic_flag.go`
- **Struct Tags**: Flags defined via struct tags: `flag:"host" env:"SFDB_DB_HOST" default:"localhost"`
- **Auto-registration**: `DynamicAddFlags(cmd, &structInstance)` automatically registers flags from struct fields
- **Struct Definitions**: Connection and option structs in `internal/structs/`

### File Organization
```
cmd/           # Cobra commands organized by feature
internal/      # Private application code
├── appconfig/ # Configuration loading and structs  
├── applog/    # Custom logging wrapper
├── backup/    # Core backup business logic
├── dbconfig/  # Database configuration management
├── structs/   # Shared data structures
└── default_value/ # Default value providers
pkg/           # Reusable packages
├── common/    # Common utilities and constants
├── compress/  # Compression handling
├── database/  # Database operations
├── encrypt/   # Encryption utilities
├── flag/      # Dynamic flag management
├── fs/        # File system operations
├── globals/   # Global dependency injection
├── input/     # User input handling
├── parsing/   # Command parsing utilities
└── ui/        # Terminal UI components
```

## Development Conventions

### File Headers & Comments
- **Required Header**: Every file must have a comment header with File, Deskripsi, Author, Tanggal, Last Modified
- **Indonesian Language**: Comments and descriptions primarily in Indonesian
- **Go Doc Format**: Use standard Go documentation conventions for exported functions

### Error Handling
- **Config Loading**: Fatal errors logged to stderr and exit with code 1
- **Service Level**: Services return errors that are handled by command layer
- **Logging**: Use injected logger for all error logging, not direct fmt.Printf

### Struct Design
- **Connection Structs**: Database connections follow `ServerDBConnection` pattern in `internal/structs/`
- **Flag Structs**: Command options organized in dedicated structs with tags
- **Service Structs**: Each module has a `Service` struct holding dependencies and state

### Logging
- **Custom Logger**: Wrapper around logrus in `internal/applog/`
- **Structured Logging**: Use logger fields for structured data
- **Access Pattern**: Always use `globals.GetLogger()` rather than direct imports

## Build & Development

### Module System
- **Module Name**: `sfDBTools` (note: not matching directory name)
- **Go Version**: 1.25.0
- **Key Dependencies**: cobra (CLI), logrus (logging), survey (prompts), tablewriter (output)

### Building
```bash
go build -o sfdbtools ./main.go
```

### Configuration
- **Required Files**: `config/config.yaml` and `config/db_list.txt`
- **Environment**: Supports environment variable overrides for all config values
- **Validation**: Configuration validated at startup

## Integration Points

### Database Operations
- **Connection Management**: Centralized in `pkg/database/`
- **MySQL/MariaDB**: Primary target databases
- **Backup Process**: Uses mysqldump with piped compression and encryption

### File System
- **Output Management**: Configurable directory structures with template patterns
- **Compression**: Pluggable compression backends (gzip, zstd, lz4)
- **Encryption**: AES encryption for backup files

### External Tools
- **mysqldump**: Core dependency for database backups
- **Compression Tools**: Various compression utilities depending on configuration

## Common Patterns

### Service Initialization
```go
svc := &Service{
    Logger: logger,
    Config: cfg,
}
```

### Command Registration
```go
func init() {
    rootCmd.AddCommand(SomeCmd)
    flags.DynamicAddFlags(SomeCmd, &defaultFlags)
}
```

### Configuration Access
```go
cfg := globals.GetConfig()
logger := globals.GetLogger()
```

### Flag Definition
```go
type SomeFlags struct {
    Host string `flag:"host" env:"SFDB_DB_HOST" default:"localhost"`
}
```

When modifying this codebase, maintain the dependency injection pattern, use the dynamic flag system for new commands, and ensure all services receive dependencies through the established patterns.
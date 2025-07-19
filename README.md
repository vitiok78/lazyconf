# lazyconf

[![Go Reference](https://pkg.go.dev/badge/github.com/vitiok78/lazyconf.svg)](https://pkg.go.dev/github.com/vitiok78/lazyconf)
[![Go Report Card](https://goreportcard.com/badge/github.com/vitiok78/lazyconf)](https://goreportcard.com/report/github.com/vitiok78/lazyconf)

A lightweight, zero-dependency Go library for parsing environment variables into Go structs using struct tags. Simple, fast, and feature-rich.

## Features

- **Zero dependencies** - Uses only Go standard library
- **Comprehensive type support** - All basic Go types, slices, time types, and complex numbers
- **Nested structs** - Recursive parsing of embedded structs
- **Custom parsing** - Setter interface and UnmarshalText/JSON support
- **Flexible tags** - Required fields, default values, custom setters, and parser options
- **Detailed errors** - Clear error messages with context

## Installation

```bash
go get github.com/vitiok78/lazyconf
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/vitiok78/lazyconf"
)

type Config struct {
    Port        int           `env:"PORT,default=8080"`
    Host        string        `env:"HOST,default=localhost"`
    Debug       bool          `env:"DEBUG"`
    Timeout     time.Duration `env:"TIMEOUT,default=30s"`
    DatabaseURL string        `env:"DATABASE_URL,required"`
}

func main() {
    // Set environment variables
    os.Setenv("DEBUG", "true")
    os.Setenv("DATABASE_URL", "postgres://localhost/mydb")
    
    var cfg Config
    if err := lazyconf.ParseEnv(&cfg); err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Config: %+v\n", cfg)
    // Output: Config: {Port:8080 Host:localhost Debug:true Timeout:30s DatabaseURL:postgres://localhost/mydb}
}
```

## Supported Types

### Basic Types
```go
type Config struct {
    StringField  string     `env:"STRING_VAL"`
    IntField     int        `env:"INT_VAL"`
    Int64Field   int64      `env:"INT64_VAL"`
    UintField    uint       `env:"UINT_VAL"`
    Float64Field float64    `env:"FLOAT_VAL"`
    BoolField    bool       `env:"BOOL_VAL"`
    ComplexField complex128 `env:"COMPLEX_VAL"`
}
```

**Environment Variables Setup:**
```bash
export STRING_VAL="hello world"
export INT_VAL="42"
export INT64_VAL="9223372036854775807"
export UINT_VAL="123"
export FLOAT_VAL="3.14159"
export BOOL_VAL="true"
export COMPLEX_VAL="1+2i"
```

### Time Types
```go
type Config struct {
    // RFC3339 format: "2006-01-02T15:04:05Z07:00"
    CreatedAt time.Time     `env:"CREATED_AT"`
    
    // Duration strings: "5s", "10m", "1h30m"
    Timeout   time.Duration `env:"TIMEOUT"`
}
```

**Environment Variables Setup:**
```bash
export CREATED_AT="2023-12-25T15:30:45Z"
export TIMEOUT="5m30s"
```

### Slices (Comma-separated values)
```go
type Config struct {
    Ports     []int           `env:"PORTS"`           // "8080,8081,8082"
    Hosts     []string        `env:"HOSTS"`           // "host1,host2,host3"
    Timeouts  []time.Duration `env:"TIMEOUTS"`        // "5s,10s,15s"
    Times     []time.Time     `env:"TIMES"`           // "2023-01-01T00:00:00Z,2023-01-02T00:00:00Z"
}
```

**Environment Variables Setup:**
```bash
export PORTS="8080,8081,8082"
export HOSTS="host1,host2,host3"
export TIMEOUTS="5s,10s,15s"
export TIMES="2023-01-01T00:00:00Z,2023-01-02T00:00:00Z"
```

### Nested Structs
```go
type DatabaseConfig struct {
    Host     string `env:"DB_HOST,default=localhost"`
    Port     int    `env:"DB_PORT,default=5432"`
    Username string `env:"DB_USER,required"`
    Password string `env:"DB_PASS,required"`
}

type Config struct {
    Database DatabaseConfig // Automatically parsed recursively
    AppName  string         `env:"APP_NAME,default=myapp"`
}
```

**Environment Variables Setup:**
```bash
export DB_HOST="database.example.com"
export DB_PORT="3306"
export DB_USER="admin"
export DB_PASS="secret123"
export APP_NAME="my-application"
```

## Tag Options

### Required Fields
```go
type Config struct {
    APIKey string `env:"API_KEY,required"`
}
```

**Environment Variables Setup:**
```bash
export API_KEY="your-secret-api-key-here"
```

### Default Values
```go
type Config struct {
    Port    int    `env:"PORT,default=8080"`
    Host    string `env:"HOST,default=localhost"`
    Debug   bool   `env:"DEBUG,default=false"`
    Timeout string `env:"TIMEOUT,default=30s"`
}
```

**Environment Variables Setup (optional - defaults will be used if not set):**
```bash
# Override defaults by setting environment variables
export PORT="3000"
export HOST="0.0.0.0"
export DEBUG="true"
export TIMEOUT="60s"

# Or leave unset to use defaults:
# PORT=8080, HOST=localhost, DEBUG=false, TIMEOUT=30s
```

### Custom Setters
```go
type Config struct {
    CustomField string `env:"CUSTOM_VAL,setter=SetCustomField"`
}

func (c *Config) SetCustomField(val string) error {
    // Custom parsing logic
    c.CustomField = "processed:" + val
    return nil
}
```

**Environment Variables Setup:**
```bash
export CUSTOM_VAL="raw-value"
# This will be processed by SetCustomField and result in: "processed:raw-value"
```

### Parser Options
```go
type Config struct {
    // Force text unmarshaling
    TextField CustomType `env:"TEXT_VAL,parser=text"`
    
    // Force JSON unmarshaling  
    JSONField CustomType `env:"JSON_VAL,parser=json"`
}
```

**Environment Variables Setup:**
```bash
export TEXT_VAL="plain text value"
export JSON_VAL='{"key":"value","number":42}'
```

## Custom Types

### Setter Interface
```go
type Status int

const (
    StatusActive Status = iota
    StatusInactive
)

func (s *Status) Scan(value interface{}) error {
    str := value.(string)
    switch str {
    case "active":
        *s = StatusActive
    case "inactive":
        *s = StatusInactive
    default:
        return fmt.Errorf("invalid status: %s", str)
    }
    return nil
}

type Config struct {
    Status Status `env:"STATUS"`
}
```

**Environment Variables Setup:**
```bash
export STATUS="active"
# or
export STATUS="inactive"
```

### UnmarshalText Interface
```go
type CustomID struct {
    Value string
}

func (c *CustomID) UnmarshalText(text []byte) error {
    c.Value = "id:" + string(text)
    return nil
}

type Config struct {
    ID CustomID `env:"CUSTOM_ID"`
}
```

**Environment Variables Setup:**
```bash
export CUSTOM_ID="12345"
# This will be processed by UnmarshalText and result in: Value="id:12345"
```

### UnmarshalJSON Interface
```go
type JSONConfig struct {
    Data map[string]interface{}
}

func (j *JSONConfig) UnmarshalJSON(data []byte) error {
    return json.Unmarshal(data, &j.Data)
}

type Config struct {
    Settings JSONConfig `env:"SETTINGS"` // Set as: SETTINGS='{"key":"value"}'
}
```

**Environment Variables Setup:**
```bash
export SETTINGS='{"database":{"host":"localhost","port":5432},"features":{"debug":true,"cache":false}}'
# This will be processed by UnmarshalJSON and populate the Data map
```

## Advanced Examples

### Complete Application Configuration
```go
type Config struct {
    // Server settings
    Server struct {
        Host         string        `env:"SERVER_HOST,default=0.0.0.0"`
        Port         int           `env:"SERVER_PORT,default=8080"`
        ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT,default=30s"`
        WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT,default=30s"`
    }
    
    // Database settings
    Database struct {
        URL             string `env:"DATABASE_URL,required"`
        MaxConnections  int    `env:"DB_MAX_CONNECTIONS,default=10"`
        ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME,default=1h"`
    }
    
    // Feature flags
    Features struct {
        EnableMetrics bool `env:"ENABLE_METRICS,default=true"`
        EnableTracing bool `env:"ENABLE_TRACING,default=false"`
        EnableDebug   bool `env:"ENABLE_DEBUG,default=false"`
    }
    
    // External services
    Services struct {
        RedisURL    string   `env:"REDIS_URL,required"`
        AllowedIPs  []string `env:"ALLOWED_IPS"` // "192.168.1.1,10.0.0.1"
        Timeouts    []time.Duration `env:"SERVICE_TIMEOUTS,default=5s,10s,15s"`
    }
}
```

### Environment Variable Setup
```bash
# Server
export SERVER_HOST="localhost"
export SERVER_PORT="3000"
export SERVER_READ_TIMEOUT="45s"

# Database
export DATABASE_URL="postgres://user:pass@localhost/db"
export DB_MAX_CONNECTIONS="20"

# Features
export ENABLE_METRICS="true"
export ENABLE_DEBUG="true"

# Services
export REDIS_URL="redis://localhost:6379"
export ALLOWED_IPS="192.168.1.1,10.0.0.1,172.16.0.1"
export SERVICE_TIMEOUTS="3s,7s,12s"
```

## Error Handling

lazyconf provides detailed error messages:

```go
var cfg Config
if err := lazyconf.ParseEnv(&cfg); err != nil {
    // Errors include context about which field and environment variable failed
    log.Printf("Configuration error: %v", err)
}
```

Common error types:
- Missing required environment variables
- Invalid type conversions
- Unsupported field types
- Custom setter/unmarshaler failures

## API Reference

### ParseEnv
```go
func ParseEnv(cfg any) error
```
Parses environment variables into the provided struct pointer.

**Parameters:**
- `cfg`: Pointer to struct to populate

**Returns:**
- `error`: nil on success, detailed error on failure

### Setter Interface
```go
type Setter interface {
    Scan(value interface{}) error
}
```
Implement this interface for custom field parsing.

## Best Practices

1. **Use meaningful environment variable names**
   ```go
   DatabaseURL string `env:"DATABASE_URL,required"`  // Good
   URL string `env:"URL,required"`                    // Avoid
   ```

2. **Provide sensible defaults**
   ```go
   Port int `env:"PORT,default=8080"`
   ```

3. **Group related configuration**
   ```go
   type Config struct {
       Database struct {
           Host string `env:"DB_HOST"`
           Port int    `env:"DB_PORT"`
       }
   }
   ```

4. **Use required for critical settings**
   ```go
   APIKey string `env:"API_KEY,required"`
   ```

5. **Validate after parsing**
   ```go
   if err := lazyconf.ParseEnv(&cfg); err != nil {
       return err
   }
   if cfg.Port < 1 || cfg.Port > 65535 {
       return errors.New("invalid port range")
   }
   ```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

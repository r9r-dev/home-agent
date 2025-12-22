# Extract configuration with validation

**Priority:** P3 (Low)
**Type:** Enhancement
**Component:** Backend
**Estimated Effort:** Low

## Summary

Replace scattered environment variable reads with structured configuration loading and validation.

## Current State

Environment variables are read directly throughout `main.go` without validation:

```go
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}

proxyURL := os.Getenv("CLAUDE_PROXY_URL")
// No validation that it's a valid URL
```

## Proposed Solution

```go
// config/config.go
package config

import (
    "errors"
    "net/url"

    "github.com/kelseyhightower/envconfig"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Proxy    ProxyConfig
    Upload   UploadConfig
}

type ServerConfig struct {
    Port      int    `envconfig:"PORT" default:"8080"`
    PublicDir string `envconfig:"PUBLIC_DIR" default:"./public"`
}

type DatabaseConfig struct {
    Path string `envconfig:"DATABASE_PATH" default:"./data/homeagent.db"`
}

type ProxyConfig struct {
    URL    string `envconfig:"CLAUDE_PROXY_URL" required:"true"`
    APIKey string `envconfig:"CLAUDE_PROXY_KEY"`
}

type UploadConfig struct {
    Dir         string `envconfig:"UPLOAD_DIR" default:"./data/uploads"`
    MaxSizeMB   int    `envconfig:"MAX_UPLOAD_SIZE_MB" default:"10"`
    WorkspacePath string `envconfig:"WORKSPACE_PATH"`
}

func Load() (*Config, error) {
    cfg := &Config{}
    if err := envconfig.Process("", cfg); err != nil {
        return nil, fmt.Errorf("load config: %w", err)
    }
    return cfg, cfg.Validate()
}

func (c *Config) Validate() error {
    if c.Proxy.URL == "" {
        return errors.New("CLAUDE_PROXY_URL is required")
    }

    if _, err := url.Parse(c.Proxy.URL); err != nil {
        return fmt.Errorf("invalid CLAUDE_PROXY_URL: %w", err)
    }

    if c.Server.Port < 1 || c.Server.Port > 65535 {
        return errors.New("PORT must be between 1 and 65535")
    }

    return nil
}
```

## Usage in main.go

```go
func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Configuration error: %v", err)
    }

    db, err := models.InitDB(cfg.Database.Path)
    // ...

    app := fiber.New()
    // ...

    log.Fatal(app.Listen(fmt.Sprintf(":%d", cfg.Server.Port)))
}
```

## Tasks

- [ ] Add envconfig dependency
- [ ] Create `config/` package
- [ ] Define configuration structs
- [ ] Implement `Load()` function
- [ ] Implement `Validate()` function
- [ ] Update `main.go` to use config
- [ ] Add configuration documentation
- [ ] Create example `.env.example` file

## Acceptance Criteria

- [ ] All configuration in one place
- [ ] Required fields validated at startup
- [ ] Clear error messages for missing/invalid config
- [ ] Defaults documented in struct tags
- [ ] Example `.env` file provided

## References

- `ARCHITECTURE_REVIEW.md` section "7. Extract Configuration with Validation"
- envconfig: https://github.com/kelseyhightower/envconfig

## Labels

```
priority: P3
type: enhancement
component: backend
```

package config

import (
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Logger   Logger
	App      App
	Security Security
}

type Logger struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type App struct {
	Adress string `yaml:"adress"`
	Dsn    string `yaml:"dsn"`
}

type Security struct {
	PasswordHash PasswordHash `yaml:"password_hash"`
	Tokener      Tokener      `yaml:"tokener"`
}

type PasswordHash struct {
	Algorithm string `yaml:"algorithm"`

	MemoryKiB    uint32 `yaml:"memory_kib"`
	Iterations   uint32 `yaml:"iterations"`          // или yaml:"iterations"
	Parallelism  uint8  `yaml:"parallelism"`
	SaltLen      uint32 `yaml:"salt_len"`
	KeyLen       uint32 `yaml:"key_len"`

	MaxMemoryKiB   uint32 `yaml:"max_memory_kib"`
	MaxIterations  uint32 `yaml:"max_iterations"`     // или max_iterations
	MaxParallelism uint8  `yaml:"max_parallelism"`
}

type SecondsDuration time.Duration

func (d *SecondsDuration) UnmarshalYAML(value *yaml.Node) error {
	// ожидаем число секунд (int)
	var sec int64
	if err := value.Decode(&sec); err != nil {
		return fmt.Errorf("duration must be integer seconds: %w", err)
	}
	if sec < 0 {
		return fmt.Errorf("duration seconds must be >= 0")
	}
	*d = SecondsDuration(time.Duration(sec) * time.Second)
	return nil
}

func (d SecondsDuration) Duration() time.Duration {
	return time.Duration(d)
}

type Tokener struct {
	Secret    []byte          `yaml:"-"` // секрет только из env
	TTL       SecondsDuration `yaml:"ttl_sec"`
	Issuer    string          `yaml:"issuer"`
	ClockSkew SecondsDuration `yaml:"clock_skew_sec"`
}

// NewConfig читает YAML-конфиг из файла и подтягивает SECRET_KEY из env.
func NewConfig(fileCfg string) (*Config, error) {
	raw, err := os.ReadFile(fileCfg)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	// cfg должен быть значением, а не nil указателем
	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("yaml unmarshal: %w", err)
	}

	// секрет из env
	secret := os.Getenv("SECRET_KEY")
	if secret == "" {
		return nil, fmt.Errorf("SECRET_KEY env var is required")
	}
	cfg.Security.Tokener.Secret = []byte(secret)

	// --- Минимальная валидация ---
	if cfg.App.Adress == "" {
		return nil, fmt.Errorf("app.adress is required")
	}
	if cfg.App.Dsn == "" {
		return nil, fmt.Errorf("app.dsn is required")
	}

	ttl := cfg.Security.Tokener.TTL.Duration()
	if ttl <= 0 {
		return nil, fmt.Errorf("security.tokener.ttl_sec must be > 0")
	}

	if cfg.Security.Tokener.Issuer == "" {
		return nil, fmt.Errorf("security.tokener.issuer is required")
	}

	clockSkew := cfg.Security.Tokener.ClockSkew.Duration()
	if clockSkew < 0*time.Second {
		return nil, fmt.Errorf("security.tokener.clock_skew_sec must be >= 0")
	}

	ph := cfg.Security.PasswordHash
	if ph.Algorithm == "" {
		return nil, fmt.Errorf("security.password_hash.algorithm is required")
	}

	// optional sanity checks (минимум)
	if ph.MemoryKiB <= 0 || ph.Iterations <= 0 || ph.Parallelism <= 0 {
		return nil, fmt.Errorf("security.password_hash params must be positive")
	}
	if ph.SaltLen <= 0 || ph.KeyLen <= 0 {
		return nil, fmt.Errorf("security.password_hash salt_len/key_len must be positive")
	}

	return &cfg, nil
}

package argon2hash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/IvanOplesnin/BotTradeService.git/internal/config"
	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash   = errors.New("invalid password hash format")
	ErrInvalidConfig = errors.New("invalid argon2 config")
)

type Hasher struct {
	Memory      uint32 // KiB
	Time        uint32 // iterations
	Parallelism uint8
	SaltLen     uint32
	KeyLen      uint32

	maxMemory      uint32
	maxTime        uint32
	maxParallelism uint8
}

// New creates Hasher from config.PasswordHash and validates everything here.
func New(ph config.PasswordHash) (*Hasher, error) {
	// ---- algorithm ----
	if ph.Algorithm == "" {
		return nil, fmt.Errorf("%w: algorithm is required", ErrInvalidConfig)
	}
	if ph.Algorithm != "argon2id" {
		return nil, fmt.Errorf("%w: unsupported algorithm %q", ErrInvalidConfig, ph.Algorithm)
	}

	// ---- required params ----
	if ph.MemoryKiB == 0 || ph.Iterations == 0 || ph.Parallelism == 0 || ph.SaltLen == 0 || ph.KeyLen == 0 {
		return nil, fmt.Errorf("%w: params must be positive", ErrInvalidConfig)
	}

	// ---- security minima ----
	if ph.MemoryKiB < 8*1024 {
		return nil, fmt.Errorf("%w: memory_kib must be >= 8192 (8 MiB)", ErrInvalidConfig)
	}
	if ph.Iterations < 2 {
		return nil, fmt.Errorf("%w: iterations must be >= 2", ErrInvalidConfig)
	}
	if ph.Parallelism > 32 {
		return nil, fmt.Errorf("%w: parallelism must be <= 32", ErrInvalidConfig)
	}
	if ph.SaltLen < 16 {
		return nil, fmt.Errorf("%w: salt_len must be >= 16", ErrInvalidConfig)
	}
	if ph.KeyLen < 32 {
		return nil, fmt.Errorf("%w: key_len must be >= 32", ErrInvalidConfig)
	}

	// ---- guards defaults (if not set) ----
	maxMem := ph.MaxMemoryKiB
	if maxMem == 0 {
		maxMem = ph.MemoryKiB
	}
	maxIter := ph.MaxIterations
	if maxIter == 0 {
		maxIter = ph.Iterations
	}
	maxPar := ph.MaxParallelism
	if maxPar == 0 {
		maxPar = ph.Parallelism
	}

	// ---- guards must cover base ----
	if maxMem < ph.MemoryKiB {
		return nil, fmt.Errorf("%w: max_memory_kib must be >= memory_kib", ErrInvalidConfig)
	}
	if maxIter < ph.Iterations {
		return nil, fmt.Errorf("%w: max_iterations must be >= iterations", ErrInvalidConfig)
	}
	if maxPar < ph.Parallelism {
		return nil, fmt.Errorf("%w: max_parallelism must be >= parallelism", ErrInvalidConfig)
	}

	// ---- upper bounds (anti foot-gun) ----
	if ph.MemoryKiB > 1024*1024 { // 1 GiB
		return nil, fmt.Errorf("%w: memory_kib too large", ErrInvalidConfig)
	}
	if ph.Iterations > 50 {
		return nil, fmt.Errorf("%w: iterations too large", ErrInvalidConfig)
	}
	if maxMem > 1024*1024 {
		return nil, fmt.Errorf("%w: max_memory_kib too large", ErrInvalidConfig)
	}
	if maxIter > 100 {
		return nil, fmt.Errorf("%w: max_iterations too large", ErrInvalidConfig)
	}
	if maxPar > 64 {
		return nil, fmt.Errorf("%w: max_parallelism too large", ErrInvalidConfig)
	}

	return &Hasher{
		Memory:      ph.MemoryKiB,
		Time:        ph.Iterations,
		Parallelism: ph.Parallelism,
		SaltLen:     ph.SaltLen,
		KeyLen:      ph.KeyLen,

		maxMemory:      maxMem,
		maxTime:        maxIter,
		maxParallelism: maxPar,
	}, nil
}

func (h *Hasher) Hash(password string) (string, error) {
	if password == "" {
		return "", errors.New("password is empty")
	}

	salt := make([]byte, h.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("read salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, h.Time, h.Memory, h.Parallelism, h.KeyLen)

	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	keyB64 := base64.RawStdEncoding.EncodeToString(key)

	encoded := fmt.Sprintf("argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		h.Memory, h.Time, h.Parallelism, saltB64, keyB64,
	)
	return encoded, nil
}

func (h *Hasher) CompareHash(password, hash string) (bool, error) {
	// password is plain, hash is encoded string
	params, salt, key, err := decodeHash(hash)
	if err != nil {
		return false, err
	}

	newKey := argon2.IDKey([]byte(password), salt, params.Time, params.Memory, params.Parallelism, uint32(len(key)))

	// constant-time compare
	if subtle.ConstantTimeCompare(key, newKey) == 1 {
		return true, nil
	}
	return false, nil
}

type decodedParams struct {
	Memory      uint32
	Time        uint32
	Parallelism uint8
}

func decodeHash(encoded string) (decodedParams, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	// expected: argon2id | v=19 | m=...,t=...,p=... | salt | key
	if len(parts) != 5 {
		return decodedParams{}, nil, nil, ErrInvalidHash
	}
	if parts[0] != "argon2id" || parts[1] != "v=19" {
		return decodedParams{}, nil, nil, ErrInvalidHash
	}

	var p decodedParams
	// parts[2] = "m=65536,t=3,p=1"
	for _, kv := range strings.Split(parts[2], ",") {
		kvp := strings.SplitN(kv, "=", 2)
		if len(kvp) != 2 {
			return decodedParams{}, nil, nil, ErrInvalidHash
		}
		switch kvp[0] {
		case "m":
			var v uint32
			if _, err := fmt.Sscanf(kvp[1], "%d", &v); err != nil || v == 0 {
				return decodedParams{}, nil, nil, ErrInvalidHash
			}
			p.Memory = v
		case "t":
			var v uint32
			if _, err := fmt.Sscanf(kvp[1], "%d", &v); err != nil || v == 0 {
				return decodedParams{}, nil, nil, ErrInvalidHash
			}
			p.Time = v
		case "p":
			var v uint32
			if _, err := fmt.Sscanf(kvp[1], "%d", &v); err != nil || v == 0 || v > 255 {
				return decodedParams{}, nil, nil, ErrInvalidHash
			}
			p.Parallelism = uint8(v)
		default:
			return decodedParams{}, nil, nil, ErrInvalidHash
		}
	}
	if p.Memory == 0 || p.Time == 0 || p.Parallelism == 0 {
		return decodedParams{}, nil, nil, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(salt) < 8 {
		return decodedParams{}, nil, nil, ErrInvalidHash
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(key) < 16 {
		return decodedParams{}, nil, nil, ErrInvalidHash
	}

	return p, salt, key, nil
}

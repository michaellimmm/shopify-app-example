package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

var mutex = &sync.RWMutex{}
var env = map[string]string{}

func Load(filename string) {
	if len(filename) > 0 {
		loadFromFile(filename)
	}

	loadFromOs()
}

func loadFromOs() {
	mutex.Lock()
	defer mutex.Unlock()

	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		env[pair[0]] = os.Getenv(pair[0])
	}
}

func loadFromFile(filename string) {
	mutex.Lock()
	defer mutex.Unlock()

	err := godotenv.Load(filename)
	if err != nil {
		panic(fmt.Errorf("failed to load file: %v", err))
	}
}

func Get(key string, value string) string {
	mutex.RLock()
	defer mutex.RUnlock()

	if envValue, ok := env[key]; ok {
		return envValue
	}

	return value
}

func GetInt64(key string, value int64) int64 {
	mutex.RLock()
	defer mutex.RUnlock()

	if envValue, ok := env[key]; ok {
		val, err := strconv.ParseInt(envValue, 10, 64)
		if err != nil {
			return value
		}
		return val
	}

	return value
}

func GetUint64(key string, value uint64) uint64 {
	mutex.RLock()
	defer mutex.RUnlock()

	if envValue, ok := env[key]; ok {
		val, err := strconv.ParseUint(envValue, 10, 64)
		if err != nil {
			return value
		}
		return val
	}

	return value
}

func MustGet(key string) (string, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	if value, ok := env[key]; ok {
		return value, nil
	}

	return "", fmt.Errorf("could not find ENV var with %s", key)
}

func Set(key string, value string) {
	mutex.Lock()
	defer mutex.Unlock()

	env[key] = value
}

func MustSet(key string, value string) error {
	mutex.Lock()
	defer mutex.Unlock()

	err := os.Setenv(key, value)
	if err != nil {
		return err
	}

	env[key] = value
	return nil
}

func Environ() []string {
	mutex.RLock()
	defer mutex.RUnlock()

	var e []string
	for k, v := range env {
		e = append(e, fmt.Sprintf("%s=%s", k, v))
	}

	return e
}

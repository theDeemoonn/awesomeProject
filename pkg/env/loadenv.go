package env

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadEnv функция для загрузки переменных окружения из файла .env
func LoadEnv(filename string) error {
	cwd, _ := os.Getwd()
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening .env file from %s: %v", cwd, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		os.Setenv(key, value)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %v", err)
	}
	return nil
}

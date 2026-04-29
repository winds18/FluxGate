package singbox

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PublishResult struct {
	CheckResult
	Published     bool `json:"published"`
	PreviousSaved bool `json:"previous_saved"`
}

func PublishConfig(config Config, configPath string, previousPath string) (PublishResult, error) {
	configPath = strings.TrimSpace(configPath)
	if configPath == "" {
		return PublishResult{}, errors.New("sing-box config path is required")
	}
	if strings.TrimSpace(previousPath) == "" {
		previousPath = configPath + ".previous"
	}

	check, err := CheckConfig(config)
	if err != nil {
		return PublishResult{}, err
	}
	result := PublishResult{CheckResult: check}
	if !check.Valid {
		return result, fmt.Errorf("sing-box config check failed")
	}

	body, err := Marshal(config)
	if err != nil {
		return result, err
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return result, err
	}

	existing, err := os.ReadFile(configPath)
	if err == nil {
		if err := os.MkdirAll(filepath.Dir(previousPath), 0o755); err != nil {
			return result, err
		}
		if err := os.WriteFile(previousPath, existing, 0o644); err != nil {
			return result, err
		}
		result.PreviousSaved = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return result, err
	}

	if err := writeFileAtomic(configPath, body, 0o644); err != nil {
		return result, err
	}
	result.Published = true
	return result, nil
}

func writeFileAtomic(path string, body []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	file, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := file.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := file.Write(body); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Chmod(mode); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

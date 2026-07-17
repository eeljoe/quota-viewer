package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	KimiAPIKey         string `json:"kimi_api_key"`
	XfyunCookie        string `json:"xfyun_cookie"`
	MimoCookie         string `json:"mimo_cookie"`
	RefreshIntervalMin int    `json:"refresh_interval_min"`
	BallX              int    `json:"ball_x"`
	BallY              int    `json:"ball_y"`
}

func configDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", os.ErrNotExist
	}
	dir := filepath.Join(appData, "quota-viewer")
	return dir, nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load 读取配置文件。文件不存在时返回带默认值的空配置(不报错)。
func Load() (*Config, error) {
	cfg := &Config{
		RefreshIntervalMin: 15,
		BallX:              -1,
		BallY:              -1,
	}

	path, err := configPath()
	if err != nil {
		return cfg, nil // APPDATA 不存在,返回默认配置
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // 文件不存在不算错误
		}
		return cfg, err
	}

	err = json.Unmarshal(data, cfg)
	if err != nil {
		return cfg, err
	}

	// 确保默认值
	if cfg.RefreshIntervalMin <= 0 {
		cfg.RefreshIntervalMin = 15
	}

	return cfg, nil
}

// Save 写入配置文件。目录不存在时自动创建。
func Save(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	path, err := configPath()
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

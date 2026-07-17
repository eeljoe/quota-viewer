package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FileNotExists_ReturnsDefaults(t *testing.T) {
	// 临时设置 APPDATA 到测试目录
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.RefreshIntervalMin != 15 {
		t.Errorf("expected default RefreshIntervalMin=15, got %d", cfg.RefreshIntervalMin)
	}
	if cfg.BallX != -1 || cfg.BallY != -1 {
		t.Errorf("expected default BallX=-1, BallY=-1, got %d,%d", cfg.BallX, cfg.BallY)
	}
}

func TestSaveThenLoad_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)

	original := &Config{
		KimiAPIKey:         "sk-kimi-test123",
		XfyunCookie:        "SSID=abc; token=xyz",
		MimoCookie:         "session=def",
		RefreshIntervalMin: 30,
		BallX:              100,
		BallY:              200,
	}

	err := Save(original)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// 验证文件确实创建在正确路径
	expectedPath := filepath.Join(tmpDir, "quota-viewer", "config.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("config file not created at %s", expectedPath)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.KimiAPIKey != "sk-kimi-test123" {
		t.Errorf("KimiAPIKey mismatch: got %s", loaded.KimiAPIKey)
	}
	if loaded.XfyunCookie != "SSID=abc; token=xyz" {
		t.Errorf("XfyunCookie mismatch: got %s", loaded.XfyunCookie)
	}
	if loaded.RefreshIntervalMin != 30 {
		t.Errorf("RefreshIntervalMin mismatch: got %d", loaded.RefreshIntervalMin)
	}
	if loaded.BallX != 100 || loaded.BallY != 200 {
		t.Errorf("Ball position mismatch: got %d,%d", loaded.BallX, loaded.BallY)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)

	// 确保目录不存在
	dir := filepath.Join(tmpDir, "quota-viewer")
	os.RemoveAll(dir)

	cfg := &Config{KimiAPIKey: "test"}
	err := Save(cfg)
	if err != nil {
		t.Fatalf("Save() should create directory, got error: %v", err)
	}
}

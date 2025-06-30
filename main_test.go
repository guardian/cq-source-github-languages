package main

import (
	"testing"

	pluginpkg "github.com/guardian/cq-source-github-languages/resources/plugin"
)

func TestMain(t *testing.T) {
	// Test that main function doesn't panic
	// We can't easily test main() directly since it calls serve()
	// but we can test that the package loads correctly
	t.Run("package loads", func(t *testing.T) {
		// If we get here, the package compiled and loaded successfully
		if testing.Short() {
			t.Skip("skipping in short mode")
		}
	})
}

func TestPlugin(t *testing.T) {
	// Test plugin creation
	p := pluginpkg.Plugin()
	if p == nil {
		t.Fatal("Plugin() returned nil")
	}

	// Test plugin metadata
	meta := p.Meta()
	if meta.Name == "" {
		t.Error("Plugin name should not be empty")
	}

	if meta.Name != "github-languages" {
		t.Errorf("Expected plugin name to be 'github-languages', got %s", meta.Name)
	}
}

func TestMainPackageIntegration(t *testing.T) {
	// Test that all components work together
	t.Run("plugin initialization", func(t *testing.T) {
		// This tests that the plugin can be initialized without errors
		p := pluginpkg.Plugin()
		if p == nil {
			t.Fatal("Failed to initialize plugin")
		}

		// Test that the plugin has the required components
		if p.Meta().Name == "" {
			t.Error("Plugin metadata not properly configured")
		}
	})
}

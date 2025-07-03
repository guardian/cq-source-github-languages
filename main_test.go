package main

import (
	"testing"

	pluginpkg "github.com/guardian/cq-source-github-languages/resources/plugin"
)

func TestMain(t *testing.T) {
	t.Run("main package can import and use plugin", func(t *testing.T) {
		// This is an integration test - verify main package can use plugin package
		p := pluginpkg.Plugin()
		if p == nil {
			t.Fatal("Main package failed to create plugin - integration issue")
		}

		// Just verify the integration works, don't duplicate plugin-specific tests
		t.Log("Main package successfully integrated with plugin package")
	})
}

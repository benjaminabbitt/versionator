// versionator-plugin-patch-npm patches version in package.json files.
package main

import (
	"github.com/benjaminabbitt/versionator/pkg/plugin"
	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

// jsonPatcher is the shared JSON patching function
var jsonPatcher = plugin.PatchJSON()

// NPMPatch implements the PatchPlugin interface for npm package.json.
type NPMPatch struct{}

func (p *NPMPatch) Name() string {
	return "patch-npm"
}

func (p *NPMPatch) FilePattern() string {
	return "package.json"
}

func (p *NPMPatch) Description() string {
	return "npm/Node.js package manifest (package.json)"
}

func (p *NPMPatch) Patch(content, version string) (string, error) {
	return jsonPatcher(content, version)
}

func main() {
	sdk.ServePatch(&NPMPatch{})
}

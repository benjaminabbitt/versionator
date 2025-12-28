// versionator-plugin-patch-gradle patches version in build.gradle files.
package main

import (
	"regexp"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

// GradlePatch implements the PatchPlugin interface for Gradle build files.
type GradlePatch struct{}

func (p *GradlePatch) Name() string {
	return "patch-gradle"
}

func (p *GradlePatch) FilePattern() string {
	return "build.gradle"
}

func (p *GradlePatch) Description() string {
	return "Gradle build script (Groovy/Kotlin DSL)"
}

func (p *GradlePatch) Patch(content, version string) (string, error) {
	// Match version = "..." or version = '...'
	reDouble := regexp.MustCompile(`(?m)^(\s*version\s*=\s*)"[^"]*"`)
	reSingle := regexp.MustCompile(`(?m)^(\s*version\s*=\s*)'[^']*'`)

	if reDouble.MatchString(content) {
		return reDouble.ReplaceAllString(content, `${1}"`+version+`"`), nil
	}
	if reSingle.MatchString(content) {
		return reSingle.ReplaceAllString(content, `${1}'`+version+`'`), nil
	}
	return content, nil
}

func main() {
	sdk.ServePatch(&GradlePatch{})
}

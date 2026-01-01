// versionator-plugin-patch-gradlekts patches version in build.gradle.kts files.
package main

import (
	"regexp"

	"github.com/benjaminabbitt/versionator/pkg/plugin/sdk"
)

// GradleKtsPatch implements the PatchPlugin interface for Gradle Kotlin DSL build files.
type GradleKtsPatch struct{}

func (p *GradleKtsPatch) Name() string {
	return "patch-gradlekts"
}

func (p *GradleKtsPatch) FilePattern() string {
	return "build.gradle.kts"
}

func (p *GradleKtsPatch) Description() string {
	return "Gradle build script (Kotlin DSL)"
}

func (p *GradleKtsPatch) Patch(content, version string) (string, error) {
	// Match version = "..." (Kotlin DSL uses double quotes)
	re := regexp.MustCompile(`(?m)^(\s*version\s*=\s*)"[^"]*"`)

	if re.MatchString(content) {
		return re.ReplaceAllString(content, `${1}"`+version+`"`), nil
	}
	return content, nil
}

func main() {
	sdk.ServePatch(&GradleKtsPatch{})
}

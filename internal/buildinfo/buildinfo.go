package buildinfo

// Version is the versionator version, set at build time via ldflags
// Example: go build -ldflags "-X github.com/benjaminabbitt/versionator/internal/buildinfo.Version=1.0.0"
var Version = "dev"

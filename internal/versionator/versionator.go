package versionator

import (
	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/benjaminabbitt/versionator/internal/version"
)

// GetPreReleaseTemplate returns the pre-release template from config
func GetPreReleaseTemplate() (string, error) {
	cfg, err := config.ReadConfig()
	if err != nil {
		return "", err
	}
	return cfg.PreRelease.Template, nil
}

// GetMetadataTemplate returns the metadata template from config
func GetMetadataTemplate() (string, error) {
	cfg, err := config.ReadConfig()
	if err != nil {
		return "", err
	}
	return cfg.Metadata.Template, nil
}

// RenderPreRelease renders the pre-release template with current version data
func RenderPreRelease() (string, error) {
	template, err := GetPreReleaseTemplate()
	if err != nil {
		return "", err
	}
	if template == "" {
		return "", nil
	}

	vd, err := version.Load()
	if err != nil {
		return "", err
	}

	data := emit.BuildTemplateDataFromVersion(vd)
	return emit.RenderTemplateWithData(template, data)
}

// RenderMetadata renders the metadata template with current version data
func RenderMetadata() (string, error) {
	template, err := GetMetadataTemplate()
	if err != nil {
		return "", err
	}
	if template == "" {
		return "", nil
	}

	vd, err := version.Load()
	if err != nil {
		return "", err
	}

	data := emit.BuildTemplateDataFromVersion(vd)
	return emit.RenderTemplateWithData(template, data)
}

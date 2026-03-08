---
title: Azure DevOps
description: Using versionator with Azure DevOps Pipelines
sidebar_position: 3
---

# Azure DevOps

**Platform:** [Azure DevOps Pipelines](https://azure.microsoft.com/en-us/products/devops/pipelines)

## Pipeline Variables

```yaml
steps:
  - script: |
      go install github.com/benjaminabbitt/versionator@latest
      VERSION=$(versionator version)
      echo "##vso[task.setvariable variable=VERSION]$VERSION"
    displayName: 'Get Version'

  - script: |
      echo "Building version $(VERSION)"
    displayName: 'Build'
```

## Build with Version

```yaml
steps:
  - script: |
      go install github.com/benjaminabbitt/versionator@latest
    displayName: 'Install versionator'

  - script: |
      VERSION=$(versionator version)
      go build -ldflags "-X main.VERSION=$VERSION" -o $(Build.ArtifactStagingDirectory)/app
    displayName: 'Build with version'

  - task: PublishBuildArtifacts@1
    inputs:
      pathtoPublish: '$(Build.ArtifactStagingDirectory)'
      artifactName: 'app'
```

## Full Pipeline Example

```yaml
trigger:
  - main
  - refs/tags/v*

pool:
  vmImage: 'ubuntu-latest'

variables:
  GOPATH: '$(Agent.BuildDirectory)/go'
  VERSION: ''

stages:
  - stage: Build
    jobs:
      - job: BuildJob
        steps:
          - script: |
              go install github.com/benjaminabbitt/versionator@latest
              echo "##vso[task.setvariable variable=VERSION]$(versionator version)"
            displayName: 'Get Version'

          - script: |
              echo "Building version $(VERSION)"
              go build -ldflags "-X main.VERSION=$(VERSION)" -o app
            displayName: 'Build'

  - stage: Release
    condition: startsWith(variables['Build.SourceBranch'], 'refs/tags/v')
    jobs:
      - job: ReleaseJob
        steps:
          - script: |
              echo "Releasing $(Build.SourceBranchName)"
            displayName: 'Release'
```

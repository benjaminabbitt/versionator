---
title: GitLab CI
description: Using versionator with GitLab CI/CD
sidebar_position: 2
---

# GitLab CI

**Platform:** [GitLab CI/CD](https://docs.gitlab.com/ee/ci/)

## Basic Usage

```yaml
variables:
  VERSION: ""

before_script:
  - go install github.com/benjaminabbitt/versionator@latest
  - VERSION=$(versionator version)

build:
  script:
    - echo "Building version $VERSION"
    - go build -ldflags "-X main.VERSION=$VERSION" -o app
```

## Dynamic Version

```yaml
build:
  script:
    - |
      VERSION=$(versionator version \
        -t "{{MajorMinorPatch}}{{MetadataWithPlus}}" \
        --metadata="${CI_COMMIT_SHORT_SHA}")
    - echo "Building $VERSION"
```

## Release Job

```yaml
release:
  stage: deploy
  rules:
    - if: $CI_COMMIT_TAG
  script:
    - VERSION=${CI_COMMIT_TAG#v}
    - echo "Releasing version $VERSION"
```

## Full Pipeline Example

```yaml
stages:
  - build
  - test
  - release

variables:
  VERSION: ""

.setup: &setup
  before_script:
    - go install github.com/benjaminabbitt/versionator@latest
    - export VERSION=$(versionator version)

build:
  stage: build
  <<: *setup
  script:
    - go build -ldflags "-X main.VERSION=$VERSION" -o app
  artifacts:
    paths:
      - app

test:
  stage: test
  <<: *setup
  script:
    - go test ./...

release:
  stage: release
  rules:
    - if: $CI_COMMIT_TAG =~ /^v\d+\.\d+\.\d+$/
  script:
    - echo "Creating release for ${CI_COMMIT_TAG}"
```

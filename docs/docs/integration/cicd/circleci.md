---
title: CircleCI
description: Using versionator with CircleCI
sidebar_position: 5
---

# CircleCI

**Platform:** [CircleCI](https://circleci.com/)

## Basic Usage

```yaml
version: 2.1

jobs:
  build:
    docker:
      - image: cimg/go:1.21
    steps:
      - checkout
      - run:
          name: Install versionator
          command: go install github.com/benjaminabbitt/versionator@latest
      - run:
          name: Get version
          command: |
            VERSION=$(versionator version)
            echo "export VERSION=$VERSION" >> $BASH_ENV
      - run:
          name: Build
          command: go build -ldflags "-X main.VERSION=$VERSION" -o app
```

## With Caching

```yaml
version: 2.1

jobs:
  build:
    docker:
      - image: cimg/go:1.21
    steps:
      - checkout

      - restore_cache:
          keys:
            - go-mod-v1-{{ checksum "go.sum" }}

      - run:
          name: Install dependencies
          command: |
            go install github.com/benjaminabbitt/versionator@latest
            go mod download

      - save_cache:
          key: go-mod-v1-{{ checksum "go.sum" }}
          paths:
            - ~/go/pkg/mod

      - run:
          name: Build with version
          command: |
            VERSION=$(versionator version)
            go build -ldflags "-X main.VERSION=$VERSION" -o app

      - store_artifacts:
          path: app
```

## Full Workflow

```yaml
version: 2.1

executors:
  go-executor:
    docker:
      - image: cimg/go:1.21

jobs:
  build:
    executor: go-executor
    steps:
      - checkout
      - run:
          name: Install versionator
          command: go install github.com/benjaminabbitt/versionator@latest
      - run:
          name: Get version
          command: |
            VERSION=$(versionator version)
            echo "export VERSION=$VERSION" >> $BASH_ENV
            echo "Building version $VERSION"
      - run:
          name: Build
          command: go build -ldflags "-X main.VERSION=$VERSION" -o app
      - persist_to_workspace:
          root: .
          paths:
            - app

  test:
    executor: go-executor
    steps:
      - checkout
      - run:
          name: Test
          command: go test ./...

  release:
    executor: go-executor
    steps:
      - attach_workspace:
          at: .
      - run:
          name: Release
          command: |
            echo "Releasing ${CIRCLE_TAG}"

workflows:
  version: 2
  build-test-release:
    jobs:
      - build:
          filters:
            tags:
              only: /^v.*/
      - test:
          filters:
            tags:
              only: /^v.*/
      - release:
          requires:
            - build
            - test
          filters:
            branches:
              ignore: /.*/
            tags:
              only: /^v\d+\.\d+\.\d+$/
```

---
title: Jenkins
description: Using versionator with Jenkins Pipelines
sidebar_position: 4
---

# Jenkins

**Platform:** [Jenkins](https://www.jenkins.io/)

## Declarative Pipeline

```groovy
pipeline {
    agent any

    environment {
        VERSION = ''
    }

    stages {
        stage('Get Version') {
            steps {
                script {
                    sh 'go install github.com/benjaminabbitt/versionator@latest'
                    env.VERSION = sh(script: 'versionator version', returnStdout: true).trim()
                }
            }
        }

        stage('Build') {
            steps {
                sh "go build -ldflags '-X main.VERSION=${env.VERSION}' -o app"
            }
        }
    }
}
```

## Scripted Pipeline

```groovy
node {
    def version = ''

    stage('Checkout') {
        checkout scm
    }

    stage('Get Version') {
        sh 'go install github.com/benjaminabbitt/versionator@latest'
        version = sh(script: 'versionator version', returnStdout: true).trim()
        echo "Building version: ${version}"
    }

    stage('Build') {
        sh "go build -ldflags '-X main.VERSION=${version}' -o app"
    }

    stage('Archive') {
        archiveArtifacts artifacts: 'app', fingerprint: true
    }
}
```

## Full Pipeline with Release

```groovy
pipeline {
    agent any

    environment {
        VERSION = ''
        GOPATH = "${WORKSPACE}/go"
        PATH = "${GOPATH}/bin:${env.PATH}"
    }

    stages {
        stage('Setup') {
            steps {
                sh 'go install github.com/benjaminabbitt/versionator@latest'
                script {
                    env.VERSION = sh(script: 'versionator version', returnStdout: true).trim()
                }
                echo "Version: ${env.VERSION}"
            }
        }

        stage('Build') {
            steps {
                sh "go build -ldflags '-X main.VERSION=${env.VERSION}' -o app"
            }
        }

        stage('Test') {
            steps {
                sh 'go test ./...'
            }
        }

        stage('Release') {
            when {
                tag pattern: "v\\d+\\.\\d+\\.\\d+", comparator: "REGEXP"
            }
            steps {
                echo "Releasing ${env.TAG_NAME}"
                archiveArtifacts artifacts: 'app', fingerprint: true
            }
        }
    }

    post {
        always {
            cleanWs()
        }
    }
}
```

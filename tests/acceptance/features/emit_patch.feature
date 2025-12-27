Feature: Emit Patch - Manifest File Patching
  As a software developer
  I want to patch version strings in manifest files
  So that my project files stay in sync with the VERSION file

  Background:
    Given a clean git repository
    And a VERSION file with version "1.2.3"

  # Python - pyproject.toml
  Scenario: Patch pyproject.toml version
    Given a file "pyproject.toml" with content:
      """
      [project]
      name = "myapp"
      version = "0.0.0"
      description = "My application"
      """
    When I run "versionator emit patch"
    Then the exit code should be 0
    And the file "pyproject.toml" should contain 'version = "1.2.3"'

  Scenario: Patch pyproject.toml with prerelease
    Given a file "pyproject.toml" with content:
      """
      [project]
      name = "myapp"
      version = "0.0.0"
      """
    When I run "versionator emit patch --prerelease='alpha'"
    Then the file "pyproject.toml" should contain 'version = "1.2.3-alpha"'

  # Python setuptools - setup.py
  Scenario: Patch setup.py version
    Given a file "setup.py" with content:
      """
      from setuptools import setup

      setup(
          name="myapp",
          version="0.0.0",
          description="My application",
      )
      """
    When I run "versionator emit patch"
    Then the file "setup.py" should contain 'version="1.2.3"'

  # JavaScript/TypeScript - package.json
  Scenario: Patch package.json version
    Given a file "package.json" with content:
      """
      {
        "name": "myapp",
        "version": "0.0.0",
        "description": "My application"
      }
      """
    When I run "versionator emit patch"
    Then the exit code should be 0
    And the file "package.json" should contain '"version": "1.2.3"'

  # Rust - Cargo.toml
  Scenario: Patch Cargo.toml version
    Given a file "Cargo.toml" with content:
      """
      [package]
      name = "myapp"
      version = "0.0.0"
      edition = "2021"
      """
    When I run "versionator emit patch"
    Then the exit code should be 0
    And the file "Cargo.toml" should contain 'version = "1.2.3"'

  # Java Maven - pom.xml
  Scenario: Patch pom.xml version
    Given a file "pom.xml" with content:
      """
      <?xml version="1.0" encoding="UTF-8"?>
      <project>
        <modelVersion>4.0.0</modelVersion>
        <groupId>com.example</groupId>
        <artifactId>myapp</artifactId>
        <version>0.0.0</version>
      </project>
      """
    When I run "versionator emit patch"
    Then the exit code should be 0
    And the file "pom.xml" should contain '<version>1.2.3</version>'

  # Java Gradle - build.gradle
  Scenario: Patch build.gradle version
    Given a file "build.gradle" with content:
      """
      plugins {
          id 'java'
      }

      group = 'com.example'
      version = '0.0.0'
      """
    When I run "versionator emit patch"
    Then the exit code should be 0
    And the file "build.gradle" should contain "version = '1.2.3'"

  # Gradle Kotlin DSL - build.gradle.kts
  Scenario: Patch build.gradle.kts version
    Given a file "build.gradle.kts" with content:
      """
      plugins {
          kotlin("jvm") version "1.9.0"
      }

      group = "com.example"
      version = "0.0.0"
      """
    When I run "versionator emit patch"
    Then the exit code should be 0
    And the file "build.gradle.kts" should contain 'version = "1.2.3"'

  # C# - *.csproj
  Scenario: Patch csproj version
    Given a file "MyApp.csproj" with content:
      """
      <Project Sdk="Microsoft.NET.Sdk">
        <PropertyGroup>
          <TargetFramework>net8.0</TargetFramework>
          <Version>0.0.0</Version>
        </PropertyGroup>
      </Project>
      """
    When I run "versionator emit patch"
    Then the exit code should be 0
    And the file "MyApp.csproj" should contain '<Version>1.2.3</Version>'

  # PHP - composer.json
  Scenario: Patch composer.json version
    Given a file "composer.json" with content:
      """
      {
        "name": "vendor/myapp",
        "version": "0.0.0",
        "type": "library"
      }
      """
    When I run "versionator emit patch"
    Then the exit code should be 0
    And the file "composer.json" should contain '"version": "1.2.3"'

  # Ruby - gemspec
  Scenario: Patch gemspec version
    Given a file "myapp.gemspec" with content:
      """
      Gem::Specification.new do |spec|
        spec.name          = "myapp"
        spec.version       = "0.0.0"
        spec.authors       = ["Author"]
      end
      """
    When I run "versionator emit patch"
    Then the exit code should be 0
    And the file "myapp.gemspec" should contain 'spec.version       = "1.2.3"'

  # Swift - Package.swift
  Scenario: Patch Package.swift version
    Given a file "Package.swift" with content:
      """
      // swift-tools-version:5.9
      import PackageDescription

      let package = Package(
          name: "MyApp",
          products: [
              .library(name: "MyApp", targets: ["MyApp"]),
          ],
          targets: [
              .target(name: "MyApp"),
          ]
      )

      // VERSION: 0.0.0
      """
    When I run "versionator emit patch"
    Then the exit code should be 0
    And the file "Package.swift" should contain '// VERSION: 1.2.3'

  # Multiple manifests in same directory - should patch all
  Scenario: Patch multiple manifest files
    Given a file "pyproject.toml" with content:
      """
      [project]
      name = "myapp"
      version = "0.0.0"
      """
    And a file "package.json" with content:
      """
      {
        "name": "myapp",
        "version": "0.0.0"
      }
      """
    When I run "versionator emit patch"
    Then the exit code should be 0
    And the file "pyproject.toml" should contain 'version = "1.2.3"'
    And the file "package.json" should contain '"version": "1.2.3"'

  # Selective patching by file type
  Scenario: Patch specific manifest file
    Given a file "pyproject.toml" with content:
      """
      [project]
      name = "myapp"
      version = "0.0.0"
      """
    And a file "package.json" with content:
      """
      {
        "name": "myapp",
        "version": "0.0.0"
      }
      """
    When I run "versionator emit patch pyproject.toml"
    Then the exit code should be 0
    And the file "pyproject.toml" should contain 'version = "1.2.3"'
    And the file "package.json" should contain '"version": "0.0.0"'

  # Dry run mode
  Scenario: Dry run shows changes without modifying files
    Given a file "package.json" with content:
      """
      {
        "name": "myapp",
        "version": "0.0.0"
      }
      """
    When I run "versionator emit patch --dry-run"
    Then the exit code should be 0
    And the output should contain "package.json"
    And the output should contain "1.2.3"
    And the file "package.json" should contain '"version": "0.0.0"'

  # Error handling
  Scenario: Error when no manifest files found
    When I run "versionator emit patch"
    Then the exit code should be 1
    And the output should contain "no manifest files found"

  Scenario: Error when manifest file has invalid format
    Given a file "package.json" with content:
      """
      { invalid json }
      """
    When I run "versionator emit patch"
    Then the exit code should be 1
    And the output should contain "error"

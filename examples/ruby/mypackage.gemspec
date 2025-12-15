# frozen_string_literal: true

require_relative "lib/mypackage/version"

Gem::Specification.new do |spec|
  spec.name = "mypackage"
  spec.version = Mypackage::VERSION
  spec.authors = ["Your Name"]
  spec.email = ["your.email@example.com"]

  spec.summary = "Sample gem demonstrating versionator integration"
  spec.description = "A sample Ruby gem that uses versionator for version management"
  spec.homepage = "https://github.com/example/mypackage"
  spec.license = "MIT"
  spec.required_ruby_version = ">= 2.7.0"

  spec.files = Dir.glob("{bin,lib}/**/*") + %w[README.md]
  spec.bindir = "bin"
  spec.executables = ["mypackage"]
  spec.require_paths = ["lib"]
end

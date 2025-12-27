#!/usr/bin/env ruby
# frozen_string_literal: true

# Test application that reads version from gemspec
require "rubygems/specification"

# Load the gemspec file
gemspec = Gem::Specification.load("testgem.gemspec")
puts "Version: #{gemspec.version}"

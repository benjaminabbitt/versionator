#!/usr/bin/env ruby
# Test application that requires generated version module

require_relative 'version'

puts "Version: #{Versionator::VERSION}"
puts "Major: #{Versionator::MAJOR}, Minor: #{Versionator::MINOR}, Patch: #{Versionator::PATCH}"

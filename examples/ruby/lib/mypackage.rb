# frozen_string_literal: true

require_relative "mypackage/version"

module Mypackage
  def self.hello
    puts "Sample Ruby Application"
    puts "Version: #{Versionator::VERSION}"
  end
end

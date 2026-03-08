# frozen_string_literal: true

require_relative "mypackage/version"

module Mypackage
  def self.hello
    puts "Sample Ruby Application (Custom Template)"
    puts "Version: #{VERSION}"
    puts "Git: #{GIT_HASH}"
    puts "Built: #{BUILD_DATE}"
  end
end

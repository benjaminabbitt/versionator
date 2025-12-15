# frozen_string_literal: true

require_relative "mypackage/version"

module Mypackage
  class Error < StandardError; end

  def self.hello
    puts "Sample Ruby Application"
    puts "Version: #{VERSION}"
  end
end

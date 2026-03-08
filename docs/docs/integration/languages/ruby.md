---
title: Ruby
description: Embed version in Ruby gems
sidebar_position: 12
---

# Ruby

**Location:** [`examples/ruby/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/ruby)

Ruby generates a `version.rb` module using `versionator emit`:

```ruby title="examples/ruby/lib/mypackage.rb"
require_relative "mypackage/version"

module Mypackage
  def self.hello
    puts "Sample Ruby Application"
    puts "Version: #{Versionator::VERSION}"
  end
end
```

```makefile title="examples/ruby/Makefile (excerpt)"
version-file:
    versionator emit ruby --output lib/mypackage/version.rb

run: version-file
    ruby -I lib -e "require 'mypackage'; Mypackage.hello"
```

## Run it

```bash
$ cd examples/ruby && just run
Generating version.rb using versionator emit...
Version 0.0.16 written to lib/mypackage/version.rb
ruby -I lib -e "require 'mypackage'; Mypackage.hello"
Sample Ruby Application
Version: 0.0.16
```

## Source Code

- [`lib/mypackage.rb`](https://github.com/benjaminabbitt/versionator/blob/master/examples/ruby/lib/mypackage.rb)
- [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/ruby/justfile)
- [`Makefile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/ruby/Makefile)
- [`Containerfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/ruby/Containerfile)

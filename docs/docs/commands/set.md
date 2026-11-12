---
title: set
description: Set version to an arbitrary value
---

# set

Set version to an arbitrary value

Set the VERSION file to an arbitrary version string.

Accepts any version string the parser grammar supports:
  [v|V]Major.Minor[.Patch[.Revision]][-PreRelease][+BuildMetadata]

**Examples:**

```bash
versionator set 1.2.3
versionator set v2.0.0-rc.1
versionator set 1.2.3.4
versionator set 1.0.0-alpha.1+build.42
```

## Usage

```bash
versionator set
```


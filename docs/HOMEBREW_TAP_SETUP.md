# Homebrew Tap Setup

This document explains how to set up and maintain the Homebrew tap for versionator.

## Overview

Homebrew taps are GitHub repositories that contain formulae. For versionator, you need a separate repository named `homebrew-tap` under your GitHub account.

## Initial Setup

### 1. Create the Tap Repository

Create a new GitHub repository named `homebrew-tap`:

```bash
# Via GitHub CLI
gh repo create benjaminabbitt/homebrew-tap --public --description "Homebrew formulae for benjaminabbitt projects"

# Or manually at https://github.com/new
# Repository name: homebrew-tap
# Visibility: Public
```

### 2. Initialize the Repository

```bash
git clone git@github.com:benjaminabbitt/homebrew-tap.git
cd homebrew-tap

# Create Formula directory
mkdir -p Formula

# Add a README
cat > README.md << 'EOF'
# Homebrew Tap

Homebrew formulae for benjaminabbitt projects.

## Installation

```bash
brew tap benjaminabbitt/tap
brew install versionator
```

## Available Formulae

- **versionator** - A semantic version management CLI tool
EOF

git add .
git commit -m "Initialize homebrew tap"
git push
```

### 3. Add the Formula

The release workflow generates `versionator.rb` as an artifact. To complete the setup:

**Option A: Manual (Initial Setup)**

1. After a release, download the `homebrew-formula` artifact from GitHub Actions
2. Copy `versionator.rb` to `Formula/versionator.rb` in homebrew-tap
3. Commit and push

```bash
cd homebrew-tap
cp /path/to/downloaded/versionator.rb Formula/
git add Formula/versionator.rb
git commit -m "Add versionator formula v<VERSION>"
git push
```

**Option B: Automated (Recommended)**

Add a step to the release workflow to automatically push the formula. This requires a Personal Access Token (PAT) with repo scope.

1. Create a PAT at https://github.com/settings/tokens with `repo` scope
2. Add it as a repository secret named `HOMEBREW_TAP_TOKEN`
3. Update `.github/workflows/release.yml`:

```yaml
  update-homebrew:
    runs-on: ubuntu-latest
    needs: release
    if: success()

    steps:
    - name: Checkout tap repository
      uses: actions/checkout@v4
      with:
        repository: benjaminabbitt/homebrew-tap
        token: ${{ secrets.HOMEBREW_TAP_TOKEN }}
        path: homebrew-tap

    - name: Download formula artifact
      uses: actions/download-artifact@v4
      with:
        name: homebrew-formula
        path: formula/

    - name: Update formula
      run: |
        cp formula/versionator.rb homebrew-tap/Formula/
        cd homebrew-tap
        git config user.name "GitHub Actions"
        git config user.email "actions@github.com"
        git add Formula/versionator.rb
        git commit -m "Update versionator to ${{ needs.release.outputs.version }}"
        git push
```

## Usage

Once set up, users can install versionator via:

```bash
# Add the tap (one-time)
brew tap benjaminabbitt/tap

# Install versionator
brew install versionator

# Or in one command
brew install benjaminabbitt/tap/versionator
```

## Updating

When a new release is created:

1. The release workflow generates an updated formula with new version and checksums
2. The formula is uploaded as a workflow artifact
3. (If automated) The formula is automatically pushed to homebrew-tap
4. (If manual) Download artifact and push manually

## Testing the Formula

Before publishing, test the formula locally:

```bash
# Install from local file
brew install --build-from-source ./Formula/versionator.rb

# Run formula audit
brew audit --strict Formula/versionator.rb

# Test the formula
brew test versionator
```

## Troubleshooting

### "No available formula" error
- Ensure the tap repository is public
- Check that the formula file is in `Formula/` directory
- Verify the formula filename matches the formula class name (lowercase)

### Checksum mismatch
- Regenerate the formula from the release workflow
- Verify the release assets haven't been modified

### Formula audit failures
- Run `brew audit --strict --online Formula/versionator.rb`
- Fix any issues reported by the audit

## Formula Structure

The generated formula supports:
- macOS (Intel and Apple Silicon)
- Linux (x64 and arm64)

```ruby
class Versionator < Formula
  desc "A semantic version management CLI tool"
  homepage "https://github.com/benjaminabbitt/versionator"
  version "X.Y.Z"
  license "BSD-3-Clause"

  on_macos do
    if Hardware::CPU.arm?
      url "...darwin-arm64"
      sha256 "..."
    else
      url "...darwin-amd64"
      sha256 "..."
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "...linux-arm64"
      sha256 "..."
    else
      url "...linux-amd64"
      sha256 "..."
    end
  end

  def install
    bin.install "versionator-<platform>" => "versionator"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/versionator version")
  end
end
```

## References

- [Homebrew Tap Documentation](https://docs.brew.sh/Taps)
- [Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Acceptable Formulae](https://docs.brew.sh/Acceptable-Formulae)

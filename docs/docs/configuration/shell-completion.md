---
title: Shell Completion
description: Set up tab completion for versionator
sidebar_position: 2
---

# Shell Completion

Versionator provides shell completion scripts for Bash, Zsh, Fish, and PowerShell.

## Bash

### Current Session

```bash
source <(versionator completion bash)
```

### Permanent Installation

**Linux:**

```bash
versionator completion bash > /etc/bash_completion.d/versionator
```

**macOS with Homebrew:**

```bash
versionator completion bash > $(brew --prefix)/etc/bash_completion.d/versionator
```

**Manual:**

```bash
# Add to ~/.bashrc
echo 'source <(versionator completion bash)' >> ~/.bashrc
```

## Zsh

### Enable Completions

First, ensure Zsh completion is enabled:

```zsh
echo "autoload -U compinit; compinit" >> ~/.zshrc
```

### Install Completion

```zsh
versionator completion zsh > "${fpath[1]}/_versionator"
```

Or add to `~/.zshrc`:

```zsh
source <(versionator completion zsh)
```

### Oh My Zsh

```zsh
versionator completion zsh > ~/.oh-my-zsh/completions/_versionator
```

## Fish

### Current Session

```fish
versionator completion fish | source
```

### Permanent Installation

```fish
versionator completion fish > ~/.config/fish/completions/versionator.fish
```

## PowerShell

### Current Session

```powershell
versionator completion powershell | Out-String | Invoke-Expression
```

### Permanent Installation

Add to your PowerShell profile:

```powershell
versionator completion powershell >> $PROFILE
```

Or create a dedicated file:

```powershell
versionator completion powershell > "$HOME\Documents\WindowsPowerShell\versionator.ps1"
# Add to profile:
echo '. "$HOME\Documents\WindowsPowerShell\versionator.ps1"' >> $PROFILE
```

## Completion Features

Once installed, tab completion provides:

- **Command completion**: `versionator <TAB>` shows available commands
- **Subcommand completion**: `versionator prefix <TAB>` shows subcommands
- **Flag completion**: `versionator version --<TAB>` shows available flags
- **Flag value completion**: For flags with known values

### Examples

```bash
$ versionator <TAB>
completion  config      custom      emit        help        major
metadata    minor       patch       prefix      prerelease  schema
tag         vars        version

$ versionator prefix <TAB>
disable  enable  set  status

$ versionator emit <TAB>
c         c-header   cpp       cpp-header  csharp    dump
go        java       js        json        kotlin    php
python    ruby       rust      swift       ts        yaml

$ versionator version --<TAB>
--help        --metadata    --prefix      --prerelease  --template
```

## Troubleshooting

### Bash: Completions not loading

Ensure bash-completion is installed:

```bash
# Debian/Ubuntu
sudo apt install bash-completion

# macOS with Homebrew
brew install bash-completion@2
```

### Zsh: compdef not found

Add this before the completion source:

```zsh
autoload -Uz compinit && compinit
```

### Fish: Completions not working

Verify the completions directory exists:

```fish
mkdir -p ~/.config/fish/completions
```

### PowerShell: Execution Policy

You may need to adjust the execution policy:

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

## Updating Completions

After upgrading versionator, regenerate completions to get new commands:

```bash
# Bash
versionator completion bash > /etc/bash_completion.d/versionator

# Zsh
versionator completion zsh > "${fpath[1]}/_versionator"

# Fish
versionator completion fish > ~/.config/fish/completions/versionator.fish
```

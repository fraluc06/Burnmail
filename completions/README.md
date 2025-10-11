# Shell Completions for Burnmail

This directory contains shell completion scripts for bash, zsh, and fish.

## Installation

### Bash

**Linux:**
```bash
sudo cp burnmail.bash /etc/bash_completion.d/burnmail
```

**macOS (Homebrew):**
```bash
cp burnmail.bash $(brew --prefix)/etc/bash_completion.d/burnmail
```

**Or load directly:**
```bash
source <(burnmail completion bash)
```

### Zsh

```bash
# Add to your fpath (usually ~/.zfunc or first directory in $fpath)
mkdir -p ~/.zfunc
cp burnmail.zsh ~/.zfunc/_burnmail

# Add to ~/.zshrc if not already present:
# fpath=(~/.zfunc $fpath)
# autoload -Uz compinit && compinit
```

**Or load directly:**
```bash
source <(burnmail completion zsh)
```

### Fish

```bash
cp burnmail.fish ~/.config/fish/completions/burnmail.fish
```

**Or load directly:**
```bash
burnmail completion fish | source
```

## Regenerating

To regenerate the completion files:

```bash
burnmail completion bash > burnmail.bash
burnmail completion zsh > burnmail.zsh
burnmail completion fish > burnmail.fish
```

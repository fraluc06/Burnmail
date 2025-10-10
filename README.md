# Burnmail

🔥 Burn through temporary emails straight from your terminal - written in Go.

A Go implementation inspired by [Mailsy](https://github.com/BalliAsghar/Mailsy), using the [mail.tm](https://mail.tm) API.

## Features

- 📧 Generate random disposable email addresses
- 📬 View and read messages interactively
- 🗑️ Delete accounts when done
- 📋 Automatic clipboard copy
- 🎨 Beautiful CLI interface with colors
- 🔒 No dependencies at runtime (single binary)

## Installation

### From Source

```bash
git clone https://github.com/fraluc06/burnmail.git
cd burnmail

go build -o burnmail

sudo mv burnmail /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/fraluc06/burnmail@latest
```

## Quick Start

Here's a typical workflow using burnmail:

```bash
# 1. Generate new email
$ burnmail g

🔍 Fetching available domains...
📧 Creating email address...

✓ Email created and copied to clipboard!

x9k2m5p7@mail.tm


# 2. Use the email to register (already in clipboard)
# ... go to a website, register with Ctrl+V ...


# 3. Wait a few seconds, then check inbox
$ burnmail m

📬 Fetching messages...

Select an email (Use arrow keys)
▸ Verify your email address - from noreply@example.com

# Press Enter to read


📖 Loading message...

────────────────────────────────────────────────────────────
From: noreply@example.com
Subject: Verify your email address
Date: 10/10/2025 14:38:15
────────────────────────────────────────────────────────────

Please verify your email by clicking this link:
https://example.com/verify?code=ABC123XYZ


# 4. After completing verification, delete the account
$ burnmail d

✓ Account deleted successfully
```

## Usage

### Generate a new email address

```bash
burnmail g
# or
burnmail generate
```

**Output:**

```
$ burnmail g

🔍 Fetching available domains...
📧 Creating email address...

✓ Email created and copied to clipboard!

a3f9b2c1@mail.tm
```

This will:
- Create a new temporary email address
- Copy it to your clipboard automatically
- Save it locally in `~/.burnmail.json` for future use

If an account already exists:
```
⚠ Account already exists: xyz123@mail.tm
Use 'burnmail d' to delete it first.
```

### View your inbox

```bash
burnmail m
# or
burnmail messages
```

**Output with messages:**

```
$ burnmail m

📬 Fetching messages...

Select an email (Use arrow keys)
▸ Welcome to our service! - from hello@example.com
  Your verification code is 123456 - from noreply@verify.com
  Special offer just for you - from marketing@shop.com
```

**After selecting a message:**

```
📖 Loading message...

────────────────────────────────────────────────────────────
From: hello@example.com
Subject: Welcome to our service!
Date: 10/10/2025 14:35:22
────────────────────────────────────────────────────────────

Hello,

Welcome to our amazing service! We're glad to have you here.

To get started, please verify your email by clicking the link below:
https://example.com/verify?token=abc123

Best regards,
The Team
```

**Navigation:**
- `↑↓` Arrow keys: Navigate between messages
- `Enter`: Select and read message
- `Esc` / `Ctrl+C`: Exit without selecting

**Empty inbox:**
```
📭 No messages yet. Your inbox is empty.
```

### Show account details

```bash
burnmail me
```

**Output:**

```
$ burnmail me

Email: a3f9b2c1@mail.tm
Created At: 10/10/2025, 14:30:45
```

### Delete your account

```bash
burnmail d
# or
burnmail delete
```

**Output:**

```
$ burnmail d

✓ Account deleted successfully
```

This will delete the account from mail.tm and remove local data.

## Troubleshooting

### Rate limit exceeded

```
✗ Failed to create account: status 429, body: {"message":"Rate limit exceeded"}
```

**Solution:** Wait a few minutes before trying again.

### Token expired

```
✗ Failed to get messages: status 401
```

**Solution:** Regenerate your account:
```bash
burnmail d
burnmail g
```

### No internet connection

```
✗ Failed to get domains: Get "https://api.mail.tm/domains": dial tcp: lookup api.mail.tm: no such host
```

**Solution:** Check your internet connection.

### Clipboard not available (Linux)

```
✓ Email created!

x9k2m5p7@mail.tm

# Note: email NOT copied to clipboard because xclip is not installed
```

**Solution:** Install clipboard support:
```bash
# Debian/Ubuntu
sudo apt install xclip

# Fedora
sudo dnf install xclip

# Arch
sudo pacman -S xclip
```

### No account found

```
✗ No account found. Generate one first with 'burnmail g'
```

**Solution:** Generate a new account first with `burnmail g`.

## Technical Details

### Performance

Typical command execution times:

```
burnmail g          ~1.5s   (fetch domains + create + login + save)
burnmail m          ~0.8s   (load + fetch messages + show UI)
burnmail me         ~0.01s  (local file read only)
burnmail d          ~0.5s   (delete API + remove file)
```

Times vary based on internet latency and mail.tm API load.

### Configuration file

Account data is stored in `~/.burnmail.json`:

```json
{
  "address": "x9k2m5p7@mail.tm",
  "password": "1a2b3c4d5e6f7g8h",
  "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",
  "accountId": "507f1f77bcf86cd799439011",
  "createdAt": "10/10/2025, 14:30:45"
}
```

⚠️ **DO NOT share this file** - it contains your credentials!

### HTML messages

If an email contains only HTML (no text), it will automatically open in your default browser.

### Terminal colors

The output uses ANSI colors for better readability:

- 🔵 **Cyan**: Info and operations in progress
- 🟢 **Green**: Successful operations
- 🟡 **Yellow**: Warnings and suggestions
- 🔴 **Red**: Errors

If your terminal doesn't support colors, you'll see plain text only.

### Binary size

```
burnmail (non-optimized)    ~8-10 MB
burnmail (with -ldflags)    ~4-6 MB
burnmail.exe (Windows)      ~5-7 MB
```

The binary is standalone - no runtime dependencies needed!

## How it works

Burnmail Go uses the [mail.tm](https://mail.tm) API to:
1. Create temporary email addresses
2. Receive and read emails
3. Manage your temporary inbox

All account data is stored locally in `~/.burnmail.json`.

## Building

```bash
# For your current platform
go build -o burnmail

# Using build scripts:
# Linux/macOS:
./build.sh all

# Windows PowerShell:
.\build.ps1 all

# Manual cross-compilation:
# For Linux
GOOS=linux GOARCH=amd64 go build -o burnmail-linux

# For macOS
GOOS=darwin GOARCH=amd64 go build -o burnmail-macos

# For Windows
GOOS=windows GOARCH=amd64 go build -o burnmail.exe
```

## Requirements

- Go 1.21 or higher (for building)
- Internet connection (for API access)
- Optional: `xclip` or `xsel` on Linux for clipboard support

## Dependencies

- [cobra](https://github.com/spf13/cobra) - CLI framework
- [promptui](https://github.com/manifoldco/promptui) - Interactive prompts
- [clipboard](https://github.com/atotto/clipboard) - Clipboard operations
- [color](https://github.com/fatih/color) - Terminal colors

## License

MIT License - feel free to use and modify!

## Credits

- Inspired by [Mailsy](https://github.com/BalliAsghar/Mailsy) by BalliAsghar
- Powered by [mail.tm](https://mail.tm) API

## Contributing

Pull requests are welcome! For major changes, please open an issue first to discuss what you would like to change.

# Burnmail

🔥 Temporary email addresses in your terminal - written in Go.

Inspired by [Mailsy](https://github.com/BalliAsghar/Mailsy), using the [mail.tm](https://mail.tm) API.

## Features

- 📧 Generate random disposable emails
- 📬 Read messages interactively
- 📋 Auto-copy to clipboard
- 🔒 Single binary, no runtime deps

## Installation

```bash
go install github.com/fraluc06/burnmail@latest
```

Or build from source:
```bash
git clone https://github.com/fraluc06/burnmail.git
cd burnmail
go build -o burnmail
sudo mv burnmail /usr/local/bin/
```

## Usage

```bash
# Generate email (auto-copied to clipboard)
burnmail g

# Check inbox
burnmail m

# Show account
burnmail me

# Delete account
burnmail d
```

## Example

```bash
$ burnmail g
✓ Email created and copied to clipboard!
x9k2m5p7@mail.tm

$ burnmail m
📬 Fetching messages...
Select an email (Use arrow keys)
▸ Verify your email address - from noreply@example.com

# Press Enter to read the message
```

## Troubleshooting

**Rate limit exceeded** - Wait a few minutes

**Token expired** - Regenerate: `burnmail d && burnmail g`

**Clipboard not working (Linux)** - Install xclip: `sudo apt install xclip`

## Development

```bash
# Build
make build

# Test
make test

# Cross-compile for all platforms
./build.sh all  # Linux/macOS
.\build.ps1 all # Windows
```

## License

This project is open source and available under the [MIT License](LICENSE).

## Credits

Inspired by [Mailsy](https://github.com/BalliAsghar/Mailsy) • Powered by [mail.tm](https://mail.tm)

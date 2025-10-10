# Burnmail - Command Output Examples

This file shows what the output of each command looks like.

═══════════════════════════════════════════════════════════════════════════════

## Command: burnmail g (Generate new email)

```
$ burnmail g

🔍 Fetching available domains...
📧 Creating email address...

✓ Email created and copied to clipboard!

a3f9b2c1@mail.tm

```

If an account already exists:
```
$ burnmail g

⚠ Account already exists: xyz123@mail.tm
Use 'burnmail d' to delete it first.
```

═══════════════════════════════════════════════════════════════════════════════

## Command: burnmail m (View messages)

With messages in inbox:
```
$ burnmail m

📬 Fetching messages...

Select an email (Use arrow keys)
▸ Welcome to our service! - from hello@example.com
  Your verification code is 123456 - from noreply@verify.com
  Special offer just for you - from marketing@shop.com

```

After selecting a message:
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

Empty inbox:
```
$ burnmail m

📬 Fetching messages...

📭 No messages yet. Your inbox is empty.
```

No account:
```
$ burnmail m

✗ No account found. Generate one first with 'burnmail g'
```

═══════════════════════════════════════════════════════════════════════════════

## Command: burnmail me (Account info)

```
$ burnmail me

Email: a3f9b2c1@mail.tm
Created At: 10/10/2025, 14:30:45

```

No account:
```
$ burnmail me

✗ No account found. Generate one first with 'burnmail g'
```

═══════════════════════════════════════════════════════════════════════════════

## Command: burnmail d (Delete account)

```
$ burnmail d

✓ Account deleted successfully
```

No account:
```
$ burnmail d

✗ No account found. Generate one first with 'burnmail g'
```

═══════════════════════════════════════════════════════════════════════════════

## Command: burnmail --help

```
$ burnmail --help

🔥 Burn through temporary emails straight from your terminal

Burnmail is a CLI tool to quickly generate and manage disposable email 
addresses using mail.tm API.

Usage:
  burnmail [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  d           Delete the current account
  g           Generate a new disposable email address
  help        Help about any command
  m           View inbox messages
  me          Show account details

Flags:
  -h, --help   help for burnmail

Use "burnmail [command] --help" for more information about a command.
```

═══════════════════════════════════════════════════════════════════════════════

## Complete workflow example

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


# 5. Generate a new account for another site
$ burnmail g

🔍 Fetching available domains...
📧 Creating email address...

✓ Email created and copied to clipboard!

b4c8d1e6@mail.tm

```

═══════════════════════════════════════════════════════════════════════════════

## Error handling

### Error: Rate limit

```
$ burnmail g

🔍 Fetching available domains...
📧 Creating email address...
✗ Failed to create account: status 429, body: {"message":"Rate limit exceeded"}

# Solution: wait a few minutes
```

### Error: Token expired

```
$ burnmail m

📬 Fetching messages...
✗ Failed to get messages: status 401

# Solution: regenerate account
$ burnmail d
✓ Account deleted successfully

$ burnmail g
✓ Email created and copied to clipboard!
```

### Error: No internet connection

```
$ burnmail g

🔍 Fetching available domains...
✗ Failed to get domains: Get "https://api.mail.tm/domains": dial tcp: lookup api.mail.tm: no such host

# Solution: check internet connection
```

### Error: Clipboard not available (Linux without xclip)

```
$ burnmail g

🔍 Fetching available domains...
📧 Creating email address...

✓ Email created!

x9k2m5p7@mail.tm

# Note: email NOT copied to clipboard because xclip is not installed
# Solution: sudo apt install xclip
```

═══════════════════════════════════════════════════════════════════════════════

## Interactive menu navigation (burnmail m)

When viewing messages, you can use:

```
Arrow keys ↑↓     : Navigate between messages
Enter             : Select and read message
Esc / Ctrl+C      : Exit without selecting
```

Navigation example:
```
Select an email (Use arrow keys)
  First message - from user1@mail.com
▸ Second message - from user2@mail.com    ← Cursor here
  Third message - from user3@mail.com
  Fourth message - from user4@mail.com

# Press ↓ to go down, ↑ to go up
# Press Enter to read
```

═══════════════════════════════════════════════════════════════════════════════

## Terminal colors

The output uses ANSI colors for better readability:

- 🔵 Cyan (blue): Info and operations in progress
- 🟢 Green: Successful operations
- 🟡 Yellow: Warnings and suggestions
- 🔴 Red: Errors

If your terminal doesn't support colors, you'll see plain text only.

═══════════════════════════════════════════════════════════════════════════════

## File ~/.burnmail.json

After generating an account, you'll see this file:

```bash
$ cat ~/.burnmail.json
{
  "address": "x9k2m5p7@mail.tm",
  "password": "1a2b3c4d5e6f7g8h",
  "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",
  "accountId": "507f1f77bcf86cd799439011",
  "createdAt": "10/10/2025, 14:30:45"
}
```

⚠️ DO NOT share this file - it contains your credentials!

═══════════════════════════════════════════════════════════════════════════════

## HTML messages

If an email contains only HTML (no text), it opens in browser:

```
$ burnmail m
📬 Fetching messages...

Select an email
▸ Newsletter with fancy graphics - from marketing@shop.com

[Press Enter]

📖 Loading message...

────────────────────────────────────────────────────────────
From: marketing@shop.com
Subject: Newsletter with fancy graphics
Date: 10/10/2025 15:00:00
────────────────────────────────────────────────────────────

[HTML content - opening in browser...]

# Opens automatically in default browser
```

═══════════════════════════════════════════════════════════════════════════════

## Typical performance

```
burnmail g          ~1.5s   (fetch domains + create + login + save)
burnmail m          ~0.8s   (load + fetch messages + show UI)
burnmail me         ~0.01s  (local file read only)
burnmail d          ~0.5s   (delete API + remove file)
```

Times vary based on internet latency and mail.tm API load.

═══════════════════════════════════════════════════════════════════════════════

## Binary size

```
burnmail (non-optimized)    ~8-10 MB
burnmail (with -ldflags)    ~4-6 MB
burnmail.exe (Windows)      ~5-7 MB
```

The binary is standalone - no runtime dependencies needed!

═══════════════════════════════════════════════════════════════════════════════

# Burnmail API Examples

Examples of how to use API functions directly in Go code.

## Example 1: Create an account programmatically

```go
package main

import (
    "fmt"
    "burnmail/api"
)

func main() {
    client := api.NewClient()
    
    // Get available domains
    domains, err := client.GetDomains()
    if err != nil {
        panic(err)
    }
    
    // Create account
    address := "test123@" + domains[0].Domain
    password := "MySecurePassword123!"
    
    account, err := client.CreateAccount(address, password)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Account created: %s (ID: %s)\n", account.Address, account.ID)
    
    // Login to get token
    token, err := client.Login(address, password)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Token: %s\n", token)
}
```

## Example 2: Read messages

```go
package main

import (
    "fmt"
    "burnmail/api"
)

func main() {
    client := api.NewClient()
    client.Token = "your-token-here"
    
    // Get all messages
    messages, err := client.GetMessages()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("You have %d messages:\n\n", len(messages))
    
    for i, msg := range messages {
        fmt.Printf("%d. %s\n", i+1, msg.Subject)
        fmt.Printf("   From: %s\n", msg.From.Address)
        fmt.Printf("   Date: %s\n\n", msg.CreatedAt.Format("02/01/2006 15:04"))
    }
}
```

## Example 3: Read a specific message

```go
package main

import (
    "fmt"
    "burnmail/api"
)

func main() {
    client := api.NewClient()
    client.Token = "your-token-here"
    
    messageID := "message-id-here"
    
    // Get full message
    message, err := client.GetMessage(messageID)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Subject: %s\n", message.Subject)
    fmt.Printf("From: %s\n", message.From.Address)
    fmt.Printf("Date: %s\n\n", message.CreatedAt.Format("02/01/2006 15:04:05"))
    fmt.Printf("Content:\n%s\n", message.Text)
}
```

## Example 4: Delete an account

```go
package main

import (
    "fmt"
    "burnmail/api"
)

func main() {
    client := api.NewClient()
    client.Token = "your-token-here"
    
    accountID := "account-id-here"
    
    err := client.DeleteAccount(accountID)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Account deleted successfully!")
}
```

## Example 5: Complete usage with storage

```go
package main

import (
    "fmt"
    "burnmail/api"
    "burnmail/storage"
    "time"
)

func main() {
    // Create account
    client := api.NewClient()
    
    domains, _ := client.GetDomains()
    address := "mytest@" + domains[0].Domain
    password := "SecurePass123!"
    
    account, _ := client.CreateAccount(address, password)
    token, _ := client.Login(address, password)
    
    // Save to storage
    accountData := &storage.AccountData{
        Address:   address,
        Password:  password,
        Token:     token,
        AccountID: account.ID,
        CreatedAt: time.Now().Format("02/01/2006, 15:04:05"),
    }
    
    storage.Save(accountData)
    fmt.Println("Account saved!")
    
    // Later, load from storage
    loadedAccount, _ := storage.Load()
    fmt.Printf("Loaded account: %s\n", loadedAccount.Address)
    
    // Use the loaded account
    client.Token = loadedAccount.Token
    messages, _ := client.GetMessages()
    fmt.Printf("Messages: %d\n", len(messages))
    
    // Cleanup
    client.DeleteAccount(loadedAccount.AccountID)
    storage.Delete()
    fmt.Println("Cleaned up!")
}
```

## Example 6: Monitor new emails

```go
package main

import (
    "fmt"
    "burnmail/api"
    "time"
)

func main() {
    client := api.NewClient()
    client.Token = "your-token-here"
    
    fmt.Println("Monitoring inbox...")
    
    lastCount := 0
    
    for {
        messages, err := client.GetMessages()
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            time.Sleep(10 * time.Second)
            continue
        }
        
        if len(messages) > lastCount {
            fmt.Printf("\n🔔 New message received!\n")
            newMsg := messages[0] // Most recent
            fmt.Printf("Subject: %s\n", newMsg.Subject)
            fmt.Printf("From: %s\n\n", newMsg.From.Address)
            lastCount = len(messages)
        }
        
        time.Sleep(10 * time.Second)
    }
}
```

## Main data structures

### Domain
```go
type Domain struct {
    ID        string
    Domain    string
    IsActive  bool
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Account
```go
type Account struct {
    ID        string
    Address   string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Message
```go
type Message struct {
    ID          string
    AccountID   string
    MsgID       string
    From        From
    To          []To
    Subject     string
    Intro       string
    Seen        bool
    IsDeleted   bool
    HasAttach   bool
    Size        int
    DownloadURL string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### MessageDetail
```go
type MessageDetail struct {
    Message
    Text        string
    HTML        []string
    // ... other fields
}
```

## API Endpoints used

- `GET /domains` - List available domains
- `POST /accounts` - Create new account
- `POST /token` - Get authentication token
- `GET /messages` - List messages
- `GET /messages/:id` - Get specific message
- `DELETE /accounts/:id` - Delete account
- `GET /accounts/:id` - Get account info

## Rate Limiting

The mail.tm API has limits:
- Account creation: ~10 per hour
- API requests: ~100 per hour

The client automatically handles timeouts and retries.

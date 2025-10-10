package cmd

import (
	"burnmail/api"
	"burnmail/storage"
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

const (
	htmlFileCleanupDelay = 30 * time.Second
	retryMaxAttempts     = 3
	retryBaseDelay       = 1 * time.Second
	retryMaxDelay        = 10 * time.Second
	requestTimeout       = 30 * time.Second
)

var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
)

var (
	Version string

	rootCmd = &cobra.Command{
		Use:     "burnmail",
		Short:   "🔥 Burn through temporary emails straight from your terminal",
		Long:    `Burnmail is a CLI tool to quickly generate and manage disposable email addresses using mail.tm API.`,
		Version: Version,
	}
)

var generateCmd = &cobra.Command{
	Use:     "g",
	Aliases: []string{"generate"},
	Short:   "Generate a new disposable email address",
	Run:     generateEmail,
}

var messagesCmd = &cobra.Command{
	Use:     "m",
	Aliases: []string{"messages", "inbox"},
	Short:   "View inbox messages (interactive TUI)",
	Run:     viewMessagesTUI,
}

var messagesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List messages (classic view)",
	Run:   viewMessages,
}

var deleteCmd = &cobra.Command{
	Use:     "d",
	Aliases: []string{"delete"},
	Short:   "Delete the current account",
	Run:     deleteAccount,
}

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Show account details",
	Run:   showAccount,
}

func init() {
	rootCmd.SetVersionTemplate("Burnmail v{{.Version}}\n")
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(messagesCmd)
	messagesCmd.AddCommand(messagesListCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(meCmd)
}

func Execute() {
	rootCmd.Version = Version
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generateEmail(_ *cobra.Command, _ []string) {
	if storage.Exists() {
		existingAccount, _ := storage.Load()
		if existingAccount != nil {
			fmt.Printf("%s Account already exists: %s\n", yellow("⚠"), cyan(existingAccount.Address))
			fmt.Printf("Use '%s' to delete it first.\n", yellow("burnmail d"))
			return
		}
	}

	fmt.Println(cyan("🔍 Fetching available domains..."))

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	client := api.GetClient()

	domains, err := retryWithBackoff(ctx, func() (interface{}, error) {
		return client.GetDomains()
	})
	if err != nil {
		fmt.Printf("%s Failed to get domains: %v\n", red("✗"), err)
		return
	}

	domainList := domains.([]api.Domain)
	if len(domainList) == 0 {
		fmt.Printf("%s No domains available\n", red("✗"))
		return
	}

	var selectedDomain string
	for _, d := range domainList {
		if d.IsActive {
			selectedDomain = d.Domain
			break
		}
	}

	if selectedDomain == "" {
		fmt.Printf("%s No active domains found\n", red("✗"))
		return
	}

	username := generateRandomString(8)
	address := username + "@" + selectedDomain
	password := generateRandomString(16)

	fmt.Println(cyan("📧 Creating email address..."))

	account, err := retryWithBackoff(ctx, func() (interface{}, error) {
		return client.CreateAccount(address, password)
	})
	if err != nil {
		fmt.Printf("%s Failed to create account: %v\n", red("✗"), err)
		return
	}

	token, err := retryWithBackoff(ctx, func() (interface{}, error) {
		return client.Login(address, password)
	})
	if err != nil {
		fmt.Printf("%s Failed to login: %v\n", red("✗"), err)
		return
	}

	accountData := &storage.AccountData{
		Address:   address,
		Password:  password,
		Token:     token.(string),
		AccountID: account.(*api.Account).ID,
		CreatedAt: time.Now().Format("02/01/2006, 15:04:05"),
	}

	if err := storage.Save(accountData); err != nil {
		fmt.Printf("%s Failed to save account: %v\n", red("✗"), err)
		return
	}

	if err := clipboard.WriteAll(address); err == nil {
		fmt.Printf("\n%s Email created and copied to clipboard!\n", green("✓"))
	} else {
		fmt.Printf("\n%s Email created!\n", green("✓"))
		fmt.Printf("%s Warning: Failed to copy to clipboard: %v\n", yellow("⚠"), err)
	}

	fmt.Printf("\n%s\n\n", green(address))
}

func viewMessages(_ *cobra.Command, _ []string) {
	accountData, err := storage.Load()
	if err != nil || accountData == nil {
		fmt.Printf("%s No account found. Generate one first with '%s'\n", red("✗"), yellow("burnmail g"))
		return
	}

	client := api.GetClient()
	client.SetToken(accountData.Token)

	fmt.Println(cyan("📬 Fetching messages..."))

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	result, err := retryWithBackoff(ctx, func() (interface{}, error) {
		return client.GetMessages()
	})
	if err != nil {
		fmt.Printf("%s Failed to get messages: %v\n", red("✗"), err)
		return
	}

	messages := result.([]api.Message)

	if len(messages) == 0 {
		fmt.Printf("\n%s No messages yet. Your inbox is empty.\n", yellow("📭"))
		return
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "▸ {{ .Subject | cyan }} - from {{ .From.Address | yellow }}",
		Inactive: "  {{ .Subject | cyan }} - from {{ .From.Address | yellow }}",
		Selected: "{{ .Subject | green }}",
	}

	prompt := promptui.Select{
		Label:     "Select an email",
		Items:     messages,
		Templates: templates,
		Size:      10,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return
	}

	selectedMessage := messages[idx]

	fmt.Println(cyan("\n📖 Loading message..."))
	fullMessage, err := client.GetMessage(selectedMessage.ID)
	if err != nil {
		fmt.Printf("%s Failed to get message: %v\n", red("✗"), err)
		return
	}

	fmt.Printf("\n%s\n", strings.Repeat("─", 60))
	fmt.Printf("%s: %s\n", cyan("From"), fullMessage.From.Address)
	fmt.Printf("%s: %s\n", cyan("Subject"), fullMessage.Subject)
	fmt.Printf("%s: %s\n", cyan("Date"), fullMessage.CreatedAt.Format("02/01/2006 15:04:05"))
	fmt.Printf("%s\n\n", strings.Repeat("─", 60))

	if fullMessage.Text != "" {
		fmt.Println(fullMessage.Text)
	} else if len(fullMessage.HTML) > 0 {
		fmt.Println(cyan("\n[HTML content - opening in browser...]"))
		openInBrowser(fullMessage)
	}

	fmt.Println()
}

func viewMessagesTUI(_ *cobra.Command, _ []string) {
	accountData, err := storage.Load()
	if err != nil || accountData == nil {
		fmt.Printf("%s No account found. Generate one first with '%s'\n", red("✗"), yellow("burnmail g"))
		return
	}

	client := api.GetClient()

	if err := runTUI(accountData, client); err != nil {
		fmt.Printf("%s TUI error: %v\n", red("✗"), err)
	}
}

func deleteAccount(_ *cobra.Command, _ []string) {
	accountData, err := storage.Load()
	if err != nil || accountData == nil {
		fmt.Printf("%s No account found. Generate one first with '%s'\n", red("✗"), yellow("burnmail g"))
		return
	}

	client := api.GetClient()
	client.SetToken(accountData.Token)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	_, err = retryWithBackoff(ctx, func() (interface{}, error) {
		return nil, client.DeleteAccount(accountData.AccountID)
	})
	if err != nil {
		fmt.Printf("%s Failed to delete account from server: %v\n", yellow("⚠"), err)
	}

	if err := storage.Delete(); err != nil {
		fmt.Printf("%s Failed to delete local data: %v\n", red("✗"), err)
		return
	}

	fmt.Printf("%s Account deleted successfully\n", green("✓"))
}

func showAccount(_ *cobra.Command, _ []string) {
	accountData, err := storage.Load()
	if err != nil || accountData == nil {
		fmt.Printf("%s No account found. Generate one first with '%s'\n", red("✗"), yellow("burnmail g"))
		return
	}

	fmt.Printf("\n%s: %s\n", cyan("Email"), accountData.Address)
	fmt.Printf("%s: %s\n\n", cyan("Created At"), accountData.CreatedAt)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes)
}

func openInBrowser(message *api.MessageDetail) {
	tmpFile, err := os.CreateTemp("", "burnmail-*.html")
	if err != nil {
		return
	}
	tmpFilePath := tmpFile.Name()

	var htmlBuilder strings.Builder
	for _, h := range message.HTML {
		htmlBuilder.WriteString(h)
	}

	if _, err := tmpFile.WriteString(htmlBuilder.String()); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFilePath)
		return
	}
	_ = tmpFile.Close()

	go func() {
		time.Sleep(htmlFileCleanupDelay)
		_ = os.Remove(tmpFilePath)
	}()

	var execCmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		execCmd = exec.Command("open", tmpFilePath)
	case "linux":
		execCmd = exec.Command("xdg-open", tmpFilePath)
	case "windows":
		execCmd = exec.Command("cmd", "/c", "start", tmpFilePath)
	default:
		_ = os.Remove(tmpFilePath)
		return
	}

	if err := execCmd.Start(); err != nil {
		_ = os.Remove(tmpFilePath)
	}
}

func retryWithBackoff(ctx context.Context, fn func() (any, error)) (any, error) {
	for attempt := 0; attempt < retryMaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		if attempt == retryMaxAttempts-1 {
			return nil, err
		}

		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "rate limit") {
			delay := min(time.Duration(math.Pow(2, float64(attempt)))*retryBaseDelay, retryMaxDelay)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		} else {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries exceeded")
}

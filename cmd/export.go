package cmd

import (
	"burnmail/api"
	"burnmail/storage"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type ExportData struct {
	Account    *storage.AccountData `json:"account"`
	Messages   []MessageExport      `json:"messages"`
	ExportedAt string               `json:"exportedAt"`
}

type MessageExport struct {
	*api.MessageDetail
	IsIncluded bool `json:"isIncluded"`
}

func exportData(_ *cobra.Command, _ []string) {
	accountData := loadAccountOrExit()
	if accountData == nil {
		return
	}

	client := api.GetClient()
	client.SetToken(accountData.Token)

	messages, success := fetchMessages(client)
	if !success {
		return
	}

	if len(messages) == 0 {
		fmt.Printf("\n%s No messages to export. Your inbox is empty.\n", yellow("📭"))
		return
	}

	fmt.Printf("%s Found %d messages. Fetching details...\n", cyan("📖"), len(messages))

	exportedMessages := make([]MessageExport, 0, len(messages))
	for i, msg := range messages {
		fmt.Printf("\r%s Fetching message %d/%d...", cyan("⏳"), i+1, len(messages))

		fullMessage, err := client.GetMessage(msg.ID)
		if err != nil {
			fmt.Printf("\n%s Failed to fetch message %s: %v\n", yellow("⚠"), msg.ID, err)
			continue
		}

		exportedMessages = append(exportedMessages, MessageExport{
			MessageDetail: fullMessage,
			IsIncluded:    true,
		})
	}
	fmt.Println() // New line after progress

	exportDataStruct := ExportData{
		Account:    accountData,
		Messages:   exportedMessages,
		ExportedAt: time.Now().Format("02/01/2006, 15:04:05"),
	}

	// Create filename with email address and timestamp
	filename := fmt.Sprintf("burnmail_export_%s_%d.json", strings.ReplaceAll(accountData.Address, "@", "_"), time.Now().Unix())

	jsonData, err := json.MarshalIndent(exportDataStruct, "", "  ")
	if err != nil {
		fmt.Printf("%s Failed to marshal export data: %v\n", red("✗"), err)
		return
	}

	if err := os.WriteFile(filename, jsonData, 0600); err != nil {
		fmt.Printf("%s Failed to write export file: %v\n", red("✗"), err)
		return
	}

	absPath, _ := os.Getwd()
	fullPath := fmt.Sprintf("%s/%s", absPath, filename)

	fmt.Printf("\n%s Export completed successfully!\n", green("✓"))
	fmt.Printf("%s File: %s\n", cyan("💾"), filename)
	fmt.Printf("%s Messages exported: %d\n", cyan("📧"), len(exportedMessages))
	fmt.Printf("%s Full path: %s\n\n", cyan("📍"), fullPath)
}

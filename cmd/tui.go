package cmd

import (
	"burnmail/api"
	"burnmail/storage"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaytaylor/html2text"
)

type view int

const (
	listView view = iota
	detailView
	helpView
	confirmView
)

const (
	autoRefreshInterval = 10 * time.Second
	cacheFileName       = ".burnmail-cache.json"
	cacheExpiry         = 5 * time.Minute
)

type sortMode int

const (
	sortByDate sortMode = iota
	sortBySender
	sortBySubject
)

type messageCache struct {
	Messages  []api.Message `json:"messages"`
	Timestamp time.Time     `json:"timestamp"`
}

type model struct {
	table          table.Model
	viewport       viewport.Model
	searchInput    textinput.Model
	spinner        spinner.Model
	messages       []api.Message
	filteredMsgs   []api.Message
	messageDetails map[string]*api.MessageDetail
	currentView    view
	previousView   view
	selectedMsg    *api.MessageDetail
	width          int
	height         int
	client         *api.Client
	accountData    *storage.AccountData
	loading        bool
	err            error
	retryCount     int
	searchMode     bool
	statusMessage  string
	autoRefresh    bool
	sortBy         sortMode
	selectedItems  map[int]bool
	bulkMode       bool
	confirmAction  string
	confirmData    interface{}
	lastUpdate     time.Time
}

type messagesLoadedMsg []api.Message
type messageDetailLoadedMsg *api.MessageDetail
type messageDeletedMsg struct{}
type bulkDeletedMsg struct{}
type errMsg error
type tickMsg time.Time

var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	baseStyleFocused = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#00D9FF"))

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B9D")).
			Bold(true).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(1, 0, 0, 2)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D9FF")).
			Bold(true)

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF87")).
			Padding(0, 0, 0, 2)

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D9FF")).
			Bold(true)

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC"))

	confirmBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF6B9D")).
			Padding(1, 2).
			Width(50)

	searchBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#555555")).
			Padding(0, 1)

	searchBoxFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#00D9FF")).
				Padding(0, 1)
)

func initialModel(accountData *storage.AccountData, client *api.Client) model {
	columns := []table.Column{
		{Title: "✓", Width: 3},
		{Title: "📎", Width: 3},
		{Title: "From", Width: 20},
		{Title: "Subject", Width: 30},
		{Title: "Preview", Width: 20},
		{Title: "Date", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	vp := viewport.New(100, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	ti := textinput.New()
	ti.Placeholder = "Search messages (sender, subject, content)..."
	ti.Prompt = " / "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF")).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	ti.CharLimit = 100

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	cached := loadCache()
	var msgs []api.Message
	if cached != nil && time.Since(cached.Timestamp) < cacheExpiry {
		msgs = cached.Messages
	}

	return model{
		table:          t,
		viewport:       vp,
		searchInput:    ti,
		spinner:        sp,
		currentView:    listView,
		client:         client,
		accountData:    accountData,
		loading:        len(msgs) == 0,
		autoRefresh:    true,
		filteredMsgs:   msgs,
		messages:       msgs,
		messageDetails: make(map[string]*api.MessageDetail),
		selectedItems:  make(map[int]bool),
		sortBy:         sortByDate,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		loadMessages(m.client),
		tickCmd(),
		m.spinner.Tick,
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(autoRefreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func loadMessages(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		messages, err := client.GetMessages()
		if err != nil {
			return errMsg(err)
		}
		return messagesLoadedMsg(messages)
	}
}

func loadMessageDetail(client *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		message, err := client.GetMessage(id)
		if err != nil {
			return errMsg(err)
		}
		_ = client.MarkMessageAsRead(id)
		return messageDetailLoadedMsg(message)
	}
}

func deleteMessage(client *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		err := client.DeleteMessage(id)
		if err != nil {
			return errMsg(err)
		}
		return messageDeletedMsg{}
	}
}

func bulkDeleteMessages(client *api.Client, ids []string) tea.Cmd {
	return func() tea.Msg {
		type result struct {
			err error
		}
		results := make(chan result, len(ids))

		for _, id := range ids {
			go func(msgID string) {
				err := client.DeleteMessage(msgID)
				results <- result{err: err}
			}(id)
		}

		for range ids {
			res := <-results
			if res.err != nil {
				return errMsg(res.err)
			}
		}

		return bulkDeletedMsg{}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 6
		footerHeight := 4
		availableHeight := msg.Height - headerHeight - footerHeight
		if availableHeight < 5 {
			availableHeight = 5
		}
		m.table.SetHeight(availableHeight)

		m.viewport.Width = msg.Width - 4
		m.viewport.Height = availableHeight

		m.searchInput.Width = msg.Width - 20
		if m.searchInput.Width < 20 {
			m.searchInput.Width = 20
		}

		m.updateColumnWidths(msg.Width)

		return m, nil

	case messagesLoadedMsg:
		m.messages = msg
		m.loading = false
		m.retryCount = 0
		m.lastUpdate = time.Now()
		saveCache(m.messages)
		m.filterMessages()
		m.sortMessages()
		m.updateTableRows()
		return m, nil

	case messageDetailLoadedMsg:
		m.selectedMsg = msg
		m.messageDetails[m.selectedMsg.ID] = m.selectedMsg
		m.currentView = detailView
		m.loading = false
		m.updateMessageSeen(m.selectedMsg.ID, true)

		var content strings.Builder
		content.WriteString(headerStyle.Render("From: ") + m.selectedMsg.From.Address + "\n")
		content.WriteString(headerStyle.Render("Subject: ") + m.selectedMsg.Subject + "\n")
		content.WriteString(headerStyle.Render("Date: ") + m.selectedMsg.CreatedAt.Format("02/01/2006 15:04:05") + "\n")
		content.WriteString(separatorStyle.Render(strings.Repeat("─", 80)) + "\n\n")

		if m.selectedMsg.Text != "" {
			content.WriteString(m.selectedMsg.Text)
		} else if len(m.selectedMsg.HTML) > 0 {
			var htmlBuilder strings.Builder
			for _, h := range m.selectedMsg.HTML {
				htmlBuilder.WriteString(h)
			}
			text, err := html2text.FromString(htmlBuilder.String(), html2text.Options{
				PrettyTables: true,
				OmitLinks:    false,
			})
			if err == nil {
				content.WriteString(text)
				content.WriteString("\n\n" + separatorStyle.Render(strings.Repeat("─", 80)) + "\n")
				content.WriteString(descStyle.Render("Press 'o' to open HTML in browser"))
			} else {
				content.WriteString("[HTML content - press 'o' to open in browser]\n\n")
				for _, h := range m.selectedMsg.HTML {
					content.WriteString(h)
				}
			}
		}

		if len(m.selectedMsg.Attachments) > 0 {
			content.WriteString("\n\n" + separatorStyle.Render(strings.Repeat("─", 80)) + "\n")
			content.WriteString(headerStyle.Render(fmt.Sprintf("📎 Attachments (%d)", len(m.selectedMsg.Attachments))) + "\n\n")
			for i, att := range m.selectedMsg.Attachments {
				sizeKB := float64(att.Size) / 1024.0
				content.WriteString(fmt.Sprintf("%d. %s (%s, %.1f KB)\n",
					i+1,
					keyStyle.Render(att.Filename),
					descStyle.Render(att.ContentType),
					sizeKB))
			}
			content.WriteString("\n" + descStyle.Render("Press '1-9' to download attachment, 'shift+a' to download all"))
		}

		m.viewport.SetContent(content.String())
		return m, nil

	case bulkDeletedMsg:
		deletedCount := len(m.selectedItems)
		m.statusMessage = fmt.Sprintf("%d messages deleted", deletedCount)

		selectedIDs := make(map[string]bool, deletedCount)
		for idx := range m.selectedItems {
			if idx < len(m.filteredMsgs) {
				selectedIDs[m.filteredMsgs[idx].ID] = true
			}
		}

		newMessages := make([]api.Message, 0, len(m.messages)-deletedCount)
		for _, msg := range m.messages {
			if !selectedIDs[msg.ID] {
				newMessages = append(newMessages, msg)
			}
		}
		m.messages = newMessages

		m.selectedItems = make(map[int]bool)
		m.bulkMode = false
		m.filterMessages()
		m.sortMessages()
		m.updateTableRows()
		saveCache(m.messages)
		return m, nil

	case messageDeletedMsg:
		m.statusMessage = "Message deleted"
		m.currentView = listView
		if m.selectedMsg != nil {
			msgID := m.selectedMsg.ID
			m.selectedMsg = nil
			for i := range m.messages {
				if m.messages[i].ID == msgID {
					m.messages = append(m.messages[:i], m.messages[i+1:]...)
					break
				}
			}
			m.filterMessages()
			m.sortMessages()
			m.updateTableRows()
			saveCache(m.messages)
		}
		return m, nil

	case tickMsg:
		if m.autoRefresh && m.currentView == listView && !m.loading {
			return m, tea.Batch(loadMessages(m.client), tickCmd())
		}
		return m, tickCmd()

	case errMsg:
		m.loading = false
		m.retryCount++

		if m.retryCount < 3 {
			m.statusMessage = fmt.Sprintf("Error (retry %d/3): %v", m.retryCount, msg)
			time.Sleep(time.Second * time.Duration(m.retryCount))
			return m, loadMessages(m.client)
		}

		m.err = msg
		m.statusMessage = fmt.Sprintf("Error after 3 retries: %v", msg)
		return m, nil

	case tea.KeyMsg:
		if m.currentView == confirmView {
			switch msg.String() {
			case "y", "Y":
				return m.executeConfirmedAction()
			case "n", "N", "esc", "q":
				m.currentView = m.previousView
				m.confirmAction = ""
				m.confirmData = nil
				return m, nil
			}
			return m, nil
		}

		if m.currentView == helpView {
			switch msg.String() {
			case "q", "esc", "?":
				m.currentView = m.previousView
				return m, nil
			}
			return m, nil
		}

		if m.searchMode {
			switch msg.String() {
			case "esc":
				m.searchMode = false
				m.searchInput.Blur()
				return m, nil
			case "enter":
				m.searchMode = false
				m.searchInput.Blur()
				m.filterMessages()
				m.sortMessages()
				m.updateTableRows()
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				return m, cmd
			}
		}

		switch msg.String() {
		case "?":
			m.previousView = m.currentView
			m.currentView = helpView
			return m, nil

		case "q", "ctrl+c":
			if m.currentView == listView {
				return m.showConfirm("quit", "quit the application")
			}
			return m, tea.Quit

		case "esc":
			if m.currentView == detailView {
				m.currentView = listView
				m.selectedMsg = nil
				return m, nil
			}
			return m, tea.Quit

		case "j", "down":
			if m.currentView == listView {
				m.table.MoveDown(1)
				return m, nil
			}

		case "k", "up":
			if m.currentView == listView {
				m.table.MoveUp(1)
				return m, nil
			}

		case "r":
			if m.currentView == listView {
				m.loading = true
				m.statusMessage = "Refreshing..."
				return m, loadMessages(m.client)
			}

		case "/":
			if m.currentView == listView {
				m.searchMode = true
				m.searchInput.Focus()
				return m, nil
			}

		case "a":
			if m.currentView == listView {
				m.autoRefresh = !m.autoRefresh
				if m.autoRefresh {
					m.statusMessage = "Auto-refresh enabled"
				} else {
					m.statusMessage = "Auto-refresh disabled"
				}
				return m, nil
			}

		case "s":
			if m.currentView == listView {
				m.sortBy = (m.sortBy + 1) % 3
				sortNames := []string{"Date", "Sender", "Subject"}
				m.statusMessage = fmt.Sprintf("Sorted by: %s", sortNames[m.sortBy])
				m.sortMessages()
				m.updateTableRows()
				return m, nil
			}

		case "c":
			if m.currentView == listView && len(m.filteredMsgs) > 0 {
				selectedIdx := m.table.Cursor()
				if selectedIdx < len(m.filteredMsgs) {
					_ = clipboard.WriteAll(m.filteredMsgs[selectedIdx].From.Address)
					m.statusMessage = "Email copied to clipboard"
					return m, nil
				}
			} else if m.currentView == detailView && m.selectedMsg != nil {
				_ = clipboard.WriteAll(m.selectedMsg.Text)
				m.statusMessage = "Message copied to clipboard"
				return m, nil
			}

		case "v":
			if m.currentView == listView {
				m.bulkMode = !m.bulkMode
				if !m.bulkMode {
					m.selectedItems = make(map[int]bool)
					m.updateTableRows()
				}
				m.statusMessage = fmt.Sprintf("Bulk mode: %v", m.bulkMode)
				return m, nil
			}

		case " ":
			if m.currentView == listView && m.bulkMode {
				selectedIdx := m.table.Cursor()
				if m.selectedItems[selectedIdx] {
					delete(m.selectedItems, selectedIdx)
				} else {
					m.selectedItems[selectedIdx] = true
				}
				m.updateTableRows()
				return m, nil
			}

		case "d":
			if m.currentView == detailView && m.selectedMsg != nil {
				return m.showConfirm("delete_single", fmt.Sprintf("delete message '%s'", truncate(m.selectedMsg.Subject, 30)))
			} else if m.currentView == listView && m.bulkMode && len(m.selectedItems) > 0 {
				return m.showConfirm("delete_bulk", fmt.Sprintf("delete %d selected messages", len(m.selectedItems)))
			}

		case "enter":
			if m.currentView == listView && len(m.filteredMsgs) > 0 {
				selectedIdx := m.table.Cursor()
				if selectedIdx < len(m.filteredMsgs) {
					msgID := m.filteredMsgs[selectedIdx].ID
					if cached, ok := m.messageDetails[msgID]; ok {
						m.selectedMsg = cached
						m.currentView = detailView
						var content strings.Builder
						content.WriteString(headerStyle.Render("From: ") + cached.From.Address + "\n")
						content.WriteString(headerStyle.Render("Subject: ") + cached.Subject + "\n")
						content.WriteString(headerStyle.Render("Date: ") + cached.CreatedAt.Format("02/01/2006 15:04:05") + "\n")
						content.WriteString(separatorStyle.Render(strings.Repeat("─", 80)) + "\n\n")
						if cached.Text != "" {
							content.WriteString(cached.Text)
						} else if len(cached.HTML) > 0 {
							var htmlBuilder strings.Builder
							for _, h := range cached.HTML {
								htmlBuilder.WriteString(h)
							}
							text, err := html2text.FromString(htmlBuilder.String(), html2text.Options{
								PrettyTables: true,
								OmitLinks:    false,
							})
							if err == nil {
								content.WriteString(text)
								content.WriteString("\n\n" + separatorStyle.Render(strings.Repeat("─", 80)) + "\n")
								content.WriteString(descStyle.Render("Press 'o' to open HTML in browser"))
							} else {
								content.WriteString("[HTML content - press 'o' to open in browser]\n\n")
								for _, h := range cached.HTML {
									content.WriteString(h)
								}
							}
						}

						if len(cached.Attachments) > 0 {
							content.WriteString("\n\n" + separatorStyle.Render(strings.Repeat("─", 80)) + "\n")
							content.WriteString(headerStyle.Render(fmt.Sprintf("📎 Attachments (%d)", len(cached.Attachments))) + "\n\n")
							for i, att := range cached.Attachments {
								sizeKB := float64(att.Size) / 1024.0
								content.WriteString(fmt.Sprintf("%d. %s (%s, %.1f KB)\n",
									i+1,
									keyStyle.Render(att.Filename),
									descStyle.Render(att.ContentType),
									sizeKB))
							}
							content.WriteString("\n" + descStyle.Render("Press '1-9' to download attachment, 'shift+a' to download all"))
						}

						m.viewport.SetContent(content.String())
						return m, nil
					}
					m.loading = true
					return m, loadMessageDetail(m.client, msgID)
				}
			}

		case "o":
			if m.currentView == detailView && m.selectedMsg != nil {
				if len(m.selectedMsg.HTML) > 0 {
					openInBrowser(m.selectedMsg)
				}
			}

		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if m.currentView == detailView && m.selectedMsg != nil && len(m.selectedMsg.Attachments) > 0 {
				idx := int(msg.String()[0] - '1')
				if idx < len(m.selectedMsg.Attachments) {
					go downloadAttachment(m.client, m.selectedMsg.ID, m.selectedMsg.Attachments[idx])
					m.statusMessage = fmt.Sprintf("Downloading %s...", m.selectedMsg.Attachments[idx].Filename)
				}
			}

		case "A":
			if m.currentView == detailView && m.selectedMsg != nil && len(m.selectedMsg.Attachments) > 0 {
				for _, att := range m.selectedMsg.Attachments {
					go downloadAttachment(m.client, m.selectedMsg.ID, att)
				}
				m.statusMessage = fmt.Sprintf("Downloading %d attachments...", len(m.selectedMsg.Attachments))
			}
		}
	}

	if m.currentView == listView {
		m.table, cmd = m.table.Update(msg)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
	}

	var spinnerCmd tea.Cmd
	m.spinner, spinnerCmd = m.spinner.Update(msg)

	return m, tea.Batch(cmd, spinnerCmd)
}

func (m model) View() string {
	if m.loading {
		return titleStyle.Render(fmt.Sprintf("%s Loading...", m.spinner.View())) + "\n"
	}

	if m.err != nil {
		return titleStyle.Render("Error: ") + m.err.Error() + "\n"
	}

	var s strings.Builder

	if m.currentView != helpView {
		msgCount := len(m.messages)
		msgWord := "messages"
		if msgCount == 1 {
			msgWord = "message"
		}
		title := fmt.Sprintf("Burnmail - %s (%d %s)", m.accountData.Address, msgCount, msgWord)
		s.WriteString(titleStyle.Render(title) + "\n")

		if m.statusMessage != "" {
			s.WriteString(statusStyle.Render("▸ "+m.statusMessage) + "\n")
		}
		s.WriteString("\n")
	}

	if m.currentView == listView {
		var searchBox string
		var tableStyle lipgloss.Style

		if m.searchMode {
			searchBox = searchBoxFocusedStyle.Render(m.searchInput.View())
			tableStyle = baseStyle
		} else {
			searchBox = searchBoxStyle.Render(m.searchInput.View())
			tableStyle = baseStyleFocused
		}

		s.WriteString(searchBox + "\n\n")
		s.WriteString(tableStyle.Render(m.table.View()) + "\n\n")

		sortNames := []string{"Date", "Sender", "Subject"}
		sortInfo := fmt.Sprintf("Sort: %s", sortNames[m.sortBy])
		s.WriteString(helpStyle.Render(sortInfo) + " ")
		s.WriteString(helpStyle.Render("• Press "+keyStyle.Render("?")+" for help") + "\n")

		helpText := keyStyle.Render("↑/↓") + "/" + keyStyle.Render("j/k") + ":navigate " + keyStyle.Render("enter") + ":open " + keyStyle.Render("s") + ":sort " + keyStyle.Render("c") + ":copy " + keyStyle.Render("v") + ":bulk " + keyStyle.Render("r") + ":refresh " + keyStyle.Render("/") + ":search "
		if m.autoRefresh {
			helpText += keyStyle.Render("a") + ":auto:" + keyStyle.Render("ON")
		} else {
			helpText += keyStyle.Render("a") + ":auto:" + keyStyle.Render("OFF")
		}
		if m.bulkMode {
			helpText += " " + keyStyle.Render("space") + ":select " + keyStyle.Render("d") + ":delete"
		}
		helpText += " " + keyStyle.Render("q") + ":quit"
		s.WriteString(helpStyle.Render(helpText))
	} else if m.currentView == helpView {
		s.WriteString(renderHelpScreen(m.width, m.height))
	} else if m.currentView == confirmView {
		s.WriteString(renderConfirmDialog(m.confirmData.(string)))
	} else {
		s.WriteString(baseStyle.Render(m.viewport.View()) + "\n")
		s.WriteString(helpStyle.Render("↑/↓ • " + keyStyle.Render("o") + ":browser • " + keyStyle.Render("c") + ":copy • " + keyStyle.Render("d") + ":delete • esc • " + keyStyle.Render("?") + ":help"))
	}

	return s.String()
}

func (m *model) filterMessages() {
	if m.searchInput.Value() == "" {
		m.filteredMsgs = m.messages
		return
	}

	query := strings.ToLower(m.searchInput.Value())
	m.filteredMsgs = make([]api.Message, 0, len(m.messages))
	for _, msg := range m.messages {
		fromLower := strings.ToLower(msg.From.Address)
		subjectLower := strings.ToLower(msg.Subject)
		introLower := strings.ToLower(msg.Intro)

		if strings.Contains(fromLower, query) ||
			strings.Contains(subjectLower, query) ||
			strings.Contains(introLower, query) {
			m.filteredMsgs = append(m.filteredMsgs, msg)
		}
	}
}

func (m *model) updateColumnWidths(termWidth int) {
	var newCols []table.Column

	if termWidth < 80 {
		newCols = []table.Column{
			{Title: "✓", Width: 2},
			{Title: "From", Width: 15},
			{Title: "Subject", Width: maxInt(20, termWidth-30)},
			{Title: "Date", Width: 10},
		}
	} else if termWidth < 120 {
		newCols = []table.Column{
			{Title: "✓", Width: 3},
			{Title: "📎", Width: 3},
			{Title: "From", Width: 20},
			{Title: "Subject", Width: maxInt(25, termWidth-55)},
			{Title: "Preview", Width: 15},
			{Title: "Date", Width: 12},
		}
	} else {
		newCols = []table.Column{
			{Title: "✓", Width: 3},
			{Title: "📎", Width: 3},
			{Title: "From", Width: 25},
			{Title: "Subject", Width: maxInt(30, termWidth-85)},
			{Title: "Preview", Width: 25},
			{Title: "Date", Width: 14},
		}
	}

	m.table.SetRows([]table.Row{})
	m.table.SetColumns(newCols)
	m.updateTableRows()
}

func (m *model) updateTableRows() {
	cols := m.table.Columns()
	fromWidth := 25
	subjectWidth := 35
	previewWidth := 25

	for _, col := range cols {
		switch col.Title {
		case "From":
			fromWidth = col.Width
		case "Subject":
			subjectWidth = col.Width
		case "Preview":
			previewWidth = col.Width
		}
	}

	rows := make([]table.Row, 0, len(m.filteredMsgs))
	for i, msg := range m.filteredMsgs {
		checkbox := " "
		if m.selectedItems[i] {
			checkbox = "✓"
		}

		attach := " "
		if msg.HasAttach {
			attach = "📎"
		}

		preview := truncate(msg.Intro, previewWidth)

		if len(cols) == 4 {
			rows = append(rows, table.Row{
				checkbox,
				truncate(msg.From.Address, fromWidth),
				truncate(msg.Subject, subjectWidth),
				msg.CreatedAt.Format("02/01 15:04"),
			})
		} else if len(cols) == 6 {
			rows = append(rows, table.Row{
				checkbox,
				attach,
				truncate(msg.From.Address, fromWidth),
				truncate(msg.Subject, subjectWidth),
				preview,
				msg.CreatedAt.Format("02/01 15:04"),
			})
		}
	}
	m.table.SetRows(rows)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *model) sortMessages() {
	switch m.sortBy {
	case sortByDate:
		sort.Slice(m.filteredMsgs, func(i, j int) bool {
			return m.filteredMsgs[i].CreatedAt.After(m.filteredMsgs[j].CreatedAt)
		})
	case sortBySender:
		sort.Slice(m.filteredMsgs, func(i, j int) bool {
			return m.filteredMsgs[i].From.Address < m.filteredMsgs[j].From.Address
		})
	case sortBySubject:
		sort.Slice(m.filteredMsgs, func(i, j int) bool {
			return m.filteredMsgs[i].Subject < m.filteredMsgs[j].Subject
		})
	}
}

func (m *model) updateMessageSeen(id string, seen bool) {
	for i := range m.messages {
		if m.messages[i].ID == id {
			m.messages[i].Seen = seen
			break
		}
	}
	m.filterMessages()
	m.sortMessages()
	m.updateTableRows()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}

	if max < 10 {
		return s[:max]
	}

	const ellipsis = "..."
	words := strings.Fields(s)
	result := strings.Builder{}
	result.Grow(max)
	length := 0

	for _, word := range words {
		if length+len(word)+1 > max-len(ellipsis) {
			break
		}
		if length > 0 {
			result.WriteByte(' ')
			length++
		}
		result.WriteString(word)
		length += len(word)
	}

	if result.Len() == 0 {
		return s[:max-len(ellipsis)] + ellipsis
	}

	result.WriteString(ellipsis)
	return result.String()
}

func loadCache() *messageCache {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	cacheFile := filepath.Join(homeDir, cacheFileName)
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil
	}

	var cache messageCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil
	}

	return &cache
}

func saveCache(messages []api.Message) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	cache := messageCache{
		Messages:  messages,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return
	}

	cacheFile := filepath.Join(homeDir, cacheFileName)
	_ = os.WriteFile(cacheFile, data, 0600)
}

func (m model) showConfirm(action, description string) (tea.Model, tea.Cmd) {
	m.previousView = m.currentView
	m.currentView = confirmView
	m.confirmAction = action
	m.confirmData = description
	return m, nil
}

func (m model) executeConfirmedAction() (tea.Model, tea.Cmd) {
	m.currentView = m.previousView
	action := m.confirmAction
	m.confirmAction = ""

	switch action {
	case "quit":
		return m, tea.Quit

	case "delete_single":
		if m.selectedMsg != nil {
			m.loading = true
			m.statusMessage = "Deleting message..."
			return m, deleteMessage(m.client, m.selectedMsg.ID)
		}

	case "delete_bulk":
		var ids []string
		for idx := range m.selectedItems {
			if idx < len(m.filteredMsgs) {
				ids = append(ids, m.filteredMsgs[idx].ID)
			}
		}
		m.loading = true
		m.statusMessage = fmt.Sprintf("Deleting %d messages...", len(ids))
		return m, bulkDeleteMessages(m.client, ids)
	}

	return m, nil
}

func renderHelpScreen(_, _ int) string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Burnmail - Help") + "\n\n")

	helpSections := []struct {
		title string
		items [][2]string
	}{
		{
			title: "General",
			items: [][2]string{
				{"?", "Show this help screen"},
				{"q", "Quit application (with confirmation)"},
				{"esc", "Go back / Cancel"},
				{"r", "Refresh messages"},
			},
		},
		{
			title: "List View",
			items: [][2]string{
				{"↑/↓ or j/k", "Navigate messages"},
				{"enter", "View selected message"},
				{"/", "Search messages"},
				{"s", "Cycle sort (Date → Sender → Subject)"},
				{"c", "Copy sender email to clipboard"},
				{"a", "Toggle auto-refresh (every 10s)"},
				{"v", "Toggle bulk selection mode"},
				{"space", "Select/deselect message (bulk mode)"},
				{"d", "Delete selected message(s)"},
			},
		},
		{
			title: "Detail View",
			items: [][2]string{
				{"↑/↓ or j/k", "Scroll message content"},
				{"o", "Open HTML content in browser"},
				{"c", "Copy message content to clipboard"},
				{"d", "Delete message"},
				{"1-9", "Download attachment by number"},
				{"shift+a", "Download all attachments"},
				{"esc", "Back to list"},
			},
		},
	}

	for _, section := range helpSections {
		s.WriteString(headerStyle.Render("▸ "+section.title) + "\n")
		for _, item := range section.items {
			s.WriteString("  " + keyStyle.Render(fmt.Sprintf("%-10s", item[0])) + " " + descStyle.Render(item[1]) + "\n")
		}
		s.WriteString("\n")
	}

	s.WriteString(helpStyle.Render("Press '?' or 'esc' to close this help"))

	return s.String()
}

func renderConfirmDialog(description string) string {
	var s strings.Builder

	s.WriteString("\n\n")
	s.WriteString(confirmBoxStyle.Render(
		titleStyle.Render("⚠ Confirmation Required") + "\n\n" +
			descStyle.Render("Are you sure you want to "+description+"?") + "\n\n" +
			keyStyle.Render("Y") + descStyle.Render(" - Yes, proceed") + "\n" +
			keyStyle.Render("N") + descStyle.Render(" - No, cancel"),
	))

	return s.String()
}

func getDownloadsDir() string {
	var downloadsDir string

	switch runtime.GOOS {
	case "windows":
		userProfile := os.Getenv("USERPROFILE")
		if userProfile == "" {
			userProfile = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		}
		downloadsDir = filepath.Join(userProfile, "Downloads")

	case "darwin":
		home := os.Getenv("HOME")
		downloadsDir = filepath.Join(home, "Downloads")

	case "linux":
		xdgDownload := os.Getenv("XDG_DOWNLOAD_DIR")
		if xdgDownload != "" {
			downloadsDir = xdgDownload
		} else {
			home := os.Getenv("HOME")
			downloadsDir = filepath.Join(home, "Downloads")
		}

	default:
		downloadsDir, _ = os.Getwd()
	}

	if _, err := os.Stat(downloadsDir); os.IsNotExist(err) {
		downloadsDir, _ = os.Getwd()
	}

	return downloadsDir
}

func downloadAttachment(client *api.Client, messageID string, att api.Attachment) {
	data, err := client.DownloadAttachment(messageID, att.ID)
	if err != nil {
		return
	}

	downloadsDir := getDownloadsDir()
	filePath := filepath.Join(downloadsDir, att.Filename)
	counter := 1
	baseName := strings.TrimSuffix(att.Filename, filepath.Ext(att.Filename))
	ext := filepath.Ext(att.Filename)

	for {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			break
		}
		filePath = filepath.Join(downloadsDir, fmt.Sprintf("%s_%d%s", baseName, counter, ext))
		counter++
	}

	_ = os.WriteFile(filePath, data, 0644)
}

func runTUI(accountData *storage.AccountData, client *api.Client) error {
	client.SetToken(accountData.Token)

	p := tea.NewProgram(
		initialModel(accountData, client),
		tea.WithAltScreen(),
	)

	_, err := p.Run()
	return err
}

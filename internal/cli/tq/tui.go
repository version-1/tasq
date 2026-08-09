package tq

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	tqconfig "github.com/version-1/tasq/internal/config"
	"github.com/version-1/tasq/internal/issue/domain/entity"
)

const (
	listRefreshInterval   = 15 * time.Second
	detailRefreshInterval = 5 * time.Second
	minimumTUIWidth       = 54
	minimumTUIHeight      = 14
	narrowTUIWidth        = 90
)

type fileDescriptor interface{ Fd() uintptr }

var terminalCheck = func(value any) bool {
	fd, ok := value.(fileDescriptor)
	return ok && term.IsTerminal(int(fd.Fd()))
}

func (a app) tui(ctx context.Context, args []string, cfg config) error {
	fs := newFlagSet("tq tui")
	orchestratorURL := fs.String("orchestrator-url", "", "orchestrator inspection API URL")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printTUIHelp(a.stdout)
			return nil
		}
		return usageError("usage: tq tui [--orchestrator-url URL]")
	}
	if fs.NArg() != 0 {
		return usageError("usage: tq tui [--orchestrator-url URL]")
	}
	if cfg.output != "text" {
		return usageError("tq tui supports only --output text")
	}
	if !terminalCheck(a.stdin) || !terminalCheck(a.stdout) {
		return usageError("tq tui requires an interactive TTY")
	}
	resolvedOrchestratorURL := strings.TrimSpace(*orchestratorURL)
	if resolvedOrchestratorURL == "" {
		stateURL, ok, err := tqconfig.OrchestratorURLFromState()
		if err == nil && ok {
			resolvedOrchestratorURL = stateURL
		}
	}
	orchestrator, err := newOrchestratorClient(resolvedOrchestratorURL)
	if err != nil {
		return usageError("%v", err)
	}
	model := newTUIModel(ctx, a.client, orchestrator, openBrowser)
	program := tea.NewProgram(model, tea.WithContext(ctx), tea.WithInput(a.stdin), tea.WithOutput(a.stdout), tea.WithAltScreen())
	_, err = program.Run()
	return err
}

type tuiTab int

const (
	tabOverview tuiTab = iota
	tabComments
	tabArtifacts
	tabRun
)

var tabNames = []string{"Overview", "Comments", "Artifacts", "Run"}

type tuiModel struct {
	ctx          context.Context
	tracker      *apiClient
	orchestrator *orchestratorClient
	openURL      func(string) error

	width, height     int
	issues            []entity.Issue
	projects          []entity.Project
	selection         int
	artifactSelection int
	detail            *entity.Issue
	comments          []entity.Comment
	commentCursor     *int64
	commentsLoading   bool
	runtime           *runtimeIssue
	runtimeAbsent     bool
	runtimeError      string
	runtimeErrorAt    time.Time

	states        []entity.Status
	projectIDs    []int64
	search        string
	nextOffset    *int
	generation    uint64
	listRequest   uint64
	detailRequest uint64
	loading       bool
	loadingMore   bool
	replaceList   bool
	errorText     string
	tab           tuiTab
	detailScreen  bool
	help          bool
	filter        bool
	filterCursor  int
	searching     bool
	searchInput   textinput.Model
	viewport      viewport.Model
	statusLine    string
}

type listTickMsg time.Time
type detailTickMsg time.Time
type listResultMsg struct {
	generation    uint64
	request       uint64
	page          issuePage
	projects      []entity.Project
	appendPage    bool
	refreshWindow bool
	err           error
}
type detailResultMsg struct {
	generation    uint64
	request       uint64
	issueID       int64
	issue         entity.Issue
	comments      commentPage
	runtime       runtimeIssue
	runtimeAbsent bool
	runtimeErr    error
	err           error
}
type browserResultMsg struct{ err error }
type olderCommentsMsg struct {
	generation uint64
	issueID    int64
	page       commentPage
	err        error
}

func newTUIModel(ctx context.Context, tracker *apiClient, orchestrator *orchestratorClient, opener func(string) error) tuiModel {
	input := textinput.New()
	input.Prompt = "/ "
	input.Placeholder = "search issues"
	return tuiModel{
		ctx: ctx, tracker: tracker, orchestrator: orchestrator, openURL: opener,
		states: defaultTUIStates(), loading: true, replaceList: true, searchInput: input, viewport: viewport.New(80, 20),
	}
}

func defaultTUIStates() []entity.Status {
	return []entity.Status{entity.StatusBacklog, entity.StatusReady, entity.StatusInProgress, entity.StatusReview, entity.StatusBlocked, entity.StatusFailed}
}

func (m tuiModel) Init() tea.Cmd {
	return tea.Batch(m.loadList(false), listTick(), detailTick())
}

func listTick() tea.Cmd {
	return tea.Tick(listRefreshInterval, func(t time.Time) tea.Msg { return listTickMsg(t) })
}
func detailTick() tea.Cmd {
	return tea.Tick(detailRefreshInterval, func(t time.Time) tea.Msg { return detailTickMsg(t) })
}

func (m tuiModel) loadList(appendPage bool) tea.Cmd {
	generation := m.generation
	request := m.listRequest
	refreshCount := 0
	if !appendPage && !m.replaceList {
		refreshCount = len(m.issues)
	}
	offset := 0
	if appendPage && m.nextOffset != nil {
		offset = *m.nextOffset
	}
	query := issueListQuery{States: append([]entity.Status(nil), m.states...), ProjectIDs: append([]int64(nil), m.projectIDs...), Search: m.search, Offset: offset}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 8*time.Second)
		defer cancel()
		page, err := m.tracker.listIssuesPage(ctx, query)
		for nextOffset := 50; err == nil && nextOffset < refreshCount && page.NextOffset != nil; nextOffset += 50 {
			query.Offset = nextOffset
			var nextPage issuePage
			nextPage, err = m.tracker.listIssuesPage(ctx, query)
			if err == nil {
				page.Issues = mergeIssues(page.Issues, nextPage.Issues)
				page.NextOffset = nextPage.NextOffset
			}
		}
		var projects []entity.Project
		if !appendPage && err == nil {
			projects, err = m.tracker.listProjects(ctx)
		}
		return listResultMsg{generation: generation, request: request, page: page, projects: projects, appendPage: appendPage, refreshWindow: refreshCount > 0, err: err}
	}
}

func (m tuiModel) loadDetail() tea.Cmd {
	selected, ok := m.selectedIssue()
	if !ok {
		return nil
	}
	generation, request, issueID := m.generation, m.detailRequest, selected.ID
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 4*time.Second)
		defer cancel()
		issue, err := m.tracker.getIssue(ctx, issueID)
		if err != nil {
			return detailResultMsg{generation: generation, request: request, issueID: issueID, err: err}
		}
		comments, err := m.tracker.commentsPage(ctx, issueID, 0)
		if err != nil {
			return detailResultMsg{generation: generation, request: request, issueID: issueID, err: err}
		}
		result := detailResultMsg{generation: generation, request: request, issueID: issueID, issue: issue, comments: comments}
		if m.orchestrator == nil {
			result.runtimeErr = errors.New("orchestrator URL is not configured")
			return result
		}
		result.runtime, result.runtimeErr = m.orchestrator.issue(ctx, issueID)
		result.runtimeAbsent = errors.Is(result.runtimeErr, errRuntimeNotFound)
		return result
	}
}

func (m tuiModel) loadOlderComments() tea.Cmd {
	selected, ok := m.selectedIssue()
	if !ok || m.commentCursor == nil {
		return nil
	}
	generation, issueID, cursor := m.generation, selected.ID, *m.commentCursor
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 4*time.Second)
		defer cancel()
		page, err := m.tracker.commentsPage(ctx, issueID, cursor)
		return olderCommentsMsg{generation: generation, issueID: issueID, page: page, err: err}
	}
}

func (m tuiModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	if m.searching {
		return m.updateSearch(message)
	}
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resizeViewport()
	case tea.KeyMsg:
		return m.updateKey(msg)
	case listTickMsg:
		m.listRequest++
		return m, tea.Batch(m.loadList(false), listTick())
	case detailTickMsg:
		var cmd tea.Cmd
		if m.detail != nil {
			m.detailRequest++
			cmd = m.loadDetail()
		}
		return m, tea.Batch(cmd, detailTick())
	case listResultMsg:
		if msg.generation != m.generation || msg.request != m.listRequest {
			return m, nil
		}
		m.loading, m.loadingMore = false, false
		if msg.err != nil {
			m.errorText = msg.err.Error()
			return m, nil
		}
		m.errorText = ""
		selectedID := m.selectedID()
		if msg.appendPage {
			m.issues = mergeIssues(m.issues, msg.page.Issues)
			m.nextOffset = msg.page.NextOffset
		} else if m.replaceList || msg.refreshWindow {
			m.issues = mergeIssues(nil, msg.page.Issues)
			m.projects = msg.projects
			m.nextOffset = msg.page.NextOffset
			m.replaceList = false
		} else {
			m.issues = mergeIssues(m.issues, msg.page.Issues)
			if len(msg.projects) > 0 {
				m.projects = msg.projects
			}
		}
		m.restoreSelection(selectedID)
		m.detailRequest++
		return m, m.loadDetail()
	case detailResultMsg:
		if msg.generation != m.generation || msg.request != m.detailRequest || m.selectedID() != msg.issueID {
			return m, nil
		}
		if msg.err != nil {
			m.errorText = msg.err.Error()
			return m, nil
		}
		m.errorText = ""
		previousIssueID := int64(0)
		if m.detail != nil {
			previousIssueID = m.detail.ID
		}
		m.detail = &msg.issue
		m.artifactSelection = min(m.artifactSelection, max(len(msg.issue.Artifacts)-1, 0))
		if previousIssueID == msg.issueID && len(m.comments) > 0 {
			m.comments = mergeComments(msg.comments.Comments, m.comments)
			m.commentCursor = olderCursor(m.commentCursor, msg.comments.NextCursor)
		} else {
			m.comments, m.commentCursor = msg.comments.Comments, msg.comments.NextCursor
		}
		m.runtimeAbsent = msg.runtimeAbsent
		if msg.runtimeErr != nil && !msg.runtimeAbsent {
			m.runtimeError, m.runtimeErrorAt = msg.runtimeErr.Error(), time.Now()
		} else {
			m.runtimeError = ""
			if !msg.runtimeAbsent {
				m.runtime = &msg.runtime
			}
		}
		m.refreshViewport()
	case browserResultMsg:
		if msg.err != nil {
			m.statusLine = "Open failed: " + msg.err.Error()
		} else {
			m.statusLine = "Opened artifact URL"
		}
	case olderCommentsMsg:
		if msg.generation != m.generation || m.selectedID() != msg.issueID {
			return m, nil
		}
		m.commentsLoading = false
		if msg.err != nil {
			m.statusLine = "Older comments failed: " + msg.err.Error()
			return m, nil
		}
		m.comments = mergeComments(msg.page.Comments, m.comments)
		m.commentCursor = msg.page.NextCursor
		m.refreshViewport()
	}
	return m, nil
}

func (m tuiModel) updateKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.filter {
		switch key.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc", "f":
			m.filter = false
			return m, nil
		case "up", "k":
			m.filterCursor = max(m.filterCursor-1, 0)
			return m, nil
		case "down", "j":
			m.filterCursor = min(m.filterCursor+1, len(allStatuses())+len(m.projects)-1)
			return m, nil
		case " ":
			m.toggleFilterOption()
			return m, nil
		case "enter":
			m.filter = false
			cmd := m.applyFilters()
			return m, cmd
		}
		return m, nil
	}
	switch key.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "?":
		m.help = !m.help
	case "esc":
		if m.help || m.filter {
			m.help, m.filter = false, false
		} else if m.detailScreen {
			m.detailScreen = false
		}
	case "/":
		m.searching = true
		m.searchInput.SetValue(m.search)
		m.searchInput.Focus()
		return m, textinput.Blink
	case "f":
		m.filter = !m.filter
	case "tab":
		m.tab = (m.tab + 1) % tuiTab(len(tabNames))
		m.refreshViewport()
	case "pgup":
		if m.tab == tabComments && m.commentCursor != nil && !m.commentsLoading {
			m.commentsLoading = true
			return m, m.loadOlderComments()
		}
		m.viewport.HalfViewUp()
	case "pgdown":
		m.viewport.HalfViewDown()
	case "left":
		if m.tab == tabArtifacts && m.artifactSelection > 0 {
			m.artifactSelection--
			m.refreshViewport()
		}
	case "right":
		if m.tab == tabArtifacts && m.detail != nil && m.artifactSelection+1 < len(m.detail.Artifacts) {
			m.artifactSelection++
			m.refreshViewport()
		}
	case "enter":
		if m.width < narrowTUIWidth {
			m.detailScreen = true
		}
	case "r":
		m.generation++
		m.listRequest++
		m.detailRequest++
		m.loading = true
		return m, tea.Batch(m.loadList(false), m.loadDetail())
	case "up", "k":
		if m.detailScreen && m.tab == tabArtifacts && m.artifactSelection > 0 {
			m.artifactSelection--
			m.refreshViewport()
			return m, nil
		}
		if m.selection > 0 {
			m.selection--
			m.detail = nil
			m.detailRequest++
			return m, m.loadDetail()
		}
	case "down", "j":
		if m.detailScreen && m.tab == tabArtifacts && m.detail != nil && m.artifactSelection+1 < len(m.detail.Artifacts) {
			m.artifactSelection++
			m.refreshViewport()
			return m, nil
		}
		if m.selection+1 < len(m.issues) {
			m.selection++
			m.detail = nil
			m.detailRequest++
			return m, m.loadDetail()
		}
		if m.nextOffset != nil && !m.loadingMore {
			m.loadingMore = true
			m.listRequest++
			return m, m.loadList(true)
		}
	case "o":
		return m, m.openArtifact()
	}
	return m, nil
}

func (m tuiModel) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.searching = false
			m.searchInput.Blur()
			return m, nil
		case "enter":
			m.search = strings.TrimSpace(m.searchInput.Value())
			m.searching = false
			m.searchInput.Blur()
			cmd := m.applyFilters()
			return m, cmd
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

func (m *tuiModel) applyFilters() tea.Cmd {
	m.generation++
	m.listRequest++
	m.loading = true
	m.selection = 0
	m.detail = nil
	m.nextOffset = nil
	m.replaceList = true
	return m.loadList(false)
}

func allStatuses() []entity.Status {
	return []entity.Status{entity.StatusBacklog, entity.StatusReady, entity.StatusInProgress, entity.StatusReview, entity.StatusDone, entity.StatusBlocked, entity.StatusFailed, entity.StatusCancelled, entity.StatusDuplicate}
}

func (m *tuiModel) toggleFilterOption() {
	statuses := allStatuses()
	if m.filterCursor < len(statuses) {
		status := statuses[m.filterCursor]
		for i, selected := range m.states {
			if selected == status {
				m.states = append(m.states[:i], m.states[i+1:]...)
				return
			}
		}
		m.states = append(m.states, status)
		return
	}
	project := m.projects[m.filterCursor-len(statuses)]
	for i, id := range m.projectIDs {
		if id == project.ID {
			m.projectIDs = append(m.projectIDs[:i], m.projectIDs[i+1:]...)
			return
		}
	}
	m.projectIDs = append(m.projectIDs, project.ID)
}

func (m tuiModel) openArtifact() tea.Cmd {
	if m.detail == nil || m.artifactSelection < 0 || m.artifactSelection >= len(m.detail.Artifacts) {
		return nil
	}
	artifact := m.detail.Artifacts[m.artifactSelection]
	parsed, err := url.Parse(artifact.DataValue)
	if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" && parsed.User == nil {
		return func() tea.Msg { return browserResultMsg{err: m.openURL(artifact.DataValue)} }
	}
	return func() tea.Msg { return browserResultMsg{err: errors.New("no safe artifact URL selected")} }
}

func mergeIssues(current, incoming []entity.Issue) []entity.Issue {
	byID := make(map[int64]entity.Issue, len(current)+len(incoming))
	for _, issue := range current {
		byID[issue.ID] = issue
	}
	for _, issue := range incoming {
		byID[issue.ID] = issue
	}
	result := make([]entity.Issue, 0, len(byID))
	for _, issue := range byID {
		result = append(result, issue)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].UpdatedAt.Equal(result[j].UpdatedAt) {
			return result[i].ID > result[j].ID
		}
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})
	return result
}

func mergeComments(older, current []entity.Comment) []entity.Comment {
	seen := make(map[int64]struct{}, len(older)+len(current))
	result := make([]entity.Comment, 0, len(older)+len(current))
	for _, page := range [][]entity.Comment{older, current} {
		for _, comment := range page {
			if _, ok := seen[comment.ID]; ok {
				continue
			}
			seen[comment.ID] = struct{}{}
			result = append(result, comment)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func olderCursor(current, refreshed *int64) *int64 {
	if current == nil {
		return nil
	}
	if refreshed == nil || *current < *refreshed {
		value := *current
		return &value
	}
	value := *refreshed
	return &value
}

func (m tuiModel) selectedIssue() (entity.Issue, bool) {
	if m.selection < 0 || m.selection >= len(m.issues) {
		return entity.Issue{}, false
	}
	return m.issues[m.selection], true
}
func (m tuiModel) selectedID() int64 {
	issue, ok := m.selectedIssue()
	if !ok {
		return 0
	}
	return issue.ID
}
func (m *tuiModel) restoreSelection(id int64) {
	if id == 0 {
		m.selection = min(m.selection, max(len(m.issues)-1, 0))
		return
	}
	for i := range m.issues {
		if m.issues[i].ID == id {
			m.selection = i
			return
		}
	}
	m.selection = min(m.selection, max(len(m.issues)-1, 0))
}
func (m *tuiModel) resizeViewport() {
	m.viewport.Width = max(m.width/2-4, 20)
	if m.width < narrowTUIWidth {
		m.viewport.Width = max(m.width-4, 20)
	}
	m.viewport.Height = max(m.height-8, 4)
	m.refreshViewport()
}
func (m *tuiModel) refreshViewport() { m.viewport.SetContent(m.detailContent()) }

func (m tuiModel) View() string {
	footer := "EXPERIMENTAL — interface may change"
	if m.width > 0 && (m.width < minimumTUIWidth || m.height < minimumTUIHeight) {
		return lipgloss.NewStyle().Padding(1).Render(fmt.Sprintf("Terminal too small (%dx%d). Minimum: %dx%d\n\n%s", m.width, m.height, minimumTUIWidth, minimumTUIHeight, footer))
	}
	if m.help {
		return m.overlay("Help", "q/Ctrl+C exit  ↑/↓/j/k navigate  Enter detail  Esc back\nTab tabs  / search  f filters  r refresh  o open artifact  ? help")
	}
	if m.filter {
		return m.overlay("Filters", m.filterView())
	}
	if m.searching {
		return m.overlay("Search", m.searchInput.View())
	}
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Tasq Console") + "  " + statusSummary(m.states)
	if m.errorText != "" && len(m.issues) == 0 {
		return header + "\n\nTracker unavailable: " + m.errorText + "\nPress r to retry\n\n" + footer
	}
	errorBanner := ""
	if m.errorText != "" {
		errorBanner = "Tracker unavailable: " + m.errorText + " (press r to retry)\n"
	}
	list := m.listView()
	detail := m.detailView()
	body := list
	if m.width >= narrowTUIWidth {
		body = lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(max(m.width/2-2, 30)).Render(list), detail)
	} else if m.detailScreen {
		body = detail
	}
	return header + "\n" + errorBanner + body + "\n" + m.statusLine + "\n" + footer
}

func (m tuiModel) filterView() string {
	var b strings.Builder
	b.WriteString("Statuses (none selected means all):\n")
	for i, status := range allStatuses() {
		fmt.Fprintf(&b, "%s [%s] %s\n", cursorMark(i == m.filterCursor), checkMark(containsStatus(m.states, status)), status)
	}
	b.WriteString("Projects (none selected means all):\n")
	for i, project := range m.projects {
		index := len(allStatuses()) + i
		fmt.Fprintf(&b, "%s [%s] %s\n", cursorMark(index == m.filterCursor), checkMark(containsInt64(m.projectIDs, project.ID)), project.Key)
	}
	b.WriteString("\nSpace toggles; Enter applies")
	return b.String()
}

func containsStatus(values []entity.Status, target entity.Status) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
func containsInt64(values []int64, target int64) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
func cursorMark(selected bool) string {
	if selected {
		return ">"
	}
	return " "
}
func checkMark(selected bool) string {
	if selected {
		return "x"
	}
	return " "
}

func (m tuiModel) listView() string {
	if m.loading && len(m.issues) == 0 {
		return "Loading issues…"
	}
	if len(m.issues) == 0 {
		return "No issues match the current filters."
	}
	var b strings.Builder
	for i, issue := range m.issues {
		prefix := "  "
		if i == m.selection {
			prefix = "> "
		}
		fmt.Fprintf(&b, "%s#%d [%s] %s\n", prefix, issue.ID, issue.Status, issue.Title)
	}
	if m.loadingMore {
		b.WriteString("  Loading more…\n")
	} else if m.nextOffset != nil {
		b.WriteString("  ↓ more\n")
	}
	return b.String()
}

func (m tuiModel) detailView() string {
	tabs := make([]string, len(tabNames))
	for i, name := range tabNames {
		if tuiTab(i) == m.tab {
			tabs[i] = "[" + name + "]"
		} else {
			tabs[i] = name
		}
	}
	return strings.Join(tabs, "  ") + "\n" + m.viewport.View()
}

func (m tuiModel) detailContent() string {
	if m.detail == nil {
		return "Select an issue."
	}
	switch m.tab {
	case tabOverview:
		return fmt.Sprintf("#%d %s\nProject: %s  Status: %s  Priority: %s\nAssignee: %s\nUpdated: %s\n\n%s", m.detail.ID, m.detail.Title, m.detail.ProjectKey, m.detail.Status, m.detail.Priority, emptyFallback(m.detail.Assignee, "unassigned"), m.detail.UpdatedAt.Local().Format(time.RFC3339), renderMarkdown(m.detail.Description, m.viewport.Width))
	case tabComments:
		if len(m.comments) == 0 {
			return "No comments."
		}
		var b strings.Builder
		if m.commentCursor != nil {
			b.WriteString("Older comments available (PgUp loads the next page).\n\n")
		}
		if m.commentsLoading {
			b.WriteString("Loading older comments…\n\n")
		}
		for _, comment := range m.comments {
			fmt.Fprintf(&b, "%s · %s · %s\n%s\n\n", comment.Author, comment.Type, comment.CreatedAt.Local().Format(time.RFC3339), renderMarkdown(replaceAttachments(comment.Body), m.viewport.Width))
		}
		return b.String()
	case tabArtifacts:
		if len(m.detail.Artifacts) == 0 {
			return "No artifacts."
		}
		var b strings.Builder
		for i, artifact := range m.detail.Artifacts {
			marker := " "
			if i == m.artifactSelection {
				marker = ">"
			}
			fmt.Fprintf(&b, "%s %s (%s)\n%s\nPress o to open.\n\n", marker, artifact.Type, artifact.DataType, artifact.DataValue)
		}
		return b.String()
	case tabRun:
		if m.runtimeAbsent {
			return "No run data for this issue."
		}
		if m.runtimeError != "" {
			return fmt.Sprintf("Run data unavailable (last error %s):\n%s", m.runtimeErrorAt.Local().Format(time.RFC3339), m.runtimeError)
		}
		if m.runtime == nil {
			return "Run data is loading."
		}
		var b strings.Builder
		fmt.Fprintf(&b, "Status: %s\nWorkspace: %s\nAttempt: %d (restarts %d)\n", m.runtime.Status, m.runtime.Workspace.Path, m.runtime.Attempts.CurrentRetryAttempt, m.runtime.Attempts.RestartCount)
		for _, run := range m.runtime.Runs {
			fmt.Fprintf(&b, "\n%s  %s  attempt %d", run.RunID, run.Status, run.Attempt)
		}
		for _, event := range m.runtime.RecentEvents {
			fmt.Fprintf(&b, "\n%s  %s  %s", event.At.Local().Format(time.RFC3339), event.Event, event.Message)
		}
		if m.runtime.LastError != nil {
			fmt.Fprintf(&b, "\nLast error: %v", m.runtime.LastError)
		}
		return b.String()
	}
	return ""
}

func (m tuiModel) overlay(title, body string) string {
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Render(title + "\n\n" + body + "\n\nEXPERIMENTAL — interface may change")
}
func statusSummary(states []entity.Status) string {
	if len(states) == 0 {
		return "statuses: all"
	}
	values := make([]string, len(states))
	for i, value := range states {
		values[i] = string(value)
	}
	return "statuses: " + strings.Join(values, ",")
}
func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
func replaceAttachments(value string) string {
	return strings.ReplaceAll(value, "attachment://", "attachment (terminal image unavailable): attachment://")
}
func renderMarkdown(value string, width int) string {
	if strings.TrimSpace(value) == "" {
		return "No description."
	}
	renderer, err := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"), glamour.WithWordWrap(max(width, 20)))
	if err != nil {
		return value
	}
	rendered, err := renderer.Render(value)
	if err != nil {
		return value
	}
	return strings.TrimSpace(rendered)
}

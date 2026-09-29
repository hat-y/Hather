package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hat-y/Hather/internal/app"
	"github.com/hat-y/Hather/internal/model"
)

type Core interface {
	Search(context.Context, string) ([]model.Wallpaper, error)
	Preview(context.Context, model.Wallpaper) model.OperationResult
	Apply(context.Context, app.ApplyRequest) model.OperationResult
}

type Model struct {
	Core         Core
	DefaultQuery string
	Input        textinput.Model
	Results      []model.Wallpaper
	Selected     model.Wallpaper
	Adapters     map[string]bool

	local, applying, previewing bool
	cursor                      int
	message                     string
	result                      model.OperationResult
	previewed                   model.Wallpaper
}

type searchResult struct {
	walls []model.Wallpaper
	err   error
}
type previewResult struct {
	result    model.OperationResult
	wallpaper model.Wallpaper
}
type applyResult struct{ result model.OperationResult }

func New(core Core) Model {
	input := textinput.New()
	input.Placeholder = "Search Wallhaven"
	input.Focus()
	return Model{Core: core, Input: input, Adapters: make(map[string]bool)}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case searchResult:
		m.Results, m.Selected, m.result, m.previewed, m.previewing = nil, model.Wallpaper{}, model.OperationResult{}, model.Wallpaper{}, false
		if msg.err != nil {
			m.message = "Error: " + msg.err.Error()
			return m, nil
		}
		m.Results = msg.walls
		if len(msg.walls) == 0 {
			m.message = "No Wallhaven results. Try a different query."
			return m, nil
		}
		m.cursor, m.Selected, m.message = 0, msg.walls[0], ""
		return m, nil
	case previewResult:
		if !sameWallpaper(msg.wallpaper, m.Selected) {
			return m, nil
		}
		m.previewing = false
		m.result, m.message = msg.result, ""
		m.previewed = model.Wallpaper{}
		if msg.result.Status == model.OperationComplete {
			m.previewed = msg.wallpaper
		}
		if m.local && msg.result.Status == model.OperationPrerequisiteFailure {
			m.message = "Unreadable local path. Choose a readable image."
		}
		return m, nil
	case applyResult:
		m.applying, m.result, m.message = false, msg.result, ""
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if m.Input.Focused() {
			if msg.Type == tea.KeyEnter {
				m.Input.Blur()
				if m.local {
					path := m.Input.Value()
					m.Selected, m.result, m.previewed = model.Wallpaper{SourceKind: "local", ID: path, ImageURL: path, Title: path}, model.OperationResult{}, model.Wallpaper{}
					m.previewing = path != ""
					return m, m.preview()
				}
				m.previewing, m.result, m.previewed = false, model.OperationResult{}, model.Wallpaper{}
				return m, m.search()
			}
			m.Input, _ = m.Input.Update(msg)
			return m, nil
		}
		switch string(msg.Runes) {
		case "q":
			return m, tea.Quit
		case "j", "k":
			if len(m.Results) > 0 {
				delta := 1
				if string(msg.Runes) == "k" {
					delta = -1
				}
				m.cursor = (m.cursor + delta + len(m.Results)) % len(m.Results)
				m.Selected, m.result, m.previewed, m.previewing = m.Results[m.cursor], model.OperationResult{}, model.Wallpaper{}, false
			}
			return m, nil
		case "l", "s":
			m.local, m.message, m.previewing = string(msg.Runes) == "l", "", false
			m.Input.SetValue("")
			if m.local {
				m.Input.Placeholder = "Local image path"
			} else {
				m.Input.Placeholder = "Search Wallhaven"
			}
			return m, m.Input.Focus()
		case "p":
			if m.Selected.ID == "" {
				m.message = "Select a wallpaper before previewing."
				return m, nil
			}
			m.previewing, m.message = true, ""
			return m, m.preview()
		case "1", "2", "3", "4", "5":
			ids := []string{"macos", "ghostty", "herdr", "neovim", "vscode"}
			id := ids[int(msg.Runes[0]-'1')]
			m.Adapters[id] = !m.Adapters[id]
			return m, nil
		case "a":
			if m.applying {
				m.message = "An operation is already applying."
				return m, nil
			}
			adapters := m.selectedAdapters()
			if len(adapters) == 0 {
				m.message = "Select at least one adapter before applying."
				return m, nil
			}
			if m.Selected.ID == "" {
				m.message = "Select a wallpaper before applying."
				return m, nil
			}
			if !sameWallpaper(m.previewed, m.Selected) {
				m.message = "Preview the selected wallpaper successfully before applying."
				return m, nil
			}
			m.applying, m.message = true, ""
			return m, m.apply(adapters)
		}
	}
	m.Input, _ = m.Input.Update(msg)
	return m, nil
}

func (m Model) search() tea.Cmd {
	query := m.Input.Value()
	if query == "" {
		query = m.DefaultQuery
	}
	return func() tea.Msg {
		walls, err := m.Core.Search(context.Background(), query)
		return searchResult{walls, err}
	}
}

func (m Model) preview() tea.Cmd {
	wall := m.Selected
	return func() tea.Msg {
		return previewResult{result: m.Core.Preview(context.Background(), wall), wallpaper: wall}
	}
}

func (m Model) apply(adapters []string) tea.Cmd {
	wall := m.Selected
	return func() tea.Msg {
		return applyResult{m.Core.Apply(context.Background(), app.ApplyRequest{Wallpaper: wall, Adapters: adapters})}
	}
}

func sameWallpaper(left, right model.Wallpaper) bool {
	return left.SourceKind == right.SourceKind && left.ID == right.ID && left.ImageURL == right.ImageURL
}

func (m Model) selectedAdapters() []string {
	var selected []string
	for _, id := range []string{"macos", "ghostty", "herdr", "neovim", "vscode"} {
		if m.Adapters[id] {
			selected = append(selected, id)
		}
	}
	return selected
}

func (m Model) View() string {
	hint := "s search · l local path · j/k navigate · p preview · 1-5 adapters · a apply · q quit"
	if m.Input.Focused() {
		hint = "type query/path · Enter submits · Ctrl+C quits"
	}
	lines := []string{lipgloss.NewStyle().Bold(true).Render("Hather"), m.Input.View(), hint}
	for i, id := range []string{"macos", "ghostty", "herdr", "neovim", "vscode"} {
		state := "unselected"
		if m.Adapters[id] {
			state = "selected"
		}
		lines = append(lines, fmt.Sprintf("%d %s [%s]", i+1, id, state))
	}
	if m.Selected.ID != "" {
		lines = append(lines, "Selected: "+m.Selected.Title)
		page, attribution := m.Selected.PageURL, m.Selected.Attribution
		if page == "" {
			page = m.Selected.Metadata["page_url"]
		}
		if attribution == "" {
			attribution = m.Selected.Metadata["attribution"]
		}
		for _, field := range []struct{ label, value string }{
			{"page", page}, {"attribution", attribution},
		} {
			if field.value != "" {
				lines = append(lines, field.label+": "+field.value)
			}
		}
	}
	if len(m.Results) > 0 {
		lines = append(lines, fmt.Sprintf("Results: %d/%d", m.cursor+1, len(m.Results)))
		start := m.cursor - 2
		if start < 0 {
			start = 0
		}
		if start > len(m.Results)-5 {
			start = len(m.Results) - 5
		}
		if start < 0 {
			start = 0
		}
		end := start + 5
		if end > len(m.Results) {
			end = len(m.Results)
		}
		for i := start; i < end; i++ {
			marker := "  "
			if i == m.cursor {
				marker = "> "
			}
			lines = append(lines, marker+m.Results[i].Title)
		}
		for _, key := range []string{"resolution", "category", "purity"} {
			if value := m.Selected.Metadata[key]; value != "" {
				lines = append(lines, key+": "+value)
			}
		}
	}
	if m.result.Status == model.OperationPartialFailure {
		lines = append(lines, "Partial failure: targets are independent; unavailable targets were not configured.")
	}
	for _, result := range m.result.AdapterResults {
		lines = append(lines, result.AdapterID+": "+string(result.Status))
		if result.ErrorClass != "" {
			lines = append(lines, "Error class: "+result.ErrorClass)
		}
		if result.ErrorClass == "permission" {
			lines = append(lines, "Permission required")
		}
		if len(result.ChangedPaths) > 0 {
			lines = append(lines, "Changed paths: "+strings.Join(result.ChangedPaths, ", "))
		}
		if len(result.GeneratedArtifacts) > 0 {
			lines = append(lines, "Generated artifacts: "+strings.Join(result.GeneratedArtifacts, ", "))
		}
		if len(result.Diagnostics) > 0 {
			lines = append(lines, "Diagnostics: "+strings.Join(result.Diagnostics, "; "))
		}
		if len(result.FollowUp) > 0 {
			lines = append(lines, "Follow-up: "+strings.Join(result.FollowUp, "; "))
		}
	}
	lines = append(lines, m.result.Diagnostics...)
	if p := m.result.Palette; p.Accent != "" {
		colors := make([]string, 0, 4)
		for _, color := range []struct{ label, hex string }{{"Background", p.Background}, {"Foreground", p.Foreground}, {"Muted", p.Muted}, {"Accent", p.Accent}} {
			colors = append(colors, color.label+": "+color.hex+" "+lipgloss.NewStyle().Background(lipgloss.Color(color.hex)).Render("  "))
		}
		lines = append(lines, "Palette preview: "+p.Accent+"  "+strings.Join(colors[:2], "  "), strings.Join(colors[2:], "  "))
		if len(p.Diagnostics) > 0 {
			lines = append(lines, "Palette diagnostics: "+strings.Join(p.Diagnostics, "; "))
		}
	}
	if m.result.Status != "" {
		lines = append(lines, "Status: "+string(m.result.Status))
	}
	if m.result.Status == model.OperationComplete && sameWallpaper(m.previewed, m.Selected) {
		lines = append(lines, "Preview ready")
	}
	if m.previewing {
		lines = append(lines, "Previewing selected wallpaper...")
	}
	if m.applying {
		lines = append(lines, "Applying...")
	}
	if m.message != "" {
		lines = append(lines, m.message)
	}
	return strings.Join(lines, "\n")
}

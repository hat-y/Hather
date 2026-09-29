package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hat-y/Hather/internal/app"
	"github.com/hat-y/Hather/internal/model"
)

type fakeCore struct {
	walls       []model.Wallpaper
	searchErr   error
	preview     model.OperationResult
	apply       model.OperationResult
	query       string
	previewWall model.Wallpaper
	request     app.ApplyRequest
	applies     int
}

func (f *fakeCore) Search(_ context.Context, query string) ([]model.Wallpaper, error) {
	f.query = query
	return f.walls, f.searchErr
}
func (f *fakeCore) Preview(_ context.Context, wall model.Wallpaper) model.OperationResult {
	f.previewWall = wall
	return f.preview
}
func (f *fakeCore) Apply(_ context.Context, request app.ApplyRequest) model.OperationResult {
	f.applies++
	f.request = request
	return f.apply
}

func update(m Model, msg tea.Msg) (Model, tea.Cmd) {
	next, cmd := m.Update(msg)
	return next.(Model), cmd
}
func run(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected asynchronous command")
	}
	m, _ = update(m, cmd())
	return m
}

func TestConfiguredQueryStartsSearchButUserInputWins(t *testing.T) {
	core := &fakeCore{walls: []model.Wallpaper{{ID: "w1", Title: "Forest"}}}
	m := New(core)
	m.Input.SetValue("forest")
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyEnter})
	run(t, m, cmd)
	if core.query != "forest" {
		t.Fatal(core.query)
	}
	m = New(core)
	m.Input.SetValue("ocean")
	m, cmd = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	run(t, m, cmd)
	if core.query != "ocean" {
		t.Fatal(core.query)
	}
}

func TestConfiguredQueryFallbackAfterClearingOrReturningToSearch(t *testing.T) {
	core := &fakeCore{walls: []model.Wallpaper{{ID: "w1", Title: "Forest"}}}
	m := New(core)
	m.DefaultQuery = "forest"
	m.Input.SetValue("forest")
	m.Input.SetValue("") // user clears the preloaded query
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = run(t, m, cmd)
	if core.query != "forest" {
		t.Fatalf("cleared search query = %q", core.query)
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m, cmd = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = run(t, m, cmd)
	if core.query != "forest" {
		t.Fatalf("return-to-search query = %q", core.query)
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m.Input.SetValue("ocean")
	m, cmd = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	run(t, m, cmd)
	if core.query != "ocean" {
		t.Fatalf("explicit search query = %q", core.query)
	}
}

func TestSearchRoutesCommandAndShowsResultsOrRecovery(t *testing.T) {
	wall := model.Wallpaper{ID: "w1", Title: "Forest"}
	core := &fakeCore{walls: []model.Wallpaper{wall, {ID: "w2", Title: "Ocean"}}}
	m := New(core)
	m.Input.SetValue("forest")
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = run(t, m, cmd)
	if core.query != "forest" || !strings.Contains(m.View(), "Forest") || !strings.Contains(m.View(), "Selected: Forest") {
		t.Fatalf("search view = %q, query = %q", m.View(), core.query)
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.Selected.ID != "w2" || !strings.Contains(m.View(), "Selected: Ocean") {
		t.Fatalf("selection view = %q", m.View())
	}
	m = New(&fakeCore{})
	m.Input.SetValue("nothing")
	m, cmd = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = run(t, m, cmd)
	if !strings.Contains(m.View(), "No Wallhaven results. Try a different query.") {
		t.Fatalf("empty search view = %q", m.View())
	}
}

func TestLocalPreviewAndErrorsNeverRenderBlank(t *testing.T) {
	core := &fakeCore{preview: model.OperationResult{Status: model.OperationPrerequisiteFailure, Diagnostics: []string{"cannot read image"}}}
	m := New(core)
	m.Input.Blur()
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m.Input.SetValue("/missing/image.png")
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = run(t, m, cmd)
	if core.previewWall.SourceKind != "local" || !strings.Contains(m.View(), "Unreadable local path") || !strings.Contains(m.View(), "Choose a readable image") {
		t.Fatalf("local recovery view = %q, wallpaper = %#v", m.View(), core.previewWall)
	}

	m = New(&fakeCore{searchErr: errors.New("service unavailable")})
	m.Input.SetValue("forest")
	m, cmd = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = run(t, m, cmd)
	if !strings.Contains(m.View(), "Error: service unavailable") {
		t.Fatalf("error view = %q", m.View())
	}
}

func TestPreviewSelectionTogglesApplyProgressAndDiagnostics(t *testing.T) {
	wall := model.Wallpaper{ID: "w1", Title: "Forest"}
	preview := model.OperationResult{Status: model.OperationComplete, Palette: model.Palette{Accent: "#abcdef"}}
	applied := model.OperationResult{Status: model.OperationPartialFailure, AdapterResults: []model.AdapterResult{
		{AdapterID: "macos", Status: model.AdapterFailed, ErrorClass: "permission", Diagnostics: []string{"Apple Events denied"}, FollowUp: []string{"System Settings → Privacy & Security → Automation"}},
		{AdapterID: "ghostty", Status: model.AdapterApplied, FollowUp: []string{"Reload Ghostty"}},
	}}
	core := &fakeCore{walls: []model.Wallpaper{wall}, preview: preview, apply: applied}
	m := New(core)
	m.Input.SetValue("forest")
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = run(t, m, cmd)
	m, cmd = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m = run(t, m, cmd)
	if !strings.Contains(m.View(), "Palette preview: #abcdef") {
		t.Fatalf("preview view = %q", m.View())
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	m, cmd = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if !strings.Contains(m.View(), "Applying...") || !strings.Contains(m.View(), "1 macos [selected]") {
		t.Fatalf("progress view = %q", m.View())
	}
	m = run(t, m, cmd)
	view := m.View()
	if core.applies != 1 || len(core.request.Adapters) != 1 || core.request.Adapters[0] != "macos" || !strings.Contains(view, "Partial failure") || !strings.Contains(view, "macos: failed") || !strings.Contains(view, "Apple Events denied") || !strings.Contains(view, "Permission required") || !strings.Contains(view, "Reload Ghostty") {
		t.Fatalf("apply view = %q, request = %#v", view, core.request)
	}
}

func TestSelectedWallhavenAttributionAndLocalWithoutAttribution(t *testing.T) {
	wall := model.Wallpaper{SourceKind: "wallhaven", ID: "abc", Title: "Forest", Metadata: map[string]string{"page_url": "https://wallhaven.cc/w/abc", "attribution": "Wallhaven: https://wallhaven.cc/w/abc"}}
	m := New(&fakeCore{walls: []model.Wallpaper{wall}})
	m.Input.SetValue("forest")
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = run(t, m, cmd)
	for _, want := range []string{"page: https://wallhaven.cc/w/abc", "attribution: Wallhaven: https://wallhaven.cc/w/abc"} {
		if !strings.Contains(m.View(), want) {
			t.Fatalf("missing %q in %q", want, m.View())
		}
	}
	m.Selected = model.Wallpaper{SourceKind: "local", ID: "local.png", Title: "local.png"}
	m.Results = nil
	if strings.Contains(m.View(), "attribution:") || strings.Contains(m.View(), "page:") {
		t.Fatalf("local view=%q", m.View())
	}
}

func TestWallhavenLabelsAndPreviewGateApply(t *testing.T) {
	wall := model.Wallpaper{SourceKind: "wallhaven", ID: "w1", Title: "Wallhaven w1", Metadata: map[string]string{"resolution": "1920x1080", "category": "anime"}}
	second := model.Wallpaper{SourceKind: "wallhaven", ID: "w2", Title: "Wallhaven w2"}
	core := &fakeCore{walls: []model.Wallpaper{wall, second}, preview: model.OperationResult{Status: model.OperationComplete}, apply: model.OperationResult{Status: model.OperationComplete}}
	m := New(core)
	m.Input.SetValue("forest")
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = run(t, m, cmd)
	for _, want := range []string{"Wallhaven w1", "resolution: 1920x1080", "category: anime"} {
		if !strings.Contains(m.View(), want) {
			t.Fatalf("missing %q in %q", want, m.View())
		}
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	m, cmd = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd != nil || core.applies != 0 || !strings.Contains(m.View(), "Preview the selected wallpaper successfully before applying.") {
		t.Fatalf("apply without preview view=%q calls=%d", m.View(), core.applies)
	}
	m, cmd = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m = run(t, m, cmd)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m, cmd = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd != nil || core.applies != 0 || !strings.Contains(m.View(), "Preview the selected wallpaper successfully before applying.") {
		t.Fatalf("changed selection applied without preview: %q", m.View())
	}
	m, cmd = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m = run(t, m, cmd)
	m, cmd = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = run(t, m, cmd)
	if core.applies != 1 || core.request.Wallpaper.ID != "w2" {
		t.Fatalf("apply calls=%d request=%#v", core.applies, core.request)
	}
}

func TestPreviewCompletingAfterSelectionChangesDoesNotAuthorizeNewSelection(t *testing.T) {
	first := model.Wallpaper{SourceKind: "wallhaven", ID: "w1", Title: "First"}
	second := model.Wallpaper{SourceKind: "wallhaven", ID: "w2", Title: "Second"}
	core := &fakeCore{preview: model.OperationResult{Status: model.OperationComplete}, apply: model.OperationResult{Status: model.OperationComplete}}
	m := New(core)
	m.Input.Blur()
	m.Results, m.Selected = []model.Wallpaper{first, second}, first
	m.Adapters["macos"] = true

	m, preview := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m, _ = update(m, preview())
	m, apply := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})

	if apply != nil || core.applies != 0 || !strings.Contains(m.View(), "Preview the selected wallpaper successfully before applying.") {
		t.Fatalf("stale preview authorized selection %q: view=%q calls=%d", m.Selected.ID, m.View(), core.applies)
	}
}

func TestApplyWithoutExplicitAdapterShowsGuidance(t *testing.T) {
	m := New(&fakeCore{})
	m.Input.Blur()
	m.Selected = model.Wallpaper{ID: "w1", Title: "Forest"}
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd != nil || !strings.Contains(m.View(), "Select at least one adapter before applying.") {
		t.Fatalf("empty selection view = %q", m.View())
	}
}

func TestFocusedInputOwnsControlsAndEnterBlurs(t *testing.T) {
	core := &fakeCore{walls: []model.Wallpaper{{ID: "w1", Title: "Forest"}}}
	m := New(core)
	var cmd tea.Cmd
	for _, key := range "aql12345" {
		m, cmd = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(string(key))})
		if cmd != nil {
			t.Fatalf("focused %q returned command", key)
		}
	}
	if m.Input.Value() != "aql12345" || m.local || len(m.Adapters) != 0 {
		t.Fatalf("focused input state = value %q local %t adapters %#v", m.Input.Value(), m.local, m.Adapters)
	}
	m.Input.SetValue("forest")
	m, cmd = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.Input.Focused() || cmd == nil {
		t.Fatalf("enter must commit and blur: focused=%t cmd=%v", m.Input.Focused(), cmd)
	}
	m = run(t, m, cmd)
	if core.query != "forest" {
		t.Fatalf("search query = %q", core.query)
	}
	_, cmd = update(m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("Ctrl+C must quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("Ctrl+C message = %T", cmd())
	}
}

func TestChangesClearStaleResultsAndApplyIsSingleFlight(t *testing.T) {
	core := &fakeCore{apply: model.OperationResult{Status: model.OperationComplete}}
	m := New(core)
	m.Input.Blur()
	m.Results = []model.Wallpaper{{ID: "w1", Title: "Forest"}, {ID: "w2", Title: "Ocean"}}
	m.Selected = m.Results[0]
	m.result = model.OperationResult{Palette: model.Palette{Accent: "#abcdef"}}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.Selected.ID != "w2" || m.result.Palette.Accent != "" {
		t.Fatalf("selection retained stale state: %#v", m)
	}
	m, _ = update(m, searchResult{err: errors.New("offline")})
	if len(m.Results) != 0 || m.Selected.ID != "" || m.result.Palette.Accent != "" {
		t.Fatalf("error retained stale search state: %#v", m)
	}
	m.Results, m.Selected = []model.Wallpaper{{ID: "w1", Title: "Forest"}}, model.Wallpaper{ID: "w1", Title: "Forest"}
	m.previewed = m.Selected
	m.Adapters["macos"] = true
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd == nil || !strings.Contains(m.View(), "Applying...") {
		t.Fatalf("first apply state = %q", m.View())
	}
	m, second := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if second != nil || core.applies != 0 || !strings.Contains(m.View(), "already applying") {
		t.Fatalf("second apply state = %q calls=%d", m.View(), core.applies)
	}
	m = run(t, m, cmd)
	if core.applies != 1 {
		t.Fatalf("apply calls = %d", core.applies)
	}
}

func TestViewHintsAndBoundedResults(t *testing.T) {
	m := New(&fakeCore{})
	if view := m.View(); !strings.Contains(view, "type query/path · Enter submits · Ctrl+C quits") || strings.Contains(view, "1-5 adapters") {
		t.Fatalf("focused hint: %q", view)
	}
	m.Input.Blur()
	for i := 0; i < 24; i++ {
		m.Results = append(m.Results, model.Wallpaper{ID: fmt.Sprint(i), Title: fmt.Sprintf("Result %02d", i), Metadata: map[string]string{"resolution": "large"}})
	}
	m.Selected, m.cursor = m.Results[12], 12
	view := m.View()
	if !strings.Contains(view, "1-5 adapters") || !strings.Contains(view, "13/24") || !strings.Contains(view, "> Result 12") || strings.Contains(view, "Result 00") || strings.Contains(view, "Result 23") || strings.Count(view, "Result ") > 6 || strings.Count(view, "resolution: large") != 1 {
		t.Fatalf("window: %q", view)
	}
	m.result.Status = model.OperationPartialFailure
	if !strings.Contains(m.View(), "targets are independent") || !strings.Contains(m.View(), "unavailable targets were not configured") {
		t.Fatalf("partial: %q", m.View())
	}
}

func visibleTail(view string) string {
	lines := strings.Split(view, "\n")
	if len(lines) > 20 {
		lines = lines[len(lines)-20:]
	}
	return strings.Join(lines, "\n")
}

func TestPreviewFeedbackVisibleBeforeResults(t *testing.T) {
	wall := model.Wallpaper{ID: "w1", Title: "Forest"}
	m := New(&fakeCore{preview: model.OperationResult{Status: model.OperationComplete, Palette: model.Palette{Background: "#010101", Foreground: "#fefefe", Muted: "#999999", Accent: "#abcdef"}}})
	m.Input.Blur()
	for i := 0; i < 24; i++ {
		m.Results = append(m.Results, model.Wallpaper{ID: fmt.Sprint(i), Title: fmt.Sprintf("Result %02d", i)})
	}
	m.Results[0] = wall
	m.Selected = wall
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	if view := visibleTail(m.View()); !strings.Contains(view, "Previewing selected wallpaper...") || !strings.HasSuffix(view, "Previewing selected wallpaper...") {
		t.Fatalf("progress not above adapters: %q", view)
	}
	m = run(t, m, cmd)
	view := m.View()
	view = visibleTail(view)
	for _, line := range strings.Split(view, "\n") {
		if lipgloss.Width(line) > 100 {
			t.Fatalf("preview exceeds 100 columns: %q", line)
		}
	}
	if strings.Contains(view, "Previewing") || !strings.Contains(view, "Preview ready") || !strings.Contains(strings.Join(strings.Split(view, "\n")[len(strings.Split(view, "\n"))-3:], "\n"), "Preview ready") {
		t.Fatalf("preview not above adapters: %q", view)
	}
	for _, want := range []string{"Background: #010101", "Foreground: #fefefe", "Muted: #999999", "Accent: #abcdef"} {
		if !strings.Contains(view, want) {
			t.Fatalf("missing visible swatch or hex %q: %q", want, view)
		}
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m, _ = update(m, previewResult{result: model.OperationResult{Status: model.OperationPrerequisiteFailure}, wallpaper: wall})
	if strings.Contains(m.View(), "Previewing") {
		t.Fatal("error retained progress")
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if strings.Contains(m.View(), "Previewing") {
		t.Fatal("selection retained progress")
	}
}

func TestViewShowsAllAdapterChoicesAndResultDetails(t *testing.T) {
	m := New(&fakeCore{})
	m.Input.Blur()
	m.Adapters["macos"] = true
	m.result = model.OperationResult{
		Palette:        model.Palette{Background: "#010101", Foreground: "#fefefe", Muted: "#999999", Accent: "#abcdef", Diagnostics: []string{"contrast adjusted"}},
		AdapterResults: []model.AdapterResult{{AdapterID: "macos", Status: model.AdapterFailed, ErrorClass: "permission", ChangedPaths: []string{"/a"}, GeneratedArtifacts: []string{"/b"}, Diagnostics: []string{"denied"}, FollowUp: []string{"allow Automation"}}},
	}
	view := m.View()
	for _, want := range []string{"1 macos [selected]", "2 ghostty [unselected]", "3 herdr [unselected]", "4 neovim [unselected]", "5 vscode [unselected]", "Background: #010101", "Foreground: #fefefe", "Muted: #999999", "Accent: #abcdef", "Palette diagnostics: contrast adjusted", "macos: failed", "Error class: permission", "Changed paths: /a", "Generated artifacts: /b", "Diagnostics: denied", "Follow-up: allow Automation"} {
		if !strings.Contains(view, want) {
			t.Fatalf("missing %q in view %q", want, view)
		}
	}
}

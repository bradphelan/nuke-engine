// Package tui provides an interactive terminal UI for selecting and running
// build targets. Both the target picker and the action picker use the same
// fuzzy-search widget so the interaction model is consistent throughout.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---- Styles ----------------------------------------------------------------

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Background(lipgloss.Color("237"))
	kindStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	promptStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	matchStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
)

// ---- Public types ----------------------------------------------------------

// TargetEntry is a display record for one build target.
type TargetEntry struct {
	Name  string
	IsExe bool // true → "run" action is available
}

// Selection is returned by Run when the user commits to an action.
type Selection struct {
	Index  int    // index into the targets slice passed to Run
	Action string // "build" | "clean" | "tree" | "tree-graphviz" | "run"
}

// Run launches the interactive picker and blocks until the user selects a
// target + action or quits (Esc / Ctrl+C from the target screen).
// Returns nil when the user quits without making a selection.
func Run(targets []*TargetEntry) (*Selection, error) {
	if len(targets) == 0 {
		return nil, nil
	}
	p := tea.NewProgram(newModel(targets))
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	if fm, ok := final.(model); ok {
		return fm.result, nil
	}
	return nil, nil
}

// ---- Fuzzy list ------------------------------------------------------------

// item is a single searchable row in the fuzzy list.
type item struct {
	label string // display text
	sub   string // secondary annotation (kind badge, etc.) — may be empty
}

// fuzzyList is a reusable component: a text query input + filtered item list.
type fuzzyList struct {
	all      []item
	query    string
	cursor   int
	filtered []item
	indices  []int // filtered[i] came from all[indices[i]]
}

func newFuzzyList(items []item) fuzzyList {
	fl := fuzzyList{all: items}
	fl.refilter()
	return fl
}

func (fl *fuzzyList) refilter() {
	fl.filtered = fl.filtered[:0]
	fl.indices = fl.indices[:0]
	q := strings.ToLower(fl.query)
	for i, it := range fl.all {
		if q == "" || strings.Contains(strings.ToLower(it.label), q) {
			fl.filtered = append(fl.filtered, it)
			fl.indices = append(fl.indices, i)
		}
	}
	if fl.cursor >= len(fl.filtered) {
		fl.cursor = max(0, len(fl.filtered)-1)
	}
}

// selectedIndex returns the original index of the currently highlighted item,
// or -1 if the list is empty.
func (fl *fuzzyList) selectedIndex() int {
	if len(fl.indices) == 0 {
		return -1
	}
	return fl.indices[fl.cursor]
}

// handleKey processes a keyboard event. Returns committed=true when Enter was
// pressed on a non-empty list, quit=true when Ctrl+C was pressed.
func (fl *fuzzyList) handleKey(key tea.KeyMsg) (committed bool, quit bool) {
	switch key.Type {
	case tea.KeyRunes:
		fl.query += string(key.Runes)
		fl.refilter()
	case tea.KeyBackspace:
		runes := []rune(fl.query)
		if len(runes) > 0 {
			fl.query = string(runes[:len(runes)-1])
			fl.refilter()
		}
	case tea.KeyUp:
		if fl.cursor > 0 {
			fl.cursor--
		}
	case tea.KeyDown:
		if fl.cursor < len(fl.filtered)-1 {
			fl.cursor++
		}
	case tea.KeyEnter:
		if len(fl.filtered) > 0 {
			return true, false
		}
	case tea.KeyCtrlC:
		return false, true
	}
	return false, false
}

// view renders the fuzzy list. renderItem is called for each visible row.
func (fl *fuzzyList) view(title, hint string, renderItem func(it item, originalIdx int, selected bool) string) string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render(title) + "\n\n")
	sb.WriteString(promptStyle.Render("  > ") + fl.query + "█\n\n")

	for i, it := range fl.filtered {
		sb.WriteString(renderItem(it, fl.indices[i], i == fl.cursor) + "\n")
	}
	if len(fl.filtered) == 0 {
		sb.WriteString(dimStyle.Render("  (nothing matches)") + "\n")
	}
	sb.WriteString("\n" + helpStyle.Render(hint))
	return sb.String()
}

// highlightMatch bolds the matching substring inside label.
func highlightMatch(label, query string) string {
	if query == "" {
		return label
	}
	lo := strings.ToLower(label)
	q := strings.ToLower(query)
	idx := strings.Index(lo, q)
	if idx < 0 {
		return label
	}
	return label[:idx] + matchStyle.Render(label[idx:idx+len(query)]) + label[idx+len(query):]
}

// ---- Actions ---------------------------------------------------------------

var stdActions = []item{
	{label: "build"},
	{label: "clean"},
	{label: "tree"},
	{label: "tree-graphviz"},
}
var exeActions = []item{
	{label: "build"},
	{label: "clean"},
	{label: "tree"},
	{label: "tree-graphviz"},
	{label: "run"},
}

func actionsFor(t *TargetEntry) []item {
	if t.IsExe {
		return exeActions
	}
	return stdActions
}

// ---- Model -----------------------------------------------------------------

type screenKind int

const (
	screenTargets screenKind = iota
	screenActions
)

type model struct {
	targets    []*TargetEntry
	screen     screenKind
	targetList fuzzyList
	actionList fuzzyList
	activeIdx  int // original target index selected on the target screen
	result     *Selection
}

func newModel(targets []*TargetEntry) model {
	items := make([]item, len(targets))
	for i, t := range targets {
		kind := "lib"
		if t.IsExe {
			kind = "exe"
		}
		items[i] = item{label: t.Name, sub: kind}
	}
	return model{
		targets:    targets,
		screen:     screenTargets,
		targetList: newFuzzyList(items),
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch m.screen {
	case screenTargets:
		if key.Type == tea.KeyEsc {
			return m, tea.Quit
		}
		committed, quit := m.targetList.handleKey(key)
		if quit {
			return m, tea.Quit
		}
		if committed {
			m.activeIdx = m.targetList.selectedIndex()
			m.actionList = newFuzzyList(actionsFor(m.targets[m.activeIdx]))
			m.screen = screenActions
		}

	case screenActions:
		if key.Type == tea.KeyEsc {
			// Back to target search, preserving the previous query.
			m.screen = screenTargets
			return m, nil
		}
		committed, quit := m.actionList.handleKey(key)
		if quit {
			return m, tea.Quit
		}
		if committed {
			actIdx := m.actionList.selectedIndex()
			m.result = &Selection{
				Index:  m.activeIdx,
				Action: actionsFor(m.targets[m.activeIdx])[actIdx].label,
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	switch m.screen {
	case screenTargets:
		return m.targetList.view(
			"nuke — select target",
			"type to filter  ↑↓ navigate  enter select  esc quit",
			func(it item, _ int, selected bool) string {
				badge := kindStyle.Render("[" + it.sub + "]")
				label := highlightMatch(it.label, m.targetList.query)
				line := fmt.Sprintf("%-24s  %s", label, badge)
				if selected {
					return selectedStyle.Render("> " + line)
				}
				return "    " + line
			},
		)

	case screenActions:
		t := m.targets[m.activeIdx]
		return m.actionList.view(
			fmt.Sprintf("nuke — %s", t.Name),
			"type to filter  ↑↓ navigate  enter run  esc back",
			func(it item, _ int, selected bool) string {
				label := highlightMatch(it.label, m.actionList.query)
				if selected {
					return selectedStyle.Render("> " + label)
				}
				return "    " + label
			},
		)
	}
	return ""
}

package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type tuiFocus int

const (
	boardFocus tuiFocus = iota
	commandFocus
)

type tuiModel struct {
	sudoku     *Sudoku
	original   string
	solution   string
	given      [PuzzleDimension][PuzzleDimension]bool
	row        int
	col        int
	focus      tuiFocus
	input      textinput.Model
	logs       []string
	logOffset  int
	logFollow  bool
	width      int
	height     int
	checkpoint map[string]string
	solved     bool
}

type cellStyle func(row int, col int, text string, given bool, selected bool) string

func runSudokuTUI(puzzle string, solution string) error {
	model, err := newTUIModel(puzzle, solution)
	if err != nil {
		return err
	}

	_, err = tea.NewProgram(model).Run()
	return err
}

func newTUIModel(puzzle string, solution string) (tuiModel, error) {
	if puzzle == "" {
		return tuiModel{}, fmt.Errorf("tui requires --puzzle")
	}

	sudoku := NewSudoku()
	if err := sudoku.Load(puzzle); err != nil {
		return tuiModel{}, fmt.Errorf("could not load puzzle: %w", err)
	}

	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "/help"
	input.SetWidth(72)
	input.Blur()

	model := tuiModel{
		sudoku:     sudoku,
		original:   puzzle,
		solution:   solution,
		input:      input,
		checkpoint: make(map[string]string),
		logs:       []string{"Loaded puzzle. Press / for commands, arrow keys to move, digits to edit."},
		logFollow:  true,
		width:      80,
		height:     24,
	}
	model.loadGivenMask(puzzle)
	model.advanceToEditable()
	return model, nil
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.clampLogOffset()
	case tea.KeyPressMsg:
		if m.focus == commandFocus {
			return m.updateCommand(msg)
		}
		return m.updateBoard(msg)
	}
	return m, nil
}

func (m tuiModel) View() tea.View {
	board := renderSudokuBoard(m.sudoku, m.given, m.row, m.col, styledCell)
	logHeight := m.logHeight()
	logWidth := m.logWidth()

	var builder strings.Builder
	builder.WriteString(titleStyle.Render("Sudoku Solver"))
	builder.WriteString("\n\n")

	if m.wideLayout() {
		left := lipgloss.NewStyle().Width(boardWidth).Render(panelTitleStyle.Render("Puzzle") + "\n" + board)
		right := m.renderLog(logWidth, logHeight)
		builder.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right))
	} else {
		builder.WriteString(panelTitleStyle.Render("Puzzle"))
		builder.WriteByte('\n')
		builder.WriteString(board)
		builder.WriteString("\n")
		builder.WriteString(m.renderLog(logWidth, logHeight))
	}

	builder.WriteString("\n\n")
	builder.WriteString(logTitleStyle.Render("Command"))
	builder.WriteByte('\n')
	builder.WriteString(commandStyle.Width(commandWidth(m.width)).Render(m.input.View()))
	builder.WriteString("\n")
	builder.WriteString(helpHintStyle.Render("Board: arrows/hjkl move, 1-9 set, 0/backspace clears, / commands, pgup/pgdn scroll log, end follows, q quits"))

	view := tea.NewView(builder.String())
	view.AltScreen = true
	view.WindowTitle = "Sudoku Solver"
	return view
}

func (m tuiModel) updateBoard(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		m.move(-1, 0)
	case "down", "j":
		m.move(1, 0)
	case "left", "h":
		m.move(0, -1)
	case "right", "l":
		m.move(0, 1)
	case "pgup", "ctrl+u":
		m.scrollLog(-m.logPageSize())
	case "pgdown", "ctrl+d":
		m.scrollLog(m.logPageSize())
	case "home":
		m.scrollLogToStart()
	case "end":
		m.scrollLogToEnd()
	case "/":
		m.focus = commandFocus
		m.input.Focus()
		m.input.SetValue("/")
	case "0", "backspace", "delete":
		m.clearFocusedCell()
	default:
		if value, ok := digitKey(msg.String()); ok {
			m.setFocusedCell(value)
		}
	}
	return m, nil
}

func (m tuiModel) updateCommand(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.focus = boardFocus
		m.input.Blur()
		m.input.SetValue("")
		return m, nil
	case "enter":
		commandText := strings.TrimSpace(m.input.Value())
		m.appendCommandLog(commandText)
		finished := m.runCommand(commandText)
		m.input.SetValue("")
		m.input.Blur()
		m.focus = boardFocus
		if finished {
			return m, tea.Quit
		}
		return m, nil
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

func (m *tuiModel) loadGivenMask(puzzle string) {
	digits, err := parseDigits(puzzle)
	if err != nil {
		return
	}
	for i, digit := range digits {
		if digit == 0 {
			continue
		}
		row := i / PuzzleDimension
		col := i % PuzzleDimension
		m.given[row][col] = true
	}
}

func (m *tuiModel) appendLog(lines ...string) {
	for _, line := range lines {
		m.logs = append(m.logs, line)
	}
	if m.logFollow {
		m.logOffset = m.maxLogOffset()
	}
}

func (m *tuiModel) appendCommandLog(commandText string) {
	if strings.TrimSpace(commandText) == "" {
		return
	}
	if len(m.logs) > 0 {
		m.logs = append(m.logs, "")
	}
	m.appendLog(commandText)
}

func (m tuiModel) renderLog(width int, height int) string {
	lines := m.logs
	if len(lines) == 0 {
		lines = []string{"No messages yet."}
	}
	visibleLines := m.visibleLogLines(height)
	start, end := m.logWindow(height)
	if start < end {
		visibleLines = lines[start:end]
	}
	scrollState := "follow"
	if !m.logFollow {
		scrollState = fmt.Sprintf("%d/%d", min(start+1, len(lines)), len(lines))
	}
	title := logTitleStyle.Render("Log") + helpHintStyle.Render("  "+scrollState)
	body := logStyle.Width(width).Height(height).Render(strings.Join(visibleLines, "\n"))
	return title + "\n" + body
}

func (m *tuiModel) move(rowDelta int, colDelta int) {
	m.row = clamp(m.row+rowDelta, 0, PuzzleDimension-1)
	m.col = clamp(m.col+colDelta, 0, PuzzleDimension-1)
}

func (m tuiModel) wideLayout() bool {
	return m.width >= boardWidth+minSideLogWidth+layoutGutterWidth
}

func (m tuiModel) logWidth() int {
	if m.wideLayout() {
		return max(minSideLogWidth, m.width-boardWidth-layoutGutterWidth)
	}
	return commandWidth(m.width)
}

func (m tuiModel) logHeight() int {
	if m.wideLayout() {
		return 20
	}
	available := m.height - 21
	return clamp(available, 6, 12)
}

func (m tuiModel) logPageSize() int {
	return max(1, m.logHeight()-1)
}

func (m tuiModel) visibleLogLines(height int) []string {
	lines := make([]string, height)
	for i := range lines {
		lines[i] = ""
	}
	return lines
}

func (m tuiModel) logWindow(height int) (int, int) {
	if len(m.logs) == 0 {
		return 0, 0
	}
	maxOffset := max(0, len(m.logs)-height)
	start := clamp(m.logOffset, 0, maxOffset)
	if m.logFollow {
		start = maxOffset
	}
	end := min(len(m.logs), start+height)
	return start, end
}

func (m tuiModel) maxLogOffset() int {
	return max(0, len(m.logs)-m.logHeight())
}

func (m *tuiModel) clampLogOffset() {
	m.logOffset = clamp(m.logOffset, 0, m.maxLogOffset())
	if m.logOffset == m.maxLogOffset() {
		m.logFollow = true
	}
}

func (m *tuiModel) scrollLog(delta int) {
	m.logFollow = false
	m.logOffset = clamp(m.logOffset+delta, 0, m.maxLogOffset())
	if m.logOffset == m.maxLogOffset() {
		m.logFollow = true
	}
}

func (m *tuiModel) scrollLogToStart() {
	m.logFollow = false
	m.logOffset = 0
}

func (m *tuiModel) scrollLogToEnd() {
	m.logOffset = m.maxLogOffset()
	m.logFollow = true
}

func (m *tuiModel) advanceToEditable() {
	for row := 0; row < PuzzleDimension; row++ {
		for col := 0; col < PuzzleDimension; col++ {
			if !m.given[row][col] {
				m.row = row
				m.col = col
				return
			}
		}
	}
}

func (m *tuiModel) setFocusedCell(value int) {
	if m.given[m.row][m.col] {
		m.appendLog(fmt.Sprintf("Cannot change original clue at (%d, %d).", m.row, m.col))
		return
	}
	m.setCell(m.row, m.col, value)
}

func (m *tuiModel) clearFocusedCell() {
	if m.given[m.row][m.col] {
		m.appendLog(fmt.Sprintf("Cannot clear original clue at (%d, %d).", m.row, m.col))
		return
	}
	if m.sudoku.ClearValue(m.row, m.col) {
		m.appendLog(fmt.Sprintf("Cleared (%d, %d).", m.row, m.col))
	}
}

func (m *tuiModel) setCell(row int, col int, value int) {
	if !inBounds(row, col, PuzzleDimension) {
		m.appendLog(fmt.Sprintf("Cell (%d, %d) is out of bounds.", row, col))
		return
	}
	if value < 1 || value > PuzzleDimension {
		m.appendLog(fmt.Sprintf("%d is not a valid Sudoku value.", value))
		return
	}
	if m.given[row][col] {
		m.appendLog(fmt.Sprintf("Cannot change original clue at (%d, %d).", row, col))
		return
	}
	current, _ := m.sudoku.Value(row, col)
	if current == value {
		m.appendLog(fmt.Sprintf("Cell (%d, %d) is already %d.", row, col, value))
		return
	}
	if current > 0 {
		m.sudoku.ClearValue(row, col)
	}
	if !m.sudoku.IsCandidate(row, col, value) {
		if current > 0 {
			m.sudoku.SetValue(row, col, current)
		}
		m.appendLog(fmt.Sprintf("%d is not valid at (%d, %d).", value, row, col))
		return
	}
	m.sudoku.SetValue(row, col, value)
	m.solved = m.sudoku.IsSolved()
	m.appendLog(fmt.Sprintf("Set (%d, %d) = %d.", row, col, value))
}

func (m *tuiModel) runCommand(commandText string) bool {
	if commandText == "" {
		return false
	}
	if !strings.HasPrefix(commandText, "/") {
		m.appendLog("Commands must start with /. Try /help.")
		return false
	}

	fields := strings.Fields(strings.TrimPrefix(commandText, "/"))
	if len(fields) == 0 {
		return false
	}

	switch fields[0] {
	case "set":
		row, col, value, ok := parseThreeInts(fields[1:])
		if !ok {
			m.appendLog("Usage: /set x y value")
			return false
		}
		m.setCell(row, col, value)
	case "get":
		row, col, ok := parseTwoInts(fields[1:])
		if !ok {
			m.appendLog("Usage: /get x y")
			return false
		}
		value, valid := m.sudoku.Value(row, col)
		m.appendLog(fmt.Sprintf("get: x = %d, y = %d, value (valid=%t) = %d", row, col, valid, value))
	case "clear":
		if err := m.sudoku.Load(m.original); err != nil {
			m.appendLog(fmt.Sprintf("Could not reload puzzle: %v", err))
			return false
		}
		m.solved = false
		m.appendLog("Reset to original puzzle.")
	case "solve":
		if m.sudoku.Solve() {
			m.solved = true
			m.appendLog("Puzzle solved.")
			if m.solution != "" && m.sudoku.Representation() != m.solution {
				m.appendLog("Solved puzzle does not match expected solution.")
			}
		} else {
			m.appendLog("No solution based on current configuration. Try /clear.")
		}
	case "status":
		full, size := m.sudoku.IsFull()
		m.solved = full && m.sudoku.IsSolved()
		m.appendLog(fmt.Sprintf("Solved=%t Full=%t Filled=%d Representation=%s", m.solved, full, size, m.sudoku.Representation()))
	case "save":
		if len(fields) != 2 {
			m.appendLog("Usage: /save name")
			return false
		}
		m.checkpoint[fields[1]] = m.sudoku.Representation()
		m.appendLog(fmt.Sprintf("Saved checkpoint %q.", fields[1]))
	case "load":
		if len(fields) != 2 {
			m.appendLog("Usage: /load name")
			return false
		}
		checkpoint := m.checkpoint[fields[1]]
		if checkpoint == "" {
			m.appendLog(fmt.Sprintf("No checkpoint named %q.", fields[1]))
			return false
		}
		if err := m.sudoku.Load(checkpoint); err != nil {
			m.appendLog(fmt.Sprintf("Could not load checkpoint: %v", err))
			return false
		}
		m.appendLog(fmt.Sprintf("Loaded checkpoint %q.", fields[1]))
	case "checkpoints":
		if len(m.checkpoint) == 0 {
			m.appendLog("No checkpoints saved.")
			return false
		}
		names := make([]string, 0, len(m.checkpoint))
		for name := range m.checkpoint {
			names = append(names, name)
		}
		sort.Strings(names)
		m.appendLog("Checkpoints: " + strings.Join(names, ", "))
	case "help":
		m.appendLog(commandHelpLines()...)
	case "quit":
		return true
	default:
		m.appendLog(fmt.Sprintf("%s: unknown command. Try /help.", fields[0]))
	}
	return false
}

func renderSudokuBoard(sudoku *Sudoku, given [PuzzleDimension][PuzzleDimension]bool, selectedRow int, selectedCol int, style cellStyle) string {
	var builder strings.Builder
	builder.WriteString(boardLine("top"))
	builder.WriteByte('\n')
	for row := 0; row < PuzzleDimension; row++ {
		builder.WriteString(boardRow(sudoku, given, row, selectedRow, selectedCol, style))
		if sum, ok := sudoku.RowSum(row); ok {
			builder.WriteString(fmt.Sprintf("  %2d", sum))
		}
		builder.WriteByte('\n')
		switch row {
		case PuzzleDimension - 1:
			builder.WriteString(boardLine("bottom"))
		case NonetDimension - 1, NonetDimension*2 - 1:
			builder.WriteString(boardLine("double"))
		default:
			builder.WriteString(boardLine("single"))
		}
		builder.WriteByte('\n')
	}
	builder.WriteString(columnSumsLine(sudoku))
	return builder.String()
}

func boardRow(sudoku *Sudoku, given [PuzzleDimension][PuzzleDimension]bool, row int, selectedRow int, selectedCol int, style cellStyle) string {
	var builder strings.Builder
	builder.WriteString("║")
	for col := 0; col < PuzzleDimension; col++ {
		value, _ := sudoku.Value(row, col)
		text := "   "
		if value > 0 {
			text = fmt.Sprintf(" %d ", value)
		}
		if style != nil {
			text = style(row, col, text, given[row][col], row == selectedRow && col == selectedCol)
		}
		builder.WriteString(text)
		if col == PuzzleDimension-1 {
			builder.WriteString("║")
		} else if (col+1)%NonetDimension == 0 {
			builder.WriteString("║")
		} else {
			builder.WriteString("│")
		}
	}
	return builder.String()
}

func boardLine(kind string) string {
	chars := map[string][6]string{
		"top":    {"╔", "╗", "╤", "╦", "═══", ""},
		"single": {"╟", "╢", "┼", "╫", "───", ""},
		"double": {"╠", "╣", "╪", "╬", "═══", ""},
		"bottom": {"╚", "╝", "╧", "╩", "═══", ""},
	}
	selected := chars[kind]
	var builder strings.Builder
	builder.WriteString(selected[0])
	for col := 0; col < PuzzleDimension; col++ {
		builder.WriteString(selected[4])
		if col == PuzzleDimension-1 {
			builder.WriteString(selected[1])
		} else if (col+1)%NonetDimension == 0 {
			builder.WriteString(selected[3])
		} else {
			builder.WriteString(selected[2])
		}
	}
	return builder.String()
}

func columnSumsLine(sudoku *Sudoku) string {
	var builder strings.Builder
	builder.WriteString(" ")
	for col := 0; col < PuzzleDimension; col++ {
		sum, _ := sudoku.ColumnSum(col)
		builder.WriteString(fmt.Sprintf("%3d", sum))
		if col == PuzzleDimension-1 {
			builder.WriteString(" ")
		} else if (col+1)%NonetDimension == 0 {
			builder.WriteString(" ")
		} else {
			builder.WriteString(" ")
		}
	}
	return builder.String()
}

func styledCell(row int, col int, text string, given bool, selected bool) string {
	style := editableCellStyle
	if given {
		style = givenCellStyle
	}
	if selected {
		style = style.Reverse(true)
	}
	return style.Render(text)
}

func commandHelpLines() []string {
	return []string{
		"Commands:",
		"/set x y value     Set editable zero-based row x, column y.",
		"/get x y           Show the value at a cell.",
		"/clear             Reset to the original puzzle.",
		"/solve             Solve from the current board.",
		"/status            Show solved/full state and representation.",
		"/save name         Save current board as a checkpoint.",
		"/load name         Restore a checkpoint.",
		"/checkpoints       List saved checkpoints.",
		"/help              Show this help.",
		"/quit              Exit the TUI.",
	}
}

func digitKey(key string) (int, bool) {
	if len(key) != 1 || key[0] < '1' || key[0] > '9' {
		return 0, false
	}
	return int(key[0] - '0'), true
}

func parseTwoInts(values []string) (int, int, bool) {
	if len(values) != 2 {
		return 0, 0, false
	}
	first, err := strconv.Atoi(values[0])
	if err != nil {
		return 0, 0, false
	}
	second, err := strconv.Atoi(values[1])
	if err != nil {
		return 0, 0, false
	}
	return first, second, true
}

func parseThreeInts(values []string) (int, int, int, bool) {
	if len(values) != 3 {
		return 0, 0, 0, false
	}
	first, second, ok := parseTwoInts(values[:2])
	if !ok {
		return 0, 0, 0, false
	}
	third, err := strconv.Atoi(values[2])
	if err != nil {
		return 0, 0, 0, false
	}
	return first, second, third, true
}

func clamp(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func commandWidth(terminalWidth int) int {
	return max(40, terminalWidth-4)
}

var (
	boardWidth        = utf8.RuneCountInString(boardLine("top")) + len("  45")
	minSideLogWidth   = 40
	layoutGutterWidth = 4

	titleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	panelTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	logTitleStyle   = panelTitleStyle
	helpHintStyle   = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
	commandStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(0, 1)
	logStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(0, 1)
	givenCellStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	editableCellStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
)

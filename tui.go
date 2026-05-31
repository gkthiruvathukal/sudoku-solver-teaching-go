package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
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
	sudoku         *Sudoku
	original       string
	solution       string
	given          [PuzzleDimension][PuzzleDimension]bool
	row            int
	col            int
	focus          tuiFocus
	input          textinput.Model
	logs           []string
	logOffset      int
	logFollow      bool
	width          int
	height         int
	checkpoint     map[string]string
	solved         bool
	traceBase      string
	trace          []TraceEvent
	traceIndex     int
	tracePlay      bool
	traceDelay     time.Duration
	progressActive bool
	progressLabel  string
	progressPath   string
	progressDone   int
	progressTotal  int
}

type cellStyle func(row int, col int, text string, given bool, selected bool) string

type traceTickMsg struct{}

type traceSaveProgressMsg struct {
	label   string
	path    string
	done    int
	total   int
	finish  bool
	err     error
	updates <-chan traceSaveProgressMsg
}

type traceLoadProgressMsg struct {
	path    string
	done    int
	total   int
	finish  bool
	puzzle  string
	events  []TraceEvent
	err     error
	updates <-chan traceLoadProgressMsg
}

type traceFileRecord struct {
	Record string      `json:"record"`
	Puzzle string      `json:"puzzle,omitempty"`
	Event  *TraceEvent `json:"event,omitempty"`
}

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
		traceDelay: defaultTraceDelay,
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
	case traceTickMsg:
		if !m.tracePlay {
			return m, nil
		}
		m.stepTrace(1)
		if m.tracePlay {
			return m, m.traceTick()
		}
		return m, nil
	case traceSaveProgressMsg:
		m.progressLabel = msg.label
		m.progressPath = msg.path
		m.progressDone = msg.done
		m.progressTotal = msg.total
		if msg.finish {
			m.progressActive = false
			if msg.err != nil {
				m.appendLog(fmt.Sprintf("Could not save trace: %v", msg.err))
			} else {
				m.appendLog(fmt.Sprintf("Saved trace to %s.", msg.path))
			}
			return m, nil
		}
		m.progressActive = true
		return m, waitTraceSaveProgress(msg.updates)
	case traceLoadProgressMsg:
		m.progressLabel = "Loading trace"
		m.progressPath = msg.path
		m.progressDone = msg.done
		m.progressTotal = msg.total
		if msg.finish {
			m.progressActive = false
			if msg.err != nil {
				m.appendLog(fmt.Sprintf("Could not load trace: %v", msg.err))
				return m, nil
			}
			if err := m.finishLoadTrace(msg.path, msg.puzzle, msg.events); err != nil {
				m.appendLog(fmt.Sprintf("Could not load trace: %v", err))
			}
			return m, nil
		}
		m.progressActive = true
		return m, waitTraceLoadProgress(msg.updates)
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
	if m.progressActive {
		builder.WriteString(m.renderTraceProgress(commandWidth(m.width)))
		builder.WriteString("\n\n")
	}
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
		finished, cmd := m.runCommandWithCmd(commandText)
		m.input.SetValue("")
		m.input.Blur()
		m.focus = boardFocus
		if finished {
			return m, tea.Quit
		}
		if cmd != nil {
			return m, cmd
		}
		if m.tracePlay {
			return m, m.traceTick()
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
	finished, _ := m.runCommandWithCmd(commandText)
	return finished
}

func (m *tuiModel) runCommandWithCmd(commandText string) (bool, tea.Cmd) {
	if commandText == "" {
		return false, nil
	}
	if !strings.HasPrefix(commandText, "/") {
		m.appendLog("Commands must start with /. Try /help.")
		return false, nil
	}

	fields := strings.Fields(strings.TrimPrefix(commandText, "/"))
	if len(fields) == 0 {
		return false, nil
	}

	switch fields[0] {
	case "set":
		row, col, value, ok := parseThreeInts(fields[1:])
		if !ok {
			m.appendLog("Usage: /set x y value")
			return false, nil
		}
		m.setCell(row, col, value)
	case "get":
		row, col, ok := parseTwoInts(fields[1:])
		if !ok {
			m.appendLog("Usage: /get x y")
			return false, nil
		}
		value, valid := m.sudoku.Value(row, col)
		m.appendLog(fmt.Sprintf("get: x = %d, y = %d, value (valid=%t) = %d", row, col, valid, value))
	case "clear":
		if err := m.sudoku.Load(m.original); err != nil {
			m.appendLog(fmt.Sprintf("Could not reload puzzle: %v", err))
			return false, nil
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
			return false, nil
		}
		m.checkpoint[fields[1]] = m.sudoku.Representation()
		m.appendLog(fmt.Sprintf("Saved checkpoint %q.", fields[1]))
	case "load":
		if len(fields) != 2 {
			m.appendLog("Usage: /load name")
			return false, nil
		}
		checkpoint := m.checkpoint[fields[1]]
		if checkpoint == "" {
			m.appendLog(fmt.Sprintf("No checkpoint named %q.", fields[1]))
			return false, nil
		}
		if err := m.sudoku.Load(checkpoint); err != nil {
			m.appendLog(fmt.Sprintf("Could not load checkpoint: %v", err))
			return false, nil
		}
		m.appendLog(fmt.Sprintf("Loaded checkpoint %q.", fields[1]))
	case "checkpoints":
		if len(m.checkpoint) == 0 {
			m.appendLog("No checkpoints saved.")
			return false, nil
		}
		names := make([]string, 0, len(m.checkpoint))
		for name := range m.checkpoint {
			names = append(names, name)
		}
		sort.Strings(names)
		m.appendLog("Checkpoints: " + strings.Join(names, ", "))
	case "trace":
		return false, m.runTraceCommand(fields[1:])
	case "help":
		m.appendLog(commandHelpLines()...)
	case "quit":
		return true, nil
	default:
		m.appendLog(fmt.Sprintf("%s: unknown command. Try /help.", fields[0]))
	}
	return false, nil
}

func (m *tuiModel) runTraceCommand(args []string) tea.Cmd {
	if len(args) == 0 {
		m.appendLog(traceUsage())
		return nil
	}

	switch args[0] {
	case "solve":
		m.buildTrace()
	case "next":
		m.tracePlay = false
		m.stepTrace(1)
	case "prev":
		m.tracePlay = false
		m.stepTrace(-1)
	case "reset":
		m.tracePlay = false
		m.resetTracePlayback()
	case "play":
		if len(m.trace) == 0 {
			m.appendLog("No trace loaded. Run /trace solve first.")
			return nil
		}
		if m.traceIndex >= len(m.trace) {
			m.appendLog("Trace is already at the end. Run /trace reset to replay.")
			return nil
		}
		m.tracePlay = true
		m.appendLog("Trace playback started.")
	case "pause":
		m.tracePlay = false
		m.appendLog("Trace playback paused.")
	case "status":
		m.appendLog(m.traceStatus())
	case "delay":
		m.setTraceDelay(args[1:])
	case "save":
		return m.saveTrace(args[1:])
	case "load":
		return m.loadTrace(args[1:])
	default:
		m.appendLog(traceUsage())
	}
	return nil
}

func traceUsage() string {
	return "Usage: /trace solve|next|prev|reset|play|pause|status|delay|save|load"
}

func (m *tuiModel) setTraceDelay(args []string) {
	if len(args) != 1 {
		m.appendLog(fmt.Sprintf("Trace delay is %d us. Usage: /trace delay microseconds", m.traceDelay.Microseconds()))
		return
	}
	delay, err := strconv.Atoi(args[0])
	if err != nil || delay < 0 {
		m.appendLog("Usage: /trace delay microseconds")
		return
	}
	m.traceDelay = time.Duration(delay) * time.Microsecond
	m.appendLog(fmt.Sprintf("Trace delay set to %d us.", delay))
}

func (m *tuiModel) saveTrace(args []string) tea.Cmd {
	if len(args) != 1 {
		m.appendLog("Usage: /trace save path.jsonl")
		return nil
	}
	if len(m.trace) == 0 || m.traceBase == "" {
		m.appendLog("No trace loaded. Run /trace solve first.")
		return nil
	}
	m.appendLog(fmt.Sprintf("Saving %d trace events to %s.", len(m.trace), args[0]))
	m.progressActive = true
	m.progressLabel = "Saving trace"
	m.progressPath = args[0]
	m.progressDone = 0
	m.progressTotal = len(m.trace)
	return startTraceSave(args[0], m.traceBase, m.trace)
}

func (m *tuiModel) loadTrace(args []string) tea.Cmd {
	if len(args) != 1 {
		m.appendLog("Usage: /trace load path.jsonl")
		return nil
	}
	m.appendLog(fmt.Sprintf("Loading trace from %s.", args[0]))
	m.progressActive = true
	m.progressLabel = "Loading trace"
	m.progressPath = args[0]
	m.progressDone = 0
	m.progressTotal = 0
	return startTraceLoad(args[0])
}

func (m *tuiModel) finishLoadTrace(path string, puzzle string, events []TraceEvent) error {
	if err := m.sudoku.Load(puzzle); err != nil {
		return fmt.Errorf("trace initial puzzle is invalid: %w", err)
	}
	m.traceBase = puzzle
	m.trace = events
	m.traceIndex = 0
	m.tracePlay = false
	m.solved = false
	m.loadGivenMask(puzzle)
	m.advanceToEditable()
	m.appendLog(fmt.Sprintf("Loaded trace from %s with %d events.", path, len(events)))
	return nil
}

func (m *tuiModel) buildTrace() {
	base := m.original
	working := NewSudoku()
	if err := working.Load(base); err != nil {
		m.appendLog(fmt.Sprintf("Could not trace original puzzle: %v", err))
		return
	}
	events, solved := working.TraceSolve()
	if err := m.sudoku.Load(base); err != nil {
		m.appendLog(fmt.Sprintf("Could not reset trace playback: %v", err))
		return
	}
	m.traceBase = base
	m.trace = events
	m.traceIndex = 0
	m.tracePlay = false
	m.solved = false
	if !solved {
		m.appendLog("Trace did not find a solution.")
		return
	}
	m.appendLog(fmt.Sprintf("Trace ready: %d events. Use /trace next or /trace play.", len(events)))
}

func (m *tuiModel) stepTrace(delta int) {
	if len(m.trace) == 0 {
		m.appendLog("No trace loaded. Run /trace solve first.")
		m.tracePlay = false
		return
	}
	switch {
	case delta > 0:
		if m.traceIndex >= len(m.trace) {
			m.appendLog("Trace is already at the end.")
			m.tracePlay = false
			return
		}
		event := m.trace[m.traceIndex]
		m.applyTraceEvent(event)
		m.traceIndex++
		m.appendLog(m.describeTraceEvent(event))
		if m.traceIndex >= len(m.trace) {
			m.tracePlay = false
			m.appendLog("Trace playback complete.")
		}
	case delta < 0:
		if m.traceIndex == 0 {
			m.appendLog("Trace is already at the beginning.")
			return
		}
		m.replayTrace(m.traceIndex - 1)
		m.appendLog(fmt.Sprintf("Trace rewound to %d/%d.", m.traceIndex, len(m.trace)))
	}
}

func (m *tuiModel) resetTracePlayback() {
	if len(m.trace) == 0 {
		m.appendLog("No trace loaded. Run /trace solve first.")
		return
	}
	if err := m.sudoku.Load(m.traceBase); err != nil {
		m.appendLog(fmt.Sprintf("Could not reset trace playback: %v", err))
		return
	}
	m.traceIndex = 0
	m.solved = false
	m.appendLog("Trace reset to starting board.")
}

func (m *tuiModel) replayTrace(count int) {
	if err := m.sudoku.Load(m.traceBase); err != nil {
		m.appendLog(fmt.Sprintf("Could not rewind trace: %v", err))
		return
	}
	m.traceIndex = 0
	m.solved = false
	for m.traceIndex < count && m.traceIndex < len(m.trace) {
		m.applyTraceEvent(m.trace[m.traceIndex])
		m.traceIndex++
	}
}

func (m *tuiModel) applyTraceEvent(event TraceEvent) {
	switch event.Type {
	case TracePlace:
		m.sudoku.SetValue(event.Row, event.Col, event.Value)
		m.row = event.Row
		m.col = event.Col
	case TraceBacktrack:
		m.sudoku.ClearValue(event.Row, event.Col)
		m.row = event.Row
		m.col = event.Col
	case TraceSolved:
		m.solved = m.sudoku.IsSolved()
	}
}

func (m tuiModel) describeTraceEvent(event TraceEvent) string {
	switch event.Type {
	case TracePlace:
		return fmt.Sprintf("Trace %d/%d: place %d at (%d, %d).", m.traceIndex+1, len(m.trace), event.Value, event.Row, event.Col)
	case TraceBacktrack:
		return fmt.Sprintf("Trace %d/%d: backtrack from (%d, %d), remove %d.", m.traceIndex+1, len(m.trace), event.Row, event.Col, event.Value)
	case TraceSolved:
		return fmt.Sprintf("Trace %d/%d: solved.", m.traceIndex+1, len(m.trace))
	default:
		return fmt.Sprintf("Trace %d/%d: %s.", m.traceIndex+1, len(m.trace), event.Type)
	}
}

func (m tuiModel) traceStatus() string {
	if len(m.trace) == 0 {
		return "No trace loaded. Run /trace solve first."
	}
	state := "paused"
	if m.tracePlay {
		state = "playing"
	}
	return fmt.Sprintf("Trace %s: %d/%d events, delay=%d us.", state, m.traceIndex, len(m.trace), m.traceDelay.Microseconds())
}

func (m tuiModel) traceTick() tea.Cmd {
	return tea.Tick(m.traceDelay, func(time.Time) tea.Msg {
		return traceTickMsg{}
	})
}

func (m tuiModel) renderTraceProgress(width int) string {
	total := max(1, m.progressTotal)
	done := clamp(m.progressDone, 0, total)
	barWidth := clamp(width-34, 10, 60)
	filled := done * barWidth / total
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", barWidth-filled)
	label := m.progressLabel
	if label == "" {
		label = "Trace"
	}
	totalText := "?"
	if m.progressTotal > 0 {
		totalText = strconv.Itoa(m.progressTotal)
	}
	return helpHintStyle.Render(fmt.Sprintf("%s [%s] %d/%s", label, bar, done, totalText))
}

func startTraceSave(path string, puzzle string, events []TraceEvent) tea.Cmd {
	updates := make(chan traceSaveProgressMsg, 1)
	go func() {
		err := writeTraceFileWithProgress(path, puzzle, events, func(done int, total int) {
			sendTraceSaveProgress(updates, traceSaveProgressMsg{label: "Saving trace", path: path, done: done, total: total})
		})
		sendTraceSaveProgress(updates, traceSaveProgressMsg{label: "Saving trace", path: path, done: len(events), total: len(events), finish: true, err: err})
		close(updates)
	}()
	return waitTraceSaveProgress(updates)
}

func startTraceLoad(path string) tea.Cmd {
	updates := make(chan traceLoadProgressMsg, 1)
	go func() {
		puzzle, events, err := readTraceFileWithProgress(path, func(done int, total int) {
			sendTraceLoadProgress(updates, traceLoadProgressMsg{path: path, done: done, total: total})
		})
		sendTraceLoadProgress(updates, traceLoadProgressMsg{path: path, done: len(events), total: len(events), finish: true, puzzle: puzzle, events: events, err: err})
		close(updates)
	}()
	return waitTraceLoadProgress(updates)
}

func waitTraceSaveProgress(updates <-chan traceSaveProgressMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-updates
		if !ok {
			return traceSaveProgressMsg{finish: true}
		}
		msg.updates = updates
		return msg
	}
}

func waitTraceLoadProgress(updates <-chan traceLoadProgressMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-updates
		if !ok {
			return traceLoadProgressMsg{finish: true}
		}
		msg.updates = updates
		return msg
	}
}

func sendTraceSaveProgress(updates chan traceSaveProgressMsg, msg traceSaveProgressMsg) {
	select {
	case updates <- msg:
	default:
		select {
		case <-updates:
		default:
		}
		updates <- msg
	}
}

func sendTraceLoadProgress(updates chan traceLoadProgressMsg, msg traceLoadProgressMsg) {
	select {
	case updates <- msg:
	default:
		select {
		case <-updates:
		default:
		}
		updates <- msg
	}
}

func writeTraceFile(path string, puzzle string, events []TraceEvent) error {
	return writeTraceFileWithProgress(path, puzzle, events, nil)
}

func writeTraceFileWithProgress(path string, puzzle string, events []TraceEvent, progress func(done int, total int)) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)
	if err := encoder.Encode(traceFileRecord{Record: "trace", Puzzle: puzzle}); err != nil {
		return err
	}
	total := len(events)
	if progress != nil {
		progress(0, total)
	}
	progressEvery := max(1, total/100)
	for i, event := range events {
		event := event
		if err := encoder.Encode(traceFileRecord{Record: "event", Event: &event}); err != nil {
			return err
		}
		done := i + 1
		if progress != nil && (done == total || done%progressEvery == 0) {
			progress(done, total)
		}
	}
	return writer.Flush()
}

func readTraceFile(path string) (string, []TraceEvent, error) {
	return readTraceFileWithProgress(path, nil)
}

func readTraceFileWithProgress(path string, progress func(done int, total int)) (string, []TraceEvent, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	totalRecords := countTraceFileRecords(path)
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	recordCount := 0
	puzzle := ""
	events := make([]TraceEvent, 0)
	progressEvery := max(1, totalRecords/100)
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		recordCount++
		var record traceFileRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return "", nil, fmt.Errorf("line %d: %w", lineNumber, err)
		}
		switch record.Record {
		case "trace":
			if puzzle != "" {
				return "", nil, fmt.Errorf("line %d: duplicate trace header", lineNumber)
			}
			if _, err := parseDigits(record.Puzzle); err != nil {
				return "", nil, fmt.Errorf("line %d: invalid puzzle: %w", lineNumber, err)
			}
			if len(record.Puzzle) != PuzzleDimension*PuzzleDimension {
				return "", nil, fmt.Errorf("line %d: expected puzzle length %d, got %d", lineNumber, PuzzleDimension*PuzzleDimension, len(record.Puzzle))
			}
			puzzle = record.Puzzle
		case "event":
			if puzzle == "" {
				return "", nil, fmt.Errorf("line %d: event before trace header", lineNumber)
			}
			if record.Event == nil {
				return "", nil, fmt.Errorf("line %d: missing event", lineNumber)
			}
			if err := validateTraceEvent(*record.Event); err != nil {
				return "", nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			events = append(events, *record.Event)
		default:
			return "", nil, fmt.Errorf("line %d: unknown record type %q", lineNumber, record.Record)
		}
		if progress != nil && (recordCount == totalRecords || recordCount%progressEvery == 0) {
			progress(recordCount, totalRecords)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", nil, err
	}
	if puzzle == "" {
		return "", nil, fmt.Errorf("missing trace header")
	}
	return puzzle, events, nil
}

func countTraceFileRecords(path string) int {
	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}
	return count
}

func validateTraceEvent(event TraceEvent) error {
	switch event.Type {
	case TraceSolved:
		return nil
	case TracePlace, TraceBacktrack:
		if !inBounds(event.Row, event.Col, PuzzleDimension) {
			return fmt.Errorf("event cell (%d, %d) is out of bounds", event.Row, event.Col)
		}
		if event.Value < 1 || event.Value > PuzzleDimension {
			return fmt.Errorf("event value %d is invalid", event.Value)
		}
		return nil
	default:
		return fmt.Errorf("unknown trace event type %q", event.Type)
	}
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
		"/trace solve       Record recursive solve events for playback.",
		"/trace next        Apply the next trace event.",
		"/trace prev        Rewind one trace event.",
		"/trace play        Play trace events automatically.",
		"/trace pause       Pause trace playback.",
		"/trace reset       Return to the trace starting board.",
		"/trace status      Show trace playback progress.",
		"/trace delay us    Set automatic trace playback delay.",
		"/trace save path   Save the current trace to JSONL.",
		"/trace load path   Load a JSONL trace and its starting puzzle.",
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
	defaultTraceDelay = 1000 * time.Microsecond

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

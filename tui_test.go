package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestTUICommandCannotChangeOriginalClue(t *testing.T) {
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}

	model.runCommand("/set 0 0 2")

	value, ok := model.sudoku.Value(0, 0)
	if !ok {
		t.Fatal("Value(0, 0) failed")
	}
	if value != 1 {
		t.Fatalf("Value(0, 0) = %d, want original clue 1", value)
	}
	if !strings.Contains(strings.Join(model.logs, "\n"), "Cannot change original clue") {
		t.Fatal("expected log message for read-only clue")
	}
}

func TestTUICommandSetsEditableCell(t *testing.T) {
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}

	model.runCommand("/set 0 1 2")

	value, ok := model.sudoku.Value(0, 1)
	if !ok {
		t.Fatal("Value(0, 1) failed")
	}
	if value != 2 {
		t.Fatalf("Value(0, 1) = %d, want 2", value)
	}
}

func TestTUIRandomPuzzleReplacesOriginalAndMask(t *testing.T) {
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}

	model.runCommand("/random easy")

	if model.original == "1"+strings.Repeat("0", 80) {
		t.Fatal("expected random puzzle to replace original")
	}
	if len(model.solution) != PuzzleDimension*PuzzleDimension {
		t.Fatalf("len(solution) = %d", len(model.solution))
	}
	if len(model.checkpoint) != 0 {
		t.Fatal("expected random puzzle to clear checkpoints")
	}
	clues := 0
	for row := 0; row < PuzzleDimension; row++ {
		for col := 0; col < PuzzleDimension; col++ {
			if model.given[row][col] {
				clues++
			}
		}
	}
	if clues != 40 {
		t.Fatalf("given clues = %d, want 40", clues)
	}
}

func TestTUIAllowsBlankInitialPuzzleForRandomGeneration(t *testing.T) {
	model, err := newTUIModel("", "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel(blank) error = %v", err)
	}
	if got := model.sudoku.Representation(); got != strings.Repeat("0", 81) {
		t.Fatalf("initial blank puzzle = %q", got)
	}
}

func TestTUIInvalidChangeKeepsCurrentEditableValue(t *testing.T) {
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}
	model.runCommand("/set 0 1 2")

	model.runCommand("/set 0 1 1")

	value, ok := model.sudoku.Value(0, 1)
	if !ok {
		t.Fatal("Value(0, 1) failed")
	}
	if value != 2 {
		t.Fatalf("Value(0, 1) = %d, want existing value 2", value)
	}
}

func TestRenderSudokuBoardUsesDoubleNonetBordersAndSums(t *testing.T) {
	sudoku := NewSudoku()
	if err := sudoku.Load("123000000400000000500000000000000000000000000000000000000000000000000000000000000"); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	board := renderSudokuBoard(sudoku, [PuzzleDimension][PuzzleDimension]bool{}, 0, 0, nil)

	for _, expected := range []string{"╔", "╦", "╬", "╚", "║", "│", "   6", " 10"} {
		if !strings.Contains(board, expected) {
			t.Fatalf("rendered board does not contain %q:\n%s", expected, board)
		}
	}
	for _, unexpected := range []string{"Row Sum", "Column Sum"} {
		if strings.Contains(board, unexpected) {
			t.Fatalf("rendered board unexpectedly contains %q:\n%s", unexpected, board)
		}
	}
}

func TestCommandHelpIncludesSlashCommands(t *testing.T) {
	help := strings.Join(commandHelpLines(), "\n")

	for _, expected := range []string{"/set x y value", "/get x y", "/checkpoints", "/random difficulty", "/state save path", "/help advanced", "/quit"} {
		if !strings.Contains(help, expected) {
			t.Fatalf("help does not contain %q:\n%s", expected, help)
		}
	}
	for _, unexpected := range []string{"/trace solve", "/strategy", "/solve"} {
		if strings.Contains(help, unexpected) {
			t.Fatalf("basic help unexpectedly contains %q:\n%s", unexpected, help)
		}
	}
}

func TestAdvancedHelpIncludesTraceAndStrategyCommands(t *testing.T) {
	help := strings.Join(helpLines([]string{"advanced"}), "\n")

	for _, expected := range []string{"/trace solve", "/trace play", "/strategy", "/solve"} {
		if !strings.Contains(help, expected) {
			t.Fatalf("advanced help does not contain %q:\n%s", expected, help)
		}
	}
}

func TestTUILogRetainsHistoryAndScrolls(t *testing.T) {
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}
	model.height = 24

	for i := 0; i < 20; i++ {
		model.appendLog("line " + string(rune('a'+i)))
	}

	if len(model.logs) != 21 {
		t.Fatalf("len(logs) = %d, want 21", len(model.logs))
	}
	if !model.logFollow {
		t.Fatal("expected appended logs to auto-follow")
	}

	model.scrollLog(-model.logPageSize())
	if model.logFollow {
		t.Fatal("expected manual upward scroll to disable follow")
	}
	before := model.logOffset
	model.appendLog("newest")
	if model.logOffset != before {
		t.Fatalf("logOffset changed while follow disabled: got %d, want %d", model.logOffset, before)
	}

	model.scrollLogToEnd()
	if !model.logFollow {
		t.Fatal("expected scroll-to-end to restore follow")
	}
}

func TestTUICommandLogSeparatesCommandBlocks(t *testing.T) {
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}

	model.appendCommandLog("/help")

	if len(model.logs) < 3 {
		t.Fatalf("len(logs) = %d, want at least 3", len(model.logs))
	}
	if model.logs[len(model.logs)-2] != "" {
		t.Fatalf("log separator = %q, want blank line", model.logs[len(model.logs)-2])
	}
	if model.logs[len(model.logs)-1] != "/help" {
		t.Fatalf("last log line = %q, want /help", model.logs[len(model.logs)-1])
	}
}

func TestTUILayoutSwitchesWhenWideEnough(t *testing.T) {
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}

	wideThreshold := boardWidth + minSideLogWidth + layoutGutterWidth
	model.width = wideThreshold - 1
	if model.wideLayout() {
		t.Fatal("expected stacked layout below wide threshold")
	}

	model.width = wideThreshold
	if !model.wideLayout() {
		t.Fatal("expected side-by-side layout at wide threshold")
	}
}

func TestTUIWideLayoutGivesLogRemainingWidthAndFullCommandWidth(t *testing.T) {
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}
	model.width = 180

	if got, want := model.logWidth(), 180-boardWidth-layoutGutterWidth; got != want {
		t.Fatalf("logWidth() = %d, want %d", got, want)
	}
	if got, want := commandWidth(model.width), 176; got != want {
		t.Fatalf("commandWidth() = %d, want %d", got, want)
	}
}

func TestTUIViewLabelsPuzzlePanel(t *testing.T) {
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}

	if view := model.View().Content; !strings.Contains(view, "Puzzle") {
		t.Fatalf("view does not contain Puzzle label:\n%s", view)
	}
}

func TestTUIViewShowsStatusAndFilledCount(t *testing.T) {
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}

	view := model.View().Content
	for _, expected := range []string{"Status: Unsolved", "Filled: 1/81"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("view does not contain %q:\n%s", expected, view)
		}
	}
}

func TestTUITraceStepAndReset(t *testing.T) {
	model, err := newTUIModel("123456780456789123789123456214365897365897214897214365531642978642978531978531640", "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}

	model = runCommandAndDrain(t, model, "/trace solve")
	if len(model.trace) != 3 {
		t.Fatalf("len(trace) = %d, want 3", len(model.trace))
	}
	if model.traceIndex != 0 {
		t.Fatalf("traceIndex = %d, want 0", model.traceIndex)
	}

	model.runCommand("/trace next")
	if value, _ := model.sudoku.Value(0, 8); value != 9 {
		t.Fatalf("Value(0, 8) = %d, want 9", value)
	}

	model.runCommand("/trace next")
	if value, _ := model.sudoku.Value(8, 8); value != 2 {
		t.Fatalf("Value(8, 8) = %d, want 2", value)
	}

	model.runCommand("/trace prev")
	if value, _ := model.sudoku.Value(8, 8); value != 0 {
		t.Fatalf("Value(8, 8) after prev = %d, want 0", value)
	}
	if value, _ := model.sudoku.Value(0, 8); value != 9 {
		t.Fatalf("Value(0, 8) after prev = %d, want 9", value)
	}

	model.runCommand("/trace reset")
	if value, _ := model.sudoku.Value(0, 8); value != 0 {
		t.Fatalf("Value(0, 8) after reset = %d, want 0", value)
	}
}

func TestTUITraceSolveStartsFromOriginalPuzzle(t *testing.T) {
	model, err := newTUIModel("123456780456789123789123456214365897365897214897214365531642978642978531978531640", "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}
	model.sudoku.SetValue(0, 8, 9)

	model = runCommandAndDrain(t, model, "/trace solve")

	if model.traceBase != model.original {
		t.Fatalf("traceBase = %q, want original puzzle", model.traceBase)
	}
	if value, _ := model.sudoku.Value(0, 8); value != 0 {
		t.Fatalf("Value(0, 8) after /trace solve = %d, want original empty cell 0", value)
	}
}

func TestTUITracePlayAdvancesOnTick(t *testing.T) {
	model, err := newTUIModel("123456780456789123789123456214365897365897214897214365531642978642978531978531640", "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}
	model = runCommandAndDrain(t, model, "/trace solve")
	model.runCommand("/trace play")

	updated, _ := model.Update(traceTickMsg{})
	model = updated.(tuiModel)

	if value, _ := model.sudoku.Value(0, 8); value != 9 {
		t.Fatalf("Value(0, 8) = %d, want 9", value)
	}
	if !model.tracePlay {
		t.Fatal("expected trace playback to continue after first tick")
	}
}

func TestTUITraceDelayCommand(t *testing.T) {
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}

	if model.traceDelay != defaultTraceDelay {
		t.Fatalf("traceDelay = %v, want %v", model.traceDelay, defaultTraceDelay)
	}

	model.runCommand("/trace delay 750")

	if model.traceDelay != 750*time.Microsecond {
		t.Fatalf("traceDelay = %v, want 750us", model.traceDelay)
	}

	model = runCommandAndDrain(t, model, "/trace solve")
	if status := model.traceStatus(); !strings.Contains(status, "delay=750 us") {
		t.Fatalf("traceStatus() = %q, want delay mention", status)
	}
}

func TestTUITraceSaveLoadRestoresInitialPuzzle(t *testing.T) {
	puzzle := "123456780456789123789123456214365897365897214897214365531642978642978531978531640"
	model, err := newTUIModel(puzzle, "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}
	model = runCommandAndDrain(t, model, "/trace solve")
	model.runCommand("/trace next")

	path := filepath.Join(t.TempDir(), "trace.jsonl")
	_, cmd := model.runCommandWithCmd("/trace save " + path)
	model = drainTUICommand(t, model, cmd)
	if model.progressActive {
		t.Fatal("expected trace save to finish")
	}

	loaded, err := newTUIModel(strings.Repeat("0", 81), "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel(empty) error = %v", err)
	}
	_, loadCmd := loaded.runCommandWithCmd("/trace load " + path)
	loaded = drainTUICommand(t, loaded, loadCmd)

	if loaded.traceBase != puzzle {
		t.Fatalf("traceBase = %q, want saved puzzle", loaded.traceBase)
	}
	if loaded.progressActive {
		t.Fatal("expected trace load progress to finish")
	}
	if got := loaded.sudoku.Representation(); got != puzzle {
		t.Fatalf("loaded board = %q, want initial puzzle", got)
	}
	if len(loaded.trace) != len(model.trace) {
		t.Fatalf("len(trace) = %d, want %d", len(loaded.trace), len(model.trace))
	}

	loaded.runCommand("/trace next")
	if value, _ := loaded.sudoku.Value(0, 8); value != 9 {
		t.Fatalf("Value(0, 8) after loaded trace next = %d, want 9", value)
	}
}

func TestTUIStrategyCommand(t *testing.T) {
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}

	if model.strategy != "row-major" {
		t.Fatalf("initial strategy = %q, want row-major", model.strategy)
	}

	model.runCommand("/strategy nonet-first")
	if model.strategy != "nonet-first" {
		t.Fatalf("strategy after set = %q, want nonet-first", model.strategy)
	}
	if !strings.Contains(strings.Join(model.logs, "\n"), "nonet-first") {
		t.Fatal("expected log to confirm strategy change")
	}

	model.runCommand("/strategy row-major")
	if model.strategy != "row-major" {
		t.Fatalf("strategy after reset = %q, want row-major", model.strategy)
	}

	model.runCommand("/strategy unknown")
	if model.strategy != "row-major" {
		t.Fatal("unknown strategy should not change current strategy")
	}
}

func TestTUIStateSaveLoadRestoresProgress(t *testing.T) {
	original := "100000000020000000003000000000400000000050000000006000000000700000000080000000009"
	model, err := newTUIModel(original, strings.Repeat("1", 81), "nonet-first")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}
	model.runCommand("/set 0 1 4")
	model.runCommand("/save progress")

	path := filepath.Join(t.TempDir(), "state.json")
	model.runCommand("/state save " + path)

	loaded, err := newTUIModel(strings.Repeat("0", 81), "", "row-major")
	if err != nil {
		t.Fatalf("newTUIModel(empty) error = %v", err)
	}
	loaded.runCommand("/state load " + path)

	if loaded.original != original {
		t.Fatalf("original = %q, want %q", loaded.original, original)
	}
	if loaded.solution != strings.Repeat("1", 81) {
		t.Fatal("solution was not restored")
	}
	if loaded.strategy != "nonet-first" {
		t.Fatalf("strategy = %q, want nonet-first", loaded.strategy)
	}
	if loaded.checkpoint["progress"] == "" {
		t.Fatal("checkpoint was not restored")
	}
	if value, _ := loaded.sudoku.Value(0, 1); value != 4 {
		t.Fatalf("loaded Value(0, 1) = %d, want 4", value)
	}
	if !loaded.given[0][0] || loaded.given[0][1] {
		t.Fatal("given mask was not restored from original puzzle")
	}
}

func TestTUIStrategyAffectsSolve(t *testing.T) {
	puzzle := "300401620100080400005020830057800000000700503002904007480530010203090000070006090"
	expected := "398471625126385479745629831657813942914762583832954167489537216263198754571246398"

	for _, strategy := range []string{"row-major", "nonet-first"} {
		t.Run(strategy, func(t *testing.T) {
			model, err := newTUIModel(puzzle, expected, strategy)
			if err != nil {
				t.Fatalf("newTUIModel() error = %v", err)
			}
			model = runCommandAndDrain(t, model, "/solve")
			if !model.solved {
				t.Fatal("expected puzzle to be solved")
			}
			if got := model.sudoku.Representation(); got != expected {
				t.Fatalf("solution = %q, want %q", got, expected)
			}
		})
	}
}

func runCommandAndDrain(t *testing.T, model tuiModel, command string) tuiModel {
	t.Helper()
	_, cmd := model.runCommandWithCmd(command)
	return drainTUICommand(t, model, cmd)
}

func drainTUICommand(t *testing.T, model tuiModel, cmd tea.Cmd) tuiModel {
	t.Helper()
	for cmd != nil {
		msg := cmd()
		updated, next := model.Update(msg)
		var ok bool
		model, ok = updated.(tuiModel)
		if !ok {
			t.Fatalf("Update returned %T, want tuiModel", updated)
		}
		cmd = next
	}
	return model
}

func TestReadTraceFileRejectsEventBeforeInitialPuzzle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.jsonl")
	if err := os.WriteFile(path, []byte(`{"record":"event","event":{"type":"solved"}}`+"\n"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, _, err := readTraceFile(path); err == nil {
		t.Fatal("expected readTraceFile to reject event before trace header")
	}
}

func TestWriteTraceFileHandlesLargeTrace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large-trace.jsonl")
	events := make([]TraceEvent, 5000)
	for i := range events {
		events[i] = TraceEvent{Type: TracePlace, Row: i % PuzzleDimension, Col: (i / PuzzleDimension) % PuzzleDimension, Value: i%PuzzleDimension + 1}
	}

	if err := writeTraceFile(path, strings.Repeat("0", 81), events); err != nil {
		t.Fatalf("writeTraceFile() error = %v", err)
	}

	_, loaded, err := readTraceFile(path)
	if err != nil {
		t.Fatalf("readTraceFile() error = %v", err)
	}
	if len(loaded) != len(events) {
		t.Fatalf("len(loaded) = %d, want %d", len(loaded), len(events))
	}
}

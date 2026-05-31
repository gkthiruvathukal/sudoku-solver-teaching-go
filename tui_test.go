package main

import (
	"strings"
	"testing"
)

func TestTUICommandCannotChangeOriginalClue(t *testing.T) {
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "")
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
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "")
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

func TestTUIInvalidChangeKeepsCurrentEditableValue(t *testing.T) {
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "")
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

	for _, expected := range []string{"/set x y value", "/get x y", "/checkpoints", "/quit"} {
		if !strings.Contains(help, expected) {
			t.Fatalf("help does not contain %q:\n%s", expected, help)
		}
	}
}

func TestTUILogRetainsHistoryAndScrolls(t *testing.T) {
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "")
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
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "")
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
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "")
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
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "")
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
	model, err := newTUIModel("1"+strings.Repeat("0", 80), "")
	if err != nil {
		t.Fatalf("newTUIModel() error = %v", err)
	}

	if view := model.View().Content; !strings.Contains(view, "Puzzle") {
		t.Fatalf("view does not contain Puzzle label:\n%s", view)
	}
}

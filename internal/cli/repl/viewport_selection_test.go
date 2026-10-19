package repl

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestViewportSelection_SelectedTextAcrossLines(t *testing.T) {
	var selection viewportSelection
	selection.setContent("alpha beta\ngamma delta\nthird")
	selection.start(6, 0, 0)
	selection.drag(5, 1, 0)

	if got := selection.selectedText(); got != "beta\ngamma" {
		t.Fatalf("expected selected text, got %q", got)
	}
}

func TestViewportSelection_SelectWord(t *testing.T) {
	var selection viewportSelection
	selection.setContent("hello world")

	if !selection.selectWord(7, 0, 0) {
		t.Fatal("expected word selection")
	}
	if got := selection.selectedText(); got != "world" {
		t.Fatalf("expected selected word, got %q", got)
	}
}

func TestViewportSelection_RenderHighlightsVisibleSelection(t *testing.T) {
	var selection viewportSelection
	selection.setContent("hello world")
	selection.start(0, 0, 0)
	selection.drag(5, 0, 0)

	rendered := selection.render("hello world", 20, 1, 0)
	if rendered == "hello world" {
		t.Fatal("expected rendered selection to add styling")
	}
	if !strings.Contains(rendered, "hello") {
		t.Fatalf("expected highlighted view to preserve content, got %q", rendered)
	}
}

func TestReplSelectionMouseDragCopiesOnRelease(t *testing.T) {
	defer setClipboardWriter(func(string) error { return nil })()

	m := newTestModel()
	m.output.AddLine("hello world")
	m.updateViewportContent()
	m.blurInput()

	updated, cmd := m.updateNormalMode(tea.MouseClickMsg(tea.Mouse{X: 0, Y: 0, Button: tea.MouseLeft}))
	if cmd != nil {
		t.Fatal("expected no command on mouse down")
	}
	updated, cmd = updated.updateNormalMode(tea.MouseMotionMsg(tea.Mouse{X: 5, Y: 0, Button: tea.MouseLeft}))
	if cmd != nil {
		t.Fatal("expected no command while dragging")
	}
	updated, cmd = updated.updateNormalMode(tea.MouseReleaseMsg(tea.Mouse{X: 5, Y: 0, Button: tea.MouseLeft}))
	if cmd == nil {
		t.Fatal("expected copy command on mouse release")
	}
	copyMsg := cmd().(clipboardCopyMsg)
	updated, _ = updated.updateNormalMode(copyMsg)
	if updated.notification.text != copyNotificationMessage {
		t.Fatalf("expected copy notification, got %q", updated.notification.text)
	}
	if got := updated.selection.selectedText(); got != "hello" {
		t.Fatalf("expected selected text to remain available, got %q", got)
	}
}

func TestReplSelectionMouseDragReportsFailureOnNativeCopyError(t *testing.T) {
	defer setClipboardWriter(func(string) error { return errors.New("no clipboard") })()

	m := newTestModel()
	m.output.AddLine("hello world")
	m.updateViewportContent()
	m.blurInput()

	m.selection.start(0, 0, 0)
	m.selection.drag(5, 0, 0)

	updated, cmd := m.updateNormalMode(tea.MouseReleaseMsg(tea.Mouse{X: 5, Y: 0, Button: tea.MouseLeft}))
	if cmd == nil {
		t.Fatal("expected copy command on mouse release")
	}
	copyMsg := cmd().(clipboardCopyMsg)
	updated, followUp := updated.updateNormalMode(copyMsg)
	if updated.notification.text != copyFailedNotificationMessage {
		t.Fatalf("expected copy-failed notification, got %q", updated.notification.text)
	}
	batch, ok := followUp().(tea.BatchMsg)
	if !ok || len(batch) != 2 {
		t.Fatalf("expected fallback batch with notification and OSC 52, got %T", followUp)
	}
}

func TestReplSelectionMouseClickFocusesViewport(t *testing.T) {
	m := newTestModel()
	m.output.AddLine("hello world")
	m.updateViewportContent()

	updated, cmd := m.updateNormalMode(tea.MouseClickMsg(tea.Mouse{X: 0, Y: 0, Button: tea.MouseLeft}))
	if cmd != nil {
		t.Fatal("expected no command on viewport mouse down")
	}
	if updated.textarea.Focused() {
		t.Fatal("expected viewport click to blur input")
	}
}

func TestReplSelectionCtrlCCopiesSelection(t *testing.T) {
	defer setClipboardWriter(func(string) error { return nil })()

	m := newTestModel()
	m.output.AddLine("hello world")
	m.updateViewportContent()
	m.selection.start(0, 0, 0)
	m.selection.drag(5, 0, 0)

	updated, cmd := m.updateNormalMode(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("expected ctrl+c to copy the active selection")
	}
	if updated.quitting {
		t.Fatal("expected ctrl+c with an active selection not to quit")
	}
	if got := updated.selection.selectedText(); got != "hello" {
		t.Fatalf("expected selection to remain available, got %q", got)
	}
	copyMsg, ok := cmd().(clipboardCopyMsg)
	if !ok {
		t.Fatalf("expected clipboard copy message, got %T", cmd())
	}
	updated, _ = updated.updateNormalMode(copyMsg)
	if updated.notification.text != copyNotificationMessage {
		t.Fatalf("expected copy notification, got %q", updated.notification.text)
	}
}

func TestReplSelectionCmdCCopiesSelection(t *testing.T) {
	defer setClipboardWriter(func(string) error { return nil })()

	m := newTestModel()
	m.output.AddLine("hello world")
	m.updateViewportContent()
	m.selection.start(0, 0, 0)
	m.selection.drag(5, 0, 0)

	updated, cmd := m.updateNormalMode(tea.KeyPressMsg{Code: 'c', Mod: tea.ModSuper})
	if cmd == nil {
		t.Fatal("expected cmd+c to copy the active selection")
	}
	if updated.quitting {
		t.Fatal("expected cmd+c not to quit")
	}
	if got := updated.selection.selectedText(); got != "hello" {
		t.Fatalf("expected selection to remain available, got %q", got)
	}
}

func TestReplSelectionCtrlCWithoutSelectionQuits(t *testing.T) {
	m := newTestModel()
	m.output.AddLine("hello world")
	m.updateViewportContent()

	updated, cmd := m.updateNormalMode(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("expected ctrl+c without a selection to keep normal quit behavior")
	}
	if !updated.quitting {
		t.Fatal("expected ctrl+c without a selection to quit")
	}
}

func TestReplInputSelectionMouseDragCopiesOnRelease(t *testing.T) {
	defer setClipboardWriter(func(string) error { return nil })()

	m := newTestModel()
	m.textarea.SetValue("hello world")
	m.blurInput()
	textY := m.inputAreaTop() + 1

	updated, cmd := m.updateNormalMode(tea.MouseClickMsg(tea.Mouse{X: inputPromptWidth, Y: textY, Button: tea.MouseLeft}))
	if cmd == nil {
		t.Fatal("expected focus command on input mouse down")
	}
	updated, cmd = updated.updateNormalMode(tea.MouseMotionMsg(tea.Mouse{X: inputPromptWidth + 5, Y: textY, Button: tea.MouseLeft}))
	if cmd != nil {
		t.Fatal("expected no command while dragging input selection")
	}
	updated, cmd = updated.updateNormalMode(tea.MouseReleaseMsg(tea.Mouse{X: inputPromptWidth + 5, Y: textY, Button: tea.MouseLeft}))
	if cmd == nil {
		t.Fatal("expected copy command on input mouse release")
	}
	copyMsg := cmd().(clipboardCopyMsg)
	updated, _ = updated.updateNormalMode(copyMsg)
	if updated.notification.text != copyNotificationMessage {
		t.Fatalf("expected copy notification, got %q", updated.notification.text)
	}
	if got := updated.inputSelection.selectedText(); got != "hello" {
		t.Fatalf("expected input selection to remain, got %q", got)
	}
	if strings.Contains(updated.selection.selectedText(), "hello") {
		t.Fatalf("expected viewport selection to stay clear, got %q", updated.selection.selectedText())
	}
	if !updated.textarea.Focused() {
		t.Fatal("expected input click to focus textarea")
	}
}

func TestReplInputSelectionMouseDragReportsFailureOnNativeCopyError(t *testing.T) {
	defer setClipboardWriter(func(string) error { return errors.New("no clipboard") })()

	m := newTestModel()
	m.textarea.SetValue("hello world")
	m.inputSelection.setContent(m.textarea.Value())
	m.inputSelection.start(0, 0, 0)
	m.inputSelection.drag(5, 0, 0)

	updated, cmd := m.updateNormalMode(tea.MouseReleaseMsg(tea.Mouse{X: inputPromptWidth + 5, Y: m.inputAreaTop() + 1, Button: tea.MouseLeft}))
	if cmd == nil {
		t.Fatal("expected copy command on input mouse release")
	}
	copyMsg := cmd().(clipboardCopyMsg)
	updated, followUp := updated.updateNormalMode(copyMsg)
	if updated.notification.text != copyFailedNotificationMessage {
		t.Fatalf("expected copy-failed notification, got %q", updated.notification.text)
	}
	batch, ok := followUp().(tea.BatchMsg)
	if !ok || len(batch) != 2 {
		t.Fatalf("expected fallback batch with notification and OSC 52, got %T", followUp)
	}
}

func TestReplInputSelectionCtrlCCopiesSelection(t *testing.T) {
	defer setClipboardWriter(func(string) error { return nil })()

	m := newTestModel()
	m.textarea.SetValue("hello world")
	m.inputSelection.setContent(m.textarea.Value())
	m.inputSelection.start(0, 0, 0)
	m.inputSelection.drag(5, 0, 0)

	updated, cmd := m.updateNormalMode(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("expected ctrl+c to copy the active input selection")
	}
	if updated.quitting {
		t.Fatal("expected ctrl+c with an active input selection not to quit")
	}
	if got := updated.textarea.Value(); got != "hello world" {
		t.Fatalf("expected input to be preserved, got %q", got)
	}
	if got := updated.inputSelection.selectedText(); got != "hello" {
		t.Fatalf("expected input selection to remain available, got %q", got)
	}
}

func TestView_RendersInputSelection(t *testing.T) {
	m := newTestModel()
	m.textarea.SetValue("hello world")
	m.inputSelection.setContent(m.textarea.Value())
	m.inputSelection.start(0, 0, 0)
	m.inputSelection.drag(5, 0, 0)

	view := m.View().Content
	plainView := ansi.Strip(view)
	if !strings.Contains(plainView, "hello") {
		t.Fatalf("expected input content in view, got %q", view)
	}
	withoutSelection := newTestModel()
	withoutSelection.textarea.SetValue("hello world")
	if view == withoutSelection.View().Content {
		t.Fatal("expected input selection to affect rendered view")
	}
}

func TestView_ShowsCopyNotification(t *testing.T) {
	m := newTestModel()
	m.notification.text = copyNotificationMessage

	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, copyNotificationMessage) {
		t.Fatalf("expected copy notification in view, got %q", view)
	}
}

func TestReplCopyNotificationExpires(t *testing.T) {
	m := newTestModel()
	expiresAt := time.Now().Add(copyNotificationTimeout)
	m.notification.text = copyNotificationMessage
	m.notification.expiresAt = expiresAt

	updated, cmd := m.updateNormalMode(copyNotificationExpiredMsg{expiresAt: expiresAt.UnixNano()})
	if cmd != nil {
		t.Fatal("expected no command for copy notification expiry")
	}
	if updated.notification.text != "" {
		t.Fatalf("expected copy notification to clear, got %q", updated.notification.text)
	}
}

func TestReplCopyNotificationIgnoresStaleExpiry(t *testing.T) {
	m := newTestModel()
	expiresAt := time.Now().Add(copyNotificationTimeout)
	m.notification.text = copyNotificationMessage
	m.notification.expiresAt = expiresAt

	updated, _ := m.updateNormalMode(copyNotificationExpiredMsg{expiresAt: expiresAt.Add(-time.Second).UnixNano()})
	if updated.notification.text != copyNotificationMessage {
		t.Fatalf("expected copy notification to remain, got %q", updated.notification.text)
	}
}

func TestView_KeepsAltScreenMouseCaptureForSelection(t *testing.T) {
	m := newTestModel()

	view := m.View()

	if !view.AltScreen {
		t.Fatal("expected alt-screen to remain enabled")
	}
	if view.MouseMode != tea.MouseModeCellMotion {
		t.Fatalf("expected cell motion mouse capture, got %v", view.MouseMode)
	}
}

func TestViewportSelection_SelectLine(t *testing.T) {
	var selection viewportSelection
	selection.setContent("first\nsecond line")
	if !selection.selectLine(1, 0) {
		t.Fatal("selectLine did not select a line")
	}
	if got := selection.selectedText(); got != "second line" {
		t.Fatalf("selected line = %q, want second line", got)
	}
}

func TestViewportSelection_RejectsWhitespaceAndEmptyLines(t *testing.T) {
	var selection viewportSelection
	selection.setContent("hello world\n")
	if selection.selectWord(5, 0, 0) {
		t.Fatal("selectWord selected whitespace")
	}
	if selection.selectLine(1, 0) {
		t.Fatal("selectLine selected an empty line")
	}
	if selection.drag(1, 0, 0) || selection.release() {
		t.Fatal("inactive selection accepted drag or release")
	}
}

func TestSelectionHelpersHandleBoundaries(t *testing.T) {
	if start, end := findSelectionWordBoundaries("", 0); start != 0 || end != 0 {
		t.Fatalf("empty boundaries = (%d, %d)", start, end)
	}
	if start, end := findSelectionWordBoundaries("hello", -1); start != 0 || end != 0 {
		t.Fatalf("negative boundaries = (%d, %d)", start, end)
	}
	if start, end := findSelectionWordBoundaries("hello", 9); start != 9 || end != 9 {
		t.Fatalf("past-end boundaries = (%d, %d)", start, end)
	}
	if got := absInt(-3); got != 3 {
		t.Fatalf("absInt(-3) = %d", got)
	}
	if got := clampInt(5, 10, 0); got != 5 {
		t.Fatalf("clampInt with reversed bounds = %d", got)
	}
}

func TestRenderSelectionLeavesInvalidOrHiddenSelectionsUnchanged(t *testing.T) {
	view := "hello"
	for _, got := range []string{
		renderSelection(view, 0, 1, 0, 0, selectionPoint{}, selectionPoint{col: 2}),
		renderSelection(view, 10, 0, 0, 0, selectionPoint{}, selectionPoint{col: 2}),
		renderSelection(view, 10, 1, 2, 0, selectionPoint{}, selectionPoint{col: 2}),
	} {
		if got != view {
			t.Fatalf("renderSelection changed hidden or invalid view: %q", got)
		}
	}
}

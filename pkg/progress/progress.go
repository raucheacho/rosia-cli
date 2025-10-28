// Package progress provides progress bar functionality for long-running operations.
//
// The progress package wraps Bubble Tea's progress component to provide
// easy-to-use progress indicators for scanning and cleaning operations.
// It supports both CLI and TUI contexts.
//
// Example usage:
//
//	bar := progress.NewBar(100, "Scanning")
//	bar.Start()
//	for i := 0; i < 100; i++ {
//	    bar.Increment()
//	    time.Sleep(10 * time.Millisecond)
//	}
//	bar.Finish()
package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

// Bar represents a progress bar for CLI operations.
//
// The Bar displays progress for long-running operations with a label
// and percentage indicator.
type Bar struct {
	model    progress.Model
	current  int
	total    int
	label    string
	mu       sync.Mutex
	program  *tea.Program
	finished bool
}

// NewBar creates a new progress bar
func NewBar(total int, label string) *Bar {
	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	return &Bar{
		model:   prog,
		current: 0,
		total:   total,
		label:   label,
	}
}

// Start initializes and displays the progress bar
func (b *Bar) Start() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.program != nil {
		return
	}

	// Create a simple model for the progress bar
	m := &progressModel{
		bar:   b,
		label: b.label,
	}

	b.program = tea.NewProgram(m)
	go b.program.Run()
}

// Increment increases the progress by one
func (b *Bar) Increment() {
	b.IncrementBy(1)
}

// IncrementBy increases the progress by the specified amount
func (b *Bar) IncrementBy(n int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.current += n
	if b.current > b.total {
		b.current = b.total
	}

	if b.program != nil {
		b.program.Send(progressMsg{
			current: b.current,
			total:   b.total,
		})
	}
}

// SetLabel updates the progress bar label
func (b *Bar) SetLabel(label string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.label = label
	if b.program != nil {
		b.program.Send(labelMsg(label))
	}
}

// Finish completes the progress bar and cleans up
func (b *Bar) Finish() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.finished {
		return
	}

	b.finished = true
	b.current = b.total

	if b.program != nil {
		b.program.Send(tea.Quit())
		time.Sleep(50 * time.Millisecond) // Give time for final render
		b.program.Quit()
	}
}

// progressMsg is sent to update progress
type progressMsg struct {
	current int
	total   int
}

// labelMsg is sent to update the label
type labelMsg string

// progressModel is the BubbleTea model for the progress bar
type progressModel struct {
	bar      *Bar
	label    string
	progress progress.Model
	quitting bool
}

func (m *progressModel) Init() tea.Cmd {
	return nil
}

func (m *progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

	case progressMsg:
		m.bar.mu.Lock()
		m.bar.current = msg.current
		m.bar.total = msg.total
		m.bar.mu.Unlock()

		if msg.current >= msg.total {
			m.quitting = true
			return m, tea.Quit
		}

	case labelMsg:
		m.label = string(msg)

	case tea.QuitMsg:
		m.quitting = true
		return m, tea.Quit

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

func (m *progressModel) View() string {
	if m.quitting {
		return ""
	}

	m.bar.mu.Lock()
	current := m.bar.current
	total := m.bar.total
	label := m.bar.label
	m.bar.mu.Unlock()

	if total == 0 {
		return ""
	}

	percent := float64(current) / float64(total)
	if percent > 1.0 {
		percent = 1.0
	}

	return fmt.Sprintf("\n%s\n%s %d/%d\n",
		label,
		m.progress.ViewAs(percent),
		current,
		total,
	)
}

// SimpleBar is a lightweight progress bar without BubbleTea
type SimpleBar struct {
	total   int
	current int
	width   int
	label   string
	writer  io.Writer
	mu      sync.Mutex
}

// NewSimpleBar creates a simple progress bar that writes to the given writer
func NewSimpleBar(total int, label string, writer io.Writer) *SimpleBar {
	if writer == nil {
		writer = os.Stdout
	}

	return &SimpleBar{
		total:   total,
		current: 0,
		width:   40,
		label:   label,
		writer:  writer,
	}
}

// Increment increases the progress by one and renders
func (s *SimpleBar) Increment() {
	s.IncrementBy(1)
}

// IncrementBy increases the progress by the specified amount and renders
func (s *SimpleBar) IncrementBy(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.current += n
	if s.current > s.total {
		s.current = s.total
	}

	s.render()
}

// SetLabel updates the label
func (s *SimpleBar) SetLabel(label string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.label = label
}

// Finish completes the progress bar
func (s *SimpleBar) Finish() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.current = s.total
	s.render()
	fmt.Fprintln(s.writer) // Add newline after completion
}

// render draws the progress bar
func (s *SimpleBar) render() {
	if s.total == 0 {
		return
	}

	percent := float64(s.current) / float64(s.total)
	if percent > 1.0 {
		percent = 1.0
	}

	filled := int(float64(s.width) * percent)
	empty := s.width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	// Clear line and render
	fmt.Fprintf(s.writer, "\r%s [%s] %d/%d (%.0f%%)",
		s.label,
		bar,
		s.current,
		s.total,
		percent*100,
	)
}

package progress

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSimpleBar(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewSimpleBar(100, "Testing", buf)

	assert.NotNil(t, bar)
	assert.Equal(t, 100, bar.total)
	assert.Equal(t, 0, bar.current)
	assert.Equal(t, "Testing", bar.label)
	assert.Equal(t, 40, bar.width)
}

func TestSimpleBar_Increment(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewSimpleBar(10, "Progress", buf)

	bar.Increment()
	assert.Equal(t, 1, bar.current)

	bar.IncrementBy(5)
	assert.Equal(t, 6, bar.current)

	// Test overflow protection
	bar.IncrementBy(10)
	assert.Equal(t, 10, bar.current)
}

func TestSimpleBar_Render(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewSimpleBar(10, "Test", buf)

	bar.Increment()
	output := buf.String()

	// Should contain label and progress indicators
	assert.Contains(t, output, "Test")
	assert.Contains(t, output, "1/10")
	assert.Contains(t, output, "10%")
}

func TestSimpleBar_Finish(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewSimpleBar(5, "Complete", buf)

	bar.IncrementBy(3)
	bar.Finish()

	assert.Equal(t, 5, bar.current)
	output := buf.String()
	assert.Contains(t, output, "5/5")
	assert.Contains(t, output, "100%")
}

func TestSimpleBar_SetLabel(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewSimpleBar(10, "Initial", buf)

	bar.SetLabel("Updated")
	assert.Equal(t, "Updated", bar.label)
}

func TestSimpleBar_ZeroTotal(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewSimpleBar(0, "Empty", buf)

	// Should not panic
	bar.Increment()
	bar.Finish()

	// Only newline expected for zero total (from Finish)
	output := buf.String()
	assert.True(t, output == "" || output == "\n")
}

func TestSimpleBar_Concurrent(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewSimpleBar(100, "Concurrent", buf)

	// Simulate concurrent updates
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				bar.Increment()
				time.Sleep(time.Millisecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	assert.Equal(t, 100, bar.current)
}

func TestNewBar(t *testing.T) {
	bar := NewBar(50, "BubbleTea Progress")

	assert.NotNil(t, bar)
	assert.Equal(t, 50, bar.total)
	assert.Equal(t, 0, bar.current)
	assert.Equal(t, "BubbleTea Progress", bar.label)
	assert.False(t, bar.finished)
}

func TestBar_IncrementAndFinish(t *testing.T) {
	bar := NewBar(10, "Test")

	bar.Increment()
	assert.Equal(t, 1, bar.current)

	bar.IncrementBy(5)
	assert.Equal(t, 6, bar.current)

	bar.Finish()
	assert.True(t, bar.finished)
	assert.Equal(t, 10, bar.current)
}

func TestBar_SetLabel(t *testing.T) {
	bar := NewBar(10, "Initial")

	bar.SetLabel("Updated Label")
	assert.Equal(t, "Updated Label", bar.label)
}

func TestProgressModel_View(t *testing.T) {
	bar := NewBar(10, "Test Progress")
	model := &progressModel{
		bar:      bar,
		label:    "Test Progress",
		progress: bar.model,
	}

	bar.current = 5
	bar.total = 10

	view := model.View()
	assert.Contains(t, view, "Test Progress")
	assert.Contains(t, view, "5/10")
}

func TestProgressModel_ViewQuitting(t *testing.T) {
	bar := NewBar(10, "Test")
	model := &progressModel{
		bar:      bar,
		label:    "Test",
		progress: bar.model,
		quitting: true,
	}

	view := model.View()
	assert.Empty(t, view)
}

func TestProgressModel_ViewZeroTotal(t *testing.T) {
	bar := NewBar(0, "Empty")
	model := &progressModel{
		bar:      bar,
		label:    "Empty",
		progress: bar.model,
	}

	view := model.View()
	assert.Empty(t, view)
}

func TestSimpleBar_ProgressIndicators(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := NewSimpleBar(4, "Visual", buf)

	// Test at different progress levels
	bar.IncrementBy(1) // 25%
	output := buf.String()
	assert.True(t, strings.Contains(output, "█") || strings.Contains(output, "░"))

	buf.Reset()
	bar.IncrementBy(1) // 50%
	output = buf.String()
	assert.Contains(t, output, "2/4")

	buf.Reset()
	bar.Finish() // 100%
	output = buf.String()
	assert.Contains(t, output, "4/4")
	assert.Contains(t, output, "100%")
}

package printer

import (
	"time"

	"github.com/briandowns/spinner"
)

// Spinner wraps briandowns/spinner for consistent styling.
type Spinner struct {
	s *spinner.Spinner
}

// NewSpinner creates a new styled spinner with the given message.
func NewSpinner(msg string) *Spinner {
	s := spinner.New(spinner.CharSets[14], 80*time.Millisecond)
	s.Suffix = "  " + msg
	s.Color("cyan", "bold") //nolint:errcheck
	return &Spinner{s: s}
}

// Start begins spinning.
func (sp *Spinner) Start() {
	sp.s.Start()
}

// UpdateMsg updates the spinner message while running.
func (sp *Spinner) UpdateMsg(msg string) {
	sp.s.Suffix = "  " + msg
}

// Done stops the spinner and prints a success line.
func (sp *Spinner) Done(msg string) {
	sp.s.Stop()
	Success(msg)
}

// Fail stops the spinner and prints an error line.
func (sp *Spinner) Fail(msg string) {
	sp.s.Stop()
	Warn(msg)
}

// Stop stops the spinner silently.
func (sp *Spinner) Stop() {
	sp.s.Stop()
}

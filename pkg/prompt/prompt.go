// Package prompt provides interactive step-mode support for exercises.
// When step mode is enabled, the exercise pauses between major steps
// and waits for the user to press Enter before continuing.
package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var stepMode bool

// EnableStepMode activates interactive step-by-step mode.
// The exercise will pause between steps waiting for user input.
func EnableStepMode() {
	stepMode = true
	fmt.Println("⏸  Step mode enabled — press Enter to advance, 'q' to quit")
	fmt.Println()
}

// IsStepMode returns whether step mode is currently active.
func IsStepMode() bool {
	return stepMode
}

// StepPause pauses execution if step mode is enabled.
// If the user types 'q', the program exits cleanly.
// In non-step mode, this is a no-op.
func StepPause() {
	if !stepMode {
		return
	}
	fmt.Print("  [Enter] next step  [q] quit → ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "q" || input == "quit" {
		fmt.Println("  Exiting exercise.")
		os.Exit(0)
	}
	fmt.Println()
}

// WaitForEnter waits for the user to press Enter, with an optional message.
// Works regardless of step mode (useful for pre-exercise setup confirmation).
func WaitForEnter(msg string) {
	if msg == "" {
		msg = "Press Enter to continue..."
	}
	fmt.Printf("  %s ", msg)
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
	fmt.Println()
}

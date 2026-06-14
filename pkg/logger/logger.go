// Package logger provides structured console output for exercise runners.
// Each exercise creates its own logger instance, which renders sections,
// concept explanations, steps, and results with ANSI formatting.
package logger

import (
	"fmt"
	"strings"
)

// ANSI escape codes for terminal formatting.
const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	dim     = "\033[2m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
)

// Logger provides structured output for an exercise module.
type Logger struct {
	module string
}

// New creates a Logger for the given module name (e.g. "01-first-pods").
func New(module string) *Logger {
	return &Logger{module: module}
}

// Section prints a major section divider with title.
// Used to mark the start of each step or the module header.
func (l *Logger) Section(title string) {
	fmt.Println()
	bar := strings.Repeat("─", 68)
	fmt.Printf("%s%s %s %s%s\n", bold, cyan, bar, title, reset)
}

// Concept prints educational prose explaining a concept.
// Use this before showing commands or manifests to give context.
func (l *Logger) Concept(text string) {
	fmt.Printf("%s%s%s\n", dim, text, reset)
}

// Step announces an action the exercise is about to take.
func (l *Logger) Step(msg string, args ...interface{}) {
	fmt.Printf("%s▶ %s%s\n", blue, fmt.Sprintf(msg, args...), reset)
}

// Info prints an informational message.
func (l *Logger) Info(msg string, args ...interface{}) {
	fmt.Printf("  %s\n", fmt.Sprintf(msg, args...))
}

// Success prints a success result (green checkmark).
func (l *Logger) Success(msg string, args ...interface{}) {
	fmt.Printf("%s✔ %s%s\n", green, fmt.Sprintf(msg, args...), reset)
}

// Warn prints a warning (non-fatal issue to be aware of).
func (l *Logger) Warn(msg string, args ...interface{}) {
	fmt.Printf("%s⚠ %s%s\n", yellow, fmt.Sprintf(msg, args...), reset)
}

// Error prints an error message in red. Exits with code 1.
func (l *Logger) Error(msg string, args ...interface{}) {
	fmt.Printf("%s✖ %s%s\n", red, fmt.Sprintf(msg, args...), reset)
}

// Command displays a shell command the user should see (dimmed, with $ prefix).
func (l *Logger) Command(cmd string) {
	fmt.Printf("%s  $ %s%s\n", dim, cmd, reset)
}

// Output displays raw output from a command (dimmed, indented).
func (l *Logger) Output(text string) {
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		fmt.Printf("%s  │ %s%s\n", dim, line, reset)
	}
}

// KeyValue prints a key-value pair for displaying resource details.
func (l *Logger) KeyValue(key, value string) {
	fmt.Printf("  %s%-24s%s %s\n", bold, key+":", reset, value)
}

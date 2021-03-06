// Package tlog implements logging utilities for boilr
package tlog

import (
	"fmt"

	"github.com/fatih/color"
)

// TODO default to ASCII if Unicode is not supported
const (
	// Character indicating debug message.
	DebugMark = "☹"

	// Character indicating success message.
	CheckMark = "✔"

	// Character indicating information message.
	InfoMark = "i"

	// Character indicating warning message.
	WarnMark = "!"

	// Character indicating error message.
	ErrorMark = "✘"

	// TODO use for prompts
	// Character indicating prompt message.
	QuestionMark = "?"
)

func coloredPrintMsg(icon string, msg string, iC color.Attribute, mC color.Attribute) {
	fmt.Println(
		color.New(mC).SprintFunc()("["+icon+"]"),
		color.New(color.Bold, iC).SprintFunc()(msg))
}

// TODO add log levels
// Debug logs the given message as a debug message.
func Debug(msg string) {
	coloredPrintMsg(DebugMark, msg, color.FgYellow, color.FgYellow)
}

// Success logs the given message as a success message.
func Success(msg string) {
	coloredPrintMsg(CheckMark, msg, color.FgWhite, color.FgGreen)
}

// Info logs the given message as a info message.
func Info(msg string) {
	coloredPrintMsg(InfoMark, msg, color.FgBlue, color.FgBlue)
}

// Warn logs the given message as a warn message.
func Warn(msg string) {
	coloredPrintMsg(WarnMark, msg, color.FgMagenta, color.FgMagenta)
}

// Error logs the given message as a error message.
func Error(msg string) {
	coloredPrintMsg(ErrorMark, msg, color.FgRed, color.FgRed)
}

// Fatal logs the given message as a fatal message.
func Fatal(msg string) {
	Error(msg)
}

// TODO use dependency injection wrapper for fmt.Print usage in the code base

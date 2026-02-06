package errors

import (
	"fmt"
	"runtime"
	"strings"
)

const maxFrames = 32

// stack holds program counters captured at error creation time.
type stack []uintptr

// capture records the call stack, skipping the specified number of
// frames (runtime.Callers itself, capture, and the errors constructor).
func capture(skip int) *stack {
	var pcs [maxFrames]uintptr
	n := runtime.Callers(skip, pcs[:])
	s := stack(pcs[:n])
	return &s
}

// Frame represents a single frame in a call stack.
type Frame struct {
	Function string // Fully qualified function name
	File     string // Source file path
	Line     int    // Line number in source file
}

// String formats a single frame as "function\n\tfile:line".
func (f Frame) String() string {
	return fmt.Sprintf("%s\n\t%s:%d", f.Function, f.File, f.Line)
}

// frames converts program counters into structured Frame values.
func (s *stack) frames() []Frame {
	if s == nil || len(*s) == 0 {
		return nil
	}

	var frames []Frame
	ci := runtime.CallersFrames([]uintptr(*s))
	for {
		frame, more := ci.Next()
		frames = append(frames, Frame{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
		})
		if !more {
			break
		}
	}
	return frames
}

// Format returns a multi-line string of the full stack trace.
func (s *stack) Format() string {
	if s == nil {
		return ""
	}

	var b strings.Builder
	for _, f := range s.frames() {
		b.WriteString(f.String())
		b.WriteString("\n")
	}
	return b.String()
}

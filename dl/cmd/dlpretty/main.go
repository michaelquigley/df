package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// ANSI color codes matching dl's options.go
const (
	colorTimestamp = "\033[90m" // dark gray
	colorFunction  = "\033[36m" // cyan
	colorChannel   = "\033[35m" // magenta
	colorFields    = "\033[33m" // yellow
	colorError     = "\033[31m" // red
	colorWarning   = "\033[33m" // yellow
	colorInfo      = "\033[37m" // white
	colorDebug     = "\033[34m" // blue
	colorReset     = "\033[0m"
)

// Level labels matching dl's options.go
var levelLabels = map[string]string{
	"ERROR": colorError + "  ERROR" + colorReset,
	"WARN":  colorWarning + "WARNING" + colorReset,
	"INFO":  colorInfo + "   INFO" + colorReset,
	"DEBUG": colorDebug + "  DEBUG" + colorReset,
}

// SourceInfo represents the source location from slog JSON output
type SourceInfo struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// Command-line flags
var (
	absoluteTime bool
	trimPrefix   string
)

// State for relative timestamp calculation
var firstTimestamp time.Time

func main() {
	flag.BoolVar(&absoluteTime, "absolute", false, "show absolute timestamps")
	flag.BoolVar(&absoluteTime, "a", false, "show absolute timestamps (shorthand)")
	flag.StringVar(&trimPrefix, "trim", "", "trim prefix from function names")
	flag.StringVar(&trimPrefix, "t", "", "trim prefix from function names (shorthand)")
	flag.Parse()

	if flag.NArg() == 0 {
		filter(os.Stdin)
	} else {
		for _, filename := range flag.Args() {
			f, err := os.Open(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error opening %s: %v\n", filename, err)
				continue
			}
			filter(f)
			f.Close()
		}
	}
}

func filter(r io.Reader) {
	scanner := bufio.NewScanner(r)
	// Increase buffer size for long log lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		formatLine(line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
	}
}

func formatLine(line string) {
	// Find the start of JSON (first '{')
	jsonStart := strings.Index(line, "{")
	if jsonStart == -1 {
		// No JSON found, print line as-is in yellow
		fmt.Println(colorFields + line + colorReset)
		return
	}

	jsonStr := line[jsonStart:]

	// Parse JSON into a map to get all fields
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		// Invalid JSON, print line as-is in yellow
		fmt.Println(colorFields + line + colorReset)
		return
	}

	// Extract known fields
	timeStr, _ := raw["time"].(string)
	level, _ := raw["level"].(string)
	msg, _ := raw["msg"].(string)
	channel, _ := raw["channel"].(string)

	// Extract source info
	var functionName string
	if source, ok := raw["source"].(map[string]interface{}); ok {
		functionName, _ = source["function"].(string)
	}

	// Apply trim prefix to function name
	if trimPrefix != "" && functionName != "" {
		functionName = strings.TrimPrefix(functionName, trimPrefix)
	}

	// Calculate timestamp
	var timeLabel string
	if t, err := time.Parse(time.RFC3339Nano, timeStr); err == nil {
		if absoluteTime {
			timeLabel = fmt.Sprintf("[%s]", t.Format("2006-01-02 15:04:05.000"))
		} else {
			if firstTimestamp.IsZero() {
				firstTimestamp = t
			}
			delta := t.Sub(firstTimestamp).Seconds()
			timeLabel = fmt.Sprintf("[%8.3f]", delta)
		}
	} else {
		timeLabel = "[        ]"
	}

	// Get level label
	levelLabel := levelLabels[level]
	if levelLabel == "" {
		levelLabel = fmt.Sprintf("%7s", level)
	}

	// Build extra fields (exclude known keys)
	knownKeys := map[string]bool{
		"time":    true,
		"level":   true,
		"msg":     true,
		"source":  true,
		"channel": true,
	}
	extra := make(map[string]interface{})
	for k, v := range raw {
		if !knownKeys[k] {
			extra[k] = v
		}
	}

	// Build output
	var out strings.Builder

	// Timestamp
	out.WriteString(colorTimestamp + timeLabel + colorReset)

	// Level
	out.WriteString(" " + levelLabel)

	// Function name
	if functionName != "" {
		out.WriteString(" " + colorFunction + functionName + colorReset)
	}

	// Channel
	if channel != "" {
		out.WriteString(colorChannel + " |" + channel + "|" + colorReset)
	}

	// Extra fields as JSON
	if len(extra) > 0 {
		extraBytes, _ := json.Marshal(extra)
		out.WriteString(" " + colorFields + string(extraBytes) + colorReset)
	}

	// Message
	out.WriteString(" " + msg)

	fmt.Println(out.String())
}

package filelog

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Entry struct {
	ID         string                 `json:"id"`
	Timestamp  string                 `json:"timestamp"`
	Scope      string                 `json:"scope"`
	Kind       string                 `json:"kind"`
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	StatusCode int                    `json:"statusCode,omitempty"`
	Source     string                 `json:"source"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
}

type ReadOptions struct {
	Scope string
	Kind  string
	Date  string
	Level string
	Query string
	Page  int
	Limit int
}

type ReadResult struct {
	Data       []Entry `json:"data"`
	TotalCount int     `json:"totalCount"`
	TotalPages int     `json:"totalPages"`
	FileName   string  `json:"fileName"`
}

type DailyLogger struct {
	baseDir string
	source  string
	mu      sync.Mutex
}

func NewDailyLogger(baseDir, source string) *DailyLogger {
	trimmed := strings.TrimSpace(baseDir)
	if trimmed == "" {
		trimmed = "logs"
	}
	return &DailyLogger{baseDir: trimmed, source: source}
}

func normalizeScope(scope string) string {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "admin":
		return "admin"
	default:
		return "app"
	}
}

func normalizeKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "api":
		return "api"
	case "error":
		return "error"
	default:
		return "access"
	}
}

func normalizeLevel(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return "debug"
	case "warn":
		return "warn"
	case "error":
		return "error"
	default:
		return "info"
	}
}

func normalizeDate(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Now().Format("20060102")
	}
	if len(trimmed) == len("2006-01-02") {
		if t, err := time.ParseInLocation("2006-01-02", trimmed, time.Local); err == nil {
			return t.Format("20060102")
		}
	}
	if len(trimmed) == len("20060102") {
		if _, err := time.ParseInLocation("20060102", trimmed, time.Local); err == nil {
			return trimmed
		}
	}
	return time.Now().Format("20060102")
}

func toInt(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	default:
		return 0
	}
}

func containsQuery(e Entry, query string) bool {
	if query == "" {
		return true
	}
	q := strings.ToLower(query)
	if strings.Contains(strings.ToLower(e.Message), q) {
		return true
	}
	if strings.Contains(strings.ToLower(e.Path), q) {
		return true
	}
	if strings.Contains(strings.ToLower(e.Method), q) {
		return true
	}
	for key, value := range e.Fields {
		if strings.Contains(strings.ToLower(key), q) {
			return true
		}
		if strings.Contains(strings.ToLower(fmt.Sprint(value)), q) {
			return true
		}
	}
	return false
}

func (l *DailyLogger) buildFilePath(scope, kind, date string) string {
	fileName := fmt.Sprintf("%s_log_%s.log", kind, date)
	return filepath.Join(l.baseDir, scope, fileName)
}

func (l *DailyLogger) Log(
	scope, kind, level, message string,
	fields map[string]interface{},
) {
	entry := Entry{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Scope:     normalizeScope(scope),
		Kind:      normalizeKind(kind),
		Level:     normalizeLevel(level),
		Message:   strings.TrimSpace(message),
		Source:    l.source,
		Fields:    fields,
	}

	if entry.Message == "" {
		entry.Message = "log"
	}

	if fields != nil {
		if method, ok := fields["method"].(string); ok {
			entry.Method = method
		}
		if path, ok := fields["path"].(string); ok {
			entry.Path = path
		}
		entry.StatusCode = toInt(fields["statusCode"])
	}

	payload, err := json.Marshal(entry)
	if err != nil {
		return
	}

	date := time.Now().Format("20060102")
	filePath := l.buildFilePath(entry.Scope, entry.Kind, date)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	_, _ = f.Write(append(payload, '\n'))
}

func (l *DailyLogger) Read(options ReadOptions) (ReadResult, error) {
	scope := normalizeScope(options.Scope)
	kind := normalizeKind(options.Kind)
	date := normalizeDate(options.Date)
	level := strings.ToLower(strings.TrimSpace(options.Level))
	query := strings.TrimSpace(options.Query)
	page := options.Page
	limit := options.Limit
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 30
	}

	filePath := l.buildFilePath(scope, kind, date)
	fileName := filepath.Base(filePath)
	f, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ReadResult{Data: []Entry{}, TotalCount: 0, TotalPages: 0, FileName: fileName}, nil
		}
		return ReadResult{}, err
	}
	defer f.Close()

	entries := make([]Entry, 0, 256)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			entries = append(entries, Entry{
				ID:        fmt.Sprintf("broken-%d", len(entries)+1),
				Timestamp: "",
				Scope:     scope,
				Kind:      kind,
				Level:     "error",
				Message:   line,
				Source:    l.source,
			})
			continue
		}

		e.Scope = normalizeScope(e.Scope)
		e.Kind = normalizeKind(e.Kind)
		e.Level = normalizeLevel(e.Level)
		if e.Source == "" {
			e.Source = l.source
		}
		if level != "" && e.Level != level {
			continue
		}
		if !containsQuery(e, query) {
			continue
		}

		entries = append(entries, e)
	}
	if err := scanner.Err(); err != nil {
		return ReadResult{}, err
	}

	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	total := len(entries)
	if total == 0 {
		return ReadResult{Data: []Entry{}, TotalCount: 0, TotalPages: 0, FileName: fileName}, nil
	}
	totalPages := (total + limit - 1) / limit
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * limit
	if start < 0 {
		start = 0
	}
	end := start + limit
	if end > total {
		end = total
	}

	return ReadResult{
		Data:       entries[start:end],
		TotalCount: total,
		TotalPages: totalPages,
		FileName:   fileName,
	}, nil
}

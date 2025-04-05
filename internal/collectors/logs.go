package collectors

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/tiwariParth/whosay/internal/models"
	"github.com/tiwariParth/whosay/internal/ui"
)

// Common log file paths by OS
var commonLogPaths = map[string][]string{
	"linux": {
		"/var/log/syslog",
		"/var/log/auth.log",
		"/var/log/kern.log",
		"/var/log/dmesg",
		"/var/log/messages",
	},
	"darwin": {
		"/var/log/system.log",
		"/var/log/wifi.log",
		"/var/log/install.log",
	},
	"windows": {
		"C:\\Windows\\Logs\\CBS\\CBS.log",
		"C:\\Windows\\Logs\\DISM\\DISM.log",
	},
}

// Default log patterns to highlight
var logPatterns = map[string]struct {
	pattern *regexp.Regexp
	level   string
}{
	"error":   {regexp.MustCompile(`(?i)(error|fail|exception)`), "error"},
	"warning": {regexp.MustCompile(`(?i)(warning|warn)`), "warning"},
	"info":    {regexp.MustCompile(`(?i)(info|notice)`), "info"},
	"debug":   {regexp.MustCompile(`(?i)(debug)`), "debug"},
}

const defaultLogLines = 20

// Log cache for better performance
var (
	logCache      = make(map[string][]models.LogEntry)
	logCacheMutex sync.RWMutex
	logWatchers   = make(map[string]*LogWatcher)
)

// LogWatcher watches a log file for changes
type LogWatcher struct {
	filePath    string
	lastSize    int64
	lastModTime time.Time
	entries     []models.LogEntry
	stopChan    chan struct{}
	watching    bool
}

// GetLogInfo displays log information
func GetLogInfo(opts models.Options) {
	logFiles := getLogFiles()

	if opts.JSONOutput {
		allEntries := []models.LogEntry{}
		for _, logFile := range logFiles {
			entries := getLogEntries(logFile, defaultLogLines, opts.VerboseOutput)
			allEntries = append(allEntries, entries...)
		}
		
		jsonData, err := json.MarshalIndent(allEntries, "", "  ")
		if err != nil {
			fmt.Printf("Error serializing log data: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
		return
	}

	sections := GetLogInfoSections(opts)
	ui.CompactDisplay(sections)
}

// GetLogInfoSections formats log information for compact display
func GetLogInfoSections(opts models.Options) map[string][][]string {
	logFiles := getLogFiles()
	
	result := map[string][][]string{
		"System Logs": {
			{"Available Logs", fmt.Sprintf("%d", len(logFiles))},
		},
	}

	for i, logPath := range logFiles {
		if i >= 3 && !opts.VerboseOutput {
			break
		}
		
		if _, err := os.Stat(logPath); err != nil {
			continue
		}
		
		entries := getLogEntries(logPath, defaultLogLines, opts.VerboseOutput)
		if len(entries) == 0 {
			continue
		}
		
		logName := filepath.Base(logPath)
		logSection := [][]string{
			{"Path", logPath},
			{"Entries", fmt.Sprintf("%d shown (newest first)", len(entries))},
			{"", ""}, // Spacer
		}
		
		for _, entry := range entries {
			timestamp := ""
			if !entry.Timestamp.IsZero() {
				timestamp = entry.Timestamp.Format("15:04:05")
			}
			
			content := entry.Content
			if len(content) > 80 && !opts.VerboseOutput {
				content = content[:77] + "..."
			}
			
			logSection = append(logSection, []string{
				timestamp,
				content,
			})
		}
		
		result[fmt.Sprintf("Log: %s", logName)] = logSection
	}
	
	return result
}

// getLogFiles returns available log files based on OS
func getLogFiles() []string {
	var paths []string
	
	if osPaths, ok := commonLogPaths[runtime.GOOS]; ok {
		paths = append(paths, osPaths...)
	}
	
	existingPaths := []string{}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existingPaths = append(existingPaths, path)
		}
	}
	
	return existingPaths
}

// getLogEntries reads and parses log entries from a file
func getLogEntries(filePath string, numLines int, includeDebug bool) []models.LogEntry {
	logCacheMutex.RLock()
	cachedEntries, found := logCache[filePath]
	logCacheMutex.RUnlock()
	
	if found && len(cachedEntries) > 0 {
		return filterLogEntries(cachedEntries, numLines, includeDebug)
	}
	
	file, err := os.Open(filePath)
	if err != nil {
		return []models.LogEntry{}
	}
	defer file.Close()
	
	fileInfo, err := file.Stat()
	if err != nil {
		return []models.LogEntry{}
	}
	
	var reader io.Reader = file
	fileSize := fileInfo.Size()
	if fileSize > 50000 { // ~50KB
		_, err := file.Seek(-50000, io.SeekEnd)
		if err != nil {
			_, _ = file.Seek(0, io.SeekStart)
		}
	}
	
	entries := parseLogFile(reader, filePath, fileInfo.ModTime())
	
	logCacheMutex.Lock()
	logCache[filePath] = entries
	logCacheMutex.Unlock()
	
	startLogWatcher(filePath, fileSize, fileInfo.ModTime())
	
	return filterLogEntries(entries, numLines, includeDebug)
}

// parseLogFile reads a log file and extracts log entries
func parseLogFile(reader io.Reader, filePath string, modTime time.Time) []models.LogEntry {
	var entries []models.LogEntry
	
	scanner := bufio.NewScanner(reader)
	
	var parseFunc func(string) models.LogEntry
	
	switch {
	case strings.Contains(filePath, "syslog"):
		parseFunc = parseSyslogLine
	case strings.Contains(filePath, "auth.log"):
		parseFunc = parseAuthLogLine
	default:
		parseFunc = parseGenericLogLine
	}
	
	for scanner.Scan() {
		line := scanner.Text()
		entry := parseFunc(line)
		
		entry.Source = filePath
		
		if entry.Timestamp.IsZero() {
			entry.Timestamp = modTime
		}
		
		entries = append(entries, entry)
	}
	
	// Reverse the order so newest entries are first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	
	return entries
}

// Parsers for different log formats
func parseSyslogLine(line string) models.LogEntry {
	entry := models.LogEntry{
		Content: line,
		Level:   detectLogLevel(line),
	}
	
	timestampRegex := regexp.MustCompile(`^(\w{3}\s+\d+\s+\d{2}:\d{2}:\d{2})`)
	if matches := timestampRegex.FindStringSubmatch(line); len(matches) > 1 {
		if ts, err := time.Parse("Jan 2 15:04:05", matches[1]); err == nil {
			currentYear := time.Now().Year()
			entry.Timestamp = time.Date(currentYear, ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), 0, time.Local)
		}
	}
	
	return entry
}

func parseAuthLogLine(line string) models.LogEntry {
	return parseSyslogLine(line)
}

func parseGenericLogLine(line string) models.LogEntry {
	entry := models.LogEntry{
		Content: line,
		Level:   detectLogLevel(line),
	}
	
	isoRegex := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2})`)
	if matches := isoRegex.FindStringSubmatch(line); len(matches) > 1 {
		formats := []string{
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
		}
		
		for _, format := range formats {
			if ts, err := time.Parse(format, matches[1]); err == nil {
				entry.Timestamp = ts
				break
			}
		}
	}
	
	return entry
}

func detectLogLevel(content string) string {
	for _, pattern := range logPatterns {
		if pattern.pattern.MatchString(content) {
			return pattern.level
		}
	}
	return "info"
}

func filterLogEntries(entries []models.LogEntry, limit int, includeDebug bool) []models.LogEntry {
	var filtered []models.LogEntry
	
	for _, entry := range entries {
		if !includeDebug && entry.Level == "debug" {
			continue
		}
		
		filtered = append(filtered, entry)
		
		if len(filtered) >= limit {
			break
		}
	}
	
	return filtered
}

func startLogWatcher(filePath string, fileSize int64, modTime time.Time) {
	if watcher, found := logWatchers[filePath]; found && watcher.watching {
		return
	}
	
	watcher := &LogWatcher{
		filePath:    filePath,
		lastSize:    fileSize,
		lastModTime: modTime,
		stopChan:    make(chan struct{}),
		watching:    true,
	}
	
	logWatchers[filePath] = watcher
	
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				checkLogFileChanges(watcher)
			case <-watcher.stopChan:
				return
			}
		}
	}()
}

func checkLogFileChanges(watcher *LogWatcher) {
	fileInfo, err := os.Stat(watcher.filePath)
	if err != nil {
		return
	}
	
	if fileInfo.Size() != watcher.lastSize || fileInfo.ModTime() != watcher.lastModTime {
		file, err := os.Open(watcher.filePath)
		if err != nil {
			return
		}
		defer file.Close()
		
		entries := parseLogFile(file, watcher.filePath, fileInfo.ModTime())
		
		logCacheMutex.Lock()
		logCache[watcher.filePath] = entries
		logCacheMutex.Unlock()
		
		watcher.lastSize = fileInfo.Size()
		watcher.lastModTime = fileInfo.ModTime()
	}
}

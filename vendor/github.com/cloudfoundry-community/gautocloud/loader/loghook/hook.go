package loghook

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

const (
	LOG_MESSAGE_PREFIX = "gautocloud"
	DEBUG_MODE_ENV_VAR = "GAUTOCLOUD_DEBUG"
	DEBUG_MODE_JSON    = "json"
	BUF_SIZE           = 35
)

type GautocloudHook struct {
	entries       []*logrus.Entry
	nbWrite       int
	buf           *bytes.Buffer
	jsonFormatter *logrus.JSONFormatter
}

func NewGautocloudHook(buf *bytes.Buffer) *GautocloudHook {
	return &GautocloudHook{
		entries:       make([]*logrus.Entry, 0),
		nbWrite:       0,
		buf:           buf,
		jsonFormatter: &logrus.JSONFormatter{},
	}
}
func (h GautocloudHook) IsDebugModeJson() bool {
	return os.Getenv(DEBUG_MODE_ENV_VAR) == DEBUG_MODE_JSON
}
func (h GautocloudHook) IsDebugMode() bool {
	return os.Getenv(DEBUG_MODE_ENV_VAR) != ""
}
func (h GautocloudHook) traceDebugMode(entry *logrus.Entry) {
	stdLogger := logrus.StandardLogger()
	currentLvl := stdLogger.Level
	stdLogger.Level = logrus.DebugLevel
	if h.IsDebugModeJson() {
		h.trace(entry, h.jsonFormatter)
	} else {
		h.trace(entry)
	}
	stdLogger.Level = currentLvl
}
func (h GautocloudHook) toLine(entry *logrus.Entry, formatters ...logrus.Formatter) string {
	stdLogger := logrus.StandardLogger()
	currentOut := entry.Logger.Out
	entry.Logger.Out = stdLogger.Out
	formatter := stdLogger.Formatter
	if len(formatters) > 0 {
		formatter = formatters[0]
	}
	b, _ := formatter.Format(entry)
	line := string(b)
	entry.Logger.Out = currentOut
	return line
}
func (h GautocloudHook) trace(entry *logrus.Entry, formatters ...logrus.Formatter) {
	stdLogger := logrus.StandardLogger()
	fmt.Fprint(stdLogger.Out, h.toLine(entry, formatters...))
}
func (h *GautocloudHook) Fire(entry *logrus.Entry) error {
	defer h.buf.Reset()

	if !strings.HasPrefix(entry.Message, LOG_MESSAGE_PREFIX) {
		if entry.Level > logrus.GetLevel() {
			return nil
		}
		h.trace(entry)
		return nil
	}

	currentLvl := logrus.GetLevel()
	if entry.Level <= currentLvl {
		h.trace(entry)
		return nil
	}
	if h.IsDebugMode() {
		h.traceDebugMode(entry)
		return nil
	}

	if h.nbWrite == BUF_SIZE {
		h.entries = make([]*logrus.Entry, 0)
		h.nbWrite = 0
	}
	h.entries = append(h.entries, entry)
	h.nbWrite++
	return nil
}
func (h GautocloudHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
func (h *GautocloudHook) ShowPreviousLog() {
	newEntries := make([]*logrus.Entry, 0)
	stdLogger := logrus.StandardLogger()
	if len(h.entries) == 0 {
		return
	}
	stdLogger.Warn("")
	stdLogger.Warnf(
		"%s: Show previous log was called, next logs was stored between '%s' and '%s'.",
		LOG_MESSAGE_PREFIX,
		h.entries[0].Time.Format("15:04:05.999999999"),
		h.entries[len(h.entries)-1].Time.Format("15:04:05.999999999"),
	)
	for i := len(h.entries) - 1; i >= 0; i-- {
		entry := h.entries[i]
		if entry.Level > logrus.GetLevel() {
			newEntries = append(newEntries, entry)
			continue
		}
		h.trace(entry)
	}
	h.entries = newEntries
	h.nbWrite = len(newEntries)
	stdLogger.Warnf(
		"%s: Finished to show previous logs.",
		LOG_MESSAGE_PREFIX,
	)
	stdLogger.Warn("")
}

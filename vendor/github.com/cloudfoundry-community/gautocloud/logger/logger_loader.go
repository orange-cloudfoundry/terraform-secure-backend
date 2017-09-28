package logger

import (
	"bytes"
	"strings"
	"log"
)

const prefixLog = "Gautocloud - "

type Level int

const (
	Loff = Level(^uint(0) >> 1)
	Lsevere = Level(1000)
	Lerror = Level(900)
	Lwarning = Level(800)
	Linfo = Level(700)
	Ldebug = Level(600)
	Lall = Level(-Loff - 1)
)

type LoggerLoader struct {
	logger    *log.Logger
	bufLogger *bytes.Buffer
	lvl       Level
}

func NewLoggerLoader() *LoggerLoader {
	bufLogger := new(bytes.Buffer)
	logger := log.New(bufLogger, "", log.Ldate | log.Ltime)
	return &LoggerLoader{
		logger: logger,
		bufLogger: bufLogger,
		lvl: Lall,
	}
}
func (l *LoggerLoader) SetLogger(logger *log.Logger) {
	l.logger = logger
	previousData := l.sanitizeLog(l.bufLogger.String())
	if previousData != "" {
		flags := l.logger.Flags()
		l.logger.SetFlags(0)
		l.logger.Print(previousData)
		l.logger.SetFlags(flags)
	}
	l.bufLogger.Reset()
}
func (l LoggerLoader) sanitizeLog(prevLog string) string {
	if l.lvl == Loff {
		return ""
	}
	lines := strings.Split(prevLog, "\n")
	toSanitize := make([]string, 0)
	if l.lvl >= Linfo {
		toSanitize = append(toSanitize, "DEBUG:")
	}
	if l.lvl >= Lwarning {
		toSanitize = append(toSanitize, "INFO:")
	}
	if l.lvl >= Lerror {
		toSanitize = append(toSanitize, "WARNING:")
	}
	if l.lvl >= Lsevere {
		toSanitize = append(toSanitize, "ERROR:")
	}
	finalLines := make([]string, 0)
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if !l.strContainsEltSlice(line, toSanitize) {
			finalLines = append(finalLines, line)
			continue
		}
		for p := i + 1; p < len(lines); p++ {
			if !strings.HasPrefix(lines[p], "\t") {
				break
			}
			i++
		}
	}
	return strings.Join(finalLines, "\n")
}
func (l LoggerLoader) strContainsEltSlice(line string, toSanitize []string) bool {
	for _, san := range toSanitize {
		if strings.Contains(line, san) {
			return true
		}
	}
	return false
}
func (l LoggerLoader) recoverBuffer() {
	if recover() != nil {
		l.bufLogger.Reset()
	}
}
func (l LoggerLoader) Info(format string, v ...interface{}) {
	defer l.recoverBuffer()
	if l.lvl > Linfo {
		return
	}
	l.logger.Printf(prefixLog + "INFO: " + format + "\n", v...)
}
func (l LoggerLoader) Debug(format string, v ...interface{}) {
	defer l.recoverBuffer()
	if l.lvl > Ldebug {
		return
	}
	l.logger.Printf(prefixLog + "DEBUG: " + format + "\n", v...)
}
func (l LoggerLoader) Warn(format string, v ...interface{}) {
	defer l.recoverBuffer()
	if l.lvl > Lwarning {
		return
	}
	l.logger.Printf(prefixLog + "WARNING: " + format + "\n", v...)
}
func (l LoggerLoader) Error(format string, v ...interface{}) {
	defer l.recoverBuffer()
	if l.lvl > Lerror {
		return
	}
	l.logger.Printf(prefixLog + "ERROR: " + format + "\n", v...)
}
func (l LoggerLoader) Level() Level {
	return l.lvl
}
func (l *LoggerLoader) SetLevel(lvl Level) {
	l.lvl = lvl
}

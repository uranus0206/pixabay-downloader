package qlogger

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	Configure(QConfig)
}

// Configuration for logging
type Config struct {
	// Enable console logging
	ConsoleLoggingEnabled bool

	// EncodeLogsAsJson makes the log framework log JSON
	EncodeLogsAsJson bool
	// FileLoggingEnabled makes the framework log to a file
	// the fields below can be skipped if this value is false!
	FileLoggingEnabled bool
	// Directory to log to to when filelogging is enabled
	Directory string
	// Filename is the name of the logfile which will be placed inside the directory
	Filename string
	// MaxSize the max size in MB of the logfile before it's rolled
	MaxSize int
	// MaxBackups the max number of rolled files to keep
	MaxBackups int
	// MaxAge the max age in days to keep a logfile
	MaxAge int
}

type QLogger struct {
	*zerolog.Logger
}

var QConfig Config = Config{
	ConsoleLoggingEnabled: true,
	EncodeLogsAsJson:      true,
	FileLoggingEnabled:    false,
	Directory:             "./",
	Filename:              "pixabay.log",
	MaxSize:               1,
	MaxBackups:            2,
	MaxAge:                7,
}

var QLoggerShared *QLogger

// Panic fatals arguments.
func Panic(args ...interface{}) {
	QLoggerShared.Panic().Msg(fmt.Sprint(args...))
}

// Fatalf fatals formatted string with arguments.
func Panicf(format string, args ...interface{}) {
	QLoggerShared.Panic().Msgf(format, args...)
}

// Fatalln fatals and new line.
func Panicln(args ...interface{}) {
	Panic(args...)
}

// Fatal fatals arguments.
func Fatal(args ...interface{}) {
	QLoggerShared.Fatal().Msg(fmt.Sprint(args...))
}

// Fatalf fatals formatted string with arguments.
func Fatalf(format string, args ...interface{}) {
	QLoggerShared.Fatal().Msgf(format, args...)
}

// Fatalln fatals and new line.
func Fatalln(args ...interface{}) {
	Fatal(args...)
}

// Error errors arguments.
func Error(args ...interface{}) {
	QLoggerShared.Error().Msg(fmt.Sprint(args...))
}

// Errorf errors formatted string with arguments.
func Errorf(format string, args ...interface{}) {
	QLoggerShared.Error().Msgf(format, args...)
}

// Errorln errors and new line.
func Errorln(args ...interface{}) {
	QLoggerShared.Error().Msg(fmt.Sprint(args...))
}

// Debug debugs arguments.
func Debug(args ...interface{}) {
	QLoggerShared.Debug().Stack().Msg(fmt.Sprint(args...))
}

// Debugf debugs formatted string with arguments.
func Debugf(format string, args ...interface{}) {
	QLoggerShared.Debug().Msgf(format, args...)
}

// Debugln debugs and new line.
func Debugln(args ...interface{}) {
	QLoggerShared.Debug().Stack().Msg(fmt.Sprint(args...))
}

// Info infos arguments.
func Info(args ...interface{}) {
	QLoggerShared.Info().Stack().Msg(fmt.Sprint(args...))
}

// Infof infos formatted string with arguments.
func Infof(format string, args ...interface{}) {
	QLoggerShared.Info().Msgf(format, args...)
}

// Infoln infos and new line.
func Infoln(args ...interface{}) {
	QLoggerShared.Info().Stack().Msg(fmt.Sprint(args...))
}

// Warning warns arguments.
func Warning(args ...interface{}) {
	QLoggerShared.Warn().Msg(fmt.Sprint(args...))
}

// Warningf warns formatted string with arguments.
func Warningf(format string, args ...interface{}) {
	QLoggerShared.Warn().Msgf(format, args...)
}

// Warningln warns and new line.
func Warningln(args ...interface{}) {
	QLoggerShared.Warn().Msg(fmt.Sprint(args...))
}

// Print logs arguments.
func Print(args ...interface{}) {
	QLoggerShared.Info().Stack().Msg(fmt.Sprint(args...))
}

// Printf logs formatted string with arguments.
func Printf(format string, args ...interface{}) {
	QLoggerShared.Info().Msgf(format, args...)
}

// Println logs with new line.
func Println(args ...interface{}) {
	QLoggerShared.Info().Stack().Msg(fmt.Sprint(args...))
}

func SetLevel(l int8) {
	zerolog.SetGlobalLevel(zerolog.Level(l))
}

func GetLevel() int8 {
	return int8(zerolog.GlobalLevel())
}

// Configure sets up the logging framework
//
// In production, the container logs will be collected and file logging should be disabled. However,
// during development it's nicer to see logs as text and optionally write to a file when debugging
// problems in the containerized pipeline
//
// The output log file will be located at /var/log/service-xyz/service-xyz.log and
// will be rolled according to configuration set.
func Configure(config Config) *QLogger {
	QConfig = config

	var writers []io.Writer

	// Equivalent of Lshortfile
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}

	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.ErrorFieldName = "err"
	zerolog.MessageFieldName = "msg"
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// Default value is 2.
	// We wrappered as pkg function like Infoln, Println... above. Skip frame count need to add by 1.
	zerolog.CallerSkipFrameCount = 3

	if config.ConsoleLoggingEnabled {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stdout})
	}
	if config.FileLoggingEnabled {
		writers = append(writers, newRollingFile(config))
	}
	mw := io.MultiWriter(writers...)

	// zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := zerolog.New(mw).With().Timestamp().Caller().Logger()

	logger.Info().
		Bool("fileLogging", config.FileLoggingEnabled).
		Bool("jsonLogOutput", config.EncodeLogsAsJson).
		Str("logDirectory", config.Directory).
		Str("fileName", config.Filename).
		Int("maxSizeMB", config.MaxSize).
		Int("maxBackups", config.MaxBackups).
		Int("maxAgeInDays", config.MaxAge).
		Msg("logging configured")

	QLoggerShared = &QLogger{
		Logger: &logger,
	}

	return QLoggerShared
}

func newRollingFile(config Config) io.Writer {
	if err := os.MkdirAll(config.Directory, 0744); err != nil {
		log.Error().Err(err).Str("path", config.Directory).Msg("can't create log directory")
		return nil
	}

	return &lumberjack.Logger{
		Filename:   path.Join(config.Directory, config.Filename),
		MaxBackups: config.MaxBackups, // files
		MaxSize:    config.MaxSize,    // megabytes
		MaxAge:     config.MaxAge,     // days
	}
}

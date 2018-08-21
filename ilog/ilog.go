package ilog

var (
	defaultLogger *Logger
)

func init() {
	defaultLogger = NewConsoleLogger()
	defaultLogger.SetCallDepth(1)
	defaultLogger.Start()
}

// InitLogger inits the global defaultLogger.
func InitLogger(logger *Logger) {
	defaultLogger.Stop()
	defaultLogger = logger
	defaultLogger.Start()
}

// DefaultLogger returns the global defaultLogger.
func DefaultLogger() *Logger {
	return defaultLogger
}

// AddWriter adds writer to the global defaultLogger.
func AddWriter(lw LogWriter) error {
	return defaultLogger.AddWriter(lw)
}

// SetCallDepth sets the global defaultLogger's call depth.
func SetCallDepth(d int) {
	defaultLogger.SetCallDepth(d)
}

// Start starts the global defaultLogger.
func Start() {
	defaultLogger.Start()
}

// Stop stops the global defaultLogger.
func Stop() {
	defaultLogger.Stop()
}

// Flush flushes the global defaultLogger.
func Flush() {
	defaultLogger.Flush()
}

// Debug generates a debug-level log.
func Debug(format string, v ...interface{}) {
	defaultLogger.Debug(format, v...)
}

// Info generates a info-level log.
func Info(format string, v ...interface{}) {
	defaultLogger.Info(format, v...)
}

// Warn generates a warn-level log.
func Warn(format string, v ...interface{}) {
	defaultLogger.Warn(format, v...)
}

// Error generates a error-level log.
func Error(format string, v ...interface{}) {
	defaultLogger.Error(format, v...)
}

// Fatal generates a fatal-level log and exits the program.
func Fatal(format string, v ...interface{}) {
	defaultLogger.Fatal(format, v...)
}

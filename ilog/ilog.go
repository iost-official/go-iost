package ilog

var (
	defaultLogger *Logger
)

func init() {
	defaultLogger = NewConsoleLogger()
	defaultLogger.SetCallDepth(1)
	defaultLogger.Start()
}

func InitLogger(logger *Logger) {
	defaultLogger.Stop()
	defaultLogger = logger
	defaultLogger.Start()
}

func DefaultLogger() *Logger {
	return defaultLogger
}

func AddWriter(lw LogWriter) error {
	return defaultLogger.AddWriter(lw)
}

func SetCallDepth(d int) {
	defaultLogger.SetCallDepth(d)
}

func Start() {
	defaultLogger.Start()
}

func Stop() {
	defaultLogger.Stop()
}

func Flush() {
	defaultLogger.Flush()
}

func Debug(format string, v ...interface{}) {
	defaultLogger.Debug(format, v...)
}

func Info(format string, v ...interface{}) {
	defaultLogger.Info(format, v...)
}

func Warn(format string, v ...interface{}) {
	defaultLogger.Warn(format, v...)
}

func Error(format string, v ...interface{}) {
	defaultLogger.Error(format, v...)
}

func Fatal(format string, v ...interface{}) {
	defaultLogger.Fatal(format, v...)
}

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

// SetLevel sets the defaultLogger's level.
func SetLevel(l Level) {
	defaultLogger.SetLevel(l)
}

// GetLevel gets the defaultLogger's level.
func GetLevel() (l Level) {
	return defaultLogger.GetLevel()
}

// SetCallDepth sets the global defaultLogger's call depth.
func SetCallDepth(d int) {
	defaultLogger.SetCallDepth(d)
}

// AsyncWrite sets logger's syncWrite to false.
func AsyncWrite() {
	defaultLogger.AsyncWrite()
}

// HideLocation sets logger's showLocation to false.
func HideLocation() {
	defaultLogger.HideLocation()
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
func Debug(v ...interface{}) {
	defaultLogger.Debug(v...)
}

// Info generates a info-level log.
func Info(v ...interface{}) {
	defaultLogger.Info(v...)
}

// Warn generates a warn-level log.
func Warn(v ...interface{}) {
	defaultLogger.Warn(v...)
}

// Error generates a error-level log.
func Error(v ...interface{}) {
	defaultLogger.Error(v...)
}

// Fatal generates a fatal-level log and exits the program.
func Fatal(v ...interface{}) {
	defaultLogger.Fatal(v...)
}

// Debugln generates a debug-level log.
func Debugln(v ...interface{}) {
	defaultLogger.Debugln(v...)
}

// Infoln generates a info-level log.
func Infoln(v ...interface{}) {
	defaultLogger.Infoln(v...)
}

// Warnln generates a warn-level log.
func Warnln(v ...interface{}) {
	defaultLogger.Warnln(v...)
}

// Errorln generates a error-level log.
func Errorln(v ...interface{}) {
	defaultLogger.Errorln(v...)
}

// Fatalln generates a fatal-level log and exits the program.
func Fatalln(v ...interface{}) {
	defaultLogger.Fatalln(v...)
}

// Debugf generates a debug-level log.
func Debugf(format string, v ...interface{}) {
	defaultLogger.Debugf(format, v...)
}

// Infof generates a info-level log.
func Infof(format string, v ...interface{}) {
	defaultLogger.Infof(format, v...)
}

// Warnf generates a warn-level log.
func Warnf(format string, v ...interface{}) {
	defaultLogger.Warnf(format, v...)
}

// Errorf generates a error-level log.
func Errorf(format string, v ...interface{}) {
	defaultLogger.Errorf(format, v...)
}

// Fatalf generates a fatal-level log and exits the program.
func Fatalf(format string, v ...interface{}) {
	defaultLogger.Fatalf(format, v...)
}

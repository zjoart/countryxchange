package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log *zap.Logger
)

// Common field keys
const (
	UserIDKey    = "user_id"
	EmailKey     = "email"
	SessionIDKey = "session_id"
	IPKey        = "ip"
	ErrorKey     = "error"
	CallerKey    = "caller"
)

func init() {
	var err error
	config := zap.NewProductionConfig()

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.StacktraceKey = ""
	encoderConfig.CallerKey = CallerKey
	encoderConfig.MessageKey = "message"
	encoderConfig.LevelKey = "level"

	config.EncoderConfig = encoderConfig
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	Log, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
}

// Fields type for more convenient logging
type Fields map[string]interface{}

// Info logs a message at info level with optional fields
func Info(msg string, fields ...Fields) {
	if len(fields) > 0 {
		Log.Info(msg, getZapFields(fields[0])...)
		return
	}
	Log.Info(msg)
}

// Error logs a message at error level with optional fields
func Error(msg string, fields ...Fields) {
	if len(fields) > 0 {
		Log.Error(msg, getZapFields(fields[0])...)
		return
	}
	Log.Error(msg)
}

// Debug logs a message at debug level with optional fields
func Debug(msg string, fields ...Fields) {
	if len(fields) > 0 {
		Log.Debug(msg, getZapFields(fields[0])...)
		return
	}
	Log.Debug(msg)
}

// Warn logs a message at warn level with optional fields
func Warn(msg string, fields ...Fields) {
	if len(fields) > 0 {
		Log.Warn(msg, getZapFields(fields[0])...)
		return
	}
	Log.Warn(msg)
}

// Fatal logs a message at fatal level with optional fields and then calls os.Exit(1)
func Fatal(msg string, fields ...Fields) {
	if len(fields) > 0 {
		Log.Fatal(msg, getZapFields(fields[0])...)
		return
	}
	Log.Fatal(msg)
}

// WithError adds an error field to the log entry
func WithError(err error) Fields {
	return Fields{
		ErrorKey: err.Error(),
	}
}

// WithUserID adds a user ID field to the log entry
func WithUserID(userID string) Fields {
	return Fields{
		UserIDKey: userID,
	}
}

// WithEmail adds an email field to the log entry
func WithEmail(email string) Fields {
	return Fields{
		EmailKey: email,
	}
}

// WithSessionID adds a session ID field to the log entry
func WithSessionID(sessionID string) Fields {
	return Fields{
		SessionIDKey: sessionID,
	}
}

// WithIP adds an IP address field to the log entry
func WithIP(ip string) Fields {
	return Fields{
		IPKey: ip,
	}
}

// Merge combines multiple Fields into a single Fields object
func Merge(fields ...Fields) Fields {
	merged := make(Fields)
	for _, f := range fields {
		for k, v := range f {
			merged[k] = v
		}
	}
	return merged
}

// Convert Fields to zap.Field slice
func getZapFields(fields Fields) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}

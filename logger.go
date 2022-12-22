package gz

import (
	"context"
	"fmt"
	"github.com/rollbar/rollbar-go"
	"log"
	"net/http"
)

// VerbosityDebug - Debug verbosity level
// Output will include Critical + Error + Warning + Info + Debug
const VerbosityDebug = 4

// VerbosityInfo - Info verbosity level
// Output will include Critical + Error + Warning + Info
const VerbosityInfo = 3

// VerbosityWarning - Warning verbosity level
// Output will include Critical + Error + Warning
const VerbosityWarning = 2

// VerbosityError - Error verbosity level
// Output will include Critical + Error
const VerbosityError = 1

// VerbosityCritical - Critical verbosity level
// Output will include Critical
const VerbosityCritical = 0

// Logger - interface for any ign logger.
type Logger interface {
	// Output when verbosity => 4
	Debug(interfaces ...interface{})
	// Output when verbosity => 3
	Info(interfaces ...interface{})
	// Output when verbosity => 2
	Warning(interfaces ...interface{})
	// Output when verbosity => 1
	Error(interfaces ...interface{})
	// Output when verbosity => 0
	Critical(interfaces ...interface{})
	// Clone this logger and returns a copy.
	Clone(reqID string) Logger
}

// default empty logger implementation. To be returned when there is no logger
// in the context. This logger just logs to the stdout.
type defaultLogImpl struct {
}

// Debug - debug log
func (l *defaultLogImpl) Debug(interfaces ...interface{}) {
	fmt.Println("[Debug][IGN EMPTY LOGGER]", interfaces)
}

// Info - info log
func (l *defaultLogImpl) Info(interfaces ...interface{}) {
	fmt.Println("[Info][IGN EMPTY LOGGER]", interfaces)
}

// Warning - logs a warning message.
func (l *defaultLogImpl) Warning(interfaces ...interface{}) {
	fmt.Println("[Warning][IGN EMPTY LOGGER]", interfaces)
}

// Error - logs an error message.
func (l *defaultLogImpl) Error(interfaces ...interface{}) {
	fmt.Println("[Error][IGN EMPTY LOGGER]", interfaces)
}

// Critical - logs a critical message.
func (l *defaultLogImpl) Critical(interfaces ...interface{}) {
	fmt.Println("[Critical][IGN EMPTY LOGGER]", interfaces)
}

// Clone - clone this logger
func (l *defaultLogImpl) Clone(reqID string) Logger {
	return l
}

var emptyLogger *defaultLogImpl

// ignLogger - internal implementation for the ign logger interface.
// The ignLogger will log to terminal and also to rollbar, if configured.
// The ignLogger will prefix all logs with a configured request Id.
type ignLogger struct {
	reqID    string
	rollbar  bool
	logToStd bool
	// Controls the level output
	// 0 = Critical
	// 1 = Critical + Error
	// 2 = Critical + Error + Warning
	// 3 = Critical + Error + Warning + Info
	// 4 = Critical + Error + Warning + Info + Debug
	verbosity int
	// Controls the level to output to rollbar
	RollbarVerbosity int
}

type ctxLoggerType int

const loggerKey ctxLoggerType = iota

// NewLogger - creates a new logger implementation associated to the given
// request ID.
func NewLogger(reqID string, std bool, verbosity int) Logger {
	logger := ignLogger{reqID, true, std, verbosity, verbosity}
	return &logger
}

// NewLoggerWithRollbarVerbosity - creates a new logger implementation associated
// to the given request ID and also configures a minimum verbosity to send logs
// to Rollbar.
func NewLoggerWithRollbarVerbosity(reqID string, std bool, verbosity, rollbarVerbosity int) Logger {
	logger := ignLogger{reqID, true, std, verbosity, rollbarVerbosity}
	return &logger
}

// NewLoggerNoRollbar - creates a new logger implementation associated to the given
// request ID, which does not log to rollbar
func NewLoggerNoRollbar(reqID string, verbosity int) Logger {
	logger := ignLogger{reqID, false, true, verbosity, verbosity}
	return &logger
}

// NewContextWithLogger - configures the context with a new ign Logger,
func NewContextWithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// LoggerFromContext - gets an ign logger from the given context.
func LoggerFromContext(ctx context.Context) Logger {
	if ctx == nil {
		return emptyLogger
	}
	if logger, ok := ctx.Value(loggerKey).(*ignLogger); ok {
		return logger
	}
	return emptyLogger
}

// LoggerFromRequest - gets an ign logger from the given http request.
func LoggerFromRequest(r *http.Request) Logger {
	return LoggerFromContext(r.Context())
}

////////////////////////////////////////////////////////

// Clone creates a new logger based on the current logger and sets a new reqID.
// This is typically used to customize a logger and still honor the original logger
// configuration.
func (l *ignLogger) Clone(reqID string) Logger {
	logger := ignLogger{reqID, true, l.logToStd, l.verbosity, l.RollbarVerbosity}
	return &logger
}

// Debug sends a debug message to rollbar. The valid types are:
//
// *http.Request
// error
// string
// map[string]interface{}
// int
// ignerr.ErrMsg
func (l *ignLogger) Debug(interfaces ...interface{}) {
	if l.verbosity < VerbosityDebug {
		return
	}
	logMsg, msg := processLogInterfaces(l.reqID, interfaces...)
	if l.RollbarVerbosity >= VerbosityDebug && l.shouldSendToRollbar(msg) {
		rollbar.Debug(msg...)
	}
	if l.logToStd {
		log.Println(logMsg)
	}
}

// Info sends an info message to rollbar. The valid types are:
//
// *http.Request
// error
// string
// map[string]interface{}
// int
// ignerr.ErrMsg
func (l *ignLogger) Info(interfaces ...interface{}) {
	if l.verbosity < VerbosityInfo {
		return
	}
	logMsg, msg := processLogInterfaces(l.reqID, interfaces...)
	if l.RollbarVerbosity >= VerbosityInfo && l.shouldSendToRollbar(msg) {
		rollbar.Info(msg...)
	}
	if l.logToStd {
		log.Println(logMsg)
	}
}

// Warning sends a warning message to rollbar. The valid types are:
//
// *http.Request
// error
// string
// map[string]interface{}
// int
// gz.ErrMsg
func (l *ignLogger) Warning(interfaces ...interface{}) {
	if l.verbosity < VerbosityWarning {
		return
	}
	logMsg, msg := processLogInterfaces(l.reqID, interfaces...)
	if l.RollbarVerbosity >= VerbosityWarning && l.shouldSendToRollbar(msg) {
		rollbar.Warning(msg...)
	}
	if l.logToStd {
		log.Println(logMsg)
	}
}

// Error sends an error message to rollbar. The valid types are:
//
// *http.Request
// error
// string
// map[string]interface{}
// int
// ignerr.ErrMsg
// If an error is present then a stack trace is captured. If an int is also present then we skip
// that number of stack frames. If the map is present it is used as extra custom data in the
// item. If a string is present without an error, then we log a message without a stack
// trace. If a request is present we extract as much relevant information from it as we can.
func (l *ignLogger) Error(interfaces ...interface{}) {
	if l.verbosity < VerbosityError {
		return
	}
	logMsg, msg := processLogInterfaces(l.reqID, interfaces...)
	if l.RollbarVerbosity >= VerbosityError && l.shouldSendToRollbar(msg) {
		rollbar.Error(msg...)
	} else {
	}
	if l.logToStd {
		log.Println(logMsg)
	}
}

// Critical sends a critical message to rollbar. The valid types are:
//
// *http.Request
// error
// string
// map[string]interface{}
// int
// ignerr.ErrMsg
func (l *ignLogger) Critical(interfaces ...interface{}) {
	logMsg, msg := processLogInterfaces(l.reqID, interfaces...)
	if l.shouldSendToRollbar(msg) {
		rollbar.Critical(msg...)
	}

	if l.logToStd {
		log.Println(logMsg)
	}
}

// shouldSendToRollbar is a helper function that validates if a processed
// msg can be sent to rollbar.
// The 'msg' argument is expected to be the result from a previous call to
// processLogInterfaces func.
func (l *ignLogger) shouldSendToRollbar(msg []interface{}) bool {
	if !l.rollbar || rollbar.Token() == "" {
		return false
	}

	// Lastly, check if the rollbar msg includes an gz.ErrMsg. If yes, then check for
	// blacklisted error codes.
	if len(msg) > 0 {
		el := msg[len(msg)-1]
		if data, ok := el.(map[string]interface{}); ok {
			if statusCode, ok := data["Status Code"]; ok {
				// change this when having more blacklisted codes
				if statusCode == http.StatusNotFound {
					return false
				}
			}
		}
	}
	return true
}

// processLogInterfaces is a helper function for log reporting that constructs
// a final message to send to rollbar
func processLogInterfaces(prefix string, interfaces ...interface{}) (string, []interface{}) {
	var finalMsg []interface{}
	str := "[" + prefix + "]"
	logMsg := "[" + prefix + "]"
	var hasError bool
	data := map[string]interface{}{}

	// Iterate over each interface
	for index, element := range interfaces {

		var errMsg *ErrMsg

		// The element could be a pointer to ErrMsg or an instance of ErrMsg
		if e, ok := element.(*ErrMsg); ok {
			errMsg = e
		} else if e, ok := element.(ErrMsg); ok {
			errMsg = &e
		}

		// Create a special log message if the type is ErrMsg
		if errMsg != nil {
			// Append the message to the string
			str += errMsg.LogString()
			logMsg += errMsg.LogString()
			finalMsg = append(finalMsg, fmt.Sprintf("[ErrCode:%d] %s. Extra: %v. [Route:%s]",
				errMsg.ErrCode, errMsg.Msg, errMsg.Extra, errMsg.Route))

			// The map of data (Error Code, Status Code, etc) can be viewed by
			// select an occurance of this log message.
			data["Error Code"] = errMsg.ErrCode
			data["Status Code"] = errMsg.StatusCode
			data["Error ID"] = errMsg.ErrID
			data["Route"] = errMsg.Route

			if errMsg.RemoteAddress != "" {
				data["Remote Address"] = errMsg.RemoteAddress
			}

			if errMsg.UserAgent != "" {
				data["User-Agent"] = errMsg.UserAgent
			}

			// Append the base error if one exists.
			if errMsg.BaseError != nil {
				hasError = true
				logMsg = fmt.Sprintf("%s. Base error: %+v\n", logMsg, errMsg.BaseError)
				data["Trace"] = logMsg
				finalMsg = append(finalMsg, errMsg.BaseError)
			}

		} else if e, ok := element.(string); ok {
			// Concatentate strings together
			str += e
			logMsg += e
		} else if e, ok := element.(error); ok {
			// Append an error if it is present
			hasError = true
			// Get a full trace using %+v because we can't be sure of the correct depth
			logMsg = fmt.Sprintf("%s. Base Error: %+v", logMsg, e)
			data["Trace"] = logMsg
			finalMsg = append(finalMsg, e)
		} else {
			finalMsg = append(finalMsg, interfaces[index])
		}
	}

	// If there is error, then rollbar will override a string value with the
	// error data. We can save any string value by putting the string in the
	// data map.
	if hasError {
		data["Message"] = str
	} else {
		finalMsg = append(finalMsg, str)
	}

	data["Prefix"] = prefix
	finalMsg = append(finalMsg, data)

	return logMsg, finalMsg
}

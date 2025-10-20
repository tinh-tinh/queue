package queue

type LoggerType string

const (
	LoggerDefault  LoggerType = "default"
	LoggerInfo     LoggerType = "info"
	LoggerWarn     LoggerType = "warn"
	LoggerError    LoggerType = "error"
	LoggerFatal    LoggerType = "fatal"
	LoggerDisabled LoggerType = "disabled"
)

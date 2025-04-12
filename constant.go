package queue

type LoggerType string

const (
	LoggerDefault  LoggerType = "default"
	LoggerInfo     LoggerType = "info"
	LoggerFatal    LoggerType = "fatal"
	LoggerPanic    LoggerType = "panic"
	LoggerDisabled LoggerType = "disabled"
)

package queue

type Logger interface {
	Infof(msg string, metadata ...any)
	Warnf(msg string, metadata ...any)
	Errorf(msg string, metadata ...any)
	Fatalf(msg string, metadata ...any)
}

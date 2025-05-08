package sl

import (
	"log/slog"
)

var (
	MessageComponent = "component"
	MessageError     = "error"
	MessagePanic     = "panic"
	MessageConf      = "conf"
)

func Error(err error) slog.Attr {
	return slog.String(MessageError, err.Error())
}

func Panic(p any) slog.Attr {
	return slog.Any(MessagePanic, p)
}

func Component(name string) slog.Attr {
	return slog.String(MessageComponent, name)
}

func Conf(conf any) slog.Attr {
	return slog.Any(MessageConf, conf)
}

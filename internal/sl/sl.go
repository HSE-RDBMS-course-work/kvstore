package sl

import (
	"log/slog"
)

func Error(err error) slog.Attr {
	return slog.String("error", err.Error())
}

func Panic(p any) slog.Attr {
	return slog.Any("panic", p)
}

func Component(name string) slog.Attr {
	return slog.String("component", name)
}

func Conf(conf any) slog.Attr {
	return slog.Any("config", conf)
}

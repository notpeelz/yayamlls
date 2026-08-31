package lsp

import (
	"context"
	"log/slog"
	"strings"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (s *Server) Logger() *slog.Logger {
	return slog.New(logHandler{server: s})
}

type logHandler struct{ server *Server }

func (h logHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelDebug
}

const AttrShowMessage = "lsp.showMessage"

func (h logHandler) Handle(_ context.Context, record slog.Record) error {
	level := protocol.MessageTypeLog
	switch {
	case record.Level >= slog.LevelError:
		level = protocol.MessageTypeError
	case record.Level >= slog.LevelWarn:
		level = protocol.MessageTypeWarning
	case record.Level >= slog.LevelInfo:
		level = protocol.MessageTypeInfo
	}

	var showMsg bool
	record.Attrs(func(a slog.Attr) bool {
		if a.Key == AttrShowMessage {
			showMsg = a.Value.Bool()
		}
		return true
	})

	var b strings.Builder
	b.WriteString(record.Message)
	record.Attrs(func(a slog.Attr) bool {
		a.Value = a.Value.Resolve()
		if a.Key == "" || a.Key == AttrShowMessage {
			return true
		}
		b.WriteByte(' ')
		b.WriteString(a.Key)
		b.WriteByte('=')
		b.WriteString(a.Value.String())
		return true
	})

	if n := h.server.currentNotify(); n != nil {
		if showMsg {
			n(protocol.ServerWindowShowMessage, protocol.ShowMessageParams{
				Type:    level,
				Message: b.String(),
			})
		} else {
			n(protocol.ServerWindowLogMessage, protocol.LogMessageParams{
				Type:    level,
				Message: b.String(),
			})
		}
	}
	return nil
}

func (h logHandler) WithAttrs([]slog.Attr) slog.Handler {
	return h
}

func (h logHandler) WithGroup(string) slog.Handler {
	return h
}

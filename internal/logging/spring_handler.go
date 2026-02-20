package logging

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// springHandler formats slog output in a Spring Boot-like format:
//
//	2026-02-20T18:28:34.756+09:00  INFO 12345 --- [gl] [            main] cli.runServer                            : message key=value
type springHandler struct {
	w       io.Writer
	mu      *sync.Mutex
	level   slog.Level
	pid     int
	appName string
	attrs   []slog.Attr
	group   string
}

func newSpringHandler(w io.Writer, level slog.Level) *springHandler {
	return &springHandler{
		w:       w,
		mu:      &sync.Mutex{},
		level:   level,
		pid:     os.Getpid(),
		appName: "gl",
	}
}

func (h *springHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *springHandler) Handle(_ context.Context, r slog.Record) error {
	var buf bytes.Buffer

	// Timestamp
	buf.WriteString(r.Time.Format("2006-01-02T15:04:05.000-07:00"))

	// Level (right-justified, 5 chars) — same as Spring Boot's %5p
	fmt.Fprintf(&buf, " %5s", r.Level.String())

	// PID
	fmt.Fprintf(&buf, " %d --- ", h.pid)

	// Application name
	fmt.Fprintf(&buf, "[%s] ", h.appName)

	// Extract request_id from attrs (used as "thread name" equivalent)
	reqID := "main"
	var extraAttrs []slog.Attr

	for _, a := range h.attrs {
		if a.Key == "request_id" {
			reqID = a.Value.String()
		} else {
			extraAttrs = append(extraAttrs, a)
		}
	}
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "request_id" {
			reqID = a.Value.String()
		} else {
			extraAttrs = append(extraAttrs, a)
		}
		return true
	})

	// Request ID / thread (right-justified, 16 chars) — like Spring Boot's %15.15t
	fmt.Fprintf(&buf, "[%16s] ", reqID)

	// Caller (left-justified, 40 chars) — like Spring Boot's %-40.40logger
	caller := "gate-limiter"
	if r.PC != 0 {
		caller = shortCaller(r.PC)
	}
	fmt.Fprintf(&buf, "%-40s : ", caller)

	// Message
	buf.WriteString(r.Message)

	// Remaining key=value attrs
	for _, a := range extraAttrs {
		buf.WriteByte(' ')
		if h.group != "" {
			buf.WriteString(h.group)
		}
		buf.WriteString(a.Key)
		buf.WriteByte('=')
		appendAttrValue(&buf, a.Value)
	}

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(buf.Bytes())
	return err
}

func (h *springHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs), len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	newAttrs = append(newAttrs, attrs...)

	return &springHandler{
		w:       h.w,
		mu:      h.mu,
		level:   h.level,
		pid:     h.pid,
		appName: h.appName,
		attrs:   newAttrs,
		group:   h.group,
	}
}

func (h *springHandler) WithGroup(name string) slog.Handler {
	return &springHandler{
		w:       h.w,
		mu:      h.mu,
		level:   h.level,
		pid:     h.pid,
		appName: h.appName,
		attrs:   h.attrs,
		group:   h.group + name + ".",
	}
}

// shortCaller extracts a short "package.Type" or "package.Func" from the PC.
//
//	gate-limiter/internal/limiter.(*RateLimitHandler).ServeHTTP → limiter.RateLimitHandler
//	gate-limiter/internal/app.initRateLimiter                   → app.initRateLimiter
//	gate-limiter/internal/cli.runServer.func1                   → cli.runServer
func shortCaller(pc uintptr) string {
	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()

	name := f.Function
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	// Clean pointer receiver syntax: (*Type) → Type
	name = strings.Replace(name, "(*", "", 1)
	name = strings.Replace(name, ")", "", 1)

	// Take first two segments: "package.TypeOrFunc"
	parts := strings.SplitN(name, ".", 3)
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return name
}

func appendAttrValue(buf *bytes.Buffer, v slog.Value) {
	v = v.Resolve()
	if v.Kind() == slog.KindString {
		s := v.String()
		if needsQuoting(s) {
			buf.WriteString(strconv.Quote(s))
			return
		}
		buf.WriteString(s)
		return
	}
	buf.WriteString(v.String())
}

func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	for _, c := range s {
		if c <= ' ' || c == '"' || c == '=' {
			return true
		}
	}
	return false
}

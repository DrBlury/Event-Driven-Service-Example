package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestRestyLoggerMethods(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	resty := NewRestyLogger(logger)

	resty.Errorf("error %s", "msg")
	resty.Warnf("warn %s", "msg")
	resty.Infof("info %s", "msg")
	resty.Debugf("debug %s", "msg")

	out := buf.String()
	for _, keyword := range []string{"error msg", "warn msg", "info msg", "debug msg"} {
		if !strings.Contains(out, keyword) {
			t.Fatalf("expected output to contain %q, got %q", keyword, out)
		}
	}
}

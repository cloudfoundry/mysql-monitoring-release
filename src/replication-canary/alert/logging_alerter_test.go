package alert_test

import (
	"bytes"
	"log/slog"
	"time"

	. "github.com/cloudfoundry/replication-canary/alert"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoggingAlerter", func() {
	var (
		testWriter *bytes.Buffer
		alerter    *LoggingAlerter
	)

	BeforeEach(func() {
		testWriter = new(bytes.Buffer)
		t, _ := time.Parse(time.DateOnly, "2023-10-05")
		replace := TimeReplacer(t)
		// need to set 'LevelDebug' because otherwise it uses the default of Info
		testHandler := slog.NewJSONHandler(testWriter, &slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: replace})
		testLogger := slog.New(testHandler)

		alerter = &LoggingAlerter{
			Logger: testLogger,
		}
	})

	Describe("NotUnhealthy", func() {
		It("logs an INFO-level message with the current time", func() {
			alerter.NotUnhealthy()
			Expect(testWriter.String()).To(ContainSubstring(`"level":"INFO"`))
			Expect(testWriter.String()).To(ContainSubstring(`"msg":"cluster is not unhealthy"`))
			Expect(testWriter.String()).To(ContainSubstring("2023-10-05T00:00:00Z"))
		})
	})

	Describe("Unhealthy", func() {
		It("logs an ERROR-level message with the current time", func() {
			alerter.Unhealthy()
			Expect(testWriter.String()).To(ContainSubstring(`"level":"ERROR"`))
			Expect(testWriter.String()).To(ContainSubstring(`"msg":"cluster is unhealthy"`))
			Expect(testWriter.String()).To(ContainSubstring("2023-10-05T00:00:00Z"))
		})
	})
})

func TimeReplacer(time time.Time) func(groups []string, a slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			a.Value = slog.TimeValue(time)
		}
		return a
	}
}

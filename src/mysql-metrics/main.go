package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/cloudfoundry/mysql-metrics/config"
	"github.com/cloudfoundry/mysql-metrics/cpu"
	"github.com/cloudfoundry/mysql-metrics/database_client"
	"github.com/cloudfoundry/mysql-metrics/disk"
	"github.com/cloudfoundry/mysql-metrics/diskstat"
	"github.com/cloudfoundry/mysql-metrics/emit"
	"github.com/cloudfoundry/mysql-metrics/gather"
	"github.com/cloudfoundry/mysql-metrics/metrics"
	"github.com/cloudfoundry/mysql-metrics/metrics_computer"

	"code.cloudfoundry.org/go-loggregator/v9"
)

const (
	defaultConfigPath = "/var/vcap/jobs/mysql-metrics/config/mysql-config.yml"
)

var (
	configFilepath string
	logFilepath    string
)

func setupGlobalLogger() {
	var handler slog.Handler

	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Only replace top-level groups
			if len(groups) != 0 {
				return a
			}

			// Backwards compatibility with legacy slog format
			switch a.Key {
			case slog.TimeKey:
				return slog.Attr{Key: "timestamp", Value: a.Value}
			case slog.MessageKey:
				return slog.Attr{Key: "message", Value: a.Value}
			case slog.LevelKey:
				return slog.Attr{Key: "level", Value: slog.StringValue(strings.ToLower(a.Value.String()))}
			default:
				return a
			}
		},
	}
	if logFilepath == "" {
		handler = slog.NewJSONHandler(os.Stderr, handlerOpts)
	} else {
		file, err := os.OpenFile(logFilepath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
		if err != nil {
			panic(fmt.Sprintf("Could not open log file: %s\n", err))
		}
		handler = slog.NewJSONHandler(file, handlerOpts)
	}

	slog.SetDefault(slog.New(handler))
}

func main() {
	flag.StringVar(&configFilepath, "c", defaultConfigPath, "location of config file")
	flag.StringVar(&logFilepath, "l", "", "location of log file")
	flag.Parse()

	setupGlobalLogger()

	mysqlMetricsConfig := &config.Config{}
	if err := config.LoadFromFile(configFilepath, mysqlMetricsConfig); err != nil {
		slog.Error("config file is not formatted correctly", "error", err)
		panic(err)
	}

	metricMappingConfig := metrics.DefaultMetricMappingConfig()

	tlsConfig, err := loggregator.NewIngressTLSConfig(
		mysqlMetricsConfig.LoggregatorCAPath,
		mysqlMetricsConfig.LoggregatorClientCertPath,
		mysqlMetricsConfig.LoggregatorClientKeyPath,
	)
	if err != nil {
		slog.Error("loggregator tls config failed to initialize", "error", err)
		panic(err)
	}

	ingressClient, err := loggregator.NewIngressClient(
		tlsConfig,
		loggregator.WithAddr("localhost:3458"),
		loggregator.WithTag("source_id", mysqlMetricsConfig.SourceID),
		loggregator.WithTag("origin", mysqlMetricsConfig.Origin),
	)
	if err != nil {
		slog.Error("loggregator client failed to initialize", "error", err)
		panic(err)
	}
	sender := metrics.NewLoggregatorSender(ingressClient, mysqlMetricsConfig.SourceID)

	conn := Connection(mysqlMetricsConfig)
	dbClient := database_client.NewDatabaseClient(conn, mysqlMetricsConfig)
	stater := disk.NewInfo(syscall.Statfs)
	procStatFile, err := os.Open("/proc/stat")
	if err != nil {
		slog.Error("failed to open /proc/stat", "error", err)
		panic(err)
	}
	cpustater := cpu.New(procStatFile)
	monitor, err := diskstat.NewVolumeMonitor()
	if err != nil {
		metricsLogger.Error("failed to initialize volume monitor", err)
		panic(err)
	}
	gatherer := gather.NewGatherer(dbClient, stater, &cpustater, monitor)

	metricsComputer := metrics_computer.NewMetricsComputer(*metricMappingConfig)
	metricsWriter := metrics.NewMetricWriter(sender, mysqlMetricsConfig.Origin)
	processor := metrics.NewProcessor(gatherer, metricsComputer, metricsWriter, mysqlMetricsConfig)
	metricsInterval := time.Duration(mysqlMetricsConfig.MetricsFrequency) * time.Second
	emitter := emit.NewEmitter(processor, metricsInterval, time.Sleep)

	emitter.Start()
}

func Connection(config *config.Config) *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?tls=preferred", config.Username, config.Password, config.Host, config.Port)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(fmt.Sprintf("Database configuration problem: %v", err))
	}

	return conn
}

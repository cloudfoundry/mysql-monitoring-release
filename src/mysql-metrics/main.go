package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"code.cloudfoundry.org/go-loggregator/v9"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagerflags"

	"github.com/cloudfoundry/mysql-metrics/config"
	"github.com/cloudfoundry/mysql-metrics/cpu"
	"github.com/cloudfoundry/mysql-metrics/database_client"
	"github.com/cloudfoundry/mysql-metrics/disk"
	"github.com/cloudfoundry/mysql-metrics/diskstat"
	"github.com/cloudfoundry/mysql-metrics/emit"
	"github.com/cloudfoundry/mysql-metrics/gather"
	"github.com/cloudfoundry/mysql-metrics/metrics"
	"github.com/cloudfoundry/mysql-metrics/metrics_computer"
)

const (
	defaultConfigPath = "/var/vcap/jobs/mysql-metrics/config/mysql-config.yml"
)

type lagerLoggerWrapper struct {
	logger lager.Logger
}

func (d lagerLoggerWrapper) Debug(action string, message map[string]interface{}) {
	data := lager.Data{}

	for k, v := range message {
		data[k] = v
	}

	d.logger.Debug(action, data)
}

func (d lagerLoggerWrapper) Error(action string, err error) {
	d.logger.Error(action, err)
}

var configFilepath = flag.String("c", defaultConfigPath, "location of config file")
var logFilepath = flag.String("l", "", "location of log file")
var timeFormat = flag.String("timeFormat", "", "timestamp format for logs")

func setupLogging(metricsLogger lager.Logger) {
	if *logFilepath != "" {
		file, err := os.OpenFile(*logFilepath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
		if err != nil {
			panic(fmt.Sprintf("Could not open log file: %s\n", err))
		}

		var sink lager.Sink
		if *timeFormat == "rfc3339" {
			sink = lager.NewPrettySink(file, lager.DEBUG)
		} else {
			sink = lager.NewWriterSink(file, lager.DEBUG)
		}
		metricsLogger.RegisterSink(sink)
	}
}

func main() {
	flag.Parse()

	lagerConfig := lagerflags.DefaultLagerConfig()
	lagerConfig.TimeFormat.Set(*timeFormat)
	metricsLogger, _ := lagerflags.NewFromConfig("MetricsLogger", lagerConfig)

	setupLogging(metricsLogger)

	mysqlMetricsConfig := &config.Config{}
	if err := config.LoadFromFile(*configFilepath, mysqlMetricsConfig); err != nil {
		metricsLogger.Error("config file is not formatted correctly", err)
		panic(err)
	}

	metricMappingConfig := metrics.DefaultMetricMappingConfig()
	loggerWrapper := lagerLoggerWrapper{metricsLogger}

	var senders []metrics.Sender

	// 1. Prometheus Scrape Sender (Always enabled by default for pull-based scraping via prom_scraper / otel-collector)
	if mysqlMetricsConfig.EnablePrometheusExporter {
		promSender := metrics.NewPrometheusSender(
			mysqlMetricsConfig.PrometheusPort,
			mysqlMetricsConfig.SourceID,
			mysqlMetricsConfig.Origin,
			mysqlMetricsConfig.InstanceID,
			mysqlMetricsConfig.Deployment,
			mysqlMetricsConfig.JobName,
			mysqlMetricsConfig.JobIndex,
			mysqlMetricsConfig.JobIP,
			loggerWrapper,
		)
		senders = append(senders, promSender)
		metricsLogger.Info(fmt.Sprintf("Prometheus metrics server started on port %d", mysqlMetricsConfig.PrometheusPort))
	}

	// 2. Legacy Loggregator Sender (Active if enabled in config and certs exist)
	if mysqlMetricsConfig.EnableLoggregatorEmitter && mysqlMetricsConfig.LoggregatorCAPath != "" {
		tlsConfig, err := loggregator.NewIngressTLSConfig(
			mysqlMetricsConfig.LoggregatorCAPath,
			mysqlMetricsConfig.LoggregatorClientCertPath,
			mysqlMetricsConfig.LoggregatorClientKeyPath,
		)
		if err == nil {
			ingressClient, err := loggregator.NewIngressClient(
				tlsConfig,
				loggregator.WithAddr("localhost:3458"),
				loggregator.WithTag("source_id", mysqlMetricsConfig.SourceID),
				loggregator.WithTag("origin", mysqlMetricsConfig.Origin),
			)
			if err == nil {
				senders = append(senders, metrics.NewLoggregatorSender(ingressClient, mysqlMetricsConfig.SourceID))
				metricsLogger.Info("Loggregator metrics sender started")
			} else {
				metricsLogger.Error("loggregator client failed to initialize", err)
			}
		} else {
			metricsLogger.Error("loggregator tls config failed to initialize", err)
		}
	}

	if len(senders) == 0 {
		panic("no metrics sender enabled or initialized")
	}

	multiSender := metrics.NewMultiSender(senders...)
	defer func() {
		if closer, ok := interface{}(multiSender).(io.Closer); ok {
			_ = closer.Close()
		}
	}()

	conn := Connection(mysqlMetricsConfig)
	dbClient := database_client.NewDatabaseClient(conn, mysqlMetricsConfig)
	stater := disk.NewInfo(syscall.Statfs)
	procStatFile, err := os.Open("/proc/stat")
	if err != nil {
		metricsLogger.Error("failed to open /proc/stat", err)
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
	metricsWriter := metrics.NewMetricWriter(multiSender, loggerWrapper, mysqlMetricsConfig.Origin)
	processor := metrics.NewProcessor(gatherer, metricsComputer, metricsWriter, mysqlMetricsConfig)
	metricsInterval := time.Duration(mysqlMetricsConfig.MetricsFrequency) * time.Second
	emitter := emit.NewEmitter(processor, metricsInterval, time.Sleep, loggerWrapper)

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

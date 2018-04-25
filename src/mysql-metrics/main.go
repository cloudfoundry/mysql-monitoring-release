package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	"syscall"

	"path/filepath"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/dropsonde"
	"mysql-metrics/config"
	"mysql-metrics/cpu"
	"mysql-metrics/database_client"
	"mysql-metrics/disk"
	"mysql-metrics/emit"
	"mysql-metrics/gather"
	"mysql-metrics/metrics"
	"mysql-metrics/metrics_computer"
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

func (d lagerLoggerWrapper) Info(action string) {
	d.logger.Info(action)
}

func (d lagerLoggerWrapper) Error(action string, err error) {
	d.logger.Error(action, err)
}

var configFilepath = flag.String("c", defaultConfigPath, "location of config file")
var logFilepath = flag.String("l", "", "location of log file")

func setupLogging(metricsLogger lager.Logger) {
	if *logFilepath != "" {
		file, err := os.OpenFile(*logFilepath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
		if err != nil {
			panic(fmt.Sprintf("Could not open log file: %s\n", err))
		}

		sink := lager.NewWriterSink(file, lager.DEBUG)
		metricsLogger.RegisterSink(sink)
	}
}

func main() {
	flag.Parse()

	metricsLogger := lager.NewLogger("MetricsLogger")
	sink := lager.NewWriterSink(os.Stdout, lager.DEBUG)
	metricsLogger.RegisterSink(sink)

	setupLogging(metricsLogger)

	mysqlMetricsConfig := &config.Config{}
	if err := config.LoadFromFile(*configFilepath, mysqlMetricsConfig); err != nil {
		metricsLogger.Error("config file is not formatted correctly", err)
		panic(err)
	}

	metricMappingConfig := metrics.DefaultMetricMappingConfig()

	if err := dropsonde.Initialize("localhost:3457", mysqlMetricsConfig.Origin); err != nil {
		metricsLogger.Error("dropsonde failed to initialize", err)
		panic(err)
	}

	conn := Connection(mysqlMetricsConfig)
	dbClient := database_client.NewDatabaseClient(conn, mysqlMetricsConfig)
	stater := disk.NewInfo(syscall.Statfs)
	procStatFile, err := os.Open("/proc/stat")
	if err != nil {
		metricsLogger.Error("failed to open /proc/stat", err)
		panic(err)
	}
	cpustater := cpu.New(procStatFile)
	gatherer := gather.NewGatherer(dbClient, stater, &cpustater)

	loggerWrapper := lagerLoggerWrapper{metricsLogger}
	metricsComputer := metrics_computer.NewMetricsComputer(*metricMappingConfig)
	metricsWriter := metrics.NewMetricWriter(new(metrics.DropsondeSender), loggerWrapper, mysqlMetricsConfig.Origin)
	processor := metrics.NewProcessor(gatherer, metricsComputer, metricsWriter, mysqlMetricsConfig)
	metricsInterval := time.Duration(mysqlMetricsConfig.MetricsFrequency) * time.Second
	emitter := emit.NewEmitter(processor, metricsInterval, time.Sleep, loggerWrapper)

	emitter.Start()
}

func Connection(config *config.Config) *sql.DB {
	var dsn string

	if filepath.IsAbs(config.Host) {
		dsn = fmt.Sprintf("%s:%s@unix(%s)/", config.Username, config.Password, config.Host)
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/", config.Username, config.Password, config.Host, 3306)
	}

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(fmt.Sprintf("Database configuration problem: %v", err))
	}

	return conn
}

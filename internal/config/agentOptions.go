package config

import (
	"flag"
)

const (
	defaultReportInterval = 5
	defaultPollInterval   = 2
)

type AgentOptions struct {
	Host           string
	PollInterval   int
	ReportInterval int
}

func NewAgentOptions() *AgentOptions {
	hostFlag := flag.String("a", defaultHost, "Хост")
	pollIntervalFlag := flag.Int("p", defaultPollInterval, "Частота опроса метрик из пакета runtime (в секундах)")
	reportIntervalFlag := flag.Int("r", defaultReportInterval, "Частота отправки метрик на сервер (в секундах)")
	flag.Parse()

	host := *hostFlag
	pollInterval := *pollIntervalFlag
	reportInterval := *reportIntervalFlag

	host = getEnvString("ADDRESS", host)
	pollInterval = getEnvInt("POLL_INTERVAL", pollInterval)
	reportInterval = getEnvInt("REPORT_INTERVAL", reportInterval)

	return &AgentOptions{
		Host:           host,
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
	}
}

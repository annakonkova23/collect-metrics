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
	Key            string
}

func NewAgentOptions() *AgentOptions {
	hostFlag := flag.String("a", defaultHost, "Хост")
	pollIntervalFlag := flag.Int("p", defaultPollInterval, "Частота опроса метрик из пакета runtime (в секундах)")
	reportIntervalFlag := flag.Int("r", defaultReportInterval, "Частота отправки метрик на сервер (в секундах)")
	keyFlag := flag.String("k", "key", "Ключ для хеша")
	flag.Parse()

	host := *hostFlag
	pollInterval := *pollIntervalFlag
	reportInterval := *reportIntervalFlag
	key := *keyFlag

	host = getEnvString("ADDRESS", host)
	pollInterval = getEnvInt("POLL_INTERVAL", pollInterval)
	reportInterval = getEnvInt("REPORT_INTERVAL", reportInterval)
	key = getEnvString("KEY", key)

	return &AgentOptions{
		Host:           host,
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		Key:            key,
	}
}

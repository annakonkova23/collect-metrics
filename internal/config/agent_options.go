package config

import (
	"flag"
)

const (
	defaultReportInterval = 5
	defaultPollInterval   = 2
	defaultRateLimiter    = 3
)

type AgentOptions struct {
	Host           string
	PollInterval   int
	ReportInterval int
	Key            string
	RateLimiter    int
	KeyPath        string
}

// NewAgentOptions создает новый объект AgentOptions.
func NewAgentOptions() *AgentOptions {
	hostFlag := flag.String("a", defaultHost, "Хост")
	pollIntervalFlag := flag.Int("p", defaultPollInterval, "Частота опроса метрик из пакета runtime (в секундах)")
	reportIntervalFlag := flag.Int("r", defaultReportInterval, "Частота отправки метрик на сервер (в секундах)")
	keyFlag := flag.String("k", defaultKey, "Ключ для хеша")
	rateLimiterFlag := flag.Int("l", defaultRateLimiter, "Rate limiter")
	keyPathFlag := flag.String("crypto-key", "", "Путь до публичного ключа")
	flag.Parse()

	host := *hostFlag
	pollInterval := *pollIntervalFlag
	reportInterval := *reportIntervalFlag
	key := *keyFlag
	rateLimiter := *rateLimiterFlag
	keyPath := *keyPathFlag

	host = getEnvString("ADDRESS", host)
	pollInterval = getEnvInt("POLL_INTERVAL", pollInterval)
	reportInterval = getEnvInt("REPORT_INTERVAL", reportInterval)
	key = getEnvString("KEY", key)
	rateLimiter = getEnvInt("RATE_LIMIT", rateLimiter)
	keyPath = getEnvString("CRYPTO_KEY", keyPath)

	return &AgentOptions{
		Host:           host,
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		Key:            key,
		RateLimiter:    rateLimiter,
		KeyPath:        keyPath,
	}
}

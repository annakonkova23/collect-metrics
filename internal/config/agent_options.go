package config

import (
	"flag"
	"log"
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
	FileConfig     string
}

// NewAgentOptions создает новый объект AgentOptions.
func NewAgentOptions() *AgentOptions {
	agentOptions := &AgentOptions{}

	agentOptionsFlag := getAgentOptionsFromFlag()

	agentOptionsEnv := getAgentOptionsFromEnv(agentOptionsFlag)

	agentOptions = agentOptionsEnv

	if agentOptions.FileConfig != "" {
		so, err := getAgentOptionsFromFile(agentOptions.FileConfig)
		if err != nil {
			log.Printf("Не удалось считать конфигурацию из файла %s: %s", agentOptions.FileConfig, err.Error())
		}
		agentOptions.CompareAndAddValues(so)
	}

	return agentOptions
}

func getAgentOptionsFromEnv(def *AgentOptions) *AgentOptions {

	agentOptions := &AgentOptions{}

	agentOptions.Host = getEnvString("ADDRESS", def.Host)
	agentOptions.PollInterval = getEnvInt("POLL_INTERVAL", def.PollInterval)
	agentOptions.ReportInterval = getEnvInt("REPORT_INTERVAL", def.ReportInterval)
	agentOptions.Key = getEnvString("KEY", def.Key)
	agentOptions.RateLimiter = getEnvInt("RATE_LIMIT", def.RateLimiter)
	agentOptions.KeyPath = getEnvString("CRYPTO_KEY", def.KeyPath)
	agentOptions.FileConfig = getEnvString("CONFIG", def.FileConfig)

	return agentOptions
}

func getAgentOptionsFromFile(path string) (*AgentOptions, error) {
	config, err := ReadJSONConfig(path)
	if err != nil {
		return nil, err
	}

	ao := &AgentOptions{}

	for key, value := range config {
		switch key {
		case "address":
			ao.Host = value.(string)
		case "poll_interval":
			ao.PollInterval = int(value.(float64))
		case "report_interval":
			ao.ReportInterval = int(value.(float64))
		case "rate_limit":
			ao.RateLimiter = int(value.(float64))
		case "crypto_key":
			ao.KeyPath = value.(string)

		}
	}

	return ao, nil
}

func getAgentOptionsFromFlag() *AgentOptions {

	hostFlag := flag.String("a", "", "Хост")
	pollIntervalFlag := flag.Int("p", 0, "Частота опроса метрик из пакета runtime (в секундах)")
	reportIntervalFlag := flag.Int("r", 0, "Частота отправки метрик на сервер (в секундах)")
	keyFlag := flag.String("k", "", "Ключ для хеша")
	rateLimiterFlag := flag.Int("l", 0, "Rate limiter")
	keyPathFlag := flag.String("crypto-key", "", "Путь до публичного ключа")
	fileConfigFlag := flag.String("c", "server.json", "Путь до файла конфигурации")
	fileConfigFlag = flag.String("config", *fileConfigFlag, "Путь до файла конфигурации")
	flag.Parse()

	ao := &AgentOptions{}
	ao.Host = *hostFlag
	ao.PollInterval = *pollIntervalFlag
	ao.ReportInterval = *reportIntervalFlag
	ao.Key = *keyFlag
	ao.RateLimiter = *rateLimiterFlag
	ao.KeyPath = *keyPathFlag
	ao.FileConfig = *fileConfigFlag

	return ao
}

func (ao *AgentOptions) CompareAndAddValues(trg *AgentOptions) {
	if ao == nil || trg == nil {
		return
	}

	if ao.Host == "" {
		ao.Host = trg.Host
	}

	if ao.Key == "" {
		ao.Key = trg.Key
	}

	if ao.KeyPath == "" {
		ao.KeyPath = trg.KeyPath
	}

	if ao.PollInterval == 0 {
		ao.PollInterval = trg.PollInterval
	}
	if ao.ReportInterval == 0 {
		ao.ReportInterval = trg.ReportInterval
	}
	if ao.RateLimiter == 0 {
		ao.RateLimiter = trg.RateLimiter
	}

}

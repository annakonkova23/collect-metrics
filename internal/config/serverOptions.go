package config

import (
	"flag"
	"os"
	"strconv"
)

const (
	defaultHost            = "localhost:8080"
	defaultStoreInterval   = 300
	defaultRestore         = false
	defaultFileStoragePath = "metrics.json"
)

type ServerOptions struct {
	Host            string
	StoreInterval   int
	Restore         bool
	FileStoragePath string
}

func NewServerOptions() *ServerOptions {
	host := flag.String("a", defaultHost, "Хост")
	storeInterval := flag.Int("i", defaultStoreInterval, "Интервал времени, по истечении которого текущие показания сервера сохраняются на диск (в секундах)")
	restore := flag.Bool("r", defaultRestore, "Загружать ранее сохранённые значения из указанного файла при старте сервера")
	fileStoragePath := flag.String("f", defaultFileStoragePath, "Путь к файлу, в котором сохраняются значения показателей")
	flag.Parse()
	if envHost := os.Getenv("ADDRESS"); envHost != "" {
		host = &envHost
	}
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		if storeIntervalInt, err := strconv.Atoi(envStoreInterval); err == nil {
			storeInterval = &storeIntervalInt
		}
	}
	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		if restoreBool, err := strconv.ParseBool(envRestore); err == nil {
			restore = &restoreBool
		}

	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		fileStoragePath = &envFileStoragePath
	}
	return &ServerOptions{
		Host:            *host,
		StoreInterval:   *storeInterval,
		Restore:         *restore,
		FileStoragePath: *fileStoragePath,
	}

}

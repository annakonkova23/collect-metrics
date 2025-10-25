package config

import (
	"flag"
	"os"
	"strconv"
)

const (
	defaultHost            = "localhost:8080"
	defaultStoreInterval   = 300
	defaultRestore         = true
	defaultFileStoragePath = "metrics.json"
	defaultDBDSN           = "host=localhost port=5432 user=postgres password=anna dbname=metricsdb sslmode=disable"
)

type ServerOptions struct {
	Host            string
	StoreInterval   int
	Restore         bool
	FileStoragePath string
	DatabaseDSN     string
}

func getEnvString(envKey, defaultValue string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(envKey string, defaultValue int) int {
	if v := os.Getenv(envKey); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvBool(envKey string, defaultValue bool) bool {
	if v := os.Getenv(envKey); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultValue
}

func NewServerOptions() *ServerOptions {
	hostFlag := flag.String("a", defaultHost, "Хост")
	storeIntervalFlag := flag.Int("i", defaultStoreInterval, "Интервал сохранения данных (сек)")
	restoreFlag := flag.Bool("r", defaultRestore, "Восстанавливать данные из файла")
	fileStoragePathFlag := flag.String("f", defaultFileStoragePath, "Путь к файлу хранения")
	databaseDSNFlag := flag.String("d", defaultDBDSN, "Aдрес подключения к БД")

	flag.Parse()

	host := *hostFlag
	storeInterval := *storeIntervalFlag
	restore := *restoreFlag
	fileStoragePath := *fileStoragePathFlag
	databaseDSN := *databaseDSNFlag

	host = getEnvString("ADDRESS", host)
	storeInterval = getEnvInt("STORE_INTERVAL", storeInterval)
	restore = getEnvBool("RESTORE", restore)
	fileStoragePath = getEnvString("FILE_STORAGE_PATH", fileStoragePath)
	databaseDSN = getEnvString("DATABASE_DSN", databaseDSN)

	return &ServerOptions{
		Host:            host,
		StoreInterval:   storeInterval,
		Restore:         restore,
		FileStoragePath: fileStoragePath,
		DatabaseDSN:     databaseDSN,
	}

}

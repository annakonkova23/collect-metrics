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
	defaultFileStoragePath = ""
	//defaultFileStoragePath = "metrics.json"
	//defaultDBDSN = ""
	//defaultDBDSN      = "postgres://postgres:anna@localhost:5432/metricsdb?sslmode=disable" //"host=localhost port=5432 user=postgres password=anna dbname=metricsdb sslmode=disable"
	defaultKey        = "key"
	defaultBufferSize = 100
)

type ServerOptions struct {
	Host            string
	StoreInterval   int
	Restore         bool
	FileStoragePath string
	DatabaseDSN     string
	Key             string
	AuditFilePath   string
	AuditURL        string
	BufferSize      int
}

func getEnvString(envKey, defaultValue string) string {
	if v, exists := os.LookupEnv(envKey); exists {
		return v
	}
	return defaultValue
}

func getEnvInt(envKey string, defaultValue int) int {
	if v, exists := os.LookupEnv(envKey); exists {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvBool(envKey string, defaultValue bool) bool {
	if v, exists := os.LookupEnv(envKey); exists {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultValue
}

// NewServerOptions создает новый экземпляр ServerOptions с значениями по умолчанию.
func NewServerOptions() *ServerOptions {
	hostFlag := flag.String("a", defaultHost, "Хост")
	storeIntervalFlag := flag.Int("i", defaultStoreInterval, "Интервал сохранения данных (сек)")
	restoreFlag := flag.Bool("r", defaultRestore, "Восстанавливать данные из файла")
	fileStoragePathFlag := flag.String("f", defaultFileStoragePath, "Путь к файлу хранения")
	databaseDSNFlag := flag.String("d", "", "Aдрес подключения к БД")
	keyFlag := flag.String("k", defaultKey, "Ключ для хеша")
	auditFileFlag := flag.String("audit-file", "", "Путь к файлу аудита")
	auditURLflag := flag.String("audit-url", "", "URL для аудита")
	bufSizeflag := flag.Int("b", defaultBufferSize, "Размер буфера каналов")
	flag.Parse()

	host := *hostFlag
	storeInterval := *storeIntervalFlag
	restore := *restoreFlag
	fileStoragePath := *fileStoragePathFlag
	databaseDSN := *databaseDSNFlag
	key := *keyFlag
	auditFile := *auditFileFlag
	auditURL := *auditURLflag
	bufSize := *bufSizeflag

	host = getEnvString("ADDRESS", host)
	storeInterval = getEnvInt("STORE_INTERVAL", storeInterval)
	restore = getEnvBool("RESTORE", restore)
	fileStoragePath = getEnvString("FILE_STORAGE_PATH", fileStoragePath)
	databaseDSN = getEnvString("DATABASE_DSN", databaseDSN)
	key = getEnvString("KEY", key)
	auditFile = getEnvString("AUDIT_FILE", auditFile)
	auditURL = getEnvString("AUDIT_URL", auditURL)
	bufSize = getEnvInt("BUFFER_SIZE", bufSize)

	return &ServerOptions{
		Host:            host,
		StoreInterval:   storeInterval,
		Restore:         restore,
		FileStoragePath: fileStoragePath,
		DatabaseDSN:     databaseDSN,
		Key:             key,
		AuditFilePath:   auditFile,
		AuditURL:        auditURL,
		BufferSize:      bufSize,
	}

}

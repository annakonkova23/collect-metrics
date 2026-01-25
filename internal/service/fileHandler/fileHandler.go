// Package filehandler предоставляет загрузку данных в/из файла
// Особенности:
//   - Гарантирует атомарность записи: либо файл полностью обновлён, либо остался прежним
//   - Предотвращает повреждение данных при сбоях (падение процесса, нехватка места и т.д.)
//   - Поддерживает создание промежуточных директорий при необходимости
//   - Использует sync.RWMutex для потокобезопасного доступа
//   - Логирует операции с помощью zap
//
// Пример:
//
//	logger, _ := zap.NewProduction()
//	defer logger.Sync()
//
//	handler, err := filehandler.NewFileHandler("/data/metrics.json", logger)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	data := []byte(`{"gauge": 3.14, "counter": 42}`)
//	if err := handler.SaveToFile(data); err != nil {
//	    log.Printf("Ошибка сохранения: %v", err)
//	}
//
//	restored := handler.LoadFromFile()
package filehandler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// generate:reset
type FileHandler struct {
	mx       sync.RWMutex
	filePath string
	logger   *zap.Logger
}

// NewFileHandler создаёт новый экземпляр FileHandler для указанного пути.
//
// Параметры:
//   - filePath: путь к файлу (может быть абсолютным или относительным)
//   - logger: экземпляр *zap.Logger для логирования операций
//
// Возвращает:
//   - Указатель на *FileHandler и nil, если аргументы валидны
//   - nil и ошибку, если filePath пустой
//
// Примечание: файл не создаётся при инициализации. Он будет создан при первом вызове SaveToFile.
func NewFileHandler(filePath string, logger *zap.Logger) (*FileHandler, error) {
	if filePath == "" {
		return nil, fmt.Errorf("путь к файлу не может быть пустым")
	}
	return &FileHandler{
		filePath: filePath,
		logger:   logger,
	}, nil
}

// createTempFileInSameDirectory создаёт временный файл в той же директории, что и целевой.
//
// Поведение:
//   - Если путь абсолютный или содержит разделитель — временный файл создаётся в родительской директории.
//   - Если путь относительный — используется текущая рабочая директория (os.Getwd).
//   - Директория создаётся, если не существует (с правами 0755).
//
// Временный файл имеет шаблон: <basename>.tmp*
//
// Возвращает *os.File или ошибку.
func (fh *FileHandler) createTempFileInSameDirectory(targetPath string) (*os.File, error) {
	var dir string
	if filepath.IsAbs(targetPath) || strings.Contains(targetPath, string(os.PathSeparator)) {
		dir = filepath.Dir(targetPath)
	} else {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("не удалось получить текущую директорию: %w", err)
		}
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("ошибка создания директории: %w", err)
	}

	base := filepath.Base(targetPath)
	tmpFile, err := os.CreateTemp(dir, base+".tmp")
	if err != nil {
		return nil, fmt.Errorf("ошибка создания временного файла: %w", err)
	}

	return tmpFile, nil
}

// SaveToFile атомарно сохраняет данные в файл.
//
// Стратегия:
//   - Данные сначала записываются во временный файл в той же директории
//   - После успешной записи — временный файл переименовывается в целевой (os.Rename)
//   - Это гарантирует, что файл всегда в целостном состоянии
//
// Логирование:
//   - При успехе: записывается Info-лог "Файл сохранён"
//   - Ошибки детально оборачиваются с контекстом
//
// Возвращает nil при успехе, иначе ошибку.
func (fh *FileHandler) SaveToFile(data []byte) error {
	fh.mx.Lock()
	defer fh.mx.Unlock()

	tmpFile, err := fh.createTempFileInSameDirectory(fh.filePath)
	if err != nil {
		return fmt.Errorf("ошибка создания временного файла: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("ошибка записи во временный файл: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("ошибка закрытия временного файла: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), fh.filePath); err != nil {
		return fmt.Errorf("ошибка переименования файла: %w", err)
	}

	fh.logger.Info("Файл сохранён", zap.String("filename", fh.filePath))
	return nil
}

// LoadFromFile читает данные из файла.
//
// Поведение:
//   - Если файл не существует — возвращается nil, ошибки нет (логируется Info)
//   - Если файл существует и доступен — содержимое считывается целиком
//
// Возвращает:
//   - []byte с содержимым файла, если удалось прочитать
//   - nil, если файл не найден или ошибка (в последнем случае ошибка логируется)
//
// Примечание: вызывающий код должен проверять, что результат не nil, если данные обязательны.
func (fh *FileHandler) LoadFromFile() []byte {
	res, err := os.ReadFile(fh.filePath)
	if err != nil {
		fh.logger.Info("Файл не найден", zap.String("filename", fh.filePath))
		return nil
	}
	return res
}

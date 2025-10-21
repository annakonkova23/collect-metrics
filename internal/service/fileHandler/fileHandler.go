package fileHandler

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileHandler struct {
	mx       sync.Mutex
	filePath string
	logger   *zap.Logger
}

func NewFileHandler(filePath string, logger *zap.Logger) *FileHandler {
	return &FileHandler{
		filePath: filePath,
		logger:   logger,
	}
}

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

func (fh *FileHandler) SaveToFile(data []byte) error {
	fh.mx.Lock()
	defer fh.mx.Unlock()

	tmpFile, err := fh.createTempFileInSameDirectory(fh.filePath)
	if err != nil {
		return fmt.Errorf("ошибка создания временного файла: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Удаляем, если что-то пошло не так

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

func (fh *FileHandler) LoadFromFile() []byte {
	res, err := os.ReadFile(fh.filePath)
	if err != nil {
		fh.logger.Info("Файл не найден", zap.String("filename", fh.filePath))
		return nil
	}
	return res
}

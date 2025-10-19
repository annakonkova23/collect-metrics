package service

import (
	"encoding/json"
	"fmt"
	"github.com/annakonkova23/collect-metrics/internal/model"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
)

func (c *Collector) createTempFileInSameDirectory(targetPath string) (*os.File, error) {

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

func (c *Collector) SaveToFile() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	metrics := c.MemStorage.GetMetricAllValues()
	js, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	tmpFile, err := c.createTempFileInSameDirectory(c.FileStoragePath)
	if err != nil {
		return fmt.Errorf("ошибка создания временного файла: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Удаляем, если что-то пошло не так

	if _, err := tmpFile.Write(js); err != nil {
		return fmt.Errorf("ошибка записи во временный файл: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("ошибка закрытия временного файла: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), c.FileStoragePath); err != nil {
		return fmt.Errorf("ошибка переименования файла: %w", err)
	}

	c.logger.Info("Файл сохранён", zap.String("filename", c.FileStoragePath))
	return nil
}

func (c *Collector) LoadFromFile() []*model.Metrics {
	res, err := os.ReadFile(c.FileStoragePath)
	if err != nil {
		c.logger.Info("Файл не найден", zap.String("filename", c.FileStoragePath))
		return nil
	}
	metrics := []*model.Metrics{}
	err = json.Unmarshal(res, &metrics)
	if err != nil {
		c.logger.Info("Ошибка парсинга метрик:", zap.String("dataFile", string(res)))
		return nil
	}
	c.logger.Info("Данные из файла загружены", zap.String("filename", c.FileStoragePath))
	return metrics
}

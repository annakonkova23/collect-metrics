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

func createTempFileInSameDirectory(targetPath string) (*os.File, error) {
	// 1. Определяем директорию для целевого файла
	var dir string
	if filepath.IsAbs(targetPath) || strings.Contains(targetPath, string(os.PathSeparator)) {
		dir = filepath.Dir(targetPath)
	} else {
		// Используем текущую рабочую директорию
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("не удалось получить текущую директорию: %w", err)
		}
	}

	// 2. Проверяем и создаём директорию, если её нет
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("ошибка создания директории: %w", err)
	}

	// 3. Получаем базовое имя файла (без пути)
	base := filepath.Base(targetPath)

	// 4. Создаём временный файл в выбранной директории
	tmpFile, err := os.CreateTemp(dir, base+".tmp")
	if err != nil {
		return nil, fmt.Errorf("ошибка создания временного файла: %w", err)
	}

	return tmpFile, nil
}

func (c *Collector) SaveToFile() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Получаем данные и сериализуем их
	metrics := c.MemStorage.GetMetricAllValues()
	js, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	// 2. Создаём временный файл
	tmpFile, err := createTempFileInSameDirectory(c.FileStoragePath)
	if err != nil {
		return fmt.Errorf("ошибка создания временного файла: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Удаляем, если что-то пошло не так

	// 3. Записываем данные во временный файл
	if _, err := tmpFile.Write(js); err != nil {
		return fmt.Errorf("ошибка записи во временный файл: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("ошибка закрытия временного файла: %w", err)
	}

	// 4. Атомарно переименовываем временный файл в целевой
	if err := os.Rename(tmpFile.Name(), c.FileStoragePath); err != nil {
		return fmt.Errorf("ошибка переименования файла: %w", err)
	}

	c.logger.Info("Файл сохранён", zap.String("filename", c.FileStoragePath))
	return nil
}

func (c *Collector) LoadFromFile() ([]*model.Metrics, error) {
	res, err := os.ReadFile(c.FileStoragePath)
	if err != nil {
		return nil, err
	}
	metrics := []*model.Metrics{}
	err = json.Unmarshal(res, &metrics)
	if err != nil {
		return nil, err
	}
	c.logger.Info("Данные из файла загружены", zap.String("filename", c.FileStoragePath))
	return metrics, nil
}

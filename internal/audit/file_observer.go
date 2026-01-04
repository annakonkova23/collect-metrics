package audit

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"sync"

	"go.uber.org/zap"
)

// FileObserver — наблюдатель, записывающий аудит-события в файл.
type FileObserver struct {
	mx     sync.RWMutex
	f      *os.File
	w      *bufio.Writer
	logger *zap.Logger
}

// NewFileObserver создаёт новый наблюдатель, записывающий аудит-события в файл.
//
// Наблюдатель открывает файл по указанному пути с правами на дозапись.
// Используется буферизованный записыватель с размером буфера 64 КБ для
// уменьшения количества системных вызовов записи.
//
// Параметры:
//   - logger: логгер для записи внутренних ошибок и событий наблюдателя.
//   - path: путь к файлу, в который будут записываться события.
//     Файл будет создан, если не существует; открыт в режиме дозаписи.
//
// Возвращает указатель на *FileObserver и ошибку, если:
//   - нет прав на создание или запись в указанный путь;
//   - файл недоступен (например, занят другим процессом).
//
// После создания наблюдатель должен быть подписан на публикатор событий.
// Буфер сбрасывается при закрытии наблюдателя (Close) или при явном вызове Flush.
func NewFileObserver(logger *zap.Logger, path string) (*FileObserver, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &FileObserver{
		f:      f,
		w:      bufio.NewWriterSize(f, 64*1024),
		logger: logger,
	}, nil
}

// Notify дописывает событие в конец файла (append).
func (fo *FileObserver) Notify(_ context.Context, e Event) {
	fo.mx.Lock()
	defer fo.mx.Unlock()

	b, err := json.Marshal(e)
	if err != nil {
		fo.logger.Error("Ошибка json.Marshal", zap.Error(err))
		return
	}

	_, err = fo.w.Write(b)
	if err != nil {
		fo.logger.Error("Ошибка записи", zap.Error(err))
		return
	}
	err = fo.w.WriteByte('\n')
	if err != nil {
		fo.logger.Error("Ошибка записи", zap.Error(err))
		return
	}

	err = fo.w.Flush()
	if err != nil {
		fo.logger.Error("Ошибка записи", zap.Error(err))
		return
	}
}

// Close закрывает файл.
func (fo *FileObserver) Close() error {
	fo.mx.Lock()
	defer fo.mx.Unlock()

	err := fo.w.Flush()
	if err != nil {
		fo.logger.Error("Ошибка записи", zap.Error(err))
	}
	return fo.f.Close()
}

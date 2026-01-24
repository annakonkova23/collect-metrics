package audit

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

type FileObserver struct {
	mx     sync.RWMutex
	f      *os.File
	w      *bufio.Writer
	logger *zap.Logger
}

// path — путь к файлу аудита, например "./audit.log"
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
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

	fmt.Println("Отправка запроса в файл!")
	b, err := json.Marshal(e)
	if err != nil {
		// если маршалинг не удался — просто пропускаем, чтобы не падать в аудите
		return
	}

	// JSONL: событие + \n
	_, _ = fo.w.Write(b)
	_ = fo.w.WriteByte('\n')

	// чтобы точно было записано сразу (можно убрать ради производительности и делать Flush по таймеру)
	_ = fo.w.Flush()
}

func (fo *FileObserver) Close() error {
	fo.mx.Lock()
	defer fo.mx.Unlock()

	_ = fo.w.Flush()
	return fo.f.Close()
}

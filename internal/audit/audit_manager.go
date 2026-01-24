// Пакет работу с аудитом.
package audit

import (
	"context"

	"go.uber.org/zap"
)

// AuditManager — управление аудиом.
type AuditManager struct {
	pub     *Publisher
	fileObs *FileObserver
	httpObs *HTTPObserver
	logger  *zap.Logger
}

// NewAuditManager создаёт новый экземпляр AuditManager.
//
// AuditManager координирует сбор аудит-событий и их отправку в наблюдателей
// (например, файл или HTTP-эндпоинт).
//
// Параметры:
//   - logger: логгер для записи внутренних событий менеджера аудита.
//   - bufferSize: размер буфера для канала публикации событий; если 0, используется значение по умолчанию (100).
//   - path: путь к файлу для записи аудита. Если пусто — файловый вывод отключается.
//   - url: URL для отправки аудит-событий по HTTP. Если пусто — HTTP-вывод отключается.
//
// Возвращает указатель на *AuditManager и ошибку, если:
//   - не удалось создать файловый наблюдатель (например, файл не существует).
//   - URL некорректен (в случае HTTP-наблюдателя).
//
// После создания необходимо вызвать Start() для запуска обработки сообщений.
func NewAuditManager(logger *zap.Logger, bufferSize int, path string, url string) (*AuditManager, error) {
	pub := NewPublisher(bufferSize)
	auditManager := &AuditManager{
		pub:    pub,
		logger: logger,
	}
	if path != "" {
		file, err := NewFileObserver(logger, path)
		if err != nil {
			return nil, err
		}
		auditManager.pub.Subscribe(file)
		auditManager.fileObs = file
	}
	if url != "" {
		http := NewHTTPObserver(logger, url, bufferSize)
		auditManager.pub.Subscribe(http)
		auditManager.httpObs = http
	}

	return auditManager, nil

}

// Start запускает обработку событий аудита.
func (m *AuditManager) Start(ctx context.Context) {
	m.logger.Info("Запуск процессов аудирования")
	m.pub.Start(ctx)
	if m.httpObs != nil {
		m.httpObs.Start(ctx)
	}
}

// Close закрывает все открытые файлы и останавливает наблюдателей.
func (m *AuditManager) Close() error {
	if m.fileObs != nil {
		return m.fileObs.Close()
	}
	return nil
}

// Publish отправляет событие аудита.
func (m *AuditManager) Publish(ctx context.Context, e Event) {
	m.pub.Publish(ctx, e)
}

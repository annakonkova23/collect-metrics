package audit

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type AuditManager struct {
	pub     *Publisher
	fileObs *FileObserver
	httpObs *HTTPObserver
	logger  *zap.Logger
}

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

func (m *AuditManager) Start(ctx context.Context) {
	m.logger.Info("Запуск процессов аудирования")
	m.pub.Start(ctx)
	if m.httpObs != nil {
		m.httpObs.Start(ctx)
	}
}

func (m *AuditManager) Close() error {
	if m.fileObs != nil {
		return m.fileObs.Close()
	}
	return nil
}

func (m *AuditManager) Publish(ctx context.Context, e Event) {
	fmt.Println("!!!!Вызов Publish")
	m.pub.Publish(ctx, e)

}

package nuwa

import (
	"context"

	"github.com/silen/nuwa/pkg/logs"
	"github.com/sirupsen/logrus"
)

type Model struct {
	ctx  context.Context
	Logs *logrus.Entry
}

func (m *Model) WithContent(ctx context.Context) {
	m.ctx = ctx
	m.Logs = logs.WithContext(ctx)
}

func (m *Model) GetCtx() context.Context {
	if m.ctx == nil {
		m.ctx = context.TODO()
	}
	return m.ctx
}

func (m *Model) GetLogs() *logrus.Entry {
	if m.Logs == nil {
		m.Logs = logs.WithContext(m.GetCtx())
	}
	return m.Logs
}

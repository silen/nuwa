package nuwa

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	"github.com/silen/nuwa/pkg/conf"
	"github.com/silen/nuwa/pkg/logs"
)

type Service struct {
	ctx  context.Context
	Logs *logrus.Entry
}

func (s *Service) WithContent(ctx context.Context) {
	s.ctx = ctx
	s.Logs = logs.WithContext(ctx)
	//执行userInfo的转化
}

func (s *Service) GetLogs() *logrus.Entry {
	if s.Logs == nil {
		s.Logs = logs.WithContext(s.GetCtx())
	}
	return s.Logs
}

func (s *Service) GetCtx() context.Context {
	if s.ctx == nil {
		s.ctx = context.TODO()
	}
	return s.ctx
}

func (s *Service) Result(data any) *conf.ReturnMap {
	return &conf.ReturnMap{
		Data:    data,
		Status:  SUCCESSFUL,
		Message: "ok",
	}
}

func (s *Service) Message(message string, args ...any) *conf.ReturnMap {
	status := EXCEPTION
	if len(args) > 0 {
		status = cast.ToInt(args[0])
	}
	return &conf.ReturnMap{
		Data:    make([]string, 0),
		Status:  status,
		Message: message,
	}

}

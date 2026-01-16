package nuwa

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	"github.com/silen/nuwa/pkg/conf"
	"github.com/silen/nuwa/pkg/logs"
)

type Controller struct {
	ctx  context.Context
	Logs *logrus.Entry
}

func (c *Controller) WithContent(ctx *gin.Context) {
	c.ctx = ctx
	c.Logs = logs.WithContext(ctx)
}

func (c *Controller) GetCtx() (ctx context.Context) {
	if c.ctx == nil {
		c.ctx = context.TODO()
	}
	return
}

func (c *Controller) GetLogs() *logrus.Entry {
	if c.Logs == nil {
		c.Logs = logs.WithContext(c.GetCtx())
	}
	return c.Logs
}

func (c *Controller) ShouldBind(ctx *gin.Context, obj any) (err error) {
	if err = ctx.ShouldBind(obj); err != nil {
		c.Message(ctx, err.Error(), PARAMETER_ERROR)
	}
	return
}

func structToMapByTags(s any) (jsonMap map[string]any) {
	jsonMap = make(map[string]any)
	if jsonBytes, err := json.Marshal(s); err == nil {
		json.Unmarshal(jsonBytes, &jsonMap)
	}
	return
}

func (c *Controller) StructToMap(req any) (rows map[string]string) {
	rows = make(map[string]string)
	for key, value := range structToMapByTags(req) {
		if v, err := cast.ToStringE(value); err == nil {
			v = strings.Trim(v, " ")
			if v != "" {
				rows[key] = v
			}
		} else {
			logs.Error("convert.StructToStringMapByTags----key----"+key, value, err)
		}
	}
	return
}

// 返回有数据结构的json
func (c *Controller) Response(ctx *gin.Context, res *conf.ReturnMap) {
	if res.Status != SUCCESSFUL {
		c.Message(ctx, res.Message, res.Status)
		return
	}
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
	StopRun()
}

// 返回只有消息体的json
func (c *Controller) Message(ctx *gin.Context, message string, args ...any) {
	status := EXCEPTION
	if len(args) > 0 {
		status = cast.ToInt(args[0])
	}
	ctx.JSON(http.StatusOK, map[string]any{
		"status":  status,
		"message": message,
	})
	ctx.Abort()
	StopRun()
}

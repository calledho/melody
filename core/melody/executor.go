package melody

import (
	"context"
	"io"
	"melody/cmd"
	"melody/config"
	"melody/logging"
	gelf "melody/middleware/melody-gelf"
	gologging "melody/middleware/melody-gologging"
	logstash "melody/middleware/melody-logstash"
	"os"
)

//NewExecutor return an new executor
func NewExecutor(ctx context.Context) cmd.Executor {
	return func(cfg config.ServiceConfig) {
		// 1. 确定以及初始化 log有哪些输出
		var writers []io.Writer
		// 1.1 检察是否使用Gelf
		gelfWriter, err := gelf.NewWriter(cfg.ExtraConfig)
		if err == nil {
			writers = append(writers, GelfWriter{gelfWriter})
			gologging.UpdateFormatSelector(func(w io.Writer) string {
				switch w.(type) {
				case GelfWriter:
					return "%{message}"
				default:
					return gologging.DefaultPattern
				}
			})
		}
		// 2.初始化Logger

		// 2.1 是否启用logstash
		logger, enableLogstashError := logstash.NewLogger(cfg.ExtraConfig, writers...)

		if enableLogstashError != nil {
			// 2.2 是否使用gologging
			var enableGologgingError error
			logger, enableGologgingError = gologging.NewLogger(cfg.ExtraConfig, writers...)

			if enableGologgingError != nil {
				// 2.3 默认使用基础Log  Level:Debug, Output:stdout, Prefix: ""
				logger, err := logging.NewLogger("DEBUG", os.Stdout, "")
				if err != nil {
					return
				}
				logger.Error("unable to create gologging logger")
			} else {
				logger.Debug("use gologging as logger")
			}
		} else {
			logger.Debug("use logstash as logger")
		}

		logger.Info("Melody server listening on port:", cfg.Port, "🎁")

		//TODO Start Reporter (目前还不知道这在干什么)

		//TODO 加载插件
		//TODO ...
	}
}

type GelfWriter struct {
	io.Writer
}

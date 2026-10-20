package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-kratos/kratos/contrib/otel/v3/tracing"
	"github.com/go-kratos/kratos/v3/log"
)

// newLogger 构建应用的根 logger。
//
// 只给本项目代码的日志带 source=，依赖库（module cache / GOROOT）的日志不带。
// 项目根目录由本文件自身路径推导（cmd/kratos-todo-service/log.go 上三级），
// 不硬编码绝对路径，项目搬位置也不用改。
func newLogger() *slog.Logger {
	_, selfFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.ToSlash(filepath.Dir(filepath.Dir(filepath.Dir(selfFile)))) + "/"

	return log.NewLogger(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.SourceKey && len(groups) == 0 {
					if src, ok := a.Value.Any().(*slog.Source); ok {
						if !strings.HasPrefix(filepath.ToSlash(src.File), projectRoot) {
							return slog.Attr{}
						}
					}
				}
				return a
			},
		}),
		log.WithExtractor(tracing.TraceAttrs),
	).With(
		slog.String("service.id", id),
		slog.String("service.name", Name),
		slog.String("service.version", Version),
	)
}

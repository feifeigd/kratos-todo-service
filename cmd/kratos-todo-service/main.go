package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"kratos-todo-service/internal/conf"

	"github.com/go-kratos/kratos/contrib/otel/v3/tracing"
	"github.com/go-kratos/kratos/v3"
	"github.com/go-kratos/kratos/v3/config"
	"github.com/go-kratos/kratos/v3/config/env"
	"github.com/go-kratos/kratos/v3/config/file"
	"github.com/go-kratos/kratos/v3/log"
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string

	// Version is the version of the compiled software.
	Version string

	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger *slog.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
		kratos.AfterStop(func(ctx context.Context) error {
			log.InfoContext(ctx, "application stoped.")
			return nil
		}),
	)
}

func main() {
	flag.Parse()

	// 只给本项目代码的日志带 source=，依赖库（module cache / GOROOT）的日志不带。
	// 项目根目录由 main.go 自身路径推导（cmd/kratos-todo-service/main.go 上三级），
	// 不硬编码绝对路径，项目搬位置也不用改。
	_, mainFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.ToSlash(filepath.Dir(filepath.Dir(filepath.Dir(mainFile)))) + "/"

	logger := log.NewLogger(
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

	log.SetDefault(logger)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
			env.NewSource("KRATOS"),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

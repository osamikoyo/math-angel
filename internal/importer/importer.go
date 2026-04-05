package importer

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"unicode"

	// "sync"

	"github.com/osamikoyo/math-angel/internal/config"
	"github.com/osamikoyo/math-angel/internal/service"
	"github.com/osamikoyo/math-angel/pkg/logger"
	"go.uber.org/zap"
)

type Importer struct {
	service *service.Service
	source  *os.File
	logger  *logger.Logger
}

type Task struct {
	Problem  string `json:"problem"`
	Level    string `json:"level"`
	Type     string `json:"type"`
	Solution string `json:"solution"`
	Boxed    string `json:"boxed"`
}

func NewImporter(service *service.Service, cfg *config.Config, logger *logger.Logger) (*Importer, error) {
	file, err := os.Open(cfg.Importer.File)
	if err != nil {
		logger.Error("failed open file with tasks",
			zap.String("path", cfg.Importer.File),
			zap.Error(err))

		return nil, fmt.Errorf("failed open file with tasks: %w", err)
	}

	return &Importer{
		service: service,
		source:  file,
		logger:  logger,
	}, nil
}

func (im *Importer) Start(ctx context.Context) {
	scanners := bufio.NewScanner(im.source)

	// var wg sync.WaitGroup

	for scanners.Scan() {
		select {
		case <-ctx.Done():
			im.logger.Info("stopping importer...")
			return
		default:
			// wg.Go(func() {
			im.logger.Info("scan new line...")

			var task Task

			if err := json.Unmarshal(scanners.Bytes(), &task); err != nil {
				im.logger.Error("failed unmarshal task",
					zap.Error(err))
			}

			im.logger.Info("adding parsed task to db...")

			task.Type = firstToLower(task.Type)

			level := ""

			switch task.Level {
			case "Level 5":
				level = "hard"
			case "Level 4", "Level 3":
				level = "medium"
			case "Level 2", "Level 1":
				level = "easy"
			default:
				im.logger.Warn("unknown task level",
					zap.String("level", task.Level))
				level = "unknown"
			}

			if err := im.service.CreateTask(context.Background(), task.Type, task.Problem, task.Solution, task.Boxed, level); err != nil {
				im.logger.Error("failed create parsed task",
					zap.Any("task", task),
					zap.Error(err))
			}
			//	})
		}
	}

	// wg.Wait()
}

func firstToLower(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return string(append([]rune{unicode.ToLower(r[0])}, r[1:]...))
}

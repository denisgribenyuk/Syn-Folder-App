package main

import (
	"context"
	"fmt"
	"folder_sync/internal/app/sync_folder"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	logFile = "log.txt"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: sync <dir1> <dir2>")
		os.Exit(1)
	}

	srcFolder := os.Args[1]
	dstFolder := os.Args[2]

	// Открываем файл для записи логов
	logFile, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		logrus.Fatalf("failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Создаем логгер
	logger := logrus.New()
	logger.Out = io.MultiWriter(os.Stdout, logFile)
	logger.Level = logrus.DebugLevel
	logger.SetReportCaller(true)
	logger.Formatter = &logrus.TextFormatter{
		ForceColors:   true,
		ForceQuote:    true,
		FullTimestamp: true,
		PadLevelText:  true,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			fileName := path.Base(frame.File) + ":" + strconv.Itoa(frame.Line)
			return "", fileName
		},
	}

	logger.Infof("Starting sync of %s and %s", srcFolder, dstFolder)

	// Создаем контекст для отмены операции
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем WaitGroup для ожидания завершения всех горутин
	var wg sync.WaitGroup

	// Запускаем горутину для синхронизации директорий
	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				logger.Infof("Syncing %s and %s", srcFolder, dstFolder)
				wg.Add(1)
				go func() {
					err := sync_folder.SyncDirs(&wg, logger, srcFolder, dstFolder)
					if err != nil {
						os.Exit(1)
					}
				}()
				time.Sleep(5 * time.Second)
			}
		}
	}()

	// Ожидаем сигнала завершения работы программы
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	cancel()
	wg.Wait()

	logger.Info("Exiting...")
}

package sync_folder

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileInfo struct {
	Name       string
	Size       int64
	ModTime    time.Time
	IsDir      bool
	FullPath   string
	Permissons string
}

func SyncDirs(wg *sync.WaitGroup, logger *log.Logger, src, dst string) error {
	defer wg.Done()

	// Получаем список файлов и директорий в первой директории
	dstFiles, err := GetFiles(dst)
	if err != nil {
		logger.Printf("Failed to get files in %s: %v\n", dst, err)
		return err
	}

	// Получаем список файлов и директорий во второй директории
	srcFiles, err := GetFiles(src)
	if err != nil {
		logger.Printf("Failed to get files in %s: %v\n", src, err)
		return err
	}
	// Добавляем файлы и директории из второй директории в первую, если их нет
	for _, srcFile := range srcFiles {
		if _, ok := dstFiles[srcFile.FullPath]; !ok {
			err = AddFile(src, dst, srcFile, logger)
			if err != nil {
				logger.Printf(err.Error())
				return err
			}
		}
	}

	// Обновляем файлы, если они отличаются по размеру или времени изменения
	for _, dstFile := range dstFiles {
		srcFile, ok := srcFiles[dstFile.FullPath]
		if !ok {
			continue
		}
		err = UpdateFile(src, dst, dstFile, srcFile, logger)
		if err != nil {
			logger.Printf(err.Error())
			return err
		}
	}

	// Удаляем файлы, которых нет во второй директории
	for _, dstFile := range dstFiles {
		if _, ok := srcFiles[dstFile.FullPath]; !ok {
			err = DeleteFile(dst, dstFile, logger)
			if err != nil {
				logger.Printf(err.Error())
				return err
			}
		}
	}
	return nil
}

func GetFiles(dir string) (map[string]FileInfo, error) {
	files := make(map[string]FileInfo)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == dir {
			return nil
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		files[relPath] = FileInfo{
			Name:       info.Name(),
			Size:       info.Size(),
			ModTime:    info.ModTime(),
			IsDir:      info.IsDir(),
			Permissons: info.Mode().String(),
			FullPath:   relPath,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

func AddFile(src string, dst string, file FileInfo, logger *log.Logger) error {
	dstFilePath := filepath.Join(dst, file.FullPath)
	srcFilePath := filepath.Join(src, file.FullPath)

	// Копируем файлы
	if file.IsDir {
		logger.Printf("Creating directory %s\n", dstFilePath)
		if err := os.Mkdir(dstFilePath, 0755); err != nil {
			return fmt.Errorf("Failed to create directory %s: %v\n", dstFilePath, err)
		}
	} else {
		logger.Printf("Copying file %s to %s. File size - %v\n", file.Name, dst, file.Size)
		if err := copyFile(srcFilePath, dstFilePath); err != nil {
			return fmt.Errorf("Failed to copy file %s to %s: %v\n", srcFilePath, dstFilePath, err)
		}
	}
	return nil
}

func UpdateFile(src string, dst string, dstFile FileInfo, srcFile FileInfo, logger *log.Logger) error {
	dstFilePath := filepath.Join(dst, dstFile.FullPath)
	srcFilePath := filepath.Join(src, srcFile.FullPath)
	if !dstFile.IsDir && !srcFile.IsDir && (dstFile.Size != srcFile.Size) || (dstFile.Permissons != srcFile.Permissons) {
		logger.Printf("Updating file %s. File size - %v\n", dstFile, dstFile.Size)
		if err := copyFile(srcFilePath, dstFilePath); err != nil {
			return fmt.Errorf("Failed to update file %s: %v\n", dstFilePath, err)
		}
	}
	return nil
}

func DeleteFile(dst string, file FileInfo, logger *log.Logger) error {
	filePath := filepath.Join(dst, file.FullPath)

	if file.IsDir {
		logger.Printf("Removing directory %s\n", filePath)
		if err := os.RemoveAll(filePath); err != nil {
			return fmt.Errorf("Failed to remove directory %s: %v\n", filePath, err)
		}
	} else {
		logger.Printf("Removing file %s. File size - %v\n", file.Name, file.Size)
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("Failed to remove file %s: %v\n", filePath, err)
		}
	}
	return nil
}

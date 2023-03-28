package test

import (
	"fmt"
	"folder_sync/internal/app/sync_folder"
	"github.com/sirupsen/logrus"
	_ "github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestGetFiles(t *testing.T) {
	dir := "./testdata"
	expectedFiles := map[string]sync_folder.FileInfo{
		"file1.txt": sync_folder.FileInfo{
			Name:       "file1.txt",
			Size:       0,
			ModTime:    time.Now(),
			IsDir:      false,
			Permissons: os.FileMode(0777),
			FullPath:   "file1.txt",
		},
		"file2.txt": sync_folder.FileInfo{
			Name:       "file2.txt",
			Size:       0,
			ModTime:    time.Now(),
			IsDir:      false,
			Permissons: os.FileMode(0777),
			FullPath:   "file2.txt",
		},
		"file3.txt": sync_folder.FileInfo{
			Name:       "file3.txt",
			Size:       0,
			ModTime:    time.Now(),
			IsDir:      false,
			Permissons: os.FileMode(0777),
			FullPath:   "file3.txt",
		},
	}
	if err := os.Mkdir(dir, 0755); err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	for _, value := range expectedFiles {
		filePath := filepath.Join(dir, value.FullPath)
		file, err := os.Create(filePath)
		if err != nil {
			panic(err)
		}
		file.Chmod(value.Permissons)
	}

	files, err := sync_folder.GetFiles(dir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(files) != len(expectedFiles) {
		t.Errorf("expected %d files, but got %d", len(expectedFiles), len(files))
	}

	for path, expectedFileInfo := range expectedFiles {
		fileInfo, ok := files[path]
		if !ok {
			t.Errorf("expected to find file %q, but it was not found", path)
		}

		if fileInfo.Name != expectedFileInfo.Name {
			t.Errorf("expected file %q to have name %q, but got %q", path, expectedFileInfo.Name, fileInfo.Name)
		}

		if fileInfo.Size != expectedFileInfo.Size {
			t.Errorf("expected file %q to have size %d, but got %d", path, expectedFileInfo.Size, fileInfo.Size)
		}

		if fileInfo.ModTime.Round(time.Second) != expectedFileInfo.ModTime.Round(time.Second) {
			t.Errorf("expected file %q to have modification time %v, but got %v", path, expectedFileInfo.ModTime, fileInfo.ModTime)
		}

		if fileInfo.IsDir != expectedFileInfo.IsDir {
			t.Errorf("expected file %q to be a directory, but it was not", path)
		}

		if fileInfo.Permissons != expectedFileInfo.Permissons {
			t.Errorf("expected file %q to have permissions %q, but got %q", path, expectedFileInfo.Permissons, fileInfo.Permissons)
		}

		if fileInfo.FullPath != expectedFileInfo.FullPath {
			t.Errorf("expected file %q to have full path %q, but got %q", path, expectedFileInfo.FullPath, fileInfo.FullPath)
		}
	}
}

func TestCopyFile(t *testing.T) {
	srcContent := "This is the source file content."
	srcFile, err := ioutil.TempFile("", "src*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(srcFile.Name())

	if _, err := srcFile.Write([]byte(srcContent)); err != nil {
		t.Fatal(err)
	}
	if err := srcFile.Close(); err != nil {
		t.Fatal(err)
	}

	dstFile, err := ioutil.TempFile("", "dst*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dstFile.Name())

	if err := dstFile.Close(); err != nil {
		t.Fatal(err)
	}

	if err := sync_folder.CopyFile(srcFile.Name(), dstFile.Name()); err != nil {
		t.Fatal(err)
	}

	dstContent, err := ioutil.ReadFile(dstFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if string(dstContent) != srcContent {
		t.Errorf("Expected %q, but got %q", srcContent, string(dstContent))
	}
}

func TestDeleteFile(t *testing.T) {
	dir := "./testdata"
	if err := os.Mkdir(dir, 0755); err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	testCases := []struct {
		dst     string
		file    sync_folder.FileInfo
		wantErr bool
	}{
		{
			dst: "/tmp1",
			file: sync_folder.FileInfo{
				FullPath: "/testdir",
				IsDir:    true,
			},
			wantErr: false,
		},
		{
			dst: "/tmp2",
			file: sync_folder.FileInfo{
				FullPath: "testfile.txt",
				Name:     "testfile.txt",
				Size:     1024,
			},
			wantErr: false,
		},
		{
			dst: "/doesnotexist",
			file: sync_folder.FileInfo{
				FullPath: "/testfile",
				Name:     "testfile.txt",
				Size:     1024,
			},
			wantErr: true,
		},
	}

	for _, value := range testCases {
		if !value.wantErr {
			path := filepath.Join(dir, value.dst)
			err := os.Mkdir(path, 0777)
			if err != nil {
				panic(err)
			}
			if !value.file.IsDir {
				path = filepath.Join(path, value.file.Name)
				_, err = os.Create(path)
				if err != nil {
					panic(err)
				}
			}
		}

	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v,%v", tc.dst, tc.file), func(t *testing.T) {
			logger := logrus.New()
			logger.Out = io.Discard
			if err := sync_folder.DeleteFile(filepath.Join(dir, tc.dst), tc.file, logger); (err != nil) != tc.wantErr {
				t.Errorf("DeleteFile() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestUpdateFile(t *testing.T) {
	dir := "./testdata"
	if err := os.Mkdir(dir, 0755); err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	dir2 := "./testdata2"
	if err := os.Mkdir(dir2, 0755); err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir2)

	newFiles := map[string]sync_folder.FileInfo{
		"file1.txt": sync_folder.FileInfo{
			Name:       "file1.txt",
			Size:       0,
			ModTime:    time.Now(),
			IsDir:      false,
			Permissons: os.FileMode(0777),
			FullPath:   "file1.txt",
		},
		"file2.txt": sync_folder.FileInfo{
			Name:       "file1.txt",
			Size:       0,
			ModTime:    time.Now(),
			IsDir:      false,
			Permissons: os.FileMode(0707),
			FullPath:   "file1.txt",
		},
	}
	for k, value := range newFiles {
		filePath := ""
		if k == "file1.txt" {
			filePath = filepath.Join(dir2, value.Name)
		} else {
			filePath = filepath.Join(dir, value.Name)
		}

		file, err := os.Create(filePath)
		if err != nil {
			panic(err)
		}
		file.Chmod(value.Permissons)
	}

	t.Run("UpdateFile ok", func(t *testing.T) {
		logger := logrus.New()
		logger.Out = io.Discard
		err := sync_folder.UpdateFile(dir, dir2, newFiles["file1.txt"], newFiles["file2.txt"], logger)
		if err != nil {
			t.Errorf("Error in UpdateFile func: %v", err)
		}

	})

	t.Run("UpdateFile no permissions", func(t *testing.T) {
		logger := logrus.New()
		logger.Out = io.Discard
		os.Chmod("./testdata/file1.txt", 0333)
		err := sync_folder.UpdateFile(dir, dir2, newFiles["file1.txt"], newFiles["file2.txt"], logger)
		if err == nil {
			t.Errorf("Error in UpdateFile func: %v", err)
		}

	})

}

func TestAddFile(t *testing.T) {
	dir := "./testdata"
	if err := os.Mkdir(dir, 0755); err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	dir2 := "./testdata2"
	if err := os.Mkdir(dir2, 0755); err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir2)

	testCases := []struct {
		src     string
		dst     string
		file    sync_folder.FileInfo
		wantErr bool
	}{
		{
			src: dir,
			dst: dir2,
			file: sync_folder.FileInfo{
				FullPath: "/testdir",
				IsDir:    true,
			},
			wantErr: false,
		},
		{
			src: dir,
			dst: dir2,
			file: sync_folder.FileInfo{
				FullPath:   "testfile.txt",
				Name:       "testfile.txt",
				Size:       1024,
				Permissons: 0644,
			},
			wantErr: false,
		},
		{
			src: "/doesnotexist",
			dst: dir2,
			file: sync_folder.FileInfo{
				FullPath: "/testfile",
				Name:     "testfile",
				Size:     1024,
			},
			wantErr: true,
		},
	}

	for _, value := range testCases {
		if !value.wantErr {
			path := filepath.Join(dir, value.file.FullPath)
			if value.file.IsDir {
				err := os.Mkdir(path, 0777)
				if err != nil {
					panic(err)
				}
			} else {
				_, err := os.Create(path)
				if err != nil {
					panic(err)
				}
			}
		}

	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v,%v,%v", tc.src, tc.dst, tc.file), func(t *testing.T) {
			logger := logrus.New()
			logger.Out = io.Discard
			if err := sync_folder.AddFile(tc.src, tc.dst, tc.file, logger); (err != nil) != tc.wantErr {
				t.Errorf("AddFile() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestSyncDirs(t *testing.T) {
	src := "./tmp/src"
	dst := "./tmp/dst"
	logger := logrus.New()
	logger.Out = io.Discard
	defer os.RemoveAll("./tmp")

	// Create temporary files and directories in the source directory
	err := os.MkdirAll(src+"/dir1", 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	err = os.MkdirAll(dst, 0777)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	err = ioutil.WriteFile(src+"/file1.txt", []byte("Hello, world!"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Call the function being tested
	var wg sync.WaitGroup

	t.Run("ok result", func(t *testing.T) {
		wg.Add(1)
		err = sync_folder.SyncDirs(&wg, logger, src, dst)
		wg.Wait()
		if err != nil {
			t.Fatalf("SyncDirs failed: %v", err)
		}

		// Check that the files and directories in the source directory exist in the destination directory
		_, err = os.Stat(dst + "/dir1")
		if err != nil {
			t.Fatalf("Directory not found in destination: %v", err)
		}
		_, err = os.Stat(dst + "/file1.txt")
		if err != nil {
			t.Fatalf("File not found in destination: %v", err)
		}
	})

	t.Run("Run with wrong src folder path", func(t *testing.T) {
		wg.Add(1)
		err = sync_folder.SyncDirs(&wg, logger, src+"wrongpath", dst)
		wg.Wait()
		if err == nil {
			t.Fatalf("SyncDirs must be failed with wrong src folder: %v", err)
		}
	})
	t.Run("Run with wrong dst folder path", func(t *testing.T) {
		wg.Add(1)
		err = sync_folder.SyncDirs(&wg, logger, src, dst+"wrongpath")
		wg.Wait()
		if err == nil {
			t.Fatalf("SyncDirs must be failed with wrong dst folder: %v", err)
		}

	})

	t.Run("Run with no permisons for reading file", func(t *testing.T) {
		wg.Add(1)
		err = ioutil.WriteFile(src+"/file2.txt", []byte("Hello, world!"), 0333)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		err = sync_folder.SyncDirs(&wg, logger, src, dst)
		wg.Wait()
		if err == nil {
			t.Fatalf("SyncDirs must be failed with wrong permissions for dst folder: %v", err)
		}
	})

}

func BenchmarkCopyFile(b *testing.B) {
	srcContent := make([]byte, 1024*1024)
	srcFile, err := ioutil.TempFile("", "src*.txt")
	defer srcFile.Close()
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(srcFile.Name())

	if _, err := srcFile.Write(srcContent); err != nil {
		b.Fatal(err)
	}

	dstFile, err := ioutil.TempFile("", "dst*.txt")
	defer dstFile.Close()
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(dstFile.Name())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := sync_folder.CopyFile(srcFile.Name(), dstFile.Name()); err != nil {
			b.Fatal(err)
		}
		if err := dstFile.Truncate(0); err != nil {
			b.Fatal(err)
		}
		if _, err := srcFile.Seek(0, 0); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}

package handlers

import (
	"bufio"
	"crypto/sha512"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"owlsintheoven/learning-go/fileengine/workers"
	"owlsintheoven/learning-go/ggin"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const (
	maxWorker = 10
)

func Filehash(c *ggin.Context) {
	reader := bufio.NewReader(c.Request.Body)
	body, err := io.ReadAll(reader)
	if err != nil {
		c.Writer.Write([]byte("error"))
	}
	path := string(body)

	fileHashes := constructFileHashesBounded(path)
	c.Writer.Write([]byte(strings.Join(fileHashes, "\n")))
}
func constructFileHashesBounded(path string) []string {
	wp := workers.NewWorkerPool(maxWorker, fileSHA512)
	wp.Run()

	var fileHashes []string
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for res := range wp.ResultC {
			fileHashes = append(fileHashes, res)
		}
	}()
	go func() {
		defer wg.Done()
		for err := range wp.ErrorC {
			log.Println(err.Error())
		}
	}()

	filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Println("walk dir encountered error", err.Error())
			return nil
		}
		if !d.IsDir() {
			wp.AddTask(path)
		}
		return nil
	})
	wp.Stop()
	wg.Wait()
	sort.Sort(sort.StringSlice(fileHashes))
	return fileHashes
}

func constructFileHashesUnbounded(path string) []string {
	var fileHashes []string
	c := make(chan string, 1024)
	var wg, wg2 sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for res := range c {
			fileHashes = append(fileHashes, res)
		}
	}()
	filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Println("walk dir encountered error", err.Error())
			return nil
		}
		if !d.IsDir() {
			wg2.Add(1)
			go func() {
				defer wg2.Done()
				hash, err := fileSHA512(path)
				if err != nil {
					log.Println("error calculating sha512", err.Error())
				} else {
					c <- fmt.Sprintf("%s %s", path, hash)
				}
			}()
		}
		return nil
	})
	wg2.Wait()
	close(c)
	wg.Wait()
	sort.Sort(sort.StringSlice(fileHashes))
	return fileHashes
}

func fileSHA512(filePath string) (string, error) {
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		log.Printf("error reading file %s: %s\n", filePath, err.Error())
		return "", err
	}

	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Printf("error performing io copying %s: %s\n", filePath, err.Error())
		return "", err
	}

	return fmt.Sprintf("%s %x", filePath, h.Sum(nil)), nil
}

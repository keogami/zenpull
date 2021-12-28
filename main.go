package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
)

var (
	workerCount           = 5
	directoryExistenceMap = make(map[string]bool)
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("you need to provide the list file, dumbass")
		return
	}

	filename := os.Args[1]

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("couldn't open the list file, bitch")
		fmt.Println("error:", err)
		return
	}
	defer file.Close()

	tap := dispatch(file)

	wg := new(sync.WaitGroup)
	wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go retrieve(wg, tap)
	}

	wg.Wait()

	fmt.Println("done <3")
}

func dispatch(from io.Reader) <-chan string {
	out := make(chan string, workerCount)
	scanner := bufio.NewScanner(from)
	go func(s *bufio.Scanner, o chan<- string) {
		for s.Scan() {
			url := s.Text()
			ext := string(path.Ext(url)[1:])
			mkdirIfRequired(filepath.Join("./", ext))
			o <- url
		}
		close(out)
	}(scanner, out)
	return out
}

func doesDirExists(p string) bool {
	if _, err := os.Stat(p); err != nil {
		return false
	}
	return true
}

func mkdirIfRequired(p string) {
	_, has := directoryExistenceMap[p]
	if has {
		return
	}
	exist := doesDirExists(p)
	if exist {
		directoryExistenceMap[p] = true
		return
	}
	os.Mkdir(p, 0666)
	directoryExistenceMap[p] = true
}

func retrieve(wg *sync.WaitGroup, tap <-chan string) {
	defer wg.Done()

	for url := range tap {
		dest := makeDestPath(url)
		_, err := os.Stat(dest)
		if err == nil {
			fmt.Printf("skipping: %s\n", url)
			continue
		}
		fmt.Printf("retrieving: %s\n", url)
		err = netCopy(url, dest)
		if err != nil {
			fmt.Printf("failed: %s | Reason: %s\n", url, err)
		}
	}
}

func makeDestPath(url string) string {
	ext := string(path.Ext(url)[1:])
	hash := sha1.Sum([]byte(url))
	hashText := hex.EncodeToString(hash[:])
	dest := filepath.Join("./", ext, hashText) + "." + ext
	return dest
}

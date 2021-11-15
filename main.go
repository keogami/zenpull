package main

import (
  "os"
  "fmt"
  "bufio"
  "io"
  "net/http"
  "encoding/json"
  "strconv"
  "time"
  "encoding/hex"
  "crypto/sha1"
  "sync"
  "path"
  "path/filepath"
  "errors"
)

var (
  workerCount = 5
  directoryExistenceMap = make(map[string]bool)
)

// Meta is the meta information to be persisted, required by the organiser
// [note] this is no longer needed, so it should be removed
type Meta struct {
  Base string `json:"base"`
}

func (m Meta) store(path string) error {
  file, err := os.Create(path)
  if err != nil {
    return err
  }
  defer file.Close()

  err = json.NewEncoder(file).Encode(m)
  return err
} 

func makeMeta() (Meta) {
  t := time.Now().Unix()
  base := strconv.FormatInt(t, 16)
  return Meta{
    Base: base,
  }
}

func getMeta() (Meta, error) {
  cfile, err := os.Open("meta.json")
  if err != nil {
    return Meta{}, err
  }
  defer cfile.Close()
  
  var meta Meta
  err = json.NewDecoder(cfile).Decode(&meta)
  if err != nil {
    return Meta{}, err
  }

  return meta, nil
}

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

  meta, err := getMeta()
  if err != nil {
    fmt.Println("couldn't load meta from meta.json")
    fmt.Println("error:", err)
    fmt.Println("just creating a new one :p")
    meta = makeMeta()
  }

  mkdirIfRequired(meta.Base)

  tap := dispatch(file, meta)

  wg := new(sync.WaitGroup)
  wg.Add(workerCount)

  for i := 0; i < workerCount; i++ {
    go retrieve(wg, tap, meta)
  }

  wg.Wait()

  err = meta.store("meta.json")
  if err != nil {
    fmt.Println("couldn't store meta at meta.json, bummer TwT")
  }

  fmt.Println("done <3")
}

func dispatch(from io.Reader, meta Meta) (<-chan string) {
  out := make(chan string, workerCount)
  scanner := bufio.NewScanner(from)
  go func (s *bufio.Scanner, o chan<- string) {
    for s.Scan() {
      url := s.Text()
      ext := string(path.Ext(url)[1:])
      mkdirIfRequired(filepath.Join(meta.Base, ext))
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

func retrieve(wg *sync.WaitGroup, tap <-chan string, meta Meta) {
  defer wg.Done()
  base := meta.Base

  for url := range tap {
    dest := makeDestPath(url, base)
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

func makeDestPath(url, base string) string {
  ext := string(path.Ext(url)[1:])
  hash := sha1.Sum([]byte(url))
  hashText := hex.EncodeToString(hash[:])
  dest := filepath.Join(base, ext, hashText) + "." + ext
  return dest
}

func netCopy(fromUrl, toFile string) error {
  status, length, rangesAllowed, unit, err := checkoutUrl(fromUrl)
  if err != nil {
    return err
  }

  if status != http.StatusOK {
    return errors.New("response status code not ok")
  }

  dest, err := os.Create(toFile)
  if err != nil {
    return err
  }
  defer dest.Close()
  bdest := bufio.NewWriter(dest)

  if length != 0 && rangesAllowed && unit == "bytes" {
    return rangedDownload(bdest, fromUrl, length)
  }
  return plainDownload(bdest, fromUrl)
}

func rangedDownload(w io.Writer, fromUrl string, length int64) error {
  downloaded := int64(0)
  client := &http.Client{}
  for downloaded != length {
    req, err := http.NewRequest(http.MethodGet, fromUrl, nil)
    if err != nil {
      return err
    }
    req.Header.Set("Range", fmt.Sprintf("bytes=%d-", downloaded))

    resp, err := client.Do(req)
    if err != nil {
      return err
    }

    more, err := io.Copy(w, resp.Body)
    if err != nil {
      fmt.Println("err:", err, "; more:", more) // for debugging, in case i see it again
    }
    downloaded += more
    resp.Body.Close()
  }
  return nil
}

func plainDownload(w io.Writer, fromUrl string) error {
  resp, err := http.Get(fromUrl)
  if err != nil {
    return err
  }
  defer resp.Body.Close()

  _, err = io.Copy(w, resp.Body)
  return err
}

func checkoutUrl(fromUrl string) (status int, length int64, acceptsRanges bool, unit string, err error) {
  resp, err := http.Head(fromUrl)

  if err != nil {
    return
  }

  status = resp.StatusCode

  length = resp.ContentLength

  if arHeader := resp.Header.Get("Accept-Ranges"); arHeader != "" {
    acceptsRanges = true
    unit = arHeader
  }
  return
}
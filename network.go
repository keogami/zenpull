package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

var (
	httpClient = &http.Client{}
)

func GetClient() *http.Client {
	return httpClient
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
	for downloaded != length { // maybe the infinite download bug is caused by this equality test
		req, err := http.NewRequest(http.MethodGet, fromUrl, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", downloaded))

		resp, err := GetClient().Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		more, err := io.Copy(w, resp.Body)
		if err != nil {
			fmt.Println("err:", err, "; more:", more) // for debugging, in case i see it again
		}
		downloaded += more
	}
	return nil
}

func plainDownload(w io.Writer, fromUrl string) error {
	resp, err := GetClient().Get(fromUrl)
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

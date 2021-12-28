package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

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
	client := &http.Client{}   // we dont need to create a client
	for downloaded != length { // maybe the infinite download bug is caused by this equality test
		req, err := http.NewRequest(http.MethodGet, fromUrl, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", downloaded))

		resp, err := client.Do(req) // use http.DO instead
		if err != nil {
			return err
		}

		more, err := io.Copy(w, resp.Body)
		if err != nil {
			fmt.Println("err:", err, "; more:", more) // for debugging, in case i see it again
		}
		downloaded += more
		resp.Body.Close() // use defer instead of closing it here
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

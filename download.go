package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type Download struct {
	Url           string
	Targetpath    string
	TotalSections int
}

func (dm Download) Do() error {
	fmt.Println("Checking URL .....")

	// 1. HEAD request
	r, err := dm.getNewRequest("HEAD")
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	fmt.Printf("Got Response %v.\n\n", resp.StatusCode)

	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Can't proceed, respone is %v", resp.StatusCode))
	}

	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	fmt.Printf("File has a size of %d bytes or %f MB\n\n", size, float64(size)/1000000.0)

	// create file chunks
	var sections = make([][2]int, dm.TotalSections)
	eachSize := size / dm.TotalSections
	fmt.Printf("Each chunk size is %v bytes\n", eachSize)

	// [[0 10][10 20]....[99 end-1]] - total 100 byte file for ex
	for i := range sections {
		if i == 0 {
			sections[i][0] = 0
		} else {
			sections[i][0] = sections[i-1][1] + 1
		}

		if i == dm.TotalSections-1 {
			sections[i][1] = size - 1
		} else {
			sections[i][1] = sections[i][0] + eachSize
		}
	}
	log.Println(sections)

	// 2. Download each chunk concurrently
	var wg sync.WaitGroup
	for i, s := range sections {
		wg.Add(1)
		go func(i int, s [2]int) {
			defer wg.Done()
			err := dm.downloadChunk(i, s)
			if err != nil {
				panic(err)
			}
		}(i, s)
	}

	wg.Wait()
	err = dm.mergeFiles()
	if err != nil {
		return err
	}

	return nil
}

func (dm Download) getNewRequest(method string) (*http.Request, error) {
	r, err := http.NewRequest(
		method,
		dm.Url,
		nil,
	)

	if err != nil {
		return nil, err
	}

	r.Header.Set("User-Agent", "Sabuj's File Download Manager v1")
	return r, nil
}

func (dm Download) downloadChunk(idx int, sec [2]int) error {
	r, err := dm.getNewRequest("GET")
	if err != nil {
		return err
	}

	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", sec[0], sec[1]))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Can't proceed, respone is %v", resp.StatusCode))
	}

	fmt.Printf("Downloaded %v bytes for Section %v\n", resp.ContentLength, idx)
	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf("section-%v.tmp", idx), b, os.ModePerm)

	if err != nil {
		return err
	}

	return nil
}

func (dm Download) mergeFiles() error {
	fileX, err := os.OpenFile(dm.Targetpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	defer fileX.Close()

	for i := 0; i < dm.TotalSections; i++ {
		fileTmp := fmt.Sprintf("section-%v.tmp", i)
		b, err := ioutil.ReadFile(fileTmp)

		if err != nil {
			return err
		}

		bw, err := fileX.Write(b)
		if err != nil {
			return err
		}
		err = os.Remove(fileTmp)
		if err != nil {
			return err
		}
		fmt.Printf("Merged %v bytes from Section %v\n", bw, i)
	}

	return nil
}

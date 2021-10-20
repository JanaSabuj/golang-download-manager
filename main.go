package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	fmt.Println("Welcome to my Download Manager !")

	// initialise a download manager
	dm := Download{
		Url:           "https://file-examples-com.github.io/uploads/2017/11/file_example_MP3_5MG.mp3",
		Targetpath:    "abc.mp3",
		TotalSections: 10,
	}

	err := dm.Do()
	if err != nil {
		log.Printf("Error occurred: %s", err)
	}

	startTime := time.Now()
	endTime := time.Now()

	fmt.Printf("Download completed successfully in %v seconds. \n", endTime.Sub(startTime).Seconds())
}

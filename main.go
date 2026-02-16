package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	statsURL = "http://srv.msk01.gigacorp.local/_stats"
)

func main() {
	errorCount := 0

	for {
		messages, err := fetchAndCheckStats()
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				errorCount = 0
			}
		} else {
			errorCount = 0
			for _, msg := range messages {
				fmt.Println(msg)
			}
		}

		time.Sleep(time.Second)
	}
}
func fetchAndCheckStats() ([]string, error) {
	resp, err := http.Get(statsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(parts) != 7 {
		return nil, fmt.Errorf("invalid data format")
	}

	load, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, err
	}

	values := make([]int64, 6)
	for i := 1; i < 7; i++ {
		v, err := strconv.ParseInt(parts[i], 10, 64)
		if err != nil {
			return nil, err
		}
		values[i-1] = v
	}

	totalMem := values[0]
	usedMem := values[1]
	totalDisk := values[2]
	usedDisk := values[3]
	totalNet := values[4]
	usedNet := values[5]

	var messages []string

	// avg load
	if load > 30 {
		messages = append(messages,
			fmt.Sprintf("Load Average is too high: %v", load))
	}

	// mem
	if totalMem > 0 {
		memUsage := usedMem * 100 / totalMem
		if memUsage >= 80 {
			messages = append(messages,
				fmt.Sprintf("Memory usage too high: %d%%", memUsage))
		}
	}

	// disk
	if totalDisk > 0 {
		diskUsage := usedDisk * 100 / totalDisk
		if diskUsage >= 90 {
			freeBytes := totalDisk - usedDisk
			freeMb := freeBytes / 1024 / 1024
			messages = append(messages,
				fmt.Sprintf("Free disk space is too low: %d Mb left", freeMb))
		}
	}

	// network
	if totalNet > 0 {
		netUsage := usedNet * 100 / totalNet
		if netUsage >= 90 {
			freeBytes := totalNet - usedNet
			freeMbit := freeBytes / 1_000_000
			messages = append(messages,
				fmt.Sprintf("Network bandwidth usage high: %d Mbit/s available", freeMbit))
		}
	}

	return messages, nil
}

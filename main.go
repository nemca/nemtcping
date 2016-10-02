package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/montanaflynn/stats"
)

var (
	host    string
	port    = 80
	count   int
	timeout int
)

func init() {
	flag.IntVar(&count, "c", 4, "Number of requests to send")
	flag.IntVar(&timeout, "t", 1, "Timeout for each request, in seconds")
}

func usage(filename string) {
	fmt.Printf("Usage: %s [-c count] [-t timeout] <host> [<port>]\n", filename)
}

func main() {
	flag.Parse()

	filename := os.Args[0]
	args := flag.Args()

	if len(args) < 1 {
		usage(filename)
		os.Exit(255)
	}

	host = args[0]
	_, err := net.LookupIP(host)
	if err != nil {
		fmt.Println("error: unknown host")
		os.Exit(2)
	}

	if len(args) == 2 {
		port, err = strconv.Atoi(args[1])
		if err != nil || port < 1 || port > 65535 {
			fmt.Printf("Argument [%s] was not correct, <port> must be a positive integer in the range 1 - 65535\n", args[1])
			os.Exit(255)
		}
	}

	ping(host, port, count, timeout)
}

func ping(host string, port, count, timeout int) {
	successfulProbes := 0
	timeTotal := time.Duration(0)
	addr := fmt.Sprintf("%s:%d", host, port)
	i := 1
	var responseTimes []float64

	for i = 1; count >= i; i++ {
		timeStart := time.Now()
		_, err := net.DialTimeout("tcp", addr, time.Second*time.Duration(timeout))
		responseTime := time.Since(timeStart)
		if err != nil {
			fmt.Println(fmt.Sprintf("Received timeout while connecting to %s on port %d.", host, port))
		} else {
			fmt.Println(fmt.Sprintf("Connected to %s:%d, RTT=%.3f ms", host, port, float32(responseTime)/1e6))
			timeTotal += responseTime
			successfulProbes++
			responseTimes = append(responseTimes, float64(responseTime))
		}
		time.Sleep(time.Second - responseTime)
	}

	var max float64
	min := float64(1000000000)
	for _, v := range responseTimes {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}

	avg, _ := stats.Median(responseTimes)

	fmt.Printf("\n--- %s nemtcping statistic ---\n", host)
	fmt.Printf("%d packets transmitted, %d packets received, %.1f%% packet loss\n", count, successfulProbes, float64(100-(successfulProbes*100)/(i-1)))
	fmt.Printf("round-trip min/avg/max = %.3f/%.3f/%.3f ms\n", float32(min)/1e6, float32(avg)/1e6, float32(max)/1e6)
}

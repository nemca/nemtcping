package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/montanaflynn/stats"
)

var (
	host      string
	port      = 80
	count     int
	timeout   int
	ping_flag bool
	quiet     bool
)

type Ping struct {
	Host             string
	Port             int
	Timeout          int
	Count            int
	SuccessfulProbes int
	ResponseTimes    []float64
	Probes           int
}

func init() {
	flag.IntVar(&count, "c", 0, "number of requests to send")
	flag.IntVar(&timeout, "t", 1, "timeout for each request, in seconds")
	flag.BoolVar(&ping_flag, "p", false, "run ping")
	flag.BoolVar(&quiet, "q", false, "quiet mode, do not output anything (except error messages)")
}

func usage(filename string) {
	fmt.Fprintf(os.Stderr, "Usage: %s [-c count] [-t timeout] [-p] <host> [<port>]\n", filename)
}

func main() {
	flag.Parse()

	filename := filepath.Base(os.Args[0])
	args := flag.Args()

	if len(args) < 1 {
		usage(filename)
		os.Exit(255)
	}

	host = args[0]
	ips, err := net.LookupIP(host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: unknown host")
		os.Exit(2)
	}

	if len(args) == 2 {
		port, err = strconv.Atoi(args[1])
		if err != nil || port < 1 || port > 65535 {
			fmt.Fprintf(os.Stderr, "Argument [%s] was not correct, <port> must be a positive integer in the range 1 - 65535\n", args[1])
			os.Exit(255)
		}
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		<-sigs
		fmt.Println()
		done <- true
	}()

	if ping_flag {
		p := &Ping{Host: host, Port: port, Timeout: timeout, Count: count}
		ping(p, filename, ips[0], done)
		var max float64
		min := float64(1000000000 * timeout)
		for _, v := range p.ResponseTimes {
			if v > max {
				max = v
			}
			if v < min {
				min = v
			}
		}

		avg, _ := stats.Median(p.ResponseTimes)

		fmt.Printf("\n--- %s nemtcping statistic ---\n", p.Host)
		fmt.Printf("%d packets transmitted, %d packets received, %.1f%% packet loss\n", p.Probes, p.SuccessfulProbes, float64(100-(p.SuccessfulProbes*100)/p.Probes))
		fmt.Printf("round-trip min/avg/max = %.3f/%.3f/%.3f ms\n", float32(min)/1e6, float32(avg)/1e6, float32(max)/1e6)

		os.Exit(0)
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	_, err = net.DialTimeout("tcp", addr, time.Second*time.Duration(timeout))
	if err != nil {
		say("%s port %d closed.\n", host, port)
		os.Exit(1)
	}
	say("%s port %d open.\n", host, port)
}

func say(format string, a ...interface{}) {
	if !quiet {
		fmt.Fprintf(os.Stdout, format, a...)
	}
}

func ping(p *Ping, filename string, ip net.IP, done chan bool) {
	addr := fmt.Sprintf("%s:%d", p.Host, p.Port)
	// var i int

	fmt.Printf("%s %s (%s)\n", filename, p.Host, ip)
	if p.Count == 0 {
		for {
			select {
			case <-done:
				return
			default:
				p.Probes++
				timeStart := time.Now()
				_, err := net.DialTimeout("tcp", addr, time.Second*time.Duration(timeout))
				responseTime := time.Since(timeStart)
				if err != nil {
					say("Received timeout while connecting to %s on port %d\n", p.Host, p.Port)
				} else {
					say("Connected to %s:%d, RTT=%.3f ms\n", host, port, float32(responseTime)/1e6)
					p.SuccessfulProbes++
					p.ResponseTimes = append(p.ResponseTimes, float64(responseTime))
				}
				time.Sleep(time.Second - responseTime)
			}
		}
	} else {
		for ; p.Count > p.Probes; p.Probes++ {
			timeStart := time.Now()
			_, err := net.DialTimeout("tcp", addr, time.Second*time.Duration(timeout))
			responseTime := time.Since(timeStart)
			if err != nil {
				say("Received timeout while connecting to %s on port %d\n", p.Host, p.Port)
			} else {
				say("Connected to %s:%d, RTT=%.3f ms\n", host, port, float32(responseTime)/1e6)
				p.SuccessfulProbes++
				p.ResponseTimes = append(p.ResponseTimes, float64(responseTime))
			}
			time.Sleep(time.Second - responseTime)
		}
	}

}

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var sigCh = make(chan os.Signal, 1)

func AcceptSigs() {
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
}

func init() {
	AcceptSigs()
}

func Go(f func() error) chan error {
	errCh := make(chan error, 0)
	go func() {
		errCh <- f()
	}()
	return errCh
}

type HookWriter struct {
	hook func(data []byte)
}

func (w *HookWriter) Write(data []byte) (int, error) {
	w.hook(data)
	return len(data), nil
}

func main() {
	delay := flag.Duration("delay", time.Second*5, "Delay between each restart")
	killon := flag.String("killon", "", "Kill program when text appear")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("Usage: %s [-delay 5s] <commands ...>")
		return
	}
	for {
		killCh := make(chan bool, 1)
		cmd := exec.Command(flag.Arg(0), flag.Args()[1:]...)
		bufs := make([]string, 4)
		hwr := &HookWriter{
			hook: func(data []byte) {
				bufs = append(bufs, string(data))
				if len(bufs) > 4 {
					bufs = bufs[len(bufs)-4 : len(bufs)]
				}
				if strings.Contains(strings.Join(bufs, ""), *killon) {
					killCh <- true
				}
			},
		}
		cmd.Stdin = os.Stdin
		cmd.Stdout = io.MultiWriter(hwr, os.Stdout)
		cmd.Stderr = io.MultiWriter(hwr, os.Stderr)

		select {
		case err := <-Go(cmd.Run):
			if err != nil {
				log.Println(err)
			} else {
				log.Println("Normal exit")
				return
			}
		case <-killCh:
			log.Println("Trigger killon")
			if cmd.Process != nil {
				log.Println("Trigger killon, kill process")
				cmd.Process.Kill()
				return
			}
		case sig := <-sigCh:
			log.Printf("recv signal: %v", sig)
			if cmd.Process != nil {
				cmd.Process.Kill()
				return
			}
		}

		select {
		case <-time.After(*delay):
		case sig := <-sigCh:
			log.Printf("recv signal: %v", sig)
			return
		}
	}
}

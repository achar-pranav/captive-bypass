//go:build linux

// Package watcher subscribes to kernel interface/route events (netlink).
// The kernel announces link up/down and route changes within milliseconds
// of a WiFi connect/disconnect, for every manager, without privileges.
package watcher

import (
	"errors"
	"log"
	"os"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

var debounce = func() time.Duration {
	if v := os.Getenv("CAPTIVE_DEBOUNCE_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 100 {
			return time.Duration(n) * time.Millisecond
		}
	}
	return 800 * time.Millisecond
}()

const pollTickMs = -1 // block until event or Stop

var debug = os.Getenv("CAPTIVE_WATCHER_DEBUG") != ""

type Watcher struct {
	fd      int
	stopped atomic.Bool
}

func Start(onChange func()) (*Watcher, error) {
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW|unix.SOCK_NONBLOCK, unix.NETLINK_ROUTE)
	if err != nil {
		return nil, err
	}
	addr := &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
		Groups: unix.RTNLGRP_LINK | unix.RTNLGRP_IPV4_IFADDR | unix.RTNLGRP_IPV4_ROUTE,
	}
	if err := unix.Bind(fd, addr); err != nil {
		unix.Close(fd)
		return nil, err
	}
	w := &Watcher{fd: fd}
	go w.loop(onChange)
	return w, nil
}

func (w *Watcher) loop(onChange func()) {
	buf := make([]byte, 8192)
	fds := []unix.PollFd{{Fd: int32(w.fd), Events: unix.POLLIN}}
	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()
	for !w.stopped.Load() {
		n, err := unix.Poll(fds, pollTickMs)
		if err != nil {
			if errors.Is(err, syscall.EINTR) {
				continue
			}
			return
		}
		if n == 0 {
			continue
		}
		msgLen, _, _ := unix.Recvfrom(w.fd, buf, 0)
		if msgLen <= 0 {
			continue
		}
		if debug {
			log.Printf("watcher: netlink burst (%d bytes)", msgLen)
		}
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounce, func() {
			if debug {
				log.Println("watcher: firing after debounce")
			}
			onChange()
		})
	}
}

func (w *Watcher) Stop() {
	if w.stopped.Swap(true) {
		return
	}
	unix.Close(w.fd)
}

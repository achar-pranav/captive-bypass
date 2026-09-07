//go:build darwin

package macos

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Foundation -framework CoreWLAN

#include <stdlib.h>

extern int start_cw_listener(void);
extern void stop_cw_listener(void);
*/
import "C"

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"unsafe"
)

var (
	eventsChanMu sync.Mutex
	eventsChan   chan Event
)

//export onCWEvent
func onCWEvent(cSSID *C.char, cBSSID *C.char, connected C.int) {
	e := Event{Connected: connected == 1}
	if cSSID != nil {
		e.SSID = C.GoString(cSSID)
		C.free(unsafe.Pointer(cSSID))
	}
	if cBSSID != nil {
		e.BSSID = C.GoString(cBSSID)
		C.free(unsafe.Pointer(cBSSID))
	}

	eventsChanMu.Lock()
	ch := eventsChan
	eventsChanMu.Unlock()

	if ch != nil {
		select {
		case ch <- e:
		default:
		}
	}
}

func Listen(ctx context.Context, sockPath string) error {
	ch := make(chan Event, 16)
	eventsChanMu.Lock()
	eventsChan = ch
	eventsChanMu.Unlock()

	defer func() {
		eventsChanMu.Lock()
		eventsChan = nil
		eventsChanMu.Unlock()
	}()

	errCh := make(chan error, 1)
	go func() {
		ret := C.start_cw_listener()
		if ret != 0 {
			errCh <- errors.New("failed to initialize CoreWLAN event listener")
		} else {
			errCh <- nil
		}
	}()

	for {
		select {
		case <-ctx.Done():
			C.stop_cw_listener()
			return nil
		case err := <-errCh:
			return err
		case e := <-ch:
			forwardEvent(sockPath, e)
		}
	}
}

func forwardEvent(sockPath string, e Event) {
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		return
	}
	defer conn.Close()

	if e.Connected {
		fmt.Fprintf(conn, "connect %s\n", e.SSID)
	} else {
		fmt.Fprintln(conn, "disconnect")
	}
}

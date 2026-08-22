//go:build windows && (amd64 || arm64)

package windows

import (
	"context"
	"fmt"
	"net"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	wlanNotificationSourceMSM = 0x00000010
	msmConnected              = 3
	msmDisconnected           = 6
)

var (
	modwlanapi                   = windows.NewLazySystemDLL("wlanapi.dll")
	procWlanOpenHandle           = modwlanapi.NewProc("WlanOpenHandle")
	procWlanCloseHandle          = modwlanapi.NewProc("WlanCloseHandle")
	procWlanRegisterNotification = modwlanapi.NewProc("WlanRegisterNotification")
)

type notificationData struct {
	Source uint32
	Code   uint32
	Guid   windows.GUID
	Size   uint32
	_      uint32
	Data   uintptr
}

type msmData struct {
	ConnMode  uint32
	Profile   [256]uint16
	SsidLen   uint32
	Ssid      [32]byte
	BssType   uint32
	Mac       [6]byte
	SecEn     uint32
	FirstPeer uint32
	LastPeer  uint32
	Reason    uint32
}

func Listen(ctx context.Context, sockPath string) error {
	var negotiated uint32
	var handle windows.Handle
	r1, _, callErr := procWlanOpenHandle.Call(2, 0, uintptr(unsafe.Pointer(&negotiated)), uintptr(unsafe.Pointer(&handle)))
	if r1 != 0 {
		return fmt.Errorf("WlanOpenHandle: %w", callErr)
	}
	defer procWlanCloseHandle.Call(uintptr(handle), 0, 0)

	events := make(chan Event, 16)
	cb := windows.NewCallback(func(data, _ uintptr) uintptr {
		if data == 0 {
			return 0
		}
		n := (*notificationData)(unsafe.Pointer(data))
		if n.Source != wlanNotificationSourceMSM {
			return 0
		}
		switch n.Code {
		case msmConnected:
			e := Event{Connected: true}
			if n.Data != 0 && n.Size >= uint32(unsafe.Sizeof(msmData{})) {
				m := (*msmData)(unsafe.Pointer(n.Data))
				if m.SsidLen <= uint32(len(m.Ssid)) {
					e.SSID = string(m.Ssid[:m.SsidLen])
				}
				e.BSSID = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
					m.Mac[0], m.Mac[1], m.Mac[2], m.Mac[3], m.Mac[4], m.Mac[5])
			}
			select {
			case events <- e:
			default:
			}
		case msmDisconnected:
			select {
			case events <- Event{}:
			default:
			}
		}
		return 0
	})

	var prev uint32
	r1, _, callErr = procWlanRegisterNotification.Call(
		uintptr(handle),
		wlanNotificationSourceMSM,
		1,
		cb,
		0,
		0,
		uintptr(unsafe.Pointer(&prev)),
	)
	if r1 != 0 {
		return fmt.Errorf("WlanRegisterNotification: %w", callErr)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-events:
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

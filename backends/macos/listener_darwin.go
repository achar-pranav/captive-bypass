//go:build darwin

package macos

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Foundation -framework CoreWLAN

#import <Foundation/Foundation.h>
#import <CoreWLAN/CoreWLAN.h>

extern void onCWEvent(char* ssid, char* bssid, int connected);

@interface CWListenerDelegate : NSObject <CWEventDelegate>
- (void)notifyCurrentState;
@end

@implementation CWListenerDelegate

- (void)ssidDidChangeForWiFiInterfaceWithName:(NSString *)interfaceName {
	[self notifyCurrentState];
}

- (void)bssidDidChangeForWiFiInterfaceWithName:(NSString *)interfaceName {
	[self notifyCurrentState];
}

- (void)linkDidChangeForWiFiInterfaceWithName:(NSString *)interfaceName {
	[self notifyCurrentState];
}

- (void)powerStateDidChangeForWiFiInterfaceWithName:(NSString *)interfaceName {
	[self notifyCurrentState];
}

- (void)notifyCurrentState {
	@autoreleasepool {
		CWWiFiClient *client = [CWWiFiClient sharedWiFiClient];
		if (!client) return;
		CWInterface *iface = [client interface];
		if (!iface) {
			onCWEvent(NULL, NULL, 0);
			return;
		}

		BOOL powerOn = [iface powerOn];
		NSString *ssid = [iface ssid];
		NSString *bssid = [iface bssid];

		if (powerOn && ssid && [ssid length] > 0) {
			char *cSSID = strdup([ssid UTF8String]);
			char *cBSSID = bssid ? strdup([bssid UTF8String]) : NULL;
			onCWEvent(cSSID, cBSSID, 1);
		} else {
			onCWEvent(NULL, NULL, 0);
		}
	}
}

@end

static CFRunLoopRef g_runLoop = NULL;
static CWListenerDelegate *g_delegate = nil;

static int start_cw_listener() {
	@autoreleasepool {
		CWWiFiClient *client = [CWWiFiClient sharedWiFiClient];
		if (!client) return -1;

		g_delegate = [[CWListenerDelegate alloc] init];
		[client setDelegate:g_delegate];

		NSError *err = nil;
		[client startMonitoringEventWithType:CWEventTypePowerDidChange error:&err];
		[client startMonitoringEventWithType:CWEventTypeSSIDDidChange error:&err];
		[client startMonitoringEventWithType:CWEventTypeBSSIDDidChange error:&err];
		[client startMonitoringEventWithType:CWEventTypeLinkDidChange error:&err];

		g_runLoop = CFRunLoopGetCurrent();
		[g_delegate notifyCurrentState];
	}
	CFRunLoopRun();
	return 0;
}

static void stop_cw_listener() {
	if (g_runLoop) {
		CFRunLoopStop(g_runLoop);
		g_runLoop = NULL;
	}
	@autoreleasepool {
		CWWiFiClient *client = [CWWiFiClient sharedWiFiClient];
		if (client) {
			[client stopMonitoringAllEventsAndReturnError:nil];
			[client setDelegate:nil];
		}
		g_delegate = nil;
	}
}
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

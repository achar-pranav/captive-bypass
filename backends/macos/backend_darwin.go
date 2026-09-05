//go:build darwin

package macos

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Foundation -framework CoreWLAN

#import <Foundation/Foundation.h>
#import <CoreWLAN/CoreWLAN.h>
#include <stdlib.h>

typedef struct {
	char* ssid;
	char* bssid;
	int signal;
	int up;
	char* err;
} CWStateResult;

typedef struct {
	char* ssid;
	char* bssid;
	int signal;
	int secured;
} CWAP;

typedef struct {
	CWAP* aps;
	int count;
	char* err;
} CWScanResult;

static CWStateResult cw_get_state() {
	CWStateResult res;
	res.ssid = NULL;
	res.bssid = NULL;
	res.signal = 0;
	res.up = 0;
	res.err = NULL;

	@autoreleasepool {
		CWWiFiClient *client = [CWWiFiClient sharedWiFiClient];
		if (!client) {
			res.err = strdup("failed to get shared CWWiFiClient");
			return res;
		}
		CWInterface *iface = [client interface];
		if (!iface) {
			res.err = strdup("no Wi-Fi interface found");
			return res;
		}

		res.up = [iface powerOn] ? 1 : 0;
		NSString *ssid = [iface ssid];
		if (ssid && [ssid length] > 0) {
			res.ssid = strdup([ssid UTF8String]);
		}
		NSString *bssid = [iface bssid];
		if (bssid && [bssid length] > 0) {
			res.bssid = strdup([bssid UTF8String]);
		}
		res.signal = (int)[iface rssiValue];
	}
	return res;
}

static CWScanResult cw_scan() {
	CWScanResult res;
	res.aps = NULL;
	res.count = 0;
	res.err = NULL;

	@autoreleasepool {
		CWWiFiClient *client = [CWWiFiClient sharedWiFiClient];
		if (!client) {
			res.err = strdup("failed to get shared CWWiFiClient");
			return res;
		}
		CWInterface *iface = [client interface];
		if (!iface) {
			res.err = strdup("no Wi-Fi interface found");
			return res;
		}

		NSError *error = nil;
		NSSet<CWNetwork *> *networks = [iface scanForNetworksWithName:nil error:&error];
		if (error) {
			res.err = strdup([[error localizedDescription] UTF8String]);
			return res;
		}

		int count = (int)[networks count];
		if (count > 0) {
			res.aps = (CWAP*)malloc(sizeof(CWAP) * count);
			int i = 0;
			for (CWNetwork *net in networks) {
				NSString *ssid = [net ssid];
				res.aps[i].ssid = (ssid && [ssid length] > 0) ? strdup([ssid UTF8String]) : strdup("");
				NSString *bssid = [net bssid];
				res.aps[i].bssid = (bssid && [bssid length] > 0) ? strdup([bssid UTF8String]) : strdup("");
				res.aps[i].signal = (int)[net rssiValue];
				res.aps[i].secured = [net supportsSecurity:kCWSecurityNone] ? 0 : 1;
				i++;
			}
			res.count = count;
		}
	}
	return res;
}
*/
import "C"

import (
	"errors"
	"unsafe"

	"github.com/achar-pranav/captive-bypass/backends"
)

type Backend struct{}

func New() *Backend {
	return &Backend{}
}

type snapshot struct {
	SSID   string
	BSSID  string
	Signal int
	Up     bool
}

func (b *Backend) snapshot() (snapshot, error) {
	res := C.cw_get_state()
	if res.err != nil {
		defer C.free(unsafe.Pointer(res.err))
		return snapshot{}, errors.New(C.GoString(res.err))
	}

	var snap snapshot
	snap.Up = res.up == 1
	snap.Signal = int(res.signal)

	if res.ssid != nil {
		snap.SSID = C.GoString(res.ssid)
		C.free(unsafe.Pointer(res.ssid))
	}
	if res.bssid != nil {
		snap.BSSID = C.GoString(res.bssid)
		C.free(unsafe.Pointer(res.bssid))
	}

	return snap, nil
}

func (b *Backend) ActiveSSID() (string, error) {
	s, err := b.snapshot()
	if err != nil {
		return "", err
	}
	if s.SSID == "" {
		return "", errors.New("not associated")
	}
	return s.SSID, nil
}

func (b *Backend) ActiveBSSID() (string, error) {
	s, err := b.snapshot()
	if err != nil {
		return "", err
	}
	if s.BSSID == "" {
		return "", errors.New("not associated")
	}
	return s.BSSID, nil
}

func (b *Backend) Signal() (int, error) {
	s, err := b.snapshot()
	return s.Signal, err
}

func (b *Backend) Up() (bool, error) {
	s, err := b.snapshot()
	return s.Up, err
}

func (b *Backend) Scan() ([]backends.AP, error) {
	res := C.cw_scan()
	if res.err != nil {
		defer C.free(unsafe.Pointer(res.err))
		return nil, errors.New(C.GoString(res.err))
	}

	count := int(res.count)
	if count == 0 || res.aps == nil {
		return []backends.AP{}, nil
	}
	defer C.free(unsafe.Pointer(res.aps))

	cAPs := (*[1 << 20]C.CWAP)(unsafe.Pointer(res.aps))[:count:count]
	out := make([]backends.AP, 0, count)

	for i := 0; i < count; i++ {
		ap := backends.AP{
			Signal:  int(cAPs[i].signal),
			Secured: cAPs[i].secured == 1,
		}
		if cAPs[i].ssid != nil {
			ap.SSID = C.GoString(cAPs[i].ssid)
			C.free(unsafe.Pointer(cAPs[i].ssid))
		}
		if cAPs[i].bssid != nil {
			ap.BSSID = C.GoString(cAPs[i].bssid)
			C.free(unsafe.Pointer(cAPs[i].bssid))
		}
		out = append(out, ap)
	}

	return out, nil
}

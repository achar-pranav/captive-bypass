//go:build darwin

#import <Foundation/Foundation.h>
#import <CoreWLAN/CoreWLAN.h>
#include <stdlib.h>
#include <string.h>

// onCWEvent is exported by Go via cgo
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

int start_cw_listener(void) {
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

void stop_cw_listener(void) {
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

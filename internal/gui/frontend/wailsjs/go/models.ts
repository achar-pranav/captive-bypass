export namespace gui {
	
	export class CredProfileInfo {
	    name: string;
	    username: string;
	    isActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CredProfileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.username = source["username"];
	        this.isActive = source["isActive"];
	    }
	}
	export class AppState {
	    status: string;
	    statusTitle: string;
	    statusSub: string;
	    activeSSID: string;
	    signalPercent: number;
	    isAutoLoginEnabled: boolean;
	    isVanguardEnabled: boolean;
	    threshold: number;
	    activeProfile: string;
	    credsCount: number;
	    recognizedNetworks: string[];
	    credProfiles: CredProfileInfo[];
	    isFirstRun: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.statusTitle = source["statusTitle"];
	        this.statusSub = source["statusSub"];
	        this.activeSSID = source["activeSSID"];
	        this.signalPercent = source["signalPercent"];
	        this.isAutoLoginEnabled = source["isAutoLoginEnabled"];
	        this.isVanguardEnabled = source["isVanguardEnabled"];
	        this.threshold = source["threshold"];
	        this.activeProfile = source["activeProfile"];
	        this.credsCount = source["credsCount"];
	        this.recognizedNetworks = source["recognizedNetworks"];
	        this.credProfiles = this.convertValues(source["credProfiles"], CredProfileInfo);
	        this.isFirstRun = source["isFirstRun"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class ScannedNetwork {
	    ssid: string;
	    signal: number;
	    secured: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ScannedNetwork(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ssid = source["ssid"];
	        this.signal = source["signal"];
	        this.secured = source["secured"];
	    }
	}

}


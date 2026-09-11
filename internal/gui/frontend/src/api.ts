import { CredSet, WifiNetwork, LEDState } from './types';

export interface BackendState {
  status: LEDState;
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
  credProfiles: Array<{
    name: string;
    username: string;
    isActive: boolean;
  }>;
  isFirstRun: boolean;
}

export interface ScannedNetwork {
  ssid: string;
  signal: number;
  secured: boolean;
}

// Global declaration for Wails runtime injection
declare global {
  interface Window {
    go?: {
      gui?: {
        App?: {
          GetState: () => Promise<BackendState>;
          ScanNetworks: () => Promise<ScannedNetwork[]>;
          AddSSID: (ssid: string) => Promise<void>;
          RemoveSSID: (ssid: string) => Promise<void>;
          SaveCreds: (username: string, password: string, setActive: boolean) => Promise<void>;
          GetCreds: (name: string) => Promise<{ name: string; username: string; password?: string }>;
          DeleteCreds: (name: string) => Promise<void>;
          SetActiveCred: (name: string) => Promise<void>;
          ToggleAutoLogin: (enabled: boolean) => Promise<void>;
          ToggleVanguard: (enabled: boolean) => Promise<void>;
          SetThreshold: (threshold: number) => Promise<void>;
          ManualLogin: () => Promise<string>;
          ManualLogout: () => Promise<void>;
          FinishWizard: (ssids: string[]) => Promise<void>;
        };
      };
    };
    runtime?: {
      Quit: () => void;
      WindowMinimise: () => void;
    };
  }
}

// Check if running inside Wails desktop runtime
export const isWails = (): boolean => {
  return typeof window !== 'undefined' && !!window.go?.gui?.App;
};

// API Bridge with fallback for standalone browser preview
export const api = {
  async getState(): Promise<BackendState | null> {
    if (isWails()) {
      return await window.go!.gui!.App!.GetState();
    }
    return null;
  },

  async scanNetworks(): Promise<ScannedNetwork[]> {
    if (isWails()) {
      return await window.go!.gui!.App!.ScanNetworks();
    }
    // Browser preview simulated scan
    return [
      { ssid: 'PESU-RR', signal: 92, secured: true },
      { ssid: 'ELEMENT BLOCK', signal: 85, secured: true },
      { ssid: 'PES-Guest', signal: 64, secured: false },
      { ssid: 'Campus-IoT', signal: 45, secured: true },
      { ssid: 'Hostel-Mesh-5G', signal: 12, secured: true }
    ];
  },

  async addSSID(ssid: string): Promise<void> {
    if (isWails()) {
      await window.go!.gui!.App!.AddSSID(ssid);
    }
  },

  async removeSSID(ssid: string): Promise<void> {
    if (isWails()) {
      await window.go!.gui!.App!.RemoveSSID(ssid);
    }
  },

  async saveCreds(username: string, password: string, setActive: boolean = false): Promise<void> {
    if (isWails()) {
      await window.go!.gui!.App!.SaveCreds(username, password, setActive);
    }
  },

  async getCreds(name: string): Promise<{ name: string; username: string; password?: string } | null> {
    if (isWails()) {
      return await window.go!.gui!.App!.GetCreds(name);
    }
    return null;
  },

  async deleteCreds(name: string): Promise<void> {
    if (isWails()) {
      await window.go!.gui!.App!.DeleteCreds(name);
    }
  },

  async setActiveCred(name: string): Promise<void> {
    if (isWails()) {
      await window.go!.gui!.App!.SetActiveCred(name);
    }
  },

  async toggleAutoLogin(enabled: boolean): Promise<void> {
    if (isWails()) {
      await window.go!.gui!.App!.ToggleAutoLogin(enabled);
    }
  },

  async toggleVanguard(enabled: boolean): Promise<void> {
    if (isWails()) {
      await window.go!.gui!.App!.ToggleVanguard(enabled);
    }
  },

  async setThreshold(threshold: number): Promise<void> {
    if (isWails()) {
      await window.go!.gui!.App!.SetThreshold(threshold);
    }
  },

  async manualLogin(): Promise<string> {
    if (isWails()) {
      return await window.go!.gui!.App!.ManualLogin();
    }
    return 'Logged in successfully';
  },

  async manualLogout(): Promise<void> {
    if (isWails()) {
      await window.go!.gui!.App!.ManualLogout();
    }
  },

  async finishWizard(ssids: string[]): Promise<void> {
    if (isWails()) {
      await window.go!.gui!.App!.FinishWizard(ssids);
    }
  }
};

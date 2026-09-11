export type LEDState = 'green' | 'red' | 'yellow' | 'orange';

export interface CredSet {
  id: string;
  username: string;
  password?: string;
}

export interface WifiNetwork {
  ssid: string;
  signal: number; // 0 - 100%
  inRange: boolean;
}

export interface ToastMessage {
  id: string;
  text: string;
  type?: 'info' | 'success' | 'warn' | 'error';
}

export type WizardStep = 'permissions' | 'credentials' | 'ssids' | 'main';

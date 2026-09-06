import React, { useState, useEffect, useRef } from 'react';
import { RefreshCw, Trash2, Shield, Wifi, Key, Check, Info, Settings, AlertTriangle, Eye, EyeOff, Pencil, ArrowLeft } from 'lucide-react';
import { CredSet, WifiNetwork, LEDState, ToastMessage, WizardStep } from './types';
import { BottomToast } from './components/BottomToast';

const DEFAULT_NETWORKS: WifiNetwork[] = [
  { ssid: 'ELEMENT BLOCK', signal: 88, inRange: true },
  { ssid: 'PESU-STUDENT', signal: 72, inRange: true },
  { ssid: 'PES-CAMPUS-5G', signal: 45, inRange: true },
  { ssid: 'PES-FACULTY', signal: 12, inRange: true },
  { ssid: 'PESU-GUEST', signal: 60, inRange: true },
];

export default function App() {
  // Navigation & Screen State
  const [step, setStep] = useState<WizardStep>('permissions');

  // Configuration State
  const [activeCredSetId, setActiveCredSetId] = useState<string>('cred-1');
  const [credSets, setCredSets] = useState<CredSet[]>([
    { id: 'cred-1', name: 'default', username: 'PES1UG23CS001', password: 'password123' },
  ]);
  const [savedSSIDs, setSavedSSIDs] = useState<string[]>(['ELEMENT BLOCK']);
  const [captiveBypassEnabled, setCaptiveBypassEnabled] = useState<boolean>(true);
  const [vanguardEnabled, setVanguardEnabled] = useState<boolean>(false);
  const [edgeThreshold, setEdgeThreshold] = useState<number>(15);

  // Scanner & Network state
  const [networks, setNetworks] = useState<WifiNetwork[]>(DEFAULT_NETWORKS);
  const [isScanning, setIsScanning] = useState<boolean>(false);

  // LED State: 'green' | 'red' | 'yellow' | 'orange'
  const [ledState, setLedState] = useState<LEDState>('green');
  const [currentSignal, setCurrentSignal] = useState<number>(85);

  // Modals inside Main Menu
  const [activeModal, setActiveModal] = useState<'none' | 'add_ssid' | 'manage_ssid' | 'add_cred' | 'manage_cred' | 'edit_cred'>('none');
  const [editingCredId, setEditingCredId] = useState<string | null>(null);

  // Form states for Wizard & Add/Edit dialogs
  const [formCredName, setFormCredName] = useState<string>('default');
  const [formUsername, setFormUsername] = useState<string>('');
  const [formPassword, setFormPassword] = useState<string>('');
  const [showPassword, setShowPassword] = useState<boolean>(false);
  const [selectedSSIDsStaging, setSelectedSSIDsStaging] = useState<string[]>(['ELEMENT BLOCK']);

  // Bottom Toast Notifications
  const [toast, setToast] = useState<ToastMessage | null>(null);
  const toastTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Anti-spam click lock
  const [isActionLocked, setIsActionLocked] = useState<boolean>(false);

  const triggerToast = (text: string, type: 'info' | 'success' | 'warn' | 'error' = 'info') => {
    if (toastTimeoutRef.current) {
      clearTimeout(toastTimeoutRef.current);
    }
    setToast({ id: Date.now().toString(), text, type });
    toastTimeoutRef.current = setTimeout(() => {
      setToast(null);
    }, 3500);
  };

  const dismissToast = () => {
    if (toastTimeoutRef.current) {
      clearTimeout(toastTimeoutRef.current);
    }
    setToast(null);
  };

  // Anti-spam wrapper
  const handleAntiSpam = (fn: () => void) => {
    if (isActionLocked) return;
    setIsActionLocked(true);
    fn();
    setTimeout(() => {
      setIsActionLocked(false);
    }, 400);
  };

  // Scan networks
  const handleRefreshNetworks = () => {
    handleAntiSpam(() => {
      setIsScanning(true);
      triggerToast('Scanning nearby Wi-Fi broadcast...', 'info');
      setTimeout(() => {
        setNetworks([
          { ssid: 'ELEMENT BLOCK', signal: Math.floor(65 + Math.random() * 30), inRange: true },
          { ssid: 'PESU-STUDENT', signal: Math.floor(50 + Math.random() * 40), inRange: true },
          { ssid: 'PES-CAMPUS-5G', signal: Math.floor(30 + Math.random() * 50), inRange: true },
          { ssid: 'PES-FACULTY', signal: Math.floor(10 + Math.random() * 25), inRange: true },
          { ssid: 'PESU-GUEST', signal: Math.floor(40 + Math.random() * 40), inRange: true },
        ]);
        setIsScanning(false);
        triggerToast('Network scan updated', 'success');
      }, 700);
    });
  };

  // Toggle SSID checkbox in staging
  const toggleStagingSSID = (ssid: string) => {
    if (selectedSSIDsStaging.includes(ssid)) {
      setSelectedSSIDsStaging(selectedSSIDsStaging.filter((s) => s !== ssid));
    } else {
      setSelectedSSIDsStaging([...selectedSSIDsStaging, ssid]);
    }
  };

  // Edge of network detection simulation
  const checkSignalThreshold = (signal: number) => {
    setCurrentSignal(signal);
    if (signal <= edgeThreshold) {
      setLedState('orange');
      if (vanguardEnabled) {
        triggerToast(`[Vanguard Alert] Weak Wi-Fi (${signal}% ≤ ${edgeThreshold}%). Pre-buffering roaming session.`, 'warn');
      } else {
        triggerToast(`Edge of network reached (${signal}% ≤ ${edgeThreshold}%).`, 'warn');
      }
    } else if (!captiveBypassEnabled) {
      setLedState('red');
    } else {
      setLedState('green');
    }
  };

  // Compute LED info
  const getLEDInfo = () => {
    switch (ledState) {
      case 'green':
        return {
          colorClass: 'bg-[#23A55A] shadow-[0_0_12px_#23A55A]',
          text: '<Connected>',
          subtext: 'Connected & Online',
          desc: 'Connected & Online',
        };
      case 'red':
        return {
          colorClass: 'bg-[#F23F43] shadow-[0_0_12px_#F23F43]',
          text: '<Disconnected>',
          subtext: 'Disconnected / Offline',
          desc: 'Disconnected / Offline',
        };
      case 'yellow':
        return {
          colorClass: 'bg-[#FEE75C] shadow-[0_0_12px_#FEE75C]',
          text: '<In Progress>',
          subtext: 'Authenticating…',
          desc: 'Authenticating with Portal…',
        };
      case 'orange':
        return {
          colorClass: 'bg-[#FF9900] shadow-[0_0_12px_#FF9900]',
          text: '<Network Edge>',
          subtext: "You're at the edge",
          desc: "You're at the edge",
        };
    }
  };

  const ledInfo = getLEDInfo();

  return (
    <div className="min-h-screen bg-[#000000] text-[#E6EAED] flex flex-col items-center justify-center p-4 selection:bg-[#00A8FF] selection:text-black">
      {/* Outer App Container simulating AMOLED Desktop Window */}
      <div className="w-full max-w-[430px] bg-[#000000] border border-[#1E1F22] rounded-xl shadow-2xl overflow-hidden flex flex-col relative min-h-[580px]">
        {/* Subtle Window Header */}
        <div className="bg-[#0D0E10] px-4 py-2.5 border-b border-[#1E1F22] flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <span className="w-2.5 h-2.5 rounded-full bg-[#2B2D31]" />
            <span className="w-2.5 h-2.5 rounded-full bg-[#2B2D31]" />
            <span className="w-2.5 h-2.5 rounded-full bg-[#2B2D31]" />
            <span className="text-[11px] font-mono text-[#7A828A] tracking-wider ml-1">captive-bypass</span>
          </div>
          {step === 'main' && (
            <button
              onClick={() => handleAntiSpam(() => setStep('permissions'))}
              className="text-[10px] text-[#7A828A] hover:text-white px-2 py-0.5 rounded border border-[#2B2D31] hover:border-white/40 transition-colors"
              title="Reset and review wizard"
            >
              Rerun Setup
            </button>
          )}
        </div>

        {/* ========================================================================= */}
        {/* SCREEN 1: WIZARD PERMISSIONS / TRUST                                      */}
        {/* ========================================================================= */}
        {step === 'permissions' && (
          <div className="p-6 flex-1 flex flex-col justify-between">
            <div className="space-y-5">
              <div className="flex items-center space-x-3">
                <div className="w-9 h-9 rounded-lg bg-[#00A8FF]/10 flex items-center justify-center text-[#00A8FF]">
                  <Shield className="w-5 h-5" />
                </div>
                <div>
                  <h1 className="text-lg font-bold text-white tracking-tight">Permissions & Trust</h1>
                  <p className="text-xs text-[#7A828A]">System access for seamless Wi-Fi auto-login</p>
                </div>
              </div>

              {/* Trust Box */}
              <div className="border border-[#1E1F22] bg-[#0A0B0D] rounded-lg p-4 space-y-3 text-xs text-[#949BA4] leading-relaxed">
                <p>
                  <strong className="text-white">Why permissions matter:</strong> On macOS and Linux, reading active
                  Wi-Fi SSIDs and observing network transitions requires local interface permissions.
                </p>
                <p>
                  captive-bypass is 100% open-source, runs fully locally, and never exposes credentials outside of the
                  official portal handshake.
                </p>
                <div className="pt-1 flex items-center space-x-3 text-[11px]">
                  <a
                    href="https://github.com/achar-pranav/captive-bypass"
                    target="_blank"
                    rel="noreferrer"
                    className="text-[#00A8FF] hover:underline"
                  >
                    View Source on GitHub ↗
                  </a>
                  <span className="text-[#2B2D31]">•</span>
                  <span className="text-[#7A828A]">Zero Root Required</span>
                </div>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="space-y-2.5 pt-6">
              <button
                onClick={() => handleAntiSpam(() => setStep('credentials'))}
                className="w-full py-2.5 px-4 rounded-lg bg-[#00A8FF] text-black font-semibold text-sm hover:bg-[#33BAFF] active:bg-[#0090DC] transition-colors"
              >
                Continue
              </button>
              <button
                onClick={() => handleAntiSpam(() => setStep('main'))}
                className="w-full py-2.5 px-4 rounded-lg bg-black text-white text-sm border border-white/20 hover:border-white/40 hover:bg-white/5 active:bg-white/10 transition-all"
              >
                Skip
              </button>
            </div>
          </div>
        )}

        {/* ========================================================================= */}
        {/* SCREEN 2: WIZARD ADD CREDENTIALS                                          */}
        {/* ========================================================================= */}
        {step === 'credentials' && (
          <div className="p-6 flex-1 flex flex-col justify-between">
            <div className="space-y-4">
              <div>
                <h1 className="text-lg font-bold text-white tracking-tight">Add Credentials</h1>
                <p className="text-xs text-[#7A828A]">Set up your PESU portal profile</p>
              </div>

              {/* 3 Fields */}
              <div className="space-y-3 pt-2">
                <div>
                  <label className="block text-[11px] font-medium text-[#949BA4] mb-1">1. Credentials Name</label>
                  <input
                    type="text"
                    value={formCredName}
                    onChange={(e) => setFormCredName(e.target.value)}
                    placeholder="default"
                    className="w-full bg-[#0E1012] border border-[#2B2D31] rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-[#00A8FF] transition-colors"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-medium text-[#949BA4] mb-1">2. Username (SRN)</label>
                  <input
                    type="text"
                    value={formUsername}
                    onChange={(e) => setFormUsername(e.target.value)}
                    placeholder="PES1UG23CS..."
                    className="w-full bg-[#0E1012] border border-[#2B2D31] rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-[#00A8FF] transition-colors"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-medium text-[#949BA4] mb-1">3. Password</label>
                  <div className="relative">
                    <input
                      type={showPassword ? 'text' : 'password'}
                      value={formPassword}
                      onChange={(e) => setFormPassword(e.target.value)}
                      placeholder="Portal Password"
                      className="w-full bg-[#0E1012] border border-[#2B2D31] rounded-lg pl-3 pr-10 py-2 text-sm text-white focus:outline-none focus:border-[#00A8FF] transition-colors"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-[#7A828A] hover:text-white transition-colors"
                      title={showPassword ? 'Hide password' : 'Show password'}
                    >
                      {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                    </button>
                  </div>
                </div>
              </div>

              {/* Bottom Disclaimer */}
              <div className="pt-2">
                <p className="text-[11px] text-[#7A828A] leading-relaxed italic border-l-2 border-[#00A8FF]/40 pl-2.5">
                  Passwords are never stored in plaintext. We use OS hardware fingerprinting and AES-GCM encryption to
                  prevent theft by copy.
                </p>
              </div>
            </div>

            {/* Back & Continue Buttons */}
            <div className="space-y-2.5 pt-6">
              <button
                type="button"
                onClick={() => handleAntiSpam(() => setStep('permissions'))}
                className="w-full py-2.5 px-4 rounded-lg bg-black text-white text-sm border border-white/20 hover:border-white/40 hover:bg-white/5 active:bg-white/10 transition-all flex items-center justify-center space-x-1.5"
              >
                <ArrowLeft className="w-4 h-4" />
                <span>Back</span>
              </button>
              <button
                onClick={() =>
                  handleAntiSpam(() => {
                    if (!formUsername.trim()) {
                      triggerToast('Please enter your SRN username', 'error');
                      return;
                    }
                    const newId = `cred-${Date.now()}`;
                    const newSet: CredSet = {
                      id: newId,
                      name: formCredName.trim() || 'default',
                      username: formUsername.trim(),
                      password: formPassword,
                    };
                    setCredSets([newSet]);
                    setActiveCredSetId(newId);
                    setStep('ssids');
                    triggerToast('Credentials encrypted & saved', 'success');
                  })
                }
                className="w-full py-2.5 px-4 rounded-lg bg-[#00A8FF] text-black font-semibold text-sm hover:bg-[#33BAFF] active:bg-[#0090DC] transition-colors"
              >
                Continue
              </button>
            </div>
          </div>
        )}

        {/* ========================================================================= */}
        {/* SCREEN 3: WIZARD SELECT SSIDs                                             */}
        {/* ========================================================================= */}
        {step === 'ssids' && (
          <div className="p-6 flex-1 flex flex-col justify-between">
            <div className="space-y-3">
              <div>
                <h1 className="text-lg font-bold text-white tracking-tight">
                  Select at least one SSID to log into automatically
                </h1>
                <p className="text-xs text-[#7A828A]">Networks that trigger auto-login</p>
              </div>

              {/* Networks Title Card with Refresh */}
              <div className="pt-2">
                <div className="flex items-center justify-between pb-2 border-b border-[#1E1F22]">
                  <span className="text-xs font-semibold text-white tracking-wide">Networks</span>
                  <button
                    onClick={handleRefreshNetworks}
                    disabled={isScanning}
                    className="flex items-center space-x-1 text-xs text-[#00A8FF] hover:text-[#33BAFF] px-2 py-0.5 rounded border border-[#00A8FF]/30 hover:border-[#00A8FF] transition-colors"
                  >
                    <span>Networks</span>
                    <span className={isScanning ? 'animate-spin' : ''}>⟳</span>
                  </button>
                </div>

                {/* List of SSIDs with signal strength and checkbox */}
                <div className="space-y-1.5 max-h-[220px] overflow-y-auto pt-2 pr-1">
                  {networks.map((net) => {
                    const checked = selectedSSIDsStaging.includes(net.ssid);
                    return (
                      <div
                        key={net.ssid}
                        onClick={() => toggleStagingSSID(net.ssid)}
                        className={`flex items-center justify-between p-2.5 rounded-lg border cursor-pointer transition-colors ${
                          checked
                            ? 'bg-[#00A8FF]/10 border-[#00A8FF]/60 text-white'
                            : 'bg-[#0B0C0E] border-[#1E1F22] text-[#949BA4] hover:border-white/20'
                        }`}
                      >
                        <div className="flex items-center space-x-2.5">
                          <input
                            type="checkbox"
                            checked={checked}
                            onChange={() => {}}
                            className="w-4 h-4 rounded accent-[#00A8FF] cursor-pointer"
                          />
                          <span className="text-xs font-mono font-medium text-white">{net.ssid}</span>
                        </div>
                        <span className="text-[11px] font-mono text-[#7A828A]">{net.signal}%</span>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>

            {/* Back & Done Buttons */}
            <div className="space-y-2.5 pt-6">
              <button
                type="button"
                onClick={() => handleAntiSpam(() => setStep('credentials'))}
                className="w-full py-2.5 px-4 rounded-lg bg-black text-white text-sm border border-white/20 hover:border-white/40 hover:bg-white/5 active:bg-white/10 transition-all flex items-center justify-center space-x-1.5"
              >
                <ArrowLeft className="w-4 h-4" />
                <span>Back</span>
              </button>
              <button
                onClick={() =>
                  handleAntiSpam(() => {
                    if (selectedSSIDsStaging.length === 0) {
                      triggerToast('Select at least one SSID to continue', 'error');
                      return;
                    }
                    setSavedSSIDs(selectedSSIDsStaging);
                    setStep('main');
                    triggerToast('Setup finished! captive-bypass active.', 'success');
                  })
                }
                className="w-full py-2.5 px-4 rounded-lg bg-[#00A8FF] text-black font-semibold text-sm hover:bg-[#33BAFF] active:bg-[#0090DC] transition-colors"
              >
                Done
              </button>
            </div>
          </div>
        )}

        {/* ========================================================================= */}
        {/* SCREEN 4: MAIN MENU                                                       */}
        {/* ========================================================================= */}
        {step === 'main' && (
          <div className="p-5 flex-1 flex flex-col justify-between space-y-4">
            {/* Top Status LED & Title */}
            <div className="flex items-center justify-between bg-[#0A0C0E] border border-[#1E1F22] px-3.5 py-2.5 rounded-lg">
              <div className="flex items-center space-x-3 min-w-0">
                {/* 4-State LED on Top Left */}
                <div
                  className={`w-3.5 h-3.5 flex-shrink-0 rounded-full transition-all duration-300 ${ledInfo.colorClass}`}
                  title={`Status: ${ledInfo.desc}`}
                />
                <div className="min-w-0 flex flex-col justify-center">
                  <span className="text-xs font-bold text-white tracking-wide font-mono whitespace-nowrap">
                    {ledInfo.text}
                  </span>
                  <span className="text-[11px] text-[#7A828A] font-mono truncate leading-tight">
                    {ledInfo.subtext}
                  </span>
                </div>
              </div>

              {/* Status State Cycler for interactive testing */}
              <button
                onClick={() => {
                  const states: LEDState[] = ['green', 'red', 'yellow', 'orange'];
                  const next = states[(states.indexOf(ledState) + 1) % states.length];
                  setLedState(next);
                  if (next === 'orange') {
                    if (vanguardEnabled) {
                      triggerToast('[Vanguard Alert] Weak Wi-Fi detected. Session pre-caching.', 'warn');
                    } else {
                      triggerToast('Edge of network reached.', 'warn');
                    }
                  } else if (next === 'green') {
                    triggerToast('Portal authenticated & connected', 'success');
                  } else if (next === 'red') {
                    triggerToast('Disconnected from portal', 'error');
                  } else {
                    triggerToast('Authenticating with Cyberoam portal...', 'info');
                  }
                }}
                className="text-[10px] text-[#7A828A] hover:text-white px-1.5 py-0.5 rounded border border-[#2B2D31] hover:border-white/40 transition-colors"
                title="Click to cycle LED states"
              >
                Test LED
              </button>
            </div>

            {/* Two Options: Networks and Creds (Title cards one below the other) */}
            <div className="space-y-3">
              {/* Option 1: Networks */}
              <div className="bg-[#0A0C0E] border border-[#1E1F22] rounded-lg p-3.5">
                <div className="flex items-center justify-between">
                  <div>
                    <h2 className="text-sm font-bold text-white">Networks</h2>
                    <p className="text-[11px] text-[#7A828A]">
                      {savedSSIDs.length} registered {savedSSIDs.length === 1 ? 'network' : 'networks'}
                    </p>
                  </div>
                  <div className="flex items-center space-x-2">
                    <button
                      onClick={() =>
                        handleAntiSpam(() => {
                          setSelectedSSIDsStaging([...savedSSIDs]);
                          setActiveModal('add_ssid');
                        })
                      }
                      className="px-3 py-1 text-xs font-semibold rounded-md bg-[#00A8FF] text-black hover:bg-[#33BAFF] active:bg-[#0090DC] transition-colors"
                    >
                      Add
                    </button>
                    <button
                      onClick={() => handleAntiSpam(() => setActiveModal('manage_ssid'))}
                      className="px-3 py-1 text-xs rounded-md bg-black text-white border border-white/20 hover:border-white/40 hover:bg-white/5 active:bg-white/10 transition-colors"
                    >
                      Manage
                    </button>
                  </div>
                </div>

                {/* Quick preview list of saved SSIDs */}
                <div className="mt-2.5 flex flex-wrap gap-1.5">
                  {savedSSIDs.map((s) => (
                    <span
                      key={s}
                      className="inline-flex items-center text-[10px] font-mono bg-[#141619] border border-[#2B2D31] text-[#949BA4] px-2 py-0.5 rounded"
                    >
                      {s}
                    </span>
                  ))}
                  {savedSSIDs.length === 0 && (
                    <span className="text-[11px] text-[#7A828A] italic">No networks registered.</span>
                  )}
                </div>
              </div>

              {/* Option 2: Creds */}
              <div className="bg-[#0A0C0E] border border-[#1E1F22] rounded-lg p-3.5">
                <div className="flex items-center justify-between">
                  <div>
                    <h2 className="text-sm font-bold text-white">Creds</h2>
                    <p className="text-[11px] text-[#7A828A]">
                      Active:{' '}
                      <span className="text-[#00A8FF] font-mono">
                        {credSets.find((c) => c.id === activeCredSetId)?.name || 'None'}
                      </span>
                    </p>
                  </div>
                  <div className="flex items-center space-x-2">
                    <button
                      onClick={() =>
                        handleAntiSpam(() => {
                          setFormCredName('');
                          setFormUsername('');
                          setFormPassword('');
                          setActiveModal('add_cred');
                        })
                      }
                      className="px-3 py-1 text-xs font-semibold rounded-md bg-[#00A8FF] text-black hover:bg-[#33BAFF] active:bg-[#0090DC] transition-colors"
                    >
                      Add
                    </button>
                    <button
                      onClick={() => handleAntiSpam(() => setActiveModal('manage_cred'))}
                      className="px-3 py-1 text-xs rounded-md bg-black text-white border border-white/20 hover:border-white/40 hover:bg-white/5 active:bg-white/10 transition-colors"
                    >
                      Manage
                    </button>
                  </div>
                </div>

                {/* Quick preview of active username */}
                <div className="mt-2.5 flex items-center space-x-2 text-[11px] font-mono text-[#949BA4]">
                  <span>SRN:</span>
                  <span className="text-white">
                    {credSets.find((c) => c.id === activeCredSetId)?.username || 'None configured'}
                  </span>
                </div>
              </div>
            </div>

            {/* Bottom Controls: Two Checkboxes */}
            <div className="border-t border-[#1E1F22] pt-3.5 space-y-2.5">
              <label className="flex items-center space-x-2.5 cursor-pointer select-none">
                <input
                  type="checkbox"
                  checked={captiveBypassEnabled}
                  onChange={(e) => {
                    setCaptiveBypassEnabled(e.target.checked);
                    if (!e.target.checked) {
                      setLedState('red');
                      triggerToast('captive-bypass disabled', 'warn');
                    } else {
                      setLedState('green');
                      triggerToast('captive-bypass enabled', 'success');
                    }
                  }}
                  className="w-4 h-4 rounded accent-[#00A8FF] cursor-pointer"
                />
                <span className="text-xs text-white">Enable/Disable the captive-bypass</span>
              </label>

              <label className="flex items-center space-x-2.5 cursor-pointer select-none">
                <input
                  type="checkbox"
                  checked={vanguardEnabled}
                  onChange={(e) => {
                    setVanguardEnabled(e.target.checked);
                    triggerToast(
                      e.target.checked ? 'Vanguard telemetry enabled' : 'Vanguard telemetry disabled',
                      'info'
                    );
                  }}
                  className="w-4 h-4 rounded accent-[#00A8FF] cursor-pointer"
                />
                <span className="text-xs text-white">Enable/Disable Vanguard(experimental)</span>
              </label>
            </div>

            {/* Signal & Threshold Testing Strip */}
            <div className="bg-[#0A0B0D] border border-[#1E1F22] rounded-lg p-2.5 text-[11px] space-y-1.5">
              <div className="flex items-center justify-between text-[#7A828A]">
                <span>Edge threshold: {edgeThreshold}%</span>
                <span className="font-mono text-[#00A8FF]">--set-threshold {edgeThreshold}</span>
              </div>
              <input
                type="range"
                min="1"
                max="100"
                value={edgeThreshold}
                onChange={(e) => {
                  const val = Number(e.target.value);
                  setEdgeThreshold(val);
                  checkSignalThreshold(currentSignal);
                }}
                className="w-full accent-[#00A8FF] h-1.5 bg-[#1E1F22] rounded-lg appearance-none cursor-pointer"
              />
            </div>
          </div>
        )}

        {/* ========================================================================= */}
        {/* MODAL 1: ADD NETWORK                                                      */}
        {/* ========================================================================= */}
        {activeModal === 'add_ssid' && (
          <div className="absolute inset-0 bg-black/90 backdrop-blur-sm z-30 p-5 flex flex-col justify-between animate-in fade-in duration-150">
            <div className="space-y-3">
              <div className="flex items-center justify-between border-b border-[#1E1F22] pb-2">
                <span className="text-sm font-bold text-white">Add SSID</span>
                <button
                  onClick={handleRefreshNetworks}
                  disabled={isScanning}
                  className="flex items-center space-x-1 text-xs text-[#00A8FF] hover:text-[#33BAFF] px-2 py-0.5 rounded border border-[#00A8FF]/30 hover:border-[#00A8FF]"
                >
                  <span>Networks</span>
                  <span className={isScanning ? 'animate-spin' : ''}>⟳</span>
                </button>
              </div>

              <div className="space-y-1.5 max-h-[300px] overflow-y-auto pr-1">
                {networks.map((net) => {
                  const checked = selectedSSIDsStaging.includes(net.ssid);
                  return (
                    <div
                      key={net.ssid}
                      onClick={() => toggleStagingSSID(net.ssid)}
                      className={`flex items-center justify-between p-2.5 rounded-lg border cursor-pointer ${
                        checked
                          ? 'bg-[#00A8FF]/10 border-[#00A8FF]/60 text-white'
                          : 'bg-[#0B0C0E] border-[#1E1F22] text-[#949BA4] hover:border-white/20'
                      }`}
                    >
                      <div className="flex items-center space-x-2.5">
                        <input
                          type="checkbox"
                          checked={checked}
                          onChange={() => {}}
                          className="w-4 h-4 rounded accent-[#00A8FF]"
                        />
                        <span className="text-xs font-mono font-medium text-white">{net.ssid}</span>
                      </div>
                      <span className="text-[11px] font-mono text-[#7A828A]">{net.signal}%</span>
                    </div>
                  );
                })}
              </div>
            </div>

            <div className="flex items-center space-x-2 pt-4">
              <button
                onClick={() =>
                  handleAntiSpam(() => {
                    setSavedSSIDs(selectedSSIDsStaging);
                    setActiveModal('none');
                    triggerToast('Recognized networks updated', 'success');
                  })
                }
                className="flex-1 py-2 px-3 rounded-lg bg-[#00A8FF] text-black font-semibold text-xs hover:bg-[#33BAFF]"
              >
                Save
              </button>
              <button
                onClick={() => setActiveModal('none')}
                className="py-2 px-4 rounded-lg bg-black text-white text-xs border border-white/20 hover:border-white/40"
              >
                Cancel
              </button>
            </div>
          </div>
        )}

        {/* ========================================================================= */}
        {/* MODAL 2: MANAGE NETWORKS (with red square trash button)                   */}
        {/* ========================================================================= */}
        {activeModal === 'manage_ssid' && (
          <div className="absolute inset-0 bg-black/90 backdrop-blur-sm z-30 p-5 flex flex-col justify-between animate-in fade-in duration-150">
            <div className="space-y-3">
              <div className="border-b border-[#1E1F22] pb-2 flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-bold text-white">Manage Networks</h3>
                  <p className="text-[11px] text-[#7A828A]">Delete saved SSIDs (in or out of range)</p>
                </div>
                <span className="text-xs font-mono text-[#00A8FF]">{savedSSIDs.length} saved</span>
              </div>

              <div className="space-y-2 max-h-[320px] overflow-y-auto pr-1">
                {savedSSIDs.map((ssid) => (
                  <div
                    key={ssid}
                    className="flex items-center justify-between p-2.5 rounded-lg bg-[#0B0C0E] border border-[#1E1F22]"
                  >
                    <span className="text-xs font-mono text-white">{ssid}</span>
                    {/* Small red square trash button */}
                    <button
                      onClick={() =>
                        handleAntiSpam(() => {
                          const updated = savedSSIDs.filter((s) => s !== ssid);
                          setSavedSSIDs(updated);
                          triggerToast(`Deleted network: ${ssid}`, 'info');
                        })
                      }
                      className="w-7 h-7 bg-[#F23F43] hover:bg-[#FF4D50] active:bg-[#D8363A] rounded flex items-center justify-center text-white transition-colors"
                      title="Delete network"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                ))}

                {savedSSIDs.length === 0 && (
                  <div className="text-xs text-[#7A828A] italic py-4 text-center">No saved networks left.</div>
                )}
              </div>
            </div>

            <div className="pt-4">
              <button
                onClick={() => setActiveModal('none')}
                className="w-full py-2 px-3 rounded-lg bg-[#00A8FF] text-black font-semibold text-xs hover:bg-[#33BAFF]"
              >
                Done
              </button>
            </div>
          </div>
        )}

        {/* ========================================================================= */}
        {/* MODAL 3: ADD CREDS (3 fields + disclaimer + save)                          */}
        {/* ========================================================================= */}
        {activeModal === 'add_cred' && (
          <div className="absolute inset-0 bg-black/90 backdrop-blur-sm z-30 p-5 flex flex-col justify-between animate-in fade-in duration-150">
            <div className="space-y-3">
              <div>
                <h3 className="text-sm font-bold text-white">Add Credential Set</h3>
                <p className="text-[11px] text-[#7A828A]">Configure a profile for portal login</p>
              </div>

              <div className="space-y-2.5 pt-1">
                <div>
                  <label className="block text-[11px] font-medium text-[#949BA4] mb-1">1. Credentials Name</label>
                  <input
                    type="text"
                    value={formCredName}
                    onChange={(e) => setFormCredName(e.target.value)}
                    placeholder="e.g. personal, campus, lab"
                    className="w-full bg-[#0E1012] border border-[#2B2D31] rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-[#00A8FF]"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-medium text-[#949BA4] mb-1">2. Username (SRN)</label>
                  <input
                    type="text"
                    value={formUsername}
                    onChange={(e) => setFormUsername(e.target.value)}
                    placeholder="PES1UG23..."
                    className="w-full bg-[#0E1012] border border-[#2B2D31] rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-[#00A8FF]"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-medium text-[#949BA4] mb-1">3. Password</label>
                  <div className="relative">
                    <input
                      type={showPassword ? 'text' : 'password'}
                      value={formPassword}
                      onChange={(e) => setFormPassword(e.target.value)}
                      placeholder="Portal password"
                      className="w-full bg-[#0E1012] border border-[#2B2D31] rounded-lg pl-3 pr-9 py-1.5 text-xs text-white focus:outline-none focus:border-[#00A8FF]"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-[#7A828A] hover:text-white transition-colors"
                      title={showPassword ? 'Hide password' : 'Show password'}
                    >
                      {showPassword ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
                    </button>
                  </div>
                </div>
              </div>

              <p className="text-[10px] text-[#7A828A] italic leading-relaxed border-l-2 border-[#00A8FF]/40 pl-2">
                Passwords are not stored in plaintext; protected with local machine-derived keys.
              </p>
            </div>

            <div className="flex items-center space-x-2 pt-4">
              <button
                onClick={() =>
                  handleAntiSpam(() => {
                    if (!formUsername.trim()) {
                      triggerToast('Username (SRN) required', 'error');
                      return;
                    }
                    const newId = `cred-${Date.now()}`;
                    const newSet: CredSet = {
                      id: newId,
                      name: formCredName.trim() || `profile-${credSets.length + 1}`,
                      username: formUsername.trim(),
                      password: formPassword,
                    };
                    setCredSets([...credSets, newSet]);
                    setActiveCredSetId(newId);
                    setActiveModal('none');
                    triggerToast(`Added & activated: ${newSet.name}`, 'success');
                  })
                }
                className="flex-1 py-2 px-3 rounded-lg bg-[#00A8FF] text-black font-semibold text-xs hover:bg-[#33BAFF]"
              >
                Save
              </button>
              <button
                onClick={() => setActiveModal('none')}
                className="py-2 px-4 rounded-lg bg-black text-white text-xs border border-white/20 hover:border-white/40"
              >
                Cancel
              </button>
            </div>
          </div>
        )}

        {/* ========================================================================= */}
        {/* MODAL 4: MANAGE CREDS (Radio button + Edit pencil + red trash button)     */}
        {/* ========================================================================= */}
        {activeModal === 'manage_cred' && (
          <div className="absolute inset-0 bg-black/90 backdrop-blur-sm z-30 p-5 flex flex-col justify-between animate-in fade-in duration-150">
            <div className="space-y-3">
              <div className="border-b border-[#1E1F22] pb-2 flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-bold text-white">Manage Credentials</h3>
                  <p className="text-[11px] text-[#7A828A]">Pick active set, edit, or delete</p>
                </div>
                <span className="text-xs font-mono text-[#00A8FF]">{credSets.length} sets</span>
              </div>

              <div className="space-y-2 max-h-[300px] overflow-y-auto pr-1">
                {credSets.map((cred) => {
                  const isActive = cred.id === activeCredSetId;
                  return (
                    <div
                      key={cred.id}
                      onClick={() => {
                        setActiveCredSetId(cred.id);
                        triggerToast(`Switched active profile to ${cred.name}`, 'info');
                      }}
                      className={`flex items-center justify-between p-2.5 rounded-lg border cursor-pointer transition-colors ${
                        isActive
                          ? 'bg-[#00A8FF]/10 border-[#00A8FF]/60 text-white'
                          : 'bg-[#0B0C0E] border-[#1E1F22] text-[#949BA4] hover:border-white/20'
                      }`}
                    >
                      <div className="flex items-center space-x-2.5">
                        <input
                          type="radio"
                          name="activeCred"
                          checked={isActive}
                          onChange={() => {}}
                          className="w-4 h-4 accent-[#00A8FF] cursor-pointer"
                        />
                        <div>
                          <div className="text-xs font-semibold text-white">{cred.name}</div>
                          <div className="text-[10px] font-mono text-[#7A828A]">{cred.username}</div>
                        </div>
                      </div>

                      {/* Edit Pencil & Red Square Trash Buttons */}
                      <div className="flex items-center space-x-1.5">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleAntiSpam(() => {
                              setEditingCredId(cred.id);
                              setFormCredName(cred.name);
                              setFormUsername(cred.username);
                              setFormPassword(cred.password || '');
                              setShowPassword(false);
                              setActiveModal('edit_cred');
                            });
                          }}
                          className="w-7 h-7 bg-[#1E2024] hover:bg-[#2B2D31] border border-white/10 hover:border-white/30 rounded flex items-center justify-center text-white transition-colors"
                          title="Edit credentials & password"
                        >
                          <Pencil className="w-3.5 h-3.5" />
                        </button>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleAntiSpam(() => {
                              const updated = credSets.filter((c) => c.id !== cred.id);
                              setCredSets(updated);
                              if (updated.length === 0) {
                                setActiveCredSetId('');
                                setLedState('red');
                                triggerToast('Warning: All credentials deleted. Auto-login paused.', 'warn');
                              } else {
                                if (isActive) {
                                  setActiveCredSetId(updated[0].id);
                                }
                                triggerToast(`Deleted credential set: ${cred.name}`, 'info');
                              }
                            });
                          }}
                          className="w-7 h-7 bg-[#F23F43] hover:bg-[#FF4D50] active:bg-[#D8363A] rounded flex items-center justify-center text-white transition-colors"
                          title="Delete credential profile"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </div>
                    </div>
                  );
                })}

                {credSets.length === 0 && (
                  <div className="text-xs text-[#7A828A] italic py-4 text-center">
                    No credentials stored. Click Add (+) on the main screen.
                  </div>
                )}
              </div>
            </div>

            <div className="pt-4">
              <button
                onClick={() => setActiveModal('none')}
                className="w-full py-2 px-3 rounded-lg bg-[#00A8FF] text-black font-semibold text-xs hover:bg-[#33BAFF]"
              >
                Done
              </button>
            </div>
          </div>
        )}

        {/* ========================================================================= */}
        {/* MODAL 5: EDIT CREDS (with password eye toggle)                            */}
        {/* ========================================================================= */}
        {activeModal === 'edit_cred' && (
          <div className="absolute inset-0 bg-black/90 backdrop-blur-sm z-30 p-5 flex flex-col justify-between animate-in fade-in duration-150">
            <div className="space-y-3">
              <div>
                <h3 className="text-sm font-bold text-white">Edit Credentials</h3>
                <p className="text-[11px] text-[#7A828A]">Update username, password, or profile name</p>
              </div>

              <div className="space-y-2.5 pt-1">
                <div>
                  <label className="block text-[11px] font-medium text-[#949BA4] mb-1">1. Credentials Name</label>
                  <input
                    type="text"
                    value={formCredName}
                    onChange={(e) => setFormCredName(e.target.value)}
                    placeholder="Profile name"
                    className="w-full bg-[#0E1012] border border-[#2B2D31] rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-[#00A8FF]"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-medium text-[#949BA4] mb-1">2. Username (SRN)</label>
                  <input
                    type="text"
                    value={formUsername}
                    onChange={(e) => setFormUsername(e.target.value)}
                    placeholder="PES1UG23..."
                    className="w-full bg-[#0E1012] border border-[#2B2D31] rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-[#00A8FF]"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-medium text-[#949BA4] mb-1">3. Password</label>
                  <div className="relative">
                    <input
                      type={showPassword ? 'text' : 'password'}
                      value={formPassword}
                      onChange={(e) => setFormPassword(e.target.value)}
                      placeholder="Portal password"
                      className="w-full bg-[#0E1012] border border-[#2B2D31] rounded-lg pl-3 pr-9 py-1.5 text-xs text-white focus:outline-none focus:border-[#00A8FF]"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-[#7A828A] hover:text-white transition-colors"
                      title={showPassword ? 'Hide password' : 'Show password'}
                    >
                      {showPassword ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
                    </button>
                  </div>
                </div>
              </div>

              <p className="text-[10px] text-[#7A828A] italic leading-relaxed border-l-2 border-[#00A8FF]/40 pl-2">
                Passwords are not stored in plaintext; protected with local machine-derived keys.
              </p>
            </div>

            <div className="flex items-center space-x-2 pt-4">
              <button
                onClick={() =>
                  handleAntiSpam(() => {
                    if (!formUsername.trim()) {
                      triggerToast('Username (SRN) required', 'error');
                      return;
                    }
                    const updatedName = formCredName.trim() || 'default';
                    setCredSets(
                      credSets.map((c) =>
                        c.id === editingCredId
                          ? {
                              ...c,
                              name: updatedName,
                              username: formUsername.trim(),
                              password: formPassword,
                            }
                          : c
                      )
                    );
                    setActiveModal('manage_cred');
                    triggerToast(`Updated credentials: ${updatedName}`, 'success');
                  })
                }
                className="flex-1 py-2 px-3 rounded-lg bg-[#00A8FF] text-black font-semibold text-xs hover:bg-[#33BAFF]"
              >
                Save Changes
              </button>
              <button
                onClick={() => setActiveModal('manage_cred')}
                className="py-2 px-4 rounded-lg bg-black text-white text-xs border border-white/20 hover:border-white/40"
              >
                Cancel
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Bottom Toast Popup (spawns from bottom, fades on click) */}
      <BottomToast toast={toast} onDismiss={dismissToast} />
    </div>
  );
}

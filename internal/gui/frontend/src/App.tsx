import React, { useState, useEffect, useRef } from 'react';
import { RefreshCw, Trash2, Shield, Wifi, Key, Check, Info, AlertTriangle, Eye, EyeOff, Pencil, ArrowLeft, CheckCircle2 } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';
import { CredSet, WifiNetwork, LEDState, ToastMessage, WizardStep } from './types';
import { BottomToast } from './components/BottomToast';
import { api, BackendState } from './api';

export default function App() {
  // Navigation & Screen State
  const [step, setStep] = useState<WizardStep>('permissions');

  // Configuration State
  const [activeCredSetId, setActiveCredSetId] = useState<string>('');
  const [credSets, setCredSets] = useState<CredSet[]>([]);
  const [savedSSIDs, setSavedSSIDs] = useState<string[]>([]);
  const [captiveBypassEnabled, setCaptiveBypassEnabled] = useState<boolean>(true);
  const [vanguardEnabled, setVanguardEnabled] = useState<boolean>(false);
  const [edgeThreshold, setEdgeThreshold] = useState<number>(15);

  // Status & Telemetry
  const [statusTitle, setStatusTitle] = useState<string>('<Disconnected>');
  const [statusSub, setStatusSub] = useState<string>('Wi-Fi is offline');
  const [ledState, setLedState] = useState<LEDState>('red');
  const [currentSignal, setCurrentSignal] = useState<number>(0);

  // Scanner & Network state (Empty initial list - real scan only)
  const [networks, setNetworks] = useState<WifiNetwork[]>([]);
  const [isScanning, setIsScanning] = useState<boolean>(false);

  // Modals inside Main Menu
  const [activeModal, setActiveModal] = useState<'none' | 'add_ssid' | 'manage_ssid' | 'add_cred' | 'manage_cred' | 'edit_cred'>('none');
  const [editingCredId, setEditingCredId] = useState<string | null>(null);

  // Form states for Wizard & Add/Edit dialogs
  const [formUsername, setFormUsername] = useState<string>('');
  const [formPassword, setFormPassword] = useState<string>('');
  const [showPassword, setShowPassword] = useState<boolean>(false);
  const [selectedSSIDsStaging, setSelectedSSIDsStaging] = useState<string[]>([]);

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

  const handleAntiSpam = (fn: () => void) => {
    if (isActionLocked) return;
    setIsActionLocked(true);
    fn();
    setTimeout(() => {
      setIsActionLocked(false);
    }, 300);
  };

  // Sync state from Go backend
  const syncState = async () => {
    try {
      const state = await api.getState();
      if (!state) return;

      setLedState(state.status);
      setStatusTitle(state.statusTitle);
      setStatusSub(state.statusSub);
      setCurrentSignal(state.signalPercent);
      setCaptiveBypassEnabled(state.isAutoLoginEnabled);
      setVanguardEnabled(state.isVanguardEnabled);
      setSavedSSIDs(state.recognizedNetworks || []);

      if (state.credProfiles && state.credProfiles.length > 0) {
        const sets: CredSet[] = state.credProfiles.map((p) => ({
          id: p.username,
          username: p.username,
        }));
        setCredSets(sets);
        const active = state.credProfiles.find((p) => p.isActive) || state.credProfiles[0];
        setActiveCredSetId(active ? active.username : "");
      } else {
        setCredSets([]);
        setActiveCredSetId('');
      }

      // If already configured and first run is false, land on main
      if (!state.isFirstRun && step === 'permissions') {
        setStep('main');
      }
    } catch (err) {
      console.error('Failed to sync backend state:', err);
    }
  };

  useEffect(() => {
    syncState();
    const interval = setInterval(syncState, 3000);
    return () => clearInterval(interval);
  }, []);

  // Real network scan via backend API
  const handleRefreshNetworks = async () => {
    if (isScanning) return;
    setIsScanning(true);
    triggerToast('Scanning nearby Wi-Fi networks...', 'info');

    try {
      const results = await api.scanNetworks();
      if (results && results.length > 0) {
        const mapped: WifiNetwork[] = results.map((r) => ({
          ssid: r.ssid,
          signal: r.signal,
          inRange: true,
        }));
        setNetworks(mapped);
        triggerToast(`Found ${mapped.length} network${mapped.length === 1 ? '' : 's'}`, 'success');
      } else {
        setNetworks([]);
        triggerToast('No Wi-Fi networks found nearby', 'info');
      }
    } catch (err) {
      console.error('Scan failed:', err);
      setNetworks([]);
      triggerToast('Unable to scan Wi-Fi networks', 'warn');
    } finally {
      setIsScanning(false);
    }
  };

  // Trigger scan when entering SSIDs step
  useEffect(() => {
    if (step === 'ssids' || activeModal === 'add_ssid') {
      handleRefreshNetworks();
    }
  }, [step, activeModal]);

  const toggleStagingSSID = (ssid: string) => {
    if (selectedSSIDsStaging.includes(ssid)) {
      setSelectedSSIDsStaging(selectedSSIDsStaging.filter((s) => s !== ssid));
    } else {
      setSelectedSSIDsStaging([...selectedSSIDsStaging, ssid]);
    }
  };

  const getLEDInfo = () => {
    switch (ledState) {
      case 'green':
        return {
          colorClass: 'bg-[#23A55A] shadow-[0_0_12px_#23A55A]',
          text: statusTitle || '<Connected>',
          subtext: statusSub || 'Connected & Online',
          desc: 'Connected & Online',
        };
      case 'red':
        return {
          colorClass: 'bg-[#F23F43] shadow-[0_0_12px_#F23F43]',
          text: statusTitle || '<Disconnected>',
          subtext: statusSub || 'Disconnected / Offline',
          desc: 'Disconnected / Offline',
        };
      case 'yellow':
        return {
          colorClass: 'bg-[#FEE75C] shadow-[0_0_12px_#FEE75C]',
          text: statusTitle || '<In Progress>',
          subtext: statusSub || 'Authenticating…',
          desc: 'Authenticating with Portal…',
        };
      case 'orange':
        return {
          colorClass: 'bg-[#FF9900] shadow-[0_0_12px_#FF9900]',
          text: statusTitle || '<Network Edge>',
          subtext: statusSub || "You're at the edge",
          desc: "You're at the edge",
        };
    }
  };

  const ledInfo = getLEDInfo();
  const activeCred = credSets.find((c) => c.id === activeCredSetId) || credSets[0];

  return (
    <div className="w-full h-screen bg-[#000000] text-[#E6EAED] flex flex-col relative selection:bg-[#00A8FF] selection:text-black font-sans overflow-hidden">
      <div className="flex-1 flex flex-col relative overflow-hidden">
        {/* ========================================================================= */}
        {/* SCREEN 1: WIZARD PERMISSIONS / TRUST                                      */}
        {/* ========================================================================= */}
        {step === 'permissions' && (
          <motion.div
            initial={{ opacity: 0, y: 15 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3 }}
            className="p-6 flex-1 flex flex-col justify-between"
          >
            <div className="space-y-4">
              <div className="flex items-center space-x-3">
                <div className="w-9 h-9 rounded-lg bg-[#00A8FF]/10 flex items-center justify-center text-[#00A8FF] flex-shrink-0">
                  <Shield className="w-5 h-5" />
                </div>
                <div>
                  <h1 className="text-lg font-bold text-white tracking-tight">Permissions & Trust</h1>
                  <p className="text-xs text-[#7A828A]">System access for seamless Wi-Fi auto-login</p>
                </div>
              </div>

              {/* Warning-first Standalone Binary Trust Box */}
              <div className="border border-[#1E1F22] bg-[#0A0B0D] rounded-lg p-4 space-y-3 text-xs text-[#949BA4] leading-relaxed">
                <p>
                  <strong className="text-white">Unsigned Standalone Binary:</strong> captive-bypass is shipped as an
                  open-source standalone executable without commercial developer certificates.
                </p>

                <div className="space-y-1.5 text-[11px] bg-[#121316] p-2.5 rounded border border-[#2B2D31]">
                  <div>
                    <strong className="text-white">macOS:</strong> Right-click the app &gt; <em>Open</em>, or allow via{' '}
                    <em>System Settings &gt; Privacy &amp; Security &gt; Open Anyway</em>.
                  </div>
                  <div>
                    <strong className="text-white">Windows:</strong> Click <em>More info</em> &gt; <em>Run anyway</em>{' '}
                    past SmartScreen.
                  </div>
                </div>

                <p>
                  100% open-source, runs entirely locally, and never exposes credentials outside of the official portal
                  handshake.
                </p>

                <div className="pt-1 flex items-center space-x-3 text-[11px]">
                  <button
                    type="button"
                    onClick={() => api.openURL('https://github.com/achar-pranav/captive-bypass')}
                    className="text-[#00A8FF] hover:underline cursor-pointer flex items-center space-x-1"
                  >
                    <span>View Source on GitHub ↗</span>
                  </button>
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
                onClick={() =>
                  handleAntiSpam(() => {
                    // Skip setup without saving any dummy credentials
                    setStep('main');
                    triggerToast('Setup skipped. No credentials saved.', 'info');
                  })
                }
                className="w-full py-2.5 px-4 rounded-lg bg-black text-white text-sm border border-white/20 hover:border-white/40 hover:bg-white/5 active:bg-white/10 transition-all"
              >
                Skip
              </button>
            </div>
          </motion.div>
        )}

        {/* ========================================================================= */}
        {/* SCREEN 2: WIZARD ADD CREDENTIALS                                          */}
        {/* ========================================================================= */}
        {step === 'credentials' && (
          <motion.div
            initial={{ opacity: 0, y: 15 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3 }}
            className="p-6 flex-1 flex flex-col justify-between"
          >
            <div className="space-y-4">
              <div>
                <h1 className="text-lg font-bold text-white tracking-tight">Add Credentials</h1>
                <p className="text-xs text-[#7A828A]">Set up your PESU portal profile</p>
              </div>

              {/* 2 Fields */}
              <div className="space-y-3 pt-2">

                <div>
                  <label className="block text-[11px] font-medium text-[#949BA4] mb-1">1. Username (SRN)</label>
                  <input
                    type="text"
                    value={formUsername}
                    onChange={(e) => setFormUsername(e.target.value)}
                    placeholder="e.g. PES1UG23CS001"
                    className="w-full bg-[#0E1012] border border-[#2B2D31] rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-[#00A8FF] transition-colors"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-medium text-[#949BA4] mb-1">2. Password</label>
                  <div className="relative">
                    <input
                      type={showPassword ? 'text' : 'password'}
                      value={formPassword}
                      onChange={(e) => setFormPassword(e.target.value)}
                      placeholder="Portal password"
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
                  Passwords are never stored in plaintext. Encrypted locally with machine-derived keys (AES-GCM) to
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
                  handleAntiSpam(async () => {
                    const trimmedUser = formUsername.trim();
                    if (!trimmedUser) {
                      triggerToast('Please enter your SRN username', 'error');
                      return;
                    }
                    try {
                      await api.saveCreds(trimmedUser, formPassword, true);
                      const newSet: CredSet = {
                        id: trimmedUser,
                        username: trimmedUser,
                        password: formPassword,
                      };
                      setCredSets([newSet]);
                      setActiveCredSetId(trimmedUser);
                      setStep('ssids');
                      triggerToast('Credentials encrypted & saved', 'success');
                    } catch (err) {
                      console.error('Error saving creds:', err);
                      triggerToast('Error saving credentials', 'error');
                    }
                  })
                }
                className="w-full py-2.5 px-4 rounded-lg bg-[#00A8FF] text-black font-semibold text-sm hover:bg-[#33BAFF] active:bg-[#0090DC] transition-colors"
              >
                Continue
              </button>
            </div>
          </motion.div>
        )}

        {/* ========================================================================= */}
        {/* SCREEN 3: WIZARD SELECT SSIDs                                             */}
        {/* ========================================================================= */}
        {step === 'ssids' && (
          <motion.div
            initial={{ opacity: 0, y: 15 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3 }}
            className="p-6 flex-1 flex flex-col justify-between"
          >
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
                    <span>Scan</span>
                    <RefreshCw className={`w-3 h-3 ${isScanning ? 'animate-spin' : ''}`} />
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

                  {networks.length === 0 && !isScanning && (
                    <div className="p-4 text-center border border-dashed border-[#1E1F22] rounded-lg text-xs text-[#7A828A]">
                      No networks found. Click Scan to search for nearby Wi-Fi.
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* Back & Continue Buttons */}
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
                    setStep('ready');
                  })
                }
                className="w-full py-2.5 px-4 rounded-lg bg-[#00A8FF] text-black font-semibold text-sm hover:bg-[#33BAFF] active:bg-[#0090DC] transition-colors"
              >
                Continue
              </button>
            </div>
          </motion.div>
        )}

        {/* ========================================================================= */}
        {/* SCREEN 4: WIZARD READY CONFIRMATION (#44)                                 */}
        {/* ========================================================================= */}
        {step === 'ready' && (
          <motion.div
            initial={{ opacity: 0, y: 15 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3 }}
            className="p-6 flex-1 flex flex-col justify-between"
          >
            <div className="space-y-5">
              <div className="flex items-center space-x-3">
                <div className="w-10 h-10 rounded-full bg-[#23A55A]/15 border border-[#23A55A]/30 flex items-center justify-center text-[#23A55A] flex-shrink-0">
                  <CheckCircle2 className="w-6 h-6" />
                </div>
                <div>
                  <h1 className="text-lg font-bold text-white tracking-tight">Ready to Auto-Login</h1>
                  <p className="text-xs text-[#7A828A]">Credentials and trigger networks configured</p>
                </div>
              </div>

              {/* Ready Summary Card */}
              <div className="border border-[#1E1F22] bg-[#0A0B0D] rounded-lg p-4 space-y-3 text-xs">
                <div className="flex items-center justify-between border-b border-[#1E1F22] pb-2">
                  <span className="text-[#7A828A]">Active SRN:</span>
                  <span className="text-[#00A8FF] font-mono font-semibold">{activeCred?.username || formUsername || 'Configured'}</span>
                </div>

                <div>
                  <span className="text-[#7A828A] block mb-1.5">Monitored Networks ({selectedSSIDsStaging.length}):</span>
                  <div className="flex flex-wrap gap-1.5">
                    {selectedSSIDsStaging.map((s) => (
                      <span
                        key={s}
                        className="inline-flex items-center text-[10px] font-mono bg-[#141619] border border-[#2B2D31] text-[#E6EAED] px-2 py-0.5 rounded"
                      >
                        {s}
                      </span>
                    ))}
                  </div>
                </div>
              </div>

              <p className="text-xs text-[#7A828A] leading-relaxed">
                You can enable auto-login now to test authentication immediately, or head straight to the main menu.
              </p>
            </div>

            {/* Final Setup Action Buttons */}
            <div className="space-y-2.5 pt-6">
              <button
                onClick={() =>
                  handleAntiSpam(async () => {
                    try {
                      await api.toggleAutoLogin(true);
                      await api.finishWizard(selectedSSIDsStaging);
                      setCaptiveBypassEnabled(true);
                      setStep('main');
                      triggerToast(`Auto-login enabled for '${activeCred?.username || formUsername.trim()}'.`, 'success');
                      if (credSets.length > 0 || formUsername.trim()) {
                        api.manualLogin().then((res) => {
                          if (res) triggerToast(res, 'info');
                        }).catch((err) => console.error('Login failed:', err));
                      } else {
                        triggerToast('No credentials configured — please add a profile.', 'warn');
                      }
                    } catch (err) {
                      console.error('Ready screen enable failed:', err);
                      setStep('main');
                    }
                  })
                }
                className="w-full py-2.5 px-4 rounded-lg bg-[#23A55A] text-black font-semibold text-sm hover:bg-[#2bc26a] active:bg-[#1f9350] transition-colors"
              >
                Enable Auto-Login
              </button>
              <button
                onClick={() =>
                  handleAntiSpam(async () => {
                    try {
                      await api.toggleAutoLogin(false);
                      await api.finishWizard(selectedSSIDsStaging);
                      setCaptiveBypassEnabled(false);
                      setStep('main');
                      triggerToast('Setup finished. Auto-login is currently off.', 'info');
                    } catch (err) {
                      console.error('Ready screen main menu fallback failed:', err);
                      setStep('main');
                    }
                  })
                }
                className="w-full py-2.5 px-4 rounded-lg bg-black text-white text-sm border border-white/20 hover:border-white/40 hover:bg-white/5 active:bg-white/10 transition-all"
              >
                Main Menu
              </button>
            </div>
          </motion.div>
        )}

        {/* ========================================================================= */}
        {/* SCREEN 5: MAIN MENU                                                       */}
        {/* ========================================================================= */}
        {step === 'main' && (
          <motion.div
            initial={{ opacity: 0, y: 15 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3 }}
            className="p-5 flex-1 flex flex-col justify-between space-y-4"
          >
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
                  <span className="text-[11px] text-[#7A828A] truncate font-mono">{ledInfo.subtext}</span>
                </div>
              </div>

              {/* Dev-only affordance */}
              {import.meta.env.DEV && (
                <span className="text-[9px] font-mono text-[#00A8FF] border border-[#00A8FF]/30 px-1.5 py-0.5 rounded">
                  DEV
                </span>
              )}
            </div>

            {/* Two Options: Networks and Creds */}
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
                        {credSets.find((c) => c.id === activeCredSetId)?.username || 'None'}
                      </span>
                    </p>
                  </div>
                  <div className="flex items-center space-x-2">
                    <button
                      onClick={() =>
                        handleAntiSpam(() => {
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

            {/* Bottom Controls: Master Toggle & Vanguard Toggle */}
            <div className="border-t border-[#1E1F22] pt-3.5 space-y-2.5">
              <label className="flex items-center justify-between cursor-pointer select-none">
                <span className="text-xs text-white font-medium">Enable/Disable the captive-bypass</span>
                <div className="relative">
                  <input
                    type="checkbox"
                    className="peer sr-only"
                    checked={captiveBypassEnabled}
                    onChange={(e) => {
                      const checked = e.target.checked;
                      setCaptiveBypassEnabled(checked);
                      api.toggleAutoLogin(checked);
                      if (!checked) {
                        triggerToast('captive-bypass disabled', 'warn');
                      } else {
                        triggerToast('captive-bypass enabled', 'success');
                      }
                    }}
                  />
                  <div className="block h-6 w-10 rounded-full bg-[#1E1F22] transition-colors peer-checked:bg-[#00A8FF]"></div>
                  <div className="absolute left-1 top-1 h-4 w-4 rounded-full bg-white transition-transform peer-checked:translate-x-4 shadow-sm"></div>
                </div>
              </label>

              <label className="flex items-center justify-between cursor-pointer select-none">
                <span className="text-xs text-white font-medium">Enable/Disable Vanguard (experimental)</span>
                <div className="relative">
                  <input
                    type="checkbox"
                    className="peer sr-only"
                    checked={vanguardEnabled}
                    onChange={(e) => {
                      const checked = e.target.checked;
                      setVanguardEnabled(checked);
                      api.toggleVanguard(checked);
                      triggerToast(
                        checked ? 'Vanguard telemetry enabled' : 'Vanguard telemetry disabled',
                        'info'
                      );
                    }}
                  />
                  <div className="block h-6 w-10 rounded-full bg-[#1E1F22] transition-colors peer-checked:bg-[#00A8FF]"></div>
                  <div className="absolute left-1 top-1 h-4 w-4 rounded-full bg-white transition-transform peer-checked:translate-x-4 shadow-sm"></div>
                </div>
              </label>
              
              {vanguardEnabled && (
                <div className="bg-[#0A0B0D] border border-[#1E1F22] rounded-lg p-2.5 text-[11px] space-y-1.5 mt-2 transition-all">
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
                      api.setThreshold(val);
                    }}
                    className="w-full accent-[#00A8FF] h-1.5 bg-[#1E1F22] rounded-lg appearance-none cursor-pointer"
                  />
                </div>
              )}
            </div>
          </motion.div>
        )}

        {/* ========================================================================= */}
        {/* MODAL 1: ADD NETWORK                                                      */}
        {/* ========================================================================= */}
        {activeModal === 'add_ssid' && (
          <div className="absolute inset-0 bg-black/95 z-30 p-5 flex flex-col justify-between animate-in fade-in duration-150">
            <div className="space-y-3">
              <div className="flex items-center justify-between border-b border-[#1E1F22] pb-2">
                <span className="text-sm font-bold text-white">Add SSID</span>
                <button
                  onClick={handleRefreshNetworks}
                  disabled={isScanning}
                  className="flex items-center space-x-1 text-xs text-[#00A8FF] hover:text-[#33BAFF] px-2 py-0.5 rounded border border-[#00A8FF]/30 hover:border-[#00A8FF]"
                >
                  <span>Scan</span>
                  <RefreshCw className={`w-3 h-3 ${isScanning ? 'animate-spin' : ''}`} />
                </button>
              </div>

              <p className="text-[11px] text-[#7A828A]">Check networks to register for automatic login.</p>

              <div className="space-y-1.5 max-h-[360px] overflow-y-auto pr-1">
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

                {networks.length === 0 && !isScanning && (
                  <div className="p-4 text-center border border-dashed border-[#1E1F22] rounded-lg text-xs text-[#7A828A]">
                    No networks discovered. Click Scan above.
                  </div>
                )}
              </div>
            </div>

            <div className="flex items-center space-x-2 pt-4">
              <button
                onClick={() =>
                  handleAntiSpam(async () => {
                    for (const s of selectedSSIDsStaging) {
                      await api.addSSID(s);
                    }
                    setSavedSSIDs([...selectedSSIDsStaging]);
                    setActiveModal('none');
                    triggerToast('Registered networks updated', 'success');
                  })
                }
                className="flex-1 py-2 px-3 rounded-lg bg-[#00A8FF] text-black font-semibold text-xs hover:bg-[#33BAFF]"
              >
                Apply Selection
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
        {/* MODAL 2: MANAGE SAVED SSIDs (Red Square Trash Button)                     */}
        {/* ========================================================================= */}
        {activeModal === 'manage_ssid' && (
          <div className="absolute inset-0 bg-black/95 z-30 p-5 flex flex-col justify-between animate-in fade-in duration-150">
            <div className="space-y-3">
              <div className="border-b border-[#1E1F22] pb-2 flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-bold text-white">Manage SSIDs</h3>
                  <p className="text-[11px] text-[#7A828A]">Remove registered networks</p>
                </div>
                <span className="text-xs font-mono text-[#00A8FF]">{savedSSIDs.length} saved</span>
              </div>

              <div className="space-y-2 max-h-[360px] overflow-y-auto pr-1">
                {savedSSIDs.map((ssid) => (
                  <div
                    key={ssid}
                    className="flex items-center justify-between p-2.5 rounded-lg bg-[#0B0C0E] border border-[#1E1F22]"
                  >
                    <span className="text-xs font-mono text-white">{ssid}</span>
                    <button
                      onClick={() =>
                        handleAntiSpam(async () => {
                          await api.removeSSID(ssid);
                          const updated = savedSSIDs.filter((s) => s !== ssid);
                          setSavedSSIDs(updated);
                          triggerToast(`Deleted network: ${ssid}`, 'info');
                        })
                      }
                      className="w-7 h-7 bg-[#F23F43] hover:bg-[#FF4D50] active:bg-[#D8363A] rounded flex items-center justify-center text-white transition-colors cursor-pointer"
                      title="Delete network"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                ))}

                {savedSSIDs.length === 0 && (
                  <div className="text-xs text-[#7A828A] italic py-6 text-center">No saved networks left.</div>
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
          <div className="absolute inset-0 bg-black/95 z-30 p-5 flex flex-col justify-between animate-in fade-in duration-150">
            <div className="space-y-3">
              <div>
                <h3 className="text-sm font-bold text-white">Add Credential Set</h3>
                <p className="text-[11px] text-[#7A828A]">Configure a profile for portal login</p>
              </div>

              <div className="space-y-2.5 pt-1">
                <div>
                  <label className="block text-[11px] font-medium text-[#949BA4] mb-1">1. Username (SRN)</label>
                  <input
                    type="text"
                    value={formUsername}
                    onChange={(e) => setFormUsername(e.target.value)}
                    placeholder="PES1UG23..."
                    className="w-full bg-[#0E1012] border border-[#2B2D31] rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-[#00A8FF]"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-medium text-[#949BA4] mb-1">2. Password</label>
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
                  handleAntiSpam(async () => {
                    const trimmedUser = formUsername.trim();
                    if (!trimmedUser) {
                      triggerToast('Username (SRN) required', 'error');
                      return;
                    }
                    try {
                      const isFirst = credSets.length === 0;
                      await api.saveCreds(trimmedUser, formPassword, isFirst);
                      const newSet: CredSet = {
                        id: trimmedUser,
                        username: trimmedUser,
                        password: formPassword,
                      };
                      const existingIndex = credSets.findIndex((c) => c.id === trimmedUser);
                      if (existingIndex >= 0) {
                        const updated = [...credSets];
                        updated[existingIndex] = newSet;
                        setCredSets(updated);
                      } else {
                        setCredSets([...credSets, newSet]);
                      }
                      if (isFirst || !activeCredSetId) {
                        setActiveCredSetId(trimmedUser);
                      }
                      setActiveModal('none');
                      triggerToast(`Saved profile: ${trimmedUser}`, 'success');
                    } catch (err) {
                      triggerToast('Failed to save credentials', 'error');
                    }
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
        {/* MODAL 4: MANAGE CREDS (Radio + Edit pencil + red trash button) (#39)      */}
        {/* ========================================================================= */}
        {activeModal === 'manage_cred' && (
          <div className="absolute inset-0 bg-black/95 z-30 p-5 flex flex-col justify-between animate-in fade-in duration-150">
            <div className="space-y-3">
              <div className="border-b border-[#1E1F22] pb-2 flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-bold text-white">Manage Credentials</h3>
                  <p className="text-[11px] text-[#7A828A]">Pick active set, edit, or delete</p>
                </div>
                <span className="text-xs font-mono text-[#00A8FF]">{credSets.length} sets</span>
              </div>

              <div className="space-y-2 max-h-[360px] overflow-y-auto pr-1">
                {credSets.map((cred) => {
                  const isActive = cred.id === activeCredSetId;
                  return (
                    <div
                      key={cred.id}
                      onClick={() => {
                        setActiveCredSetId(cred.username);
                        api.setActiveCred(cred.username);
                        triggerToast(`Active profile: ${cred.username}`, 'info');
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
                          <div className="text-xs font-semibold font-mono text-white">{cred.username}</div>
                        </div>
                      </div>

                      {/* Edit Pencil & Red Trash Buttons */}
                      <div className="flex items-center space-x-1.5">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleAntiSpam(() => {
                              setEditingCredId(cred.id);
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
                            handleAntiSpam(async () => {
                              await api.deleteCreds(cred.username);
                              const updated = credSets.filter((c) => c.id !== cred.id);
                              setCredSets(updated);
                              if (updated.length === 0) {
                                setActiveCredSetId('');
                                triggerToast('No credentials stored. Add credentials to enable auto-login.', 'warn');
                              } else {
                                if (isActive) {
                                  setActiveCredSetId(updated[0].username);
                                  api.setActiveCred(updated[0].username);
                                }
                                triggerToast(`Deleted credentials: ${cred.username}`, 'info');
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
                  <div className="text-xs text-[#7A828A] italic py-6 text-center">
                    No credentials stored. Click Add on the main screen.
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
        {/* MODAL 5: EDIT CREDS                                                       */}
        {/* ========================================================================= */}
        {activeModal === 'edit_cred' && (
          <div className="absolute inset-0 bg-black/95 z-30 p-5 flex flex-col justify-between animate-in fade-in duration-150">
            <div className="space-y-3">
              <div>
                <h3 className="text-sm font-bold text-white">Edit Credentials</h3>
                <p className="text-[11px] text-[#7A828A]">Update username or password</p>
              </div>

              <div className="space-y-2.5 pt-1">
                <div>
                  <label className="block text-[11px] font-medium text-[#949BA4] mb-1">1. Username (SRN)</label>
                  <input
                    type="text"
                    value={formUsername}
                    onChange={(e) => setFormUsername(e.target.value)}
                    placeholder="PES1UG23..."
                    className="w-full bg-[#0E1012] border border-[#2B2D31] rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-[#00A8FF]"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-medium text-[#949BA4] mb-1">2. Password</label>
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
                  handleAntiSpam(async () => {
                    const trimmedUser = formUsername.trim();
                    if (!trimmedUser) {
                      triggerToast('Username (SRN) required', 'error');
                      return;
                    }
                    try {
                      const wasActive = activeCredSetId === editingCredId;
                      await api.saveCreds(trimmedUser, formPassword, wasActive);
                      if (editingCredId && editingCredId !== trimmedUser) {
                        await api.deleteCreds(editingCredId);
                      }
                      setCredSets(
                        credSets.map((c) =>
                          c.id === editingCredId
                            ? {
                                id: trimmedUser,
                                username: trimmedUser,
                                password: formPassword,
                              }
                            : c
                        )
                      );
                      if (wasActive) {
                        setActiveCredSetId(trimmedUser);
                      }
                      setActiveModal('manage_cred');
                      triggerToast(`Updated credentials: ${trimmedUser}`, 'success');
                    } catch (err) {
                      triggerToast('Failed to update credentials', 'error');
                    }
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

      {/* Bottom Toast Popup */}
      <BottomToast toast={toast} onDismiss={dismissToast} />
    </div>
  );
}

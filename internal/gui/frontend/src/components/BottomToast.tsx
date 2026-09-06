import React from 'react';
import { AnimatePresence, motion } from 'motion/react';
import { ToastMessage } from '../types';

interface BottomToastProps {
  toast: ToastMessage | null;
  onDismiss: () => void;
}

export const BottomToast: React.FC<BottomToastProps> = ({ toast, onDismiss }) => {
  return (
    <AnimatePresence>
      {toast && (
        <motion.div
          key={toast.id}
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: 20 }}
          transition={{ duration: 0.25, ease: 'easeOut' }}
          onClick={onDismiss}
          className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 cursor-pointer max-w-sm w-[90%]"
        >
          <div className="bg-[#111214] border border-[#00A8FF]/40 text-[#E6EAED] px-4 py-3 rounded-lg shadow-2xl flex items-center justify-between space-x-3 hover:border-[#00A8FF] transition-colors">
            <div className="flex items-center space-x-2.5">
              <span
                className={`w-2 h-2 rounded-full flex-shrink-0 ${
                  toast.type === 'error'
                    ? 'bg-[#F23F43]'
                    : toast.type === 'warn'
                    ? 'bg-[#FF9900]'
                    : toast.type === 'success'
                    ? 'bg-[#23A55A]'
                    : 'bg-[#00A8FF]'
                }`}
              />
              <span className="text-xs font-medium tracking-wide">{toast.text}</span>
            </div>
            <span className="text-[10px] text-white/40 uppercase tracking-wider flex-shrink-0">Dismiss</span>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

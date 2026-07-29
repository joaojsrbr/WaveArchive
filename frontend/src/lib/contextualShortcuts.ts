import { useEffect, useRef, useState } from 'react';

type ShortcutFeedback = {
  message: string;
  tone: 'success' | 'warning';
};

type ContextualShortcutOptions = {
  canSave: boolean;
  onNew: () => void;
  onSave: () => boolean | Promise<boolean>;
  newMessage: string;
  savedMessage: string;
  invalidMessage: string;
};

export function useContextualShortcuts(options: ContextualShortcutOptions) {
  const optionsRef = useRef(options);
  const timeoutRef = useRef<number | undefined>(undefined);
  const [feedback, setFeedback] = useState<ShortcutFeedback>();
  optionsRef.current = options;

  useEffect(() => {
    function show(message: string, tone: ShortcutFeedback['tone']) {
      window.clearTimeout(timeoutRef.current);
      setFeedback({ message, tone });
      timeoutRef.current = window.setTimeout(() => setFeedback(undefined), 2400);
    }

    function handleShortcut(event: KeyboardEvent) {
      if ((!event.ctrlKey && !event.metaKey) || event.altKey) return;

      const key = event.key.toLocaleLowerCase();
      if (key !== 'n' && key !== 's') return;
      if (isEditableTarget(event.target)) {
        event.preventDefault();
        return;
      }
      if (key === 'n') {
        event.preventDefault();
        optionsRef.current.onNew();
        show(optionsRef.current.newMessage, 'success');
        return;
      }
      event.preventDefault();
      if (!optionsRef.current.canSave) {
        show(optionsRef.current.invalidMessage, 'warning');
        return;
      }

      void Promise.resolve(optionsRef.current.onSave()).then((saved) => {
        if (saved) show(optionsRef.current.savedMessage, 'success');
      });
    }

    window.addEventListener('keydown', handleShortcut);
    return () => {
      window.removeEventListener('keydown', handleShortcut);
      window.clearTimeout(timeoutRef.current);
    };
  }, []);

  return feedback;
}

function isEditableTarget(target: EventTarget | null) {
  if (!(target instanceof HTMLElement)) return false;
  return Boolean(target.closest('input, textarea, select, [contenteditable="true"]'));
}

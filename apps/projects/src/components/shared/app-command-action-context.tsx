"use client";

import type { ReactNode } from "react";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";

type AppCommandAction = {
  disabled?: boolean;
  id: string;
  label: string;
  onSelect: () => void;
};

type AppCommandActionContextValue = {
  action: AppCommandAction | null;
  clearAction: (id: string) => void;
  registerAction: (action: AppCommandAction) => void;
};

const AppCommandActionContext =
  createContext<AppCommandActionContextValue | null>(null);

export const AppCommandActionProvider = ({
  children,
}: {
  children: ReactNode;
}) => {
  const [action, setAction] = useState<AppCommandAction | null>(null);

  const registerAction = useCallback((nextAction: AppCommandAction) => {
    setAction(nextAction);
  }, []);

  const clearAction = useCallback((id: string) => {
    setAction((currentAction) =>
      currentAction?.id === id ? null : currentAction,
    );
  }, []);

  const value = useMemo(
    () => ({ action, clearAction, registerAction }),
    [action, clearAction, registerAction],
  );

  return (
    <AppCommandActionContext.Provider value={value}>
      {children}
    </AppCommandActionContext.Provider>
  );
};

export const useAppCommandAction = ({
  disabled = false,
  id,
  label,
  onSelect,
}: AppCommandAction) => {
  const context = useContext(AppCommandActionContext);
  const clearAction = context?.clearAction;
  const registerAction = context?.registerAction;
  const onSelectRef = useRef(onSelect);

  useEffect(() => {
    onSelectRef.current = onSelect;
  }, [onSelect]);

  useEffect(() => {
    if (!clearAction || !registerAction) return;

    registerAction({
      disabled,
      id,
      label,
      onSelect: () => {
        onSelectRef.current();
      },
    });

    return () => {
      clearAction(id);
    };
  }, [clearAction, disabled, id, label, registerAction]);
};

export const useCurrentAppCommandAction = () => {
  const context = useContext(AppCommandActionContext);

  if (!context) {
    throw new Error(
      "useCurrentAppCommandAction must be used within AppCommandActionProvider",
    );
  }

  return context.action;
};

"use client";

import {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  type ReactNode,
} from "react";
import { useLocalStorage } from "@/hooks";

export interface WalkthroughStep {
  id: string;
  target: string; // CSS selector or data attribute
  title: string;
  content: ReactNode;
  position?:
    | "top"
    | "top-start"
    | "bottom"
    | "bottom-start"
    | "center"
    | "left"
    | "right";
  showSkip?: boolean;
  action?: () => void; // Optional action to perform when step is shown
}

interface WalkthroughState {
  isActive: boolean;
  currentStep: number;
  totalSteps: number;
  hasSeenWalkthrough: boolean;
  walkthroughVersion: string;
}

interface WalkthroughContextType {
  state: WalkthroughState;
  steps: WalkthroughStep[];
  currentStepData: WalkthroughStep | null;
  nextStep: () => void;
  prevStep: () => void;
  goToStep: (stepIndex: number) => void;
  startWalkthrough: () => void;
  skipWalkthrough: () => void;
  closeWalkthrough: () => void;
  setSteps: (steps: WalkthroughStep[]) => void;
}

const WalkthroughContext = createContext<WalkthroughContextType>({
  state: {
    isActive: false,
    currentStep: 0,
    totalSteps: 0,
    hasSeenWalkthrough: false,
    walkthroughVersion: "1.0.0",
  },
  steps: [],
  currentStepData: null,
  nextStep: () => {},
  prevStep: () => {},
  goToStep: () => {},
  startWalkthrough: () => {},
  skipWalkthrough: () => {},
  closeWalkthrough: () => {},
  setSteps: () => {},
});

export const useWalkthrough = () => {
  const context = useContext(WalkthroughContext);
  return context;
};

interface WalkthroughProviderProps {
  children: ReactNode;
  version?: string;
  autoStart?: boolean;
}

export const WalkthroughProvider = ({
  children,
  version = "1.0.0",
  autoStart = false,
}: WalkthroughProviderProps) => {
  const [walkthroughState, setWalkthroughState] = useLocalStorage<{
    hasSeenWalkthrough: boolean;
    version: string;
  }>("complexus:walkthrough", {
    hasSeenWalkthrough: false,
    version: "0.0.0",
  });

  const [state, setState] = useState<WalkthroughState>({
    isActive: false,
    currentStep: 0,
    totalSteps: 0,
    hasSeenWalkthrough: walkthroughState.hasSeenWalkthrough,
    walkthroughVersion: version,
  });

  const [steps, setSteps] = useState<WalkthroughStep[]>([]);

  const currentStepData = steps[state.currentStep] || null;

  const startWalkthrough = useCallback(() => {
    setState((prev) => ({
      ...prev,
      isActive: true,
      currentStep: 0,
      totalSteps: steps.length,
    }));
  }, [steps.length]);

  const nextStep = useCallback(() => {
    setState((prev) => {
      if (prev.currentStep < prev.totalSteps - 1) {
        return { ...prev, currentStep: prev.currentStep + 1 };
      }
      // Completed walkthrough
      setWalkthroughState({
        hasSeenWalkthrough: true,
        version,
      });
      return {
        ...prev,
        isActive: false,
        hasSeenWalkthrough: true,
      };
    });
  }, [setWalkthroughState, version]);

  const prevStep = useCallback(() => {
    setState((prev) => ({
      ...prev,
      currentStep: Math.max(0, prev.currentStep - 1),
    }));
  }, []);

  const goToStep = useCallback((stepIndex: number) => {
    setState((prev) => ({
      ...prev,
      currentStep: Math.max(0, Math.min(stepIndex, prev.totalSteps - 1)),
    }));
  }, []);

  const skipWalkthrough = useCallback(() => {
    setWalkthroughState({
      hasSeenWalkthrough: true,
      version,
    });
    setState((prev) => ({
      ...prev,
      isActive: false,
      hasSeenWalkthrough: true,
    }));
  }, [setWalkthroughState, version]);

  const closeWalkthrough = useCallback(() => {
    setState((prev) => ({
      ...prev,
      isActive: false,
    }));
  }, []);

  // Handle version changes - restart walkthrough if version is newer
  useEffect(() => {
    const shouldShowWalkthrough =
      !walkthroughState.hasSeenWalkthrough ||
      walkthroughState.version !== version;

    if (shouldShowWalkthrough && autoStart && steps.length > 0) {
      startWalkthrough();
    }

    setState((prev) => ({
      ...prev,
      hasSeenWalkthrough:
        walkthroughState.hasSeenWalkthrough &&
        walkthroughState.version === version,
      totalSteps: steps.length,
    }));
  }, [walkthroughState, version, autoStart, steps.length, startWalkthrough]);

  // Execute step action when current step changes
  useEffect(() => {
    if (state.isActive && currentStepData.action) {
      currentStepData.action();
    }
  }, [state.isActive, state.currentStep, currentStepData]);

  const contextValue: WalkthroughContextType = {
    state,
    steps,
    currentStepData,
    nextStep,
    prevStep,
    goToStep,
    startWalkthrough,
    skipWalkthrough,
    closeWalkthrough,
    setSteps,
  };

  return (
    <WalkthroughContext.Provider value={contextValue}>
      {children}
    </WalkthroughContext.Provider>
  );
};

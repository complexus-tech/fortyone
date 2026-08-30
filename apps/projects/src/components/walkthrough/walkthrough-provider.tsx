"use client";

import {
  createContext,
  useContext,
  useState,
  useRef,
  useCallback,
  useEffect,
  type ReactNode,
} from "react";
import { useProfile } from "@/lib/hooks/profile";
import { useUpdateProfileMutation } from "@/lib/hooks/update-profile-mutation";
import { useLocalStorage, useMediaQuery } from "@/hooks";

export interface WalkthroughStep {
  id: string;
  target: string; // CSS selector or data attribute
  title: string;
  content: ReactNode;
  position?:
    | "top"
    | "top-start"
    | "top-end"
    | "bottom"
    | "bottom-start"
    | "center"
    | "left"
    | "right";
  showSkip?: boolean;
  action?: () => void; // Optional action to perform when step is shown
  highlight?: boolean;
}

interface WalkthroughState {
  isActive: boolean;
  currentStep: number;
  totalSteps: number;
  hasSeenWalkthrough: boolean;
  walkthroughVersion: string;
}

interface WalkthroughProviderState extends WalkthroughState {
  completionVersion: number;
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

const isCooldownComplete = (walkthroughClosedAt: string | null) => {
  if (!walkthroughClosedAt) {
    return true;
  }

  try {
    const closedTimestamp = new Date(walkthroughClosedAt).getTime();
    const now = new Date().getTime();
    const fourHours = 4 * 60 * 60 * 1000;
    return now - closedTimestamp > fourHours;
  } catch {
    return true;
  }
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
  const { data: profile } = useProfile();
  const { mutate: updateProfile } = useUpdateProfileMutation();
  const [walkthroughClosedAt, setWalkthroughClosedAt] = useLocalStorage<
    string | null
  >("fortyone:walkthrough-closed-at", null);

  const [state, setState] = useState<WalkthroughProviderState>({
    isActive: false,
    currentStep: 0,
    totalSteps: 0,
    hasSeenWalkthrough: Boolean(profile?.hasSeenWalkthrough),
    walkthroughVersion: version,
    completionVersion: 0,
  });

  const [walkthroughSteps, setWalkthroughSteps] = useState<WalkthroughStep[]>(
    [],
  );
  const lastSyncedCompletionVersionRef = useRef(0);
  const isMobile = useMediaQuery("(max-width: 768px)");
  const hasSeenWalkthrough =
    state.hasSeenWalkthrough || Boolean(profile?.hasSeenWalkthrough);

  const currentStepData = walkthroughSteps[state.currentStep] || null;

  const completeWalkthrough = useCallback(() => {
    setState((prev) => ({
      ...prev,
      isActive: false,
      hasSeenWalkthrough: true,
      completionVersion: prev.completionVersion + 1,
    }));
  }, []);

  const startWalkthrough = useCallback(() => {
    setState((prev) => ({
      ...prev,
      isActive: true,
      currentStep: 0,
      totalSteps: walkthroughSteps.length,
    }));
  }, [walkthroughSteps.length]);

  const nextStep = useCallback(() => {
    setState((prev) => {
      if (prev.currentStep < prev.totalSteps - 1) {
        return { ...prev, currentStep: prev.currentStep + 1 };
      }
      return {
        ...prev,
        isActive: false,
        hasSeenWalkthrough: true,
        completionVersion: prev.completionVersion + 1,
      };
    });
  }, []);

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
    completeWalkthrough();
  }, [completeWalkthrough]);

  const closeWalkthrough = useCallback(() => {
    // Store the close timestamp using useLocalStorage hook
    const closeTimestamp = new Date().toISOString();
    setWalkthroughClosedAt(closeTimestamp);

    setState((prev) => ({
      ...prev,
      isActive: false,
    }));
  }, [setWalkthroughClosedAt]);

  const canShowWalkthrough =
    !hasSeenWalkthrough && !isMobile && isCooldownComplete(walkthroughClosedAt);

  useEffect(() => {
    if (state.completionVersion <= lastSyncedCompletionVersionRef.current) {
      return;
    }

    lastSyncedCompletionVersionRef.current = state.completionVersion;
    updateProfile({ hasSeenWalkthrough: true });
  }, [state.completionVersion, updateProfile]);

  const setSteps = useCallback(
    (nextSteps: WalkthroughStep[]) => {
      setWalkthroughSteps(nextSteps);
      setState((prev) => {
        const totalSteps = nextSteps.length;
        const boundedCurrentStep = totalSteps
          ? Math.min(prev.currentStep, totalSteps - 1)
          : 0;
        const shouldAutoStart =
          autoStart && canShowWalkthrough && totalSteps > 0 && !prev.isActive;

        return {
          ...prev,
          currentStep: shouldAutoStart ? 0 : boundedCurrentStep,
          totalSteps,
          isActive: shouldAutoStart ? true : prev.isActive,
        };
      });
    },
    [autoStart, canShowWalkthrough],
  );

  // Execute step action when current step changes
  useEffect(() => {
    if (state.isActive && currentStepData.action) {
      currentStepData.action();
    }
  }, [state.isActive, state.currentStep, currentStepData]);

  const contextState: WalkthroughState = {
    isActive: state.isActive,
    currentStep: state.currentStep,
    totalSteps: state.totalSteps,
    hasSeenWalkthrough,
    walkthroughVersion: state.walkthroughVersion,
  };

  const contextValue: WalkthroughContextType = {
    state: contextState,
    steps: walkthroughSteps,
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

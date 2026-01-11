"use client";

import {
  createContext,
  useContext,
  useState,
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
  const { data: profile } = useProfile();
  const { mutate: updateProfile } = useUpdateProfileMutation();
  const [walkthroughClosedAt, setWalkthroughClosedAt] = useLocalStorage<
    string | null
  >("fortyone:walkthrough-closed-at", null);

  const [state, setState] = useState<WalkthroughState>({
    isActive: false,
    currentStep: 0,
    totalSteps: 0,
    hasSeenWalkthrough: profile?.hasSeenWalkthrough || true,
    walkthroughVersion: version,
  });

  const [steps, setSteps] = useState<WalkthroughStep[]>([]);
  const isMobile = useMediaQuery("(max-width: 768px)");

  const currentStepData = steps[state.currentStep] || null;

  // Update state when profile changes
  useEffect(() => {
    if (profile) {
      setState((prev) => ({
        ...prev,
        hasSeenWalkthrough: profile.hasSeenWalkthrough,
      }));
    }
  }, [profile, updateProfile]);

  const markWalkthroughComplete = useCallback(() => {
    updateProfile({ hasSeenWalkthrough: true });
    setState((prev) => ({
      ...prev,
      hasSeenWalkthrough: true,
    }));
  }, [updateProfile]);

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
      markWalkthroughComplete();
      return {
        ...prev,
        isActive: false,
        hasSeenWalkthrough: true,
      };
    });
  }, [markWalkthroughComplete]);

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
    markWalkthroughComplete();
    setState((prev) => ({
      ...prev,
      isActive: false,
      hasSeenWalkthrough: true,
    }));
  }, [markWalkthroughComplete]);

  const closeWalkthrough = useCallback(() => {
    // Store the close timestamp using useLocalStorage hook
    const closeTimestamp = new Date().toISOString();
    setWalkthroughClosedAt(closeTimestamp);

    setState((prev) => ({
      ...prev,
      isActive: false,
    }));
  }, [setWalkthroughClosedAt]);

  // Check if walkthrough should be shown based on close timestamp
  const shouldShowWalkthrough = useCallback(() => {
    if (state.hasSeenWalkthrough || isMobile) {
      return false; // Already completed permanently or on mobile
    }

    if (!walkthroughClosedAt) {
      return true; // Never closed before
    }

    try {
      const closedTimestamp = new Date(walkthroughClosedAt).getTime();
      const now = new Date().getTime();
      const fourHours = 4 * 60 * 60 * 1000; // 4 hours in milliseconds

      // Show if more than 4 hours have passed since last close
      return now - closedTimestamp > fourHours;
    } catch (error) {
      // If timestamp parsing fails, default to showing
      return true;
    }
  }, [state.hasSeenWalkthrough, walkthroughClosedAt, isMobile]);

  // Calculate if should show - this will update when dependencies change
  const canShowWalkthrough = shouldShowWalkthrough();

  // Handle auto-start logic
  useEffect(() => {
    if (canShowWalkthrough && autoStart && steps.length > 0) {
      startWalkthrough();
    }

    setState((prev) => ({
      ...prev,
      totalSteps: steps.length,
    }));
  }, [canShowWalkthrough, autoStart, steps.length, startWalkthrough]);

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

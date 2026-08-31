"use client";

import {
  createContext,
  useContext,
  useState,
  useRef,
  useCallback,
  useEffect,
  useMemo,
  type ReactNode,
} from "react";
import type { WalkthroughTargetSelector } from "@/shared/walkthrough/targets";
import { useProfile } from "@/lib/hooks/profile";
import { useOnboardingTourProgress } from "@/lib/hooks/users/onboarding-tour-progress";
import { useUpdateOnboardingTourProgressMutation } from "@/lib/hooks/users/update-onboarding-tour-progress";
import { useLocalStorage, useMediaQuery, useWorkspacePath } from "@/hooks";
import type {
  OnboardingTourStatus,
  UpdateOnboardingTourProgress,
} from "@/types";
import type { WalkthroughPosition } from "./walkthrough-position";
import { WORKSPACE_GETTING_STARTED_TOUR_KEY } from "./walkthrough-tour";

const WALKTHROUGH_ACTION_PROGRESS_STORAGE_KEY =
  "fortyone:walkthrough-action-progress";

export type WalkthroughAction = "story-created" | "maya-message-completed";

type WalkthroughActionProgress = Partial<Record<string, WalkthroughAction[]>>;

export interface WalkthroughRequiredAction {
  actionLabel: string;
  id: WalkthroughAction;
}

export interface WalkthroughWelcomeChoice {
  description: string;
  id: string;
  illustration: ReactNode;
  startAction?: () => void;
  targetStepId: string;
  title: string;
}

export interface WalkthroughStep {
  id: string;
  target: WalkthroughTargetSelector | "body";
  title: string;
  content: ReactNode;
  panelLayout?: "standard" | "welcome";
  welcomeChoices?: WalkthroughWelcomeChoice[];
  nextActionLabel?: string;
  position?: WalkthroughPosition;
  showSkip?: boolean;
  action?: () => void; // Optional action to perform when step is shown
  highlight?: boolean;
  requiredAction?: WalkthroughRequiredAction;
}

interface WalkthroughState {
  isActive: boolean;
  currentStep: number;
  totalSteps: number;
  hasSeenWalkthrough: boolean;
  walkthroughVersion: string;
}

type WalkthroughProviderState = WalkthroughState;

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
  completeWalkthroughAction: (action: WalkthroughAction) => void;
  isWalkthroughActionComplete: (action: WalkthroughAction) => boolean;
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
  completeWalkthroughAction: () => {},
  isWalkthroughActionComplete: () => false,
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

const getFirstIncompleteStepIndex = (
  steps: WalkthroughStep[],
  completedStepIds: string[],
) => {
  const completedSteps = new Set(completedStepIds);
  return steps.findIndex((step) => !completedSteps.has(step.id));
};

const completedStepIdsForActions = (actions: string[]) => {
  const actionStepIds: Partial<Record<string, string>> = {
    "maya-message-completed": "maya",
    "story-created": "create-story",
  };

  return actions.flatMap((action) => {
    const stepID = actionStepIds[action];
    return stepID ? [stepID] : [];
  });
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
  const { data: profile, isPending: isProfilePending } = useProfile();
  const { workspaceSlug } = useWorkspacePath();
  const {
    data: onboardingTourProgress,
    isPending: isOnboardingTourProgressPending,
  } = useOnboardingTourProgress({
    tourKey: WORKSPACE_GETTING_STARTED_TOUR_KEY,
    tourVersion: version,
  });
  const { mutate: updateOnboardingTourProgress } =
    useUpdateOnboardingTourProgressMutation();
  const [walkthroughClosedAt, setWalkthroughClosedAt] = useLocalStorage<
    string | null
  >("fortyone:walkthrough-closed-at", null);
  const [walkthroughActionProgress, setWalkthroughActionProgress] =
    useLocalStorage<WalkthroughActionProgress>(
      WALKTHROUGH_ACTION_PROGRESS_STORAGE_KEY,
      {},
    );

  const [state, setState] = useState<WalkthroughProviderState>({
    isActive: false,
    currentStep: 0,
    totalSteps: 0,
    hasSeenWalkthrough: false,
    walkthroughVersion: version,
  });

  const [walkthroughSteps, setWalkthroughSteps] = useState<WalkthroughStep[]>(
    [],
  );
  const migratedLegacyActionProgressScopeRef = useRef<string | null>(null);
  const isMobile = useMediaQuery("(max-width: 768px)");
  const hasSeenWalkthrough =
    state.hasSeenWalkthrough ||
    (onboardingTourProgress?.status !== undefined &&
      onboardingTourProgress.status !== "active");
  const walkthroughActionScope = `${profile?.id ?? "anonymous"}:${workspaceSlug || "workspace"}:${version}`;
  const legacyCompletedActions = useMemo(
    () => walkthroughActionProgress[walkthroughActionScope] ?? [],
    [walkthroughActionProgress, walkthroughActionScope],
  );
  const completedActions = useMemo(
    () =>
      Array.from(
        new Set([
          ...(onboardingTourProgress?.completedActionIds ?? []),
          ...legacyCompletedActions,
        ]),
      ),
    [legacyCompletedActions, onboardingTourProgress?.completedActionIds],
  );
  const completedStepIds = useMemo(() => {
    const persistedStepIds = onboardingTourProgress?.completedStepIds ?? [];
    const actionStepIds = completedActions.length
      ? ["welcome", ...completedStepIdsForActions(completedActions)]
      : [];

    return Array.from(new Set([...persistedStepIds, ...actionStepIds]));
  }, [completedActions, onboardingTourProgress?.completedStepIds]);
  const isOnboardingTourReady =
    !isProfilePending && !isOnboardingTourProgressPending;

  const currentStepData = walkthroughSteps[state.currentStep] || null;

  const persistOnboardingTourProgress = useCallback(
    (
      updates: Omit<UpdateOnboardingTourProgress, "tourKey" | "tourVersion">,
    ) => {
      updateOnboardingTourProgress({
        ...updates,
        tourKey: WORKSPACE_GETTING_STARTED_TOUR_KEY,
        tourVersion: version,
      });
    },
    [updateOnboardingTourProgress, version],
  );

  const completeWalkthrough = useCallback(
    (
      status: Extract<OnboardingTourStatus, "completed" | "skipped">,
      updates: Pick<
        UpdateOnboardingTourProgress,
        "completedActionIds" | "completedStepIds"
      > = {},
    ) => {
      persistOnboardingTourProgress({ ...updates, status });

      setState((prev) => ({
        ...prev,
        hasSeenWalkthrough: true,
        isActive: false,
      }));
    },
    [persistOnboardingTourProgress],
  );

  const startWalkthrough = useCallback(() => {
    const firstIncompleteStep = getFirstIncompleteStepIndex(
      walkthroughSteps,
      completedStepIds,
    );

    if (firstIncompleteStep < 0) {
      return;
    }

    setState((prev) => ({
      ...prev,
      currentStep: firstIncompleteStep,
      isActive: true,
      totalSteps: walkthroughSteps.length,
    }));
  }, [completedStepIds, walkthroughSteps]);

  const completeWalkthroughAction = useCallback(
    (action: WalkthroughAction) => {
      setWalkthroughActionProgress((currentProgress) => {
        const progress = currentProgress;
        const completedForScope = progress[walkthroughActionScope] ?? [];

        if (completedForScope.includes(action)) {
          return progress;
        }

        return {
          ...progress,
          [walkthroughActionScope]: [...completedForScope, action],
        };
      });

      const currentStep = currentStepData;
      const resolvesCurrentStep =
        state.isActive && currentStep?.requiredAction?.id === action;

      if (!resolvesCurrentStep) {
        persistOnboardingTourProgress({
          completedActionIds: [action],
          completedStepIds: [
            "welcome",
            ...completedStepIdsForActions([action]),
          ],
        });
        return;
      }

      const completedStepIdsForAction = ["welcome", currentStep.id];
      const isLastStep = state.currentStep >= walkthroughSteps.length - 1;

      if (isLastStep) {
        completeWalkthrough("completed", {
          completedActionIds: [action],
          completedStepIds: completedStepIdsForAction,
        });
        return;
      }

      persistOnboardingTourProgress({
        completedActionIds: [action],
        completedStepIds: completedStepIdsForAction,
      });
      setState((prev) => ({
        ...prev,
        currentStep: prev.currentStep + 1,
      }));
    },
    [
      completeWalkthrough,
      currentStepData,
      persistOnboardingTourProgress,
      setWalkthroughActionProgress,
      state.currentStep,
      state.isActive,
      walkthroughActionScope,
      walkthroughSteps.length,
    ],
  );

  const isWalkthroughActionComplete = useCallback(
    (action: WalkthroughAction) => completedActions.includes(action),
    [completedActions],
  );

  const nextStep = useCallback(() => {
    const currentStep = currentStepData;

    if (!currentStep) {
      return;
    }

    if (
      currentStep.requiredAction &&
      !completedActions.includes(currentStep.requiredAction.id)
    ) {
      return;
    }

    const isLastStep = state.currentStep >= walkthroughSteps.length - 1;

    if (isLastStep) {
      completeWalkthrough("completed", {
        completedStepIds: [currentStep.id],
      });
      return;
    }

    persistOnboardingTourProgress({ completedStepIds: [currentStep.id] });
    setState((prev) => ({
      ...prev,
      currentStep: prev.currentStep + 1,
    }));
  }, [
    completeWalkthrough,
    completedActions,
    currentStepData,
    persistOnboardingTourProgress,
    state.currentStep,
    walkthroughSteps.length,
  ]);

  const prevStep = useCallback(() => {
    setState((prev) => ({
      ...prev,
      currentStep: Math.max(0, prev.currentStep - 1),
    }));
  }, []);

  const goToStep = useCallback(
    (stepIndex: number) => {
      const targetStepIndex = Math.max(
        0,
        Math.min(stepIndex, state.totalSteps - 1),
      );

      if (targetStepIndex > state.currentStep) {
        persistOnboardingTourProgress({
          completedStepIds: walkthroughSteps
            .slice(state.currentStep, targetStepIndex)
            .map((step) => step.id),
        });
      }

      setState((prev) => ({
        ...prev,
        currentStep: targetStepIndex,
      }));
    },
    [
      persistOnboardingTourProgress,
      state.currentStep,
      state.totalSteps,
      walkthroughSteps,
    ],
  );

  const skipWalkthrough = useCallback(() => {
    completeWalkthrough("skipped");
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
    isOnboardingTourReady &&
    !hasSeenWalkthrough &&
    !isMobile &&
    isCooldownComplete(walkthroughClosedAt);

  useEffect(() => {
    if (
      !onboardingTourProgress ||
      migratedLegacyActionProgressScopeRef.current === walkthroughActionScope
    ) {
      return;
    }

    migratedLegacyActionProgressScopeRef.current = walkthroughActionScope;

    if (
      onboardingTourProgress.status !== "active" ||
      onboardingTourProgress.completedActionIds.length > 0 ||
      onboardingTourProgress.completedStepIds.length > 0 ||
      legacyCompletedActions.length === 0
    ) {
      return;
    }

    persistOnboardingTourProgress({
      completedActionIds: legacyCompletedActions,
      completedStepIds: [
        "welcome",
        ...completedStepIdsForActions(legacyCompletedActions),
      ],
    });
  }, [
    legacyCompletedActions,
    onboardingTourProgress,
    persistOnboardingTourProgress,
    walkthroughActionScope,
  ]);

  const setSteps = useCallback(
    (nextSteps: WalkthroughStep[]) => {
      setWalkthroughSteps(nextSteps);
      setState((prev) => {
        const totalSteps = nextSteps.length;
        const firstIncompleteStep = getFirstIncompleteStepIndex(
          nextSteps,
          completedStepIds,
        );
        const boundedCurrentStep = totalSteps
          ? Math.min(
              prev.isActive || firstIncompleteStep < 0
                ? prev.currentStep
                : firstIncompleteStep,
              totalSteps - 1,
            )
          : 0;
        const shouldAutoStart =
          autoStart &&
          canShowWalkthrough &&
          firstIncompleteStep >= 0 &&
          totalSteps > 0 &&
          !prev.isActive;

        return {
          ...prev,
          currentStep: shouldAutoStart
            ? Math.max(0, firstIncompleteStep)
            : boundedCurrentStep,
          totalSteps,
          isActive: shouldAutoStart ? true : prev.isActive,
        };
      });
    },
    [autoStart, canShowWalkthrough, completedStepIds],
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
    completeWalkthroughAction,
    isWalkthroughActionComplete,
    setSteps,
  };

  return (
    <WalkthroughContext.Provider value={contextValue}>
      {children}
    </WalkthroughContext.Provider>
  );
};

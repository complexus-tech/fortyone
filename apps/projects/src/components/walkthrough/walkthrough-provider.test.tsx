/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { StrictMode, type ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { useLocalStorage, useMediaQuery, useWorkspacePath } from "@/hooks";
import { useProfile } from "@/lib/hooks/profile";
import { useOnboardingTourProgress } from "@/lib/hooks/users/onboarding-tour-progress";
import { useUpdateOnboardingTourProgressMutation } from "@/lib/hooks/users/update-onboarding-tour-progress";
import {
  type WalkthroughStep,
  WalkthroughProvider,
  useWalkthrough,
} from "./walkthrough-provider";

jest.mock("@/hooks", () => ({
  useLocalStorage: jest.fn(),
  useMediaQuery: jest.fn(),
  useWorkspacePath: jest.fn(),
}));

jest.mock("@/lib/hooks/profile", () => ({
  useProfile: jest.fn(),
}));

jest.mock("@/lib/hooks/users/onboarding-tour-progress", () => ({
  useOnboardingTourProgress: jest.fn(),
}));

jest.mock("@/lib/hooks/users/update-onboarding-tour-progress", () => ({
  useUpdateOnboardingTourProgressMutation: jest.fn(),
}));

const useMediaQueryMock = jest.mocked(useMediaQuery);
const useLocalStorageMock = jest.mocked(useLocalStorage);
const useWorkspacePathMock = jest.mocked(useWorkspacePath);
const useProfileMock = jest.mocked(useProfile);
const useOnboardingTourProgressMock = jest.mocked(useOnboardingTourProgress);
const useUpdateOnboardingTourProgressMutationMock = jest.mocked(
  useUpdateOnboardingTourProgressMutation,
);
const updateOnboardingTourProgressMock = jest.fn();

const walkthroughSteps: WalkthroughStep[] = [
  {
    content: "First step",
    id: "first",
    target: "body",
    title: "First",
  },
  {
    content: "Last step",
    id: "last",
    target: "body",
    title: "Last",
  },
];

const actionGatedWalkthroughSteps: WalkthroughStep[] = [
  {
    content: "Create a task.",
    id: "create-story",
    requiredAction: {
      actionLabel: "Create my first task",
      id: "story-created",
    },
    target: "body",
    title: "Create your first task",
  },
  {
    content: "Next step",
    id: "next",
    target: "body",
    title: "Next",
  },
];

const branchedWalkthroughSteps: WalkthroughStep[] = [
  {
    content: "Choose a path.",
    id: "welcome",
    target: "body",
    title: "Welcome",
  },
  {
    content: "Create work.",
    id: "create",
    target: "body",
    title: "Create",
  },
  {
    content: "Plan time.",
    id: "calendar",
    target: "body",
    title: "Calendar",
  },
  {
    content: "Get help.",
    id: "help",
    target: "body",
    title: "Help",
  },
];

const WalkthroughTestProvider = ({ children }: { children: ReactNode }) => (
  <StrictMode>
    <WalkthroughProvider>{children}</WalkthroughProvider>
  </StrictMode>
);

describe("WalkthroughProvider", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    useMediaQueryMock.mockReturnValue(false);
    useLocalStorageMock.mockImplementation((key) =>
      key === "fortyone:walkthrough-action-progress"
        ? ([{}, jest.fn()] as never)
        : ([null, jest.fn()] as never),
    );
    useWorkspacePathMock.mockReturnValue({
      withWorkspace: (path: string) => `/acme${path}`,
      workspaceSlug: "acme",
    } as ReturnType<typeof useWorkspacePath>);
    useProfileMock.mockReturnValue({ data: undefined } as ReturnType<
      typeof useProfile
    >);
    useOnboardingTourProgressMock.mockReturnValue({
      data: {
        completedActionIds: [],
        completedStepIds: [],
        status: "active",
        tourKey: "workspace-getting-started",
        tourVersion: "1.0.0",
      },
      isPending: false,
    } as unknown as ReturnType<typeof useOnboardingTourProgress>);
    useUpdateOnboardingTourProgressMutationMock.mockReturnValue({
      mutate: updateOnboardingTourProgressMock,
    } as unknown as ReturnType<typeof useUpdateOnboardingTourProgressMutation>);
  });

  it("syncs a completed walkthrough once after the final state commits", async () => {
    const { result } = renderHook(() => useWalkthrough(), {
      wrapper: WalkthroughTestProvider,
    });

    act(() => {
      result.current.setSteps(walkthroughSteps);
    });

    act(() => {
      result.current.startWalkthrough();
    });

    act(() => {
      result.current.nextStep();
    });

    expect(result.current.state).toMatchObject({
      currentStep: 1,
      hasSeenWalkthrough: false,
      isActive: true,
      totalSteps: 2,
    });
    act(() => {
      result.current.nextStep();
    });

    await waitFor(() => {
      expect(updateOnboardingTourProgressMock).toHaveBeenCalledTimes(2);
    });
    expect(updateOnboardingTourProgressMock).toHaveBeenCalledWith({
      completedStepIds: ["last"],
      status: "completed",
      tourKey: "workspace-getting-started",
      tourVersion: "1.0.0",
    });
    expect(result.current.state).toMatchObject({
      currentStep: 1,
      hasSeenWalkthrough: true,
      isActive: false,
      totalSteps: 2,
    });
    expect(result.current.state).not.toHaveProperty("completionVersion");
  });

  it("does not advance a gated step until the real action completes", () => {
    const { result } = renderHook(() => useWalkthrough(), {
      wrapper: WalkthroughTestProvider,
    });

    act(() => {
      result.current.setSteps(actionGatedWalkthroughSteps);
    });

    act(() => {
      result.current.startWalkthrough();
    });

    act(() => {
      result.current.nextStep();
    });

    expect(result.current.state.currentStep).toBe(0);

    act(() => {
      result.current.completeWalkthroughAction("story-created");
    });

    expect(result.current.state.currentStep).toBe(1);
    expect(updateOnboardingTourProgressMock).toHaveBeenCalledWith({
      completedActionIds: ["story-created"],
      completedStepIds: ["welcome", "create-story"],
      tourKey: "workspace-getting-started",
      tourVersion: "1.0.0",
    });
  });

  it("counts and navigates only the journey selected from the welcome step", () => {
    const { result } = renderHook(() => useWalkthrough(), {
      wrapper: WalkthroughTestProvider,
    });

    act(() => {
      result.current.setSteps(branchedWalkthroughSteps);
    });

    act(() => {
      result.current.startWalkthrough();
    });

    act(() => {
      result.current.goToStep(2);
    });

    expect(result.current.state).toMatchObject({
      currentStep: 2,
      progress: {
        current: 2,
        total: 3,
      },
      totalSteps: 4,
    });

    act(() => {
      result.current.prevStep();
    });

    expect(result.current.state).toMatchObject({
      currentStep: 0,
      progress: {
        current: 1,
        total: 3,
      },
    });

    act(() => {
      result.current.nextStep();
    });

    expect(result.current.state).toMatchObject({
      currentStep: 2,
      progress: {
        current: 2,
        total: 3,
      },
    });
  });

  it("persists a real action completed outside the visible tour as its resolved step", () => {
    const { result } = renderHook(() => useWalkthrough(), {
      wrapper: WalkthroughTestProvider,
    });

    act(() => {
      result.current.setSteps(actionGatedWalkthroughSteps);
      result.current.completeWalkthroughAction("story-created");
    });

    expect(updateOnboardingTourProgressMock).toHaveBeenCalledWith({
      completedActionIds: ["story-created"],
      completedStepIds: ["welcome", "create-story"],
      tourKey: "workspace-getting-started",
      tourVersion: "1.0.0",
    });
  });

  it("does not let the legacy global completion flag hide active scoped progress", () => {
    useProfileMock.mockReturnValue({
      data: { hasSeenWalkthrough: true },
      isPending: false,
    } as ReturnType<typeof useProfile>);

    const { result } = renderHook(() => useWalkthrough(), {
      wrapper: WalkthroughTestProvider,
    });

    act(() => {
      result.current.setSteps(walkthroughSteps);
    });

    act(() => {
      result.current.startWalkthrough();
    });

    expect(result.current.state).toMatchObject({
      currentStep: 0,
      isActive: true,
    });
  });

  it("maps a persisted real action to its resolved step before resuming", () => {
    useOnboardingTourProgressMock.mockReturnValue({
      data: {
        completedActionIds: ["story-created"],
        completedStepIds: [],
        status: "active",
        tourKey: "workspace-getting-started",
        tourVersion: "1.0.0",
      },
      isPending: false,
    } as unknown as ReturnType<typeof useOnboardingTourProgress>);

    const { result } = renderHook(() => useWalkthrough(), {
      wrapper: WalkthroughTestProvider,
    });

    act(() => {
      result.current.setSteps(actionGatedWalkthroughSteps);
    });

    act(() => {
      result.current.startWalkthrough();
    });

    expect(result.current.state).toMatchObject({
      currentStep: 1,
      isActive: true,
    });
  });

  it.each(["completed", "skipped"] as const)(
    "replays %s progress from the welcome step",
    (status) => {
      useOnboardingTourProgressMock.mockReturnValue({
        data: {
          completedActionIds: [],
          completedStepIds: ["first"],
          status,
          tourKey: "workspace-getting-started",
          tourVersion: "1.0.0",
        },
        isPending: false,
      } as unknown as ReturnType<typeof useOnboardingTourProgress>);

      const { result } = renderHook(() => useWalkthrough(), {
        wrapper: WalkthroughTestProvider,
      });

      act(() => {
        result.current.setSteps(walkthroughSteps);
      });

      act(() => {
        result.current.startWalkthrough();
      });

      expect(result.current.state).toMatchObject({
        currentStep: 0,
        isActive: true,
      });
    },
  );

  it("persists contextual tour dismissal against its own tour key", () => {
    const wrapper = ({ children }: { children: ReactNode }) => (
      <WalkthroughProvider
        dismissOnClose
        tourKey="workspace-module-calendar"
        version="1.0.0"
      >
        {children}
      </WalkthroughProvider>
    );
    const { result } = renderHook(() => useWalkthrough(), { wrapper });

    act(() => {
      result.current.setSteps(walkthroughSteps);
    });

    act(() => {
      result.current.startWalkthrough();
    });

    act(() => {
      result.current.closeWalkthrough();
    });

    expect(updateOnboardingTourProgressMock).toHaveBeenCalledWith({
      status: "skipped",
      tourKey: "workspace-module-calendar",
      tourVersion: "1.0.0",
    });
    expect(result.current.state.isActive).toBe(false);
    expect(result.current.state.hasSeenWalkthrough).toBe(true);
  });
});

/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { act, renderHook } from "@testing-library/react";
import type { ReactNode } from "react";
import { usePathname, useRouter } from "next/navigation";
import {
  useFeatures,
  useTerminology,
  useUserRole,
  useWorkspacePath,
} from "@/hooks";
import { useChatContext } from "@/context/chat-context";
import { useMayaMessageAvailability } from "@/modules/maya/hooks/use-maya-message-availability";
import { walkthroughTargetSelectors } from "@/shared/walkthrough/targets";
import { useWalkthroughSteps } from "./use-walkthrough-steps";

jest.mock("canvas-confetti", () => jest.fn());

jest.mock("ui", () => ({
  Box: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  Kbd: ({ children }: { children: ReactNode }) => <kbd>{children}</kbd>,
  Text: ({ children }: { children: ReactNode }) => <p>{children}</p>,
}));

jest.mock("next/navigation", () => ({
  usePathname: jest.fn(),
  useRouter: jest.fn(),
}));

jest.mock("@/hooks", () => ({
  useTerminology: jest.fn(),
  useUserRole: jest.fn(),
  useFeatures: jest.fn(),
  useWorkspacePath: jest.fn(),
}));

jest.mock("@/context/chat-context", () => ({
  useChatContext: jest.fn(),
}));

jest.mock("@/modules/maya/hooks/use-maya-message-availability", () => ({
  useMayaMessageAvailability: jest.fn(),
}));

const usePathnameMock = jest.mocked(usePathname);
const useRouterMock = jest.mocked(useRouter);
const useTerminologyMock = jest.mocked(useTerminology);
const useUserRoleMock = jest.mocked(useUserRole);
const useFeaturesMock = jest.mocked(useFeatures);
const useWorkspacePathMock = jest.mocked(useWorkspacePath);
const useChatContextMock = jest.mocked(useChatContext);
const useMayaMessageAvailabilityMock = jest.mocked(useMayaMessageAvailability);
const routerPushMock = jest.fn();
const openChatMock = jest.fn();

const getStepIds = (userRole: "guest" | "member") => {
  useUserRoleMock.mockReturnValue({ userRole } as ReturnType<
    typeof useUserRole
  >);

  const { result } = renderHook(() => useWalkthroughSteps());

  return result.current.map((step) => step.id);
};

describe("useWalkthroughSteps", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    usePathnameMock.mockReturnValue("/acme/my-work");
    useRouterMock.mockReturnValue({
      push: routerPushMock,
    } as unknown as ReturnType<typeof useRouter>);
    useWorkspacePathMock.mockReturnValue({
      withWorkspace: (path: string) => `/acme${path}`,
    } as ReturnType<typeof useWorkspacePath>);
    useTerminologyMock.mockReturnValue({
      getTermDisplay: (term: string, options?: { variant?: string }) =>
        options?.variant === "plural" ? `${term}s` : term,
    } as ReturnType<typeof useTerminology>);
    useFeaturesMock.mockReturnValue({
      isPending: false,
      objectiveEnabled: true,
    } as ReturnType<typeof useFeatures>);
    useChatContextMock.mockReturnValue({
      openChat: openChatMock,
    } as unknown as ReturnType<typeof useChatContext>);
    useMayaMessageAvailabilityMock.mockReturnValue({
      isLimited: false,
      isPending: false,
      isUnavailable: false,
    });
  });

  it("omits the creation step for guests", () => {
    expect(getStepIds("guest")).not.toContain("create-story");
  });

  it("keeps current product actions available to members", () => {
    expect(getStepIds("member")).toEqual([
      "welcome",
      "create-story",
      "calendar",
      "roadmap",
      "maya",
      "help",
      "teams",
    ]);
  });

  it("only highlights Roadmap when objectives are enabled", () => {
    useFeaturesMock.mockReturnValue({
      isPending: false,
      objectiveEnabled: false,
    } as ReturnType<typeof useFeatures>);

    expect(getStepIds("member")).not.toContain("roadmap");
    expect(getStepIds("member")).toEqual([
      "welcome",
      "create-story",
      "calendar",
      "maya",
      "help",
      "teams",
    ]);
  });

  it("uses one choice-led welcome panel and compact contextual callouts", () => {
    useUserRoleMock.mockReturnValue({ userRole: "member" } as ReturnType<
      typeof useUserRole
    >);

    const { result } = renderHook(() => useWalkthroughSteps());
    const welcome = result.current.find((step) => step.id === "welcome");
    const calendar = result.current.find((step) => step.id === "calendar");
    const maya = result.current.find((step) => step.id === "maya");

    expect(welcome).toMatchObject({ panelLayout: "welcome" });
    expect(welcome?.welcomeChoices).toHaveLength(3);
    expect(welcome).not.toHaveProperty("illustration");
    expect(calendar).not.toHaveProperty("panelLayout");
    expect(calendar).not.toHaveProperty("illustration");
    expect(maya).not.toHaveProperty("panelLayout");
    expect(maya).not.toHaveProperty("illustration");
  });

  it("adapts the three welcome choices to the features a member can use", () => {
    useUserRoleMock.mockReturnValue({ userRole: "member" } as ReturnType<
      typeof useUserRole
    >);

    const { result, rerender } = renderHook(() => useWalkthroughSteps());
    const getWelcomeChoiceIds = () =>
      result.current
        .find((step) => step.id === "welcome")
        ?.welcomeChoices?.map((choice) => choice.id);

    expect(getWelcomeChoiceIds()).toEqual([
      "create-story",
      "roadmap",
      "calendar",
    ]);

    useFeaturesMock.mockReturnValue({
      isPending: false,
      objectiveEnabled: false,
    } as ReturnType<typeof useFeatures>);
    rerender();

    expect(getWelcomeChoiceIds()).toEqual(["create-story", "calendar", "maya"]);
  });

  it("gives guests three available ways to start", () => {
    useUserRoleMock.mockReturnValue({ userRole: "guest" } as ReturnType<
      typeof useUserRole
    >);

    const { result } = renderHook(() => useWalkthroughSteps());
    const welcome = result.current.find((step) => step.id === "welcome");

    expect(
      welcome?.welcomeChoices?.map((choice) => choice.targetStepId),
    ).toEqual(["my-work", "calendar", "maya"]);
    expect(result.current.map((step) => step.id)).toEqual([
      "welcome",
      "my-work",
      "calendar",
      "maya",
      "help",
      "teams",
    ]);
  });

  it("waits for the workspace role and settings before starting", () => {
    useUserRoleMock.mockReturnValue({ userRole: undefined } as ReturnType<
      typeof useUserRole
    >);
    useFeaturesMock.mockReturnValue({
      isPending: true,
      objectiveEnabled: true,
    } as ReturnType<typeof useFeatures>);

    const { result } = renderHook(() => useWalkthroughSteps());

    expect(result.current).toEqual([]);
  });

  it("sends an objective-led setup to the Roadmap tip", () => {
    useUserRoleMock.mockReturnValue({ userRole: "member" } as ReturnType<
      typeof useUserRole
    >);

    const { result } = renderHook(() => useWalkthroughSteps());
    const objectiveChoice = result.current
      .find((step) => step.id === "welcome")
      ?.welcomeChoices?.find((choice) => choice.id === "roadmap");

    expect(objectiveChoice?.targetStepId).toBe("roadmap");
  });

  it("opens Roadmap before showing the objective tip in the guided path", () => {
    useUserRoleMock.mockReturnValue({ userRole: "member" } as ReturnType<
      typeof useUserRole
    >);

    const { result } = renderHook(() => useWalkthroughSteps());
    const roadmapStep = result.current.find((step) => step.id === "roadmap");

    act(() => {
      roadmapStep?.action?.();
    });

    expect(routerPushMock).toHaveBeenCalledWith("/acme/roadmap");
  });

  it("opens My Work before task-led guidance so create stays in task mode", () => {
    usePathnameMock.mockReturnValue("/acme/calendar");
    useUserRoleMock.mockReturnValue({ userRole: "member" } as ReturnType<
      typeof useUserRole
    >);

    const { result } = renderHook(() => useWalkthroughSteps());
    const taskChoice = result.current
      .find((step) => step.id === "welcome")
      ?.welcomeChoices?.find((choice) => choice.id === "create-story");
    const taskStep = result.current.find((step) => step.id === "create-story");

    expect(taskChoice?.targetStepId).toBe("create-story");
    expect(taskStep?.target).toBe(walkthroughTargetSelectors.createStory);

    act(() => {
      taskStep?.action?.();
    });

    expect(routerPushMock).toHaveBeenCalledWith("/acme/my-work");
  });

  it("uses action-labelled gates for a real task and Maya message", () => {
    useUserRoleMock.mockReturnValue({ userRole: "member" } as ReturnType<
      typeof useUserRole
    >);

    const { result } = renderHook(() => useWalkthroughSteps());
    const taskStep = result.current.find((step) => step.id === "create-story");
    const mayaStep = result.current.find((step) => step.id === "maya");

    expect(taskStep?.requiredAction).toMatchObject({
      actionLabel: "Create my first storyTerm",
      id: "story-created",
    });
    expect(mayaStep?.requiredAction).toMatchObject({
      actionLabel: "Write my first Maya message",
      id: "maya-message-completed",
    });

    act(() => {
      mayaStep?.action?.();
    });

    expect(openChatMock).toHaveBeenCalledTimes(1);
  });

  it("lets setup continue when Maya is unavailable at the message limit", () => {
    useUserRoleMock.mockReturnValue({ userRole: "member" } as ReturnType<
      typeof useUserRole
    >);
    useMayaMessageAvailabilityMock.mockReturnValue({
      isLimited: true,
      isPending: false,
      isUnavailable: false,
    });

    const { result } = renderHook(() => useWalkthroughSteps());
    const mayaStep = result.current.find((step) => step.id === "maya");

    expect(mayaStep).toMatchObject({
      highlight: false,
      nextActionLabel: "Continue setup",
      position: "center",
      target: "body",
    });
    expect(mayaStep?.requiredAction).toBeUndefined();
    expect(mayaStep?.action).toBeUndefined();
  });

  it("does not make Maya mandatory when availability cannot be checked", () => {
    useUserRoleMock.mockReturnValue({ userRole: "member" } as ReturnType<
      typeof useUserRole
    >);
    useMayaMessageAvailabilityMock.mockReturnValue({
      isLimited: false,
      isPending: false,
      isUnavailable: true,
    });

    const { result } = renderHook(() => useWalkthroughSteps());
    const mayaStep = result.current.find((step) => step.id === "maya");

    expect(mayaStep).toMatchObject({
      highlight: false,
      nextActionLabel: "Continue setup",
      position: "center",
      target: "body",
    });
    expect(mayaStep?.requiredAction).toBeUndefined();
  });

  it("waits for Maya availability before building a gated setup path", () => {
    useUserRoleMock.mockReturnValue({ userRole: "member" } as ReturnType<
      typeof useUserRole
    >);
    useMayaMessageAvailabilityMock.mockReturnValue({
      isLimited: false,
      isPending: true,
      isUnavailable: false,
    });

    const { result } = renderHook(() => useWalkthroughSteps());

    expect(result.current).toEqual([]);
  });

  it("brings the Teams target into view before showing its tip", () => {
    useUserRoleMock.mockReturnValue({ userRole: "member" } as ReturnType<
      typeof useUserRole
    >);
    const teamTarget = document.createElement("button");
    const scrollIntoViewMock = jest.fn();
    Object.defineProperty(teamTarget, "scrollIntoView", {
      value: scrollIntoViewMock,
    });
    teamTarget.dataset.walkthroughTarget = "teams";
    document.body.append(teamTarget);

    const { result } = renderHook(() => useWalkthroughSteps());
    const teamsStep = result.current.find((step) => step.id === "teams");

    act(() => {
      teamsStep?.action?.();
    });

    expect(scrollIntoViewMock).toHaveBeenCalledWith({ block: "nearest" });
    teamTarget.remove();
  });
});

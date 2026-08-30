/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { StrictMode, type ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { useMediaQuery, useLocalStorage } from "@/hooks";
import { useProfile } from "@/lib/hooks/profile";
import { useUpdateProfileMutation } from "@/lib/hooks/update-profile-mutation";
import {
  type WalkthroughStep,
  WalkthroughProvider,
  useWalkthrough,
} from "./walkthrough-provider";

jest.mock("@/hooks", () => ({
  useLocalStorage: jest.fn(),
  useMediaQuery: jest.fn(),
}));

jest.mock("@/lib/hooks/profile", () => ({
  useProfile: jest.fn(),
}));

jest.mock("@/lib/hooks/update-profile-mutation", () => ({
  useUpdateProfileMutation: jest.fn(),
}));

const useMediaQueryMock = jest.mocked(useMediaQuery);
const useLocalStorageMock = jest.mocked(useLocalStorage);
const useProfileMock = jest.mocked(useProfile);
const useUpdateProfileMutationMock = jest.mocked(useUpdateProfileMutation);
const updateProfileMock = jest.fn();

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

const WalkthroughTestProvider = ({ children }: { children: ReactNode }) => (
  <StrictMode>
    <WalkthroughProvider>{children}</WalkthroughProvider>
  </StrictMode>
);

describe("WalkthroughProvider", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    useMediaQueryMock.mockReturnValue(false);
    useLocalStorageMock.mockReturnValue([null, jest.fn()]);
    useProfileMock.mockReturnValue({ data: undefined } as ReturnType<
      typeof useProfile
    >);
    useUpdateProfileMutationMock.mockReturnValue({
      mutate: updateProfileMock,
    } as unknown as ReturnType<typeof useUpdateProfileMutation>);
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
    expect(updateProfileMock).not.toHaveBeenCalled();

    act(() => {
      result.current.nextStep();
    });

    await waitFor(() => {
      expect(updateProfileMock).toHaveBeenCalledTimes(1);
    });
    expect(updateProfileMock).toHaveBeenCalledWith({
      hasSeenWalkthrough: true,
    });
    expect(result.current.state).toMatchObject({
      currentStep: 1,
      hasSeenWalkthrough: true,
      isActive: false,
      totalSteps: 2,
    });
    expect(result.current.state).not.toHaveProperty("completionVersion");
  });
});

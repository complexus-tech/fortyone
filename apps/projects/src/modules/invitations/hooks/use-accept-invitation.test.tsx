/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  acceptUserInvitation,
  InvitationUnavailableError,
} from "../actions/accept-user-invitation";
import { invitationKeys } from "../keys";
import { useAcceptInvitationMutation } from "./use-accept-invitation";

jest.mock("../actions/accept-user-invitation", () => ({
  acceptUserInvitation: jest.fn(),
  InvitationUnavailableError: class extends Error {},
}));
jest.mock("sonner", () => ({
  toast: {
    loading: jest.fn(),
    info: jest.fn(),
    error: jest.fn(),
    success: jest.fn(),
  },
}));
jest.mock("next/navigation", () => ({ redirect: jest.fn() }));

const invitations = [
  { id: "selected", workspaceSlug: "workspace" },
  { id: "other" },
];

describe("account invitation acceptance", () => {
  let queryClient: QueryClient;
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
  beforeEach(() => {
    jest.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    queryClient.setQueryData(invitationKeys.mine, invitations);
  });
  afterEach(() => {
    queryClient.clear();
  });

  it("discards an unavailable invitation without restoring it or offering retry", async () => {
    jest
      .mocked(acceptUserInvitation)
      .mockRejectedValue(new InvitationUnavailableError("No longer available"));
    const { result } = renderHook(useAcceptInvitationMutation, { wrapper });
    act(() => {
      result.current.mutate("selected");
    });
    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
    expect(acceptUserInvitation).toHaveBeenCalledWith("selected");
    expect(queryClient.getQueryData(invitationKeys.mine)).toEqual([
      invitations[1],
    ]);
    expect(queryClient.getQueryState(invitationKeys.mine)?.isInvalidated).toBe(
      true,
    );
    expect(toast.info).toHaveBeenCalled();
    expect(toast.error).not.toHaveBeenCalled();
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("restores the invitation after a transient failure and offers retry", async () => {
    jest
      .mocked(acceptUserInvitation)
      .mockRejectedValue(new Error("Service unavailable"));
    const { result } = renderHook(useAcceptInvitationMutation, { wrapper });
    act(() => {
      result.current.mutate("selected");
    });
    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
    expect(queryClient.getQueryData(invitationKeys.mine)).toEqual(invitations);
    expect(toast.error).toHaveBeenCalledWith(
      "Failed to accept",
      expect.objectContaining({
        action: expect.objectContaining({ label: "Retry" }),
      }),
    );
  });

  it("accepts token-free list entries by ID", async () => {
    jest.mocked(acceptUserInvitation).mockResolvedValue({ data: null });
    const { result } = renderHook(useAcceptInvitationMutation, { wrapper });
    act(() => {
      result.current.mutate("selected");
    });
    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
    expect(queryClient.getQueryData(invitationKeys.mine)).toEqual([
      invitations[1],
    ]);
    expect(toast.success).toHaveBeenCalled();
  });
});

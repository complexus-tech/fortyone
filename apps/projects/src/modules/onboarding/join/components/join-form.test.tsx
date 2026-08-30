/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type * as ReactTypes from "react";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { toast } from "sonner";
import { redirect } from "next/navigation";
import { useWorkspaces } from "@/lib/hooks/workspaces";
import { acceptInvitation } from "@/modules/invitations/public/onboarding";
import type { Invitation } from "@/modules/invitations/public/types";
import { JoinForm } from "./join-form";

jest.mock("@/lib/hooks/workspaces", () => ({
  useWorkspaces: jest.fn(),
}));

jest.mock("@/modules/invitations/public/onboarding", () => ({
  acceptInvitation: jest.fn(),
}));

jest.mock("next/navigation", () => ({
  redirect: jest.fn(),
}));

jest.mock("sonner", () => ({
  toast: {
    error: jest.fn(),
  },
}));

jest.mock("ui", () => ({
  Avatar: ({ name }: { name: string }) => <div>{name}</div>,
  Box: ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  ),
  Button: ({
    children,
    loading = false,
    loadingText,
    onClick,
  }: {
    children: ReactTypes.ReactNode;
    loading?: boolean;
    loadingText?: string;
    onClick?: () => void;
  }) => (
    <button
      data-loading={loading ? "true" : "false"}
      onClick={onClick}
      type="button"
    >
      {loading ? loadingText : children}
    </button>
  ),
  Flex: ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  ),
  Text: ({ children }: { children: ReactTypes.ReactNode }) => (
    <span>{children}</span>
  ),
  Wrapper: ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  ),
}));

type InvitationResponse = Awaited<ReturnType<typeof acceptInvitation>>;

const createDeferred = <T,>() => {
  let resolve: (value: T) => void;
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise;
  });

  return {
    promise,
    resolve: (value: T) => {
      resolve(value);
    },
  };
};

const invitation: Invitation = {
  createdAt: "2026-08-30T00:00:00.000Z",
  email: "ada@example.com",
  expiresAt: "2026-09-30T00:00:00.000Z",
  id: "invitation-1",
  inviterId: "user-1",
  role: "member",
  teamIds: [],
  updatedAt: "2026-08-30T00:00:00.000Z",
  workspaceColor: "#000000",
  workspaceId: "workspace-1",
  workspaceName: "Acme",
  workspaceSlug: "acme",
};

const acceptInvitationMock = jest.mocked(acceptInvitation);
const redirectMock = jest.mocked(redirect);
const toastErrorMock = jest.mocked(toast.error);
const useWorkspacesMock = jest.mocked(useWorkspaces);

describe("JoinForm", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    useWorkspacesMock.mockReturnValue({
      data: [{ id: "existing-workspace" }],
    } as ReturnType<typeof useWorkspaces>);
  });

  it("ignores an older failed request while a newer join request is pending", async () => {
    const firstRequest = createDeferred<InvitationResponse>();
    const secondRequest = createDeferred<InvitationResponse>();
    acceptInvitationMock
      .mockReturnValueOnce(firstRequest.promise)
      .mockReturnValueOnce(secondRequest.promise);

    render(<JoinForm invitation={invitation} token="invite-token" />);

    const joinButton = screen.getByRole("button", {
      name: "Accept invitation",
    });
    fireEvent.click(joinButton);
    fireEvent.click(joinButton);

    await act(async () => {
      firstRequest.resolve({
        error: { message: "The earlier request failed" },
      } as InvitationResponse);
      await firstRequest.promise;
    });

    expect(joinButton).toHaveAttribute("data-loading", "true");
    expect(toastErrorMock).not.toHaveBeenCalled();
    expect(redirectMock).not.toHaveBeenCalled();

    await act(async () => {
      secondRequest.resolve({
        error: { message: "The latest request failed" },
      } as InvitationResponse);
      await secondRequest.promise;
    });

    await waitFor(() => {
      expect(toastErrorMock).toHaveBeenCalledWith("Failed to join workspace", {
        description: "The latest request failed",
      });
    });
    expect(joinButton).toHaveAttribute("data-loading", "false");
  });
});

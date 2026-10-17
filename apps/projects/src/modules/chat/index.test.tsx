/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { ChatPage } from "./index";

const mockCapabilities = {
  enabled: false,
  isPending: false,
  isError: false,
  isFetching: false,
  refetch: jest.fn(),
};
const mockUseMaya = jest.fn(() => ({ data: { avatarUrl: "/maya.png" } }));
jest.mock("@/lib/hooks/workspace/chat-capabilities", () => ({
  useChatCapabilities: () => mockCapabilities,
}));
jest.mock("@/lib/hooks/members", () => ({
  useMayaAssignee: () => mockUseMaya(),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("@/components/shared/mobile-menu", () => ({
  MobileMenuButton: () => null,
}));
jest.mock(
  "next/dynamic",
  () => () =>
    function MockChatShell() {
      return <div>Enabled chat shell</div>;
    },
);
jest.mock("ui", () => ({
  Text: ({ children }: { children: ReactNode }) => <p>{children}</p>,
  Button: ({
    children,
    onClick,
    disabled,
  }: {
    children: ReactNode;
    onClick: () => void;
    disabled: boolean;
  }) => (
    <button disabled={disabled} onClick={onClick} type="button">
      {children}
    </button>
  ),
}));

beforeEach(() => {
  Object.assign(mockCapabilities, {
    enabled: false,
    isPending: false,
    isError: false,
    isFetching: false,
  });
  mockCapabilities.refetch.mockClear();
  mockUseMaya.mockClear();
});

describe("chat rollout boundary", () => {
  it("does not mount chat or fetch Maya for a disabled direct URL", () => {
    render(<ChatPage />);
    expect(
      screen.getByText("Chat is turned off for this workspace."),
    ).toBeInTheDocument();
    expect(mockUseMaya).not.toHaveBeenCalled();
    expect(screen.queryByText("Enabled chat shell")).not.toBeInTheDocument();
  });
  it("unmounts the conversation when its rollout is revoked", () => {
    mockCapabilities.enabled = true;
    const view = render(<ChatPage />);
    expect(screen.getByText("Enabled chat shell")).toBeInTheDocument();
    mockCapabilities.enabled = false;
    view.rerender(<ChatPage />);
    expect(screen.queryByText("Enabled chat shell")).not.toBeInTheDocument();
  });
  it("distinguishes loading from an availability failure and offers retry", () => {
    mockCapabilities.isPending = true;
    const view = render(<ChatPage />);
    expect(screen.getByText("Checking chat availability…")).toBeInTheDocument();
    mockCapabilities.isPending = false;
    mockCapabilities.isError = true;
    view.rerender(<ChatPage />);
    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    expect(mockCapabilities.refetch).toHaveBeenCalledTimes(1);
  });
});

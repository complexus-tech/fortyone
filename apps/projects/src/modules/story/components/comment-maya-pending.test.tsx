/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { fireEvent, render, screen } from "@testing-library/react";
import type { CommentMayaPendingStatus } from "../types/comment-maya";
import {
  CommentMayaPending,
  CommentMayaPendingLabel,
} from "./comment-maya-pending";

const MAYA = {
  id: "55555555-5555-4555-8555-555555555555",
  fullName: "Maya",
  username: "maya",
  avatarUrl: "/maya.png",
  isActive: true,
  isSystem: true,
};
let mockMaya: typeof MAYA | undefined = MAYA;
jest.mock("@/lib/hooks/members", () => ({
  useMayaAssignee: () => ({ data: mockMaya }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ withWorkspace: (path: string) => `/acme${path}` }),
}));
beforeEach(() => {
  mockMaya = MAYA;
});

describe("Maya pending reply", () => {
  it("uses Maya's ordinary identity and announces one coherent status with decorative dots", () => {
    const { container } = render(<CommentMayaPending status="thinking" />);
    expect(screen.getByText("maya")).toBeVisible();
    expect(screen.getByAltText("Maya")).toHaveAttribute("src", "/maya.png");
    const status = screen.getByRole("status");
    expect(status).toHaveAttribute("aria-live", "polite");
    expect(status).toHaveAttribute("aria-atomic", "true");
    expect(status.querySelector(".sr-only")).toHaveTextContent(
      "Maya is thinking",
    );
    expect(status.querySelector("[aria-hidden=true]")).toHaveTextContent(
      "Thinking",
    );
    expect(status.querySelectorAll(".comment-maya-dots > span")).toHaveLength(
      3,
    );
    expect(container.querySelector("time")).toBeNull();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it("keeps the same plain Maya identity while the canonical member is loading", () => {
    mockMaya = undefined;
    const { container } = render(
      <CommentMayaPending isDialog status="thinking" />,
    );
    expect(screen.getByText("maya")).toBeVisible();
    expect(screen.getByTitle("Maya")).toHaveTextContent("MA");
    expect(container.querySelector("a")).toBeNull();
  });

  it.each<CommentMayaPendingStatus>([
    "thinking",
    "working",
    "replying",
    "delayed",
    "offline",
    "failed",
  ])(
    "provides a compact noninteractive %s label and limits dots to active states",
    (status) => {
      const { container } = render(
        <CommentMayaPendingLabel compact status={status} />,
      );
      const visible = screen
        .getByRole("status")
        .querySelector("[aria-hidden=true]")!;
      expect(visible).toHaveTextContent(/Maya/);
      const active = ["thinking", "working", "replying"].includes(status);
      expect(
        container.querySelectorAll(".comment-maya-dots > span"),
      ).toHaveLength(active ? 3 : 0);
      expect(screen.queryByRole("button")).not.toBeInTheDocument();
    },
  );

  it("offers a guarded Check again only when the response is delayed or unavailable", () => {
    const onCheckAgain = jest.fn();
    const { rerender } = render(
      <CommentMayaPending onCheckAgain={onCheckAgain} status="thinking" />,
    );
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
    rerender(
      <CommentMayaPending onCheckAgain={onCheckAgain} status="delayed" />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Check again" }));
    expect(onCheckAgain).toHaveBeenCalledTimes(1);
    rerender(
      <CommentMayaPending
        checkAgainDisabled
        onCheckAgain={onCheckAgain}
        status="offline"
      />,
    );
    expect(screen.getByText("Couldn’t check for a response.")).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "Check again" }));
    expect(onCheckAgain).toHaveBeenCalledTimes(1);
  });

  it("offers Retry only for a failed request with an allowed handler", () => {
    const onRetry = jest.fn();
    const { rerender } = render(<CommentMayaPending status="failed" />);
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
    rerender(<CommentMayaPending onRetry={onRetry} status="failed" />);
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    expect(onRetry).toHaveBeenCalledTimes(1);
    rerender(
      <CommentMayaPending onRetry={onRetry} retryDisabled status="failed" />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    expect(onRetry).toHaveBeenCalledTimes(1);
    rerender(<CommentMayaPending onRetry={onRetry} status="working" />);
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });
});

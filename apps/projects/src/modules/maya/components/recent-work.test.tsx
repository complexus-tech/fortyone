/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { RecentWork } from "./recent-work";

let mockTab = "created";
let mockRole = "admin";
const mockSetTab = jest.fn();
jest.mock("../hooks/use-work-tab", () => ({
  useWorkTab: () => [mockTab, mockSetTab],
}));
jest.mock("../hooks/use-recent-work", () => ({
  useRecentWork: () => ({
    stories: [],
    isPending: false,
    isError: false,
    retry: jest.fn(),
  }),
}));
jest.mock("@/hooks/role", () => ({
  useUserRole: () => ({ userRole: mockRole }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ withWorkspace: (path: string) => `/acme${path}` }),
}));
jest.mock("@/hooks/use-terminology-display", () => ({
  useTerminology: () => ({
    getTermDisplay: (
      _: string,
      options: { variant?: string; capitalize?: boolean } = {},
    ) => {
      const term = options.variant === "plural" ? "issues" : "issue";
      return options.capitalize ? term[0].toUpperCase() + term.slice(1) : term;
    },
  }),
}));
jest.mock("./work-attention", () => ({ WorkAttention: () => null }));
jest.mock("@/components/ui/stories-board", () => ({
  StoriesBoard: () => null,
}));
jest.mock("@/components/ui/illustrations/empty-state-illustrations", () => ({
  WorkEmptyIllustration: () => null,
}));
jest.mock("next/dynamic", () => () => {
  const MockDialog = ({ onCreated }: { onCreated: () => void }) => (
    <div role="dialog">
      <button onClick={onCreated} type="button">
        Complete creation
      </button>
    </div>
  );
  return MockDialog;
});
jest.mock("next/link", () => {
  const MockLink = ({
    children,
    href,
  }: {
    children: ReactNode;
    href: string;
  }) => <a href={href}>{children}</a>;
  return MockLink;
});
jest.mock("ui", () => {
  const Container = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );
  return {
    Box: Container,
    Text: Container,
    Skeleton: Container,
    Button: ({
      children,
      onClick,
    }: {
      children: ReactNode;
      onClick?: () => void;
    }) => (
      <button onClick={onClick} type="button">
        {children}
      </button>
    ),
    Tabs: Object.assign(Container, {
      List: Container,
      Tab: Container,
      Panel: Container,
    }),
  };
});
beforeEach(() => {
  mockTab = "created";
  mockRole = "admin";
  mockSetTab.mockReset();
});

describe("Maya empty-state actions", () => {
  it("opens the creation dialog with workspace terminology and selects Created after success", () => {
    render(<RecentWork />);
    fireEvent.click(screen.getByRole("button", { name: "Create issue" }));
    expect(screen.getByRole("dialog")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Complete creation" }));
    expect(mockSetTab).toHaveBeenCalledWith("created");
  });
  it("offers browsing for an empty Assigned tab", () => {
    mockTab = "assigned";
    render(<RecentWork />);
    expect(
      screen
        .getByRole("link", { name: "Browse all issues" })
        .getAttribute("href"),
    ).toBe("/acme/my-work?tab=all");
    expect(screen.queryByRole("button", { name: "Create issue" })).toBeNull();
  });
  it("does not offer creation to guests", () => {
    mockRole = "guest";
    render(<RecentWork />);
    expect(screen.queryByRole("button", { name: "Create issue" })).toBeNull();
    expect(
      screen.getByRole("link", { name: "Browse all issues" }),
    ).toBeTruthy();
  });
});

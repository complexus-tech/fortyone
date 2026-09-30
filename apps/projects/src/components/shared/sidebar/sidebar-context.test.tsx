/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import { useLocalStorage, useWorkspacePath } from "@/hooks";
import { SidebarProvider, useSidebar } from "./sidebar-context";

jest.mock("@/hooks", () => ({
  useLocalStorage: jest.fn(),
  useWorkspacePath: jest.fn(),
}));

jest.mock("react-hotkeys-hook", () => ({
  useHotkeys: jest.fn(),
}));

const useLocalStorageMock = jest.mocked(useLocalStorage);
const useWorkspacePathMock = jest.mocked(useWorkspacePath);

const SidebarState = () => {
  const { isCollapsed } = useSidebar();

  return <span>{isCollapsed ? "collapsed" : "expanded"}</span>;
};

const AssistantCardState = () => {
  const { hasAssistantCards, setHasAssistantCards } = useSidebar();

  return (
    <>
      <span>
        {hasAssistantCards ? "assistant cards" : "no assistant cards"}
      </span>
      <button
        onClick={() => {
          setHasAssistantCards(true);
        }}
        type="button"
      >
        Show assistant cards
      </button>
      <button
        onClick={() => {
          setHasAssistantCards(false);
        }}
        type="button"
      >
        Hide assistant cards
      </button>
    </>
  );
};

describe("SidebarProvider", () => {
  beforeEach(() => {
    useLocalStorageMock.mockImplementation((_key, initialValue) => [
      initialValue,
      jest.fn(),
    ]);
    useWorkspacePathMock.mockReturnValue({
      workspaceSlug: "acme",
    } as ReturnType<typeof useWorkspacePath>);
  });

  it("collapses the sidebar when no preference has been saved", () => {
    render(
      <SidebarProvider>
        <SidebarState />
      </SidebarProvider>,
    );

    expect(screen.getByText("collapsed")).toBeInTheDocument();
    expect(useLocalStorageMock).toHaveBeenCalledWith(
      "sidebar:acme:collapsed",
      true,
    );
  });

  it("respects a saved expanded preference", () => {
    useLocalStorageMock.mockReturnValue([false, jest.fn()]);

    render(
      <SidebarProvider>
        <SidebarState />
      </SidebarProvider>,
    );

    expect(screen.getByText("expanded")).toBeInTheDocument();
  });

  it("shares assistant-card visibility independently of the collapse preference", () => {
    render(
      <SidebarProvider>
        <AssistantCardState />
        <SidebarState />
      </SidebarProvider>,
    );

    expect(screen.getByText("no assistant cards")).toBeInTheDocument();

    fireEvent.click(
      screen.getByRole("button", { name: "Show assistant cards" }),
    );
    expect(screen.getByText("assistant cards")).toBeInTheDocument();
    expect(screen.getByText("collapsed")).toBeInTheDocument();

    fireEvent.click(
      screen.getByRole("button", { name: "Hide assistant cards" }),
    );
    expect(screen.getByText("no assistant cards")).toBeInTheDocument();
  });
});

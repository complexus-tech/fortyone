/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { Root } from "react-dom/client";
import { act, fireEvent, render, screen, within } from "@testing-library/react";
import { hydrateRoot } from "react-dom/client";
import { renderToString } from "react-dom/server";
import { useLocalStorage, useWorkspacePath } from "@/hooks";
import { useLocalStorage as useStoredPreference } from "@/hooks/local-storage";
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

const SidebarControls = () => {
  const { isCollapsed, toggleSidebar } = useSidebar();

  return (
    <aside aria-label="Sidebar" data-collapsed={isCollapsed}>
      <SidebarState />
      <button onClick={toggleSidebar} type="button">
        Toggle sidebar
      </button>
    </aside>
  );
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
    localStorage.clear();
    jest.clearAllMocks();
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
      { initializeWithValue: false },
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

  it("hydrates the default sidebar before restoring saved expansion and keeps workspace preferences isolated", async () => {
    useLocalStorageMock.mockImplementation(useStoredPreference);
    localStorage.setItem("sidebar:acme:collapsed", "false");
    localStorage.setItem("sidebar:other:collapsed", "true");
    const tree = (
      <SidebarProvider>
        <SidebarControls />
      </SidebarProvider>
    );

    const serverStorage = jest
      .spyOn(Storage.prototype, "getItem")
      .mockReturnValue(null);
    let serverHtml: string;
    try {
      serverHtml = renderToString(tree);
      expect(serverStorage).not.toHaveBeenCalled();
    } finally {
      serverStorage.mockRestore();
    }

    const container = document.createElement("div");
    container.innerHTML = serverHtml;
    document.body.appendChild(container);
    const sidebar = () =>
      within(container).getByRole("complementary", { name: "Sidebar" });
    const toggle = () =>
      within(container).getByRole("button", { name: "Toggle sidebar" });
    expect(sidebar()).toHaveAttribute("data-collapsed", "true");

    const onRecoverableError = jest.fn();
    const consoleError = jest
      .spyOn(console, "error")
      .mockImplementation(() => {});
    let root: Root | undefined;
    try {
      await act(async () => {
        root = hydrateRoot(container, tree, { onRecoverableError });
      });
      expect(onRecoverableError).not.toHaveBeenCalled();
      expect(consoleError).not.toHaveBeenCalled();
      expect(sidebar()).toHaveAttribute("data-collapsed", "false");
      expect(localStorage.getItem("sidebar:acme:collapsed")).toBe("false");

      fireEvent.click(toggle());
      expect(sidebar()).toHaveAttribute("data-collapsed", "true");
      expect(localStorage.getItem("sidebar:acme:collapsed")).toBe("true");
      fireEvent.click(toggle());
      expect(sidebar()).toHaveAttribute("data-collapsed", "false");

      useWorkspacePathMock.mockReturnValue({
        workspaceSlug: "other",
      } as ReturnType<typeof useWorkspacePath>);
      await act(async () => {
        root?.render(
          <SidebarProvider>
            <SidebarControls />
          </SidebarProvider>,
        );
      });
      expect(sidebar()).toHaveAttribute("data-collapsed", "true");
      fireEvent.click(toggle());
      expect(sidebar()).toHaveAttribute("data-collapsed", "false");
      expect(localStorage.getItem("sidebar:other:collapsed")).toBe("false");

      useWorkspacePathMock.mockReturnValue({
        workspaceSlug: "acme",
      } as ReturnType<typeof useWorkspacePath>);
      await act(async () => {
        root?.render(
          <SidebarProvider>
            <SidebarControls />
          </SidebarProvider>,
        );
      });
      expect(sidebar()).toHaveAttribute("data-collapsed", "false");
      expect(localStorage.getItem("sidebar:acme:collapsed")).toBe("false");
      expect(onRecoverableError).not.toHaveBeenCalled();
      expect(consoleError).not.toHaveBeenCalled();
    } finally {
      await act(async () => {
        root?.unmount();
      });
      container.remove();
      consoleError.mockRestore();
    }
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

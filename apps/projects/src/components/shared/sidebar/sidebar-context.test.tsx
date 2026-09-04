/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { render, screen } from "@testing-library/react";
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
});

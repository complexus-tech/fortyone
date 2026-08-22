/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { SidebarEdgeToggle } from "./sidebar-edge-toggle";

let mockIsCollapsed = false;
const mockToggleSidebar = jest.fn();

jest.mock("ui", () => ({
  Tooltip: ({ children }: { children: ReactNode }) => children,
}));

jest.mock("icons", () => ({
  SidebarEdgeToggleIcon: ({ direction }: { direction: "left" | "right" }) => (
    <svg aria-label={`${direction} direction`} />
  ),
}));

jest.mock("./sidebar-context", () => ({
  useSidebar: () => ({
    isCollapsed: mockIsCollapsed,
    toggleSidebar: mockToggleSidebar,
  }),
}));

describe("SidebarEdgeToggle", () => {
  beforeEach(() => {
    mockIsCollapsed = false;
    mockToggleSidebar.mockClear();
  });

  it("collapses the expanded sidebar", () => {
    render(<SidebarEdgeToggle />);

    const toggle = screen.getByRole("button", { name: "Collapse sidebar" });
    expect(screen.getByLabelText("left direction")).toBeInTheDocument();

    fireEvent.click(toggle);

    expect(mockToggleSidebar).toHaveBeenCalledTimes(1);
  });

  it("expands the collapsed sidebar", () => {
    mockIsCollapsed = true;

    render(<SidebarEdgeToggle />);

    expect(
      screen.getByRole("button", { name: "Expand sidebar" }),
    ).toBeInTheDocument();
    expect(screen.getByLabelText("right direction")).toBeInTheDocument();
  });
});

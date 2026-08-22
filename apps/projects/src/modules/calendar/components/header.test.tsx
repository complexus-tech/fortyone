/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { CalendarHeader } from "./header";

const mockUseAppCommandAction = jest.fn();

jest.mock("icons", () => ({
  CalendarIcon: () => <span aria-hidden>Calendar icon</span>,
  PlusIcon: () => <span aria-hidden>Plus icon</span>,
}));

jest.mock("ui", () => ({
  Box: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  BreadCrumbs: ({ breadCrumbs }: { breadCrumbs: { name: string }[] }) => (
    <div>{breadCrumbs.map((item) => item.name).join(" / ")}</div>
  ),
  Button: ({
    children,
    color,
    onClick,
  }: {
    children: ReactNode;
    color?: string;
    onClick: () => void;
  }) => (
    <button data-color={color} onClick={onClick} type="button">
      {children}
    </button>
  ),
  Flex: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

jest.mock("@/components/shared", () => ({
  HeaderContainer: ({ children }: { children: ReactNode }) => (
    <header>{children}</header>
  ),
  MobileMenuButton: () => <button type="button">Open navigation</button>,
  useAppCommandAction: (action: unknown) => mockUseAppCommandAction(action),
}));

jest.mock("@/hooks", () => ({
  useTerminology: () => ({
    getTermDisplay: () => "task",
  }),
  useUserRole: () => ({ userRole: "admin" }),
}));

describe("CalendarHeader", () => {
  it("labels the standalone page and registers task scheduling", () => {
    const onSchedule = jest.fn();

    render(<CalendarHeader onSchedule={onSchedule} />);

    expect(screen.getByText("Calendar")).toBeInTheDocument();
    expect(mockUseAppCommandAction).toHaveBeenCalledWith({
      disabled: false,
      id: "calendar:schedule-story",
      label: "Schedule task",
      onSelect: onSchedule,
    });
  });
});

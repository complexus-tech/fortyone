/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { CalendarHeader } from "./header";

const mockUseAppCommandAction = jest.fn();

jest.mock("icons", () => ({
  ArrowDown2Icon: () => <span aria-hidden>Open views</span>,
  CalendarIcon: () => <span data-testid="calendar-icon" />,
  ChevronLeftIcon: () => <span aria-hidden>Previous</span>,
  ChevronRightIcon: () => <span aria-hidden>Next</span>,
  TimeScheduleIcon: () => <span aria-hidden>Focus time</span>,
}));

jest.mock("ui", () => ({
  Box: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  BreadCrumbs: ({
    breadCrumbs,
  }: {
    breadCrumbs: { icon?: ReactNode; name: string }[];
  }) => (
    <div>
      {breadCrumbs.map((item) => item.name).join(" / ")}
      {breadCrumbs.map((item) => (
        <span key={item.name}>{item.icon}</span>
      ))}
    </div>
  ),
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
  Flex: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  Menu: Object.assign(
    ({ children }: { children: ReactNode }) => <div>{children}</div>,
    {
      Button: ({ children }: { children: ReactNode }) => <>{children}</>,
      Group: ({ children }: { children: ReactNode }) => <div>{children}</div>,
      Item: ({ children }: { children: ReactNode }) => (
        <button type="button">{children}</button>
      ),
      Items: ({ children }: { children: ReactNode }) => <div>{children}</div>,
    },
  ),
  Text: ({ children }: { children: ReactNode }) => <div>{children}</div>,
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

    render(
      <CalendarHeader
        canNavigateNext
        canNavigatePrevious
        currentView="week"
        onFocus={jest.fn()}
        onNext={jest.fn()}
        onPrevious={jest.fn()}
        onSchedule={onSchedule}
        onToday={jest.fn()}
        onViewChange={jest.fn()}
        title="August 2026"
      />,
    );

    expect(screen.getByText("Calendar / Week")).toBeInTheDocument();
    expect(screen.getByTestId("calendar-icon")).toBeInTheDocument();
    expect(screen.getByText("August 2026")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Today" })).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Block focus time" }),
    ).toBeInTheDocument();
    expect(mockUseAppCommandAction).toHaveBeenCalledWith({
      disabled: false,
      id: "calendar:schedule-story",
      label: "Schedule task",
      onSelect: onSchedule,
    });
  });
});

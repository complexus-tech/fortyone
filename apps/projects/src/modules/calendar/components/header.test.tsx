/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { CalendarHeader } from "./header";

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
    onClick,
  }: {
    children: ReactNode;
    onClick: () => void;
  }) => (
    <button onClick={onClick} type="button">
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
}));

jest.mock("@/hooks", () => ({
  useTerminology: () => ({
    getTermDisplay: () => "task",
  }),
}));

describe("CalendarHeader", () => {
  it("labels the standalone page and opens task scheduling", () => {
    const onSchedule = jest.fn();

    render(<CalendarHeader onSchedule={onSchedule} />);

    expect(screen.getByText("Calendar")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Schedule task" }));
    expect(onSchedule).toHaveBeenCalledTimes(1);
  });
});

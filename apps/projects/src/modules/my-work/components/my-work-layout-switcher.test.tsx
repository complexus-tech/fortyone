/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { MyWorkLayoutSwitcher } from "./my-work-layout-switcher";

jest.mock("icons", () => ({
  CalendarIcon: () => <span aria-hidden>Calendar icon</span>,
  KanbanIcon: () => <span aria-hidden>Board icon</span>,
  ListIcon: () => <span aria-hidden>List icon</span>,
}));

jest.mock("lib", () => ({
  cn: (...values: unknown[]) => values.filter(Boolean).join(" "),
}));

jest.mock("ui", () => ({
  Flex: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

describe("MyWorkLayoutSwitcher", () => {
  it("offers calendar without changing the shared story layout switcher", () => {
    const setLayout = jest.fn();

    render(<MyWorkLayoutSwitcher layout="list" setLayout={setLayout} />);

    expect(screen.getByRole("button", { name: "List" })).toHaveAttribute(
      "aria-pressed",
      "true",
    );
    fireEvent.click(screen.getByRole("button", { name: "Calendar" }));
    expect(setLayout).toHaveBeenCalledWith("calendar");
  });

  it("keeps the calendar option desktop-only", () => {
    render(
      <MyWorkLayoutSwitcher
        layout="list"
        setLayout={jest.fn()}
        showCalendar={false}
      />,
    );

    expect(screen.queryByRole("button", { name: "Calendar" })).toBeNull();
  });
});

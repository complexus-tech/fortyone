/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import { CalendarPage } from "./index";

jest.mock("next/dynamic", () => ({
  __esModule: true,
  default: () =>
    function MockPersonalCalendar({
      isScheduleDialogOpen,
      onScheduleDialogOpenChange,
    }: {
      isScheduleDialogOpen: boolean;
      onScheduleDialogOpenChange: (open: boolean) => void;
    }) {
      return (
        <>
          <div data-testid="personal-calendar">
            {isScheduleDialogOpen ? "Schedule open" : "Schedule closed"}
          </div>
          <button
            onClick={() => {
              onScheduleDialogOpenChange(true);
            }}
            type="button"
          >
            Schedule task
          </button>
        </>
      );
    },
}));

jest.mock("ui", () => ({
  Box: () => null,
  Skeleton: () => null,
}));

describe("CalendarPage", () => {
  it("passes scheduling state to the calendar", () => {
    render(<CalendarPage />);

    expect(screen.getByTestId("personal-calendar")).toHaveTextContent(
      "Schedule closed",
    );
    fireEvent.click(screen.getByRole("button", { name: "Schedule task" }));
    expect(screen.getByTestId("personal-calendar")).toHaveTextContent(
      "Schedule open",
    );
  });
});

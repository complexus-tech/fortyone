/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import { CalendarPage } from "./index";

jest.mock("next/dynamic", () => ({
  __esModule: true,
  default: () =>
    function MockPersonalCalendar({
      isScheduleDialogOpen,
    }: {
      isScheduleDialogOpen: boolean;
    }) {
      return (
        <div data-testid="personal-calendar">
          {isScheduleDialogOpen ? "Schedule open" : "Schedule closed"}
        </div>
      );
    },
}));

jest.mock("ui", () => ({
  Box: () => null,
  Skeleton: () => null,
}));

jest.mock("./components/header", () => ({
  CalendarHeader: ({ onSchedule }: { onSchedule: () => void }) => (
    <button onClick={onSchedule} type="button">
      Schedule task
    </button>
  ),
}));

describe("CalendarPage", () => {
  it("opens scheduling from the standalone page header", () => {
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

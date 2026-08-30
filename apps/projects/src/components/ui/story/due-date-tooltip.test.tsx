/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type * as ReactModule from "react";
import { render, screen } from "@testing-library/react";
import { getDueDateMessage } from "./due-date-tooltip";

jest.mock("ui", () => {
  const React = jest.requireActual<typeof ReactModule>("react");

  return {
    Text: ({ children }: { children?: ReactModule.ReactNode }) =>
      React.createElement("span", null, children),
  };
});

describe("getDueDateMessage", () => {
  it("uses deterministic absolute copy before hydration", () => {
    render(
      <div>{getDueDateMessage(new Date(2026, 7, 30), "story", null)}</div>,
    );

    expect(screen.getByText("Due on Aug 30, 2026")).toBeInTheDocument();
    expect(screen.queryByText(/Due in|overdue|yesterday/i)).toBeNull();
  });

  it("retains the relative due-date copy after hydration", () => {
    render(
      <div>
        {getDueDateMessage(
          new Date(2026, 7, 31),
          "story",
          new Date(2026, 7, 30, 12),
        )}
      </div>,
    );

    expect(screen.getByText("Due tomorrow")).toBeInTheDocument();
  });
});

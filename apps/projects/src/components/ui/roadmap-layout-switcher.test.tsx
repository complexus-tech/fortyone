/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { render, screen } from "@testing-library/react";
import { RoadmapLayoutSwitcher } from "./roadmap-layout-switcher";

jest.mock("ui", () => ({ Flex: "div" }));
jest.mock("icons", () => ({
  GanttIcon: "span",
  KanbanIcon: "span",
  ListIcon: "span",
}));

describe("RoadmapLayoutSwitcher", () => {
  it("keeps Timeline, Board, and List available while page content loads", () => {
    render(<RoadmapLayoutSwitcher layout="gantt" setLayout={jest.fn()} />);

    expect(screen.getByRole("button", { name: "Timeline" })).toBeVisible();
    expect(screen.getByRole("button", { name: "Board" })).toBeVisible();
    expect(screen.getByRole("button", { name: "List" })).toBeVisible();
  });
});

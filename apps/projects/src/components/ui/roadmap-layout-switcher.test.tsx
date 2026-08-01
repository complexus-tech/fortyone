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
  it("keeps Board, Timeline, and List in the expected order", () => {
    render(<RoadmapLayoutSwitcher layout="gantt" setLayout={jest.fn()} />);

    const viewButtons = screen.getAllByRole("button");

    expect(viewButtons).toHaveLength(3);
    expect(viewButtons[0]).toHaveAccessibleName("Board");
    expect(viewButtons[1]).toHaveAccessibleName("Timeline");
    expect(viewButtons[2]).toHaveAccessibleName("List");
  });
});

/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { Objective } from "@/modules/objectives/types";
import { RoadmapObjectiveDetails } from "./objective-details";

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ withWorkspace: (path: string) => `/acme${path}` }),
}));
jest.mock("@/modules/objectives/hooks/use-objective", () => ({
  useObjective: () => ({ data: undefined }),
}));
jest.mock("@/modules/objectives/hooks/use-key-results", () => ({
  useKeyResults: () => ({ data: [] }),
}));
jest.mock("@/modules/objectives/hooks/objective-analytics", () => ({
  useObjectiveAnalytics: () => ({ data: undefined }),
}));
jest.mock("@/modules/objectives/components/objective-forecast-risk", () => ({
  ObjectiveForecastRiskBanner: () => null,
}));
jest.mock("@/modules/objectives/stories/progress-chart", () => ({
  ProgressChart: () => null,
}));
jest.mock("./objective-details-key-results", () => ({
  ObjectiveDetailsKeyResults: () => null,
}));
jest.mock("./objective-details-properties", () => ({
  ObjectiveDetailsProperties: () => (
    <button
      onKeyDown={(event) => {
        if (event.key === "Escape") event.preventDefault();
      }}
      type="button"
    >
      Nested menu control
    </button>
  ),
}));

const OBJECTIVE = {
  id: "objective-1",
  teamId: "team-1",
  name: "Improve reliability",
  color: "#f97316",
} as Objective;

describe("timeline objective details", () => {
  it("names the preview, exposes the full-page link and restores its opener focus", () => {
    const opener = document.createElement("button");
    document.body.append(opener);
    opener.focus();
    const { unmount } = render(
      <RoadmapObjectiveDetails objective={OBJECTIVE} onClose={jest.fn()} />,
    );
    expect(screen.getByRole("region", { name: OBJECTIVE.name })).toHaveFocus();
    expect(screen.getByRole("link", { name: OBJECTIVE.name })).toHaveAttribute(
      "href",
      "/acme/teams/team-1/objectives/objective-1",
    );
    unmount();
    expect(opener).toHaveFocus();
    opener.remove();
  });

  it("closes with Escape while leaving nested menu Escape handling intact", () => {
    const onClose = jest.fn();
    render(<RoadmapObjectiveDetails objective={OBJECTIVE} onClose={onClose} />);
    fireEvent.keyDown(
      screen.getByRole("button", { name: "Nested menu control" }),
      {
        key: "Escape",
      },
    );
    expect(onClose).not.toHaveBeenCalled();
    fireEvent.keyDown(screen.getByRole("region", { name: OBJECTIVE.name }), {
      key: "Escape",
    });
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("does not steal focus from another control when the preview closes", () => {
    const { unmount } = render(
      <RoadmapObjectiveDetails objective={OBJECTIVE} onClose={jest.fn()} />,
    );
    const outside = document.createElement("button");
    document.body.append(outside);
    outside.focus();
    unmount();
    expect(outside).toHaveFocus();
    outside.remove();
  });
});

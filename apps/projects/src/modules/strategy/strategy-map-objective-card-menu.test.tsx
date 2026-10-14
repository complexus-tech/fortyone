/* global afterEach, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { ContextMenu } from "ui";
import { StrategyMapObjectiveContextMenu } from "./strategy-map-objective-card-menu";

afterEach(cleanup);

const renderMenu = (canEdit = true, expanded = false, keyResultCount = 2) => {
  const onAddKeyResult = jest.fn();
  const onOpenDetails = jest.fn();
  const onToggleKeyResults = jest.fn();
  render(
    <ContextMenu>
      <ContextMenu.Trigger>
        <button type="button">Objective card</button>
      </ContextMenu.Trigger>
      <StrategyMapObjectiveContextMenu
        canEdit={canEdit}
        currentPillarId={null}
        isKeyResultsExpanded={expanded}
        keyResultCount={keyResultCount}
        objectiveId="objective-1"
        objectivePath="/acme/teams/team-1/objectives/objective-1"
        objectiveStatusId="status-1"
        onAddKeyResult={onAddKeyResult}
        onAlign={jest.fn()}
        onOpenDetails={onOpenDetails}
        onSetStatus={jest.fn()}
        onToggleKeyResults={onToggleKeyResults}
        pillars={[]}
        statuses={[]}
      />
    </ContextMenu>,
  );
  fireEvent.contextMenu(screen.getByRole("button", { name: "Objective card" }));
  return { onAddKeyResult, onOpenDetails, onToggleKeyResults };
};

it("opens objective details without navigating away from the map", () => {
  const { onOpenDetails } = renderMenu();
  fireEvent.click(screen.getByRole("menuitem", { name: "Open details" }));
  expect(onOpenDetails).toHaveBeenCalledTimes(1);
});

it("offers key-result creation even when the objective has no key results", () => {
  const { onAddKeyResult } = renderMenu(true, false, 0);
  expect(
    screen.queryByRole("menuitem", { name: "Expand key results" }),
  ).toBeNull();
  fireEvent.click(screen.getByRole("menuitem", { name: "Add key result..." }));
  expect(onAddKeyResult).toHaveBeenCalledTimes(1);
});

it.each([false, true])(
  "toggles key-result visibility from the menu (expanded=%s)",
  (expanded) => {
    const { onToggleKeyResults } = renderMenu(true, expanded);
    fireEvent.click(
      screen.getByRole("menuitem", {
        name: expanded ? "Collapse key results" : "Expand key results",
      }),
    );
    expect(onToggleKeyResults).toHaveBeenCalledTimes(1);
  },
);

it("keeps navigation and local visibility actions available to read-only users", () => {
  const { onToggleKeyResults } = renderMenu(false);
  expect(
    screen.queryByRole("menuitem", { name: "Add key result..." }),
  ).toBeNull();
  expect(
    screen.getByRole("menuitem", { name: "Open objective" }),
  ).toHaveAttribute("href", "/acme/teams/team-1/objectives/objective-1");
  fireEvent.click(screen.getByRole("menuitem", { name: "Expand key results" }));
  expect(onToggleKeyResults).toHaveBeenCalledTimes(1);
});

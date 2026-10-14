import { memo } from "react";
import type { ComponentProps } from "react";
import { render, screen } from "@testing-library/react";
import { BoardProvider, useBoard, useBoardView } from "./board-context";

type BoardValue = ComponentProps<typeof BoardProvider>["value"];

const settings: Omit<BoardValue, "selectedStories"> = {
  embedded: false,
  setSelectedStories: jest.fn(),
  viewOptions: {
    groupBy: "status",
    orderBy: "created",
    orderDirection: "desc",
    showEmptyGroups: true,
    showSubStories: true,
    displayColumns: ["Status"],
  },
  isColumnVisible: () => true,
  updateStory: jest.fn(),
  newStoryDefaults: { teamId: "team-a" },
};

it("keeps 250 view subscribers idle during selection changes and updates them when display settings change", () => {
  const renderCard = jest.fn();
  const Card = memo(function Card() {
    const { viewOptions } = useBoardView();
    renderCard(viewOptions.groupBy);
    return null;
  });
  const Selection = memo(function Selection() {
    const { selectedStories } = useBoard();
    return <span>{selectedStories.length} selected</span>;
  });
  const consumers = (
    <>
      {Array.from({ length: 250 }, (_, index) => (
        <Card key={index} />
      ))}
      <Selection />
    </>
  );
  const { rerender } = render(
    <BoardProvider value={{ ...settings, selectedStories: [] }}>
      {consumers}
    </BoardProvider>,
  );
  expect(renderCard).toHaveBeenCalledTimes(250);

  rerender(
    <BoardProvider value={{ ...settings, selectedStories: ["story-a"] }}>
      {consumers}
    </BoardProvider>,
  );
  expect(screen.getByText("1 selected")).toBeInTheDocument();
  expect(renderCard).toHaveBeenCalledTimes(250);

  rerender(
    <BoardProvider
      value={{
        ...settings,
        selectedStories: ["story-a"],
        viewOptions: { ...settings.viewOptions, groupBy: "priority" },
      }}
    >
      {consumers}
    </BoardProvider>,
  );
  expect(renderCard).toHaveBeenCalledTimes(500);
  expect(renderCard).toHaveBeenLastCalledWith("priority");
});

it("publishes updated actions and creation defaults to view subscribers", () => {
  const observe = jest.fn();
  const Consumer = memo(function Consumer() {
    observe(useBoardView());
    return null;
  });
  const { rerender } = render(
    <BoardProvider value={{ ...settings, selectedStories: [] }}>
      <Consumer />
    </BoardProvider>,
  );
  const updateStory = jest.fn();
  const newStoryDefaults = { teamId: "team-b" };
  rerender(
    <BoardProvider
      value={{
        ...settings,
        selectedStories: [],
        updateStory,
        newStoryDefaults,
      }}
    >
      <Consumer />
    </BoardProvider>,
  );
  expect(observe).toHaveBeenLastCalledWith(
    expect.objectContaining({ updateStory, newStoryDefaults }),
  );
});

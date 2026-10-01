import type { ReactNode } from "react";
import { createContext, useContext, useState } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import type {
  BoardPropertyProviderProps,
  BoardPropertySlots,
  StoryPropertyBadgesProps,
  StoryPropertyDisplayPickerProps,
  WorkflowCountsProviderProps,
} from "./board-property-slots";
import {
  BoardPropertySlotsProvider,
  useBoardPropertySlots,
} from "./board-property-slots";

const story = {
  id: "deal-1",
  teamId: "marketing",
  subStories: [{ id: "follow-up-1", teamId: "marketing" }],
};
const SelectionContext = createContext<string[]>([]);

const FeatureProvider = ({
  children,
  selectedIds = [],
  stories,
}: BoardPropertyProviderProps) => (
  <SelectionContext.Provider value={selectedIds}>
    <output>{stories[0].subStories?.[0].id}</output>
    {children}
  </SelectionContext.Provider>
);

const FeatureBadges = ({ storyId, teamId }: StoryPropertyBadgesProps) => {
  const selectedIds = useContext(SelectionContext);
  return <output>{`${teamId}/${storyId}: ${selectedIds.join(", ")}`}</output>;
};

const FeaturePicker = ({ onChange }: StoryPropertyDisplayPickerProps) => {
  const [changes, setChanges] = useState(0);
  return (
    <button
      onClick={() => {
        setChanges((count) => count + 1);
        onChange(["amount", "customer"]);
      }}
      type="button"
    >
      Show properties ({changes})
    </button>
  );
};

const FeatureWorkflowCounts = ({
  children,
  enabled,
  teamId,
}: WorkflowCountsProviderProps) => (
  <>{children(new Map([["in-progress", enabled && teamId ? 42 : 0]]))}</>
);

const slots: BoardPropertySlots = {
  Provider: FeatureProvider,
  Badges: FeatureBadges,
  DisplayPicker: FeaturePicker,
  WorkflowCounts: FeatureWorkflowCounts,
};

const TaskView = ({ onChange }: { onChange: (ids: string[]) => void }) => {
  const { Provider, Badges, DisplayPicker, WorkflowCounts } =
    useBoardPropertySlots();
  return (
    <WorkflowCounts enabled teamId={story.teamId}>
      {(counts) => (
        <Provider selectedIds={["amount"]} stories={[story]}>
          <p>Follow up with customer</p>
          <output>{`Workflow count: ${counts.get("in-progress") ?? "unavailable"}`}</output>
          <Badges asList storyId={story.id} teamId={story.teamId} />
          <DisplayPicker onChange={onChange} selectedIds={["amount"]} />
        </Provider>
      )}
    </WorkflowCounts>
  );
};

const ComposedView = ({
  children,
  value,
}: {
  children: ReactNode;
  value: BoardPropertySlots;
}) => (
  <BoardPropertySlotsProvider slots={value}>
    {children}
  </BoardPropertySlotsProvider>
);

describe("board property composition", () => {
  it("preserves native tasks when optional feature slots are absent", () => {
    render(<TaskView onChange={jest.fn()} />);

    expect(screen.getByText("Follow up with customer")).toBeInTheDocument();
    expect(screen.getByText("Workflow count: unavailable")).toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it("composes the feature provider, nested task data, badges and picker", () => {
    const onChange = jest.fn();
    render(
      <ComposedView value={slots}>
        <TaskView onChange={onChange} />
      </ComposedView>,
    );

    expect(screen.getByText("follow-up-1")).toBeInTheDocument();
    expect(screen.getByText("Workflow count: 42")).toBeInTheDocument();
    expect(screen.getByText("marketing/deal-1: amount")).toBeInTheDocument();
    fireEvent.click(
      screen.getByRole("button", { name: "Show properties (0)" }),
    );
    expect(onChange).toHaveBeenCalledWith(["amount", "customer"]);
  });

  it("keeps feature interaction state when the shell refreshes its slot object", () => {
    const onChange = jest.fn();
    const { rerender } = render(
      <ComposedView value={slots}>
        <TaskView onChange={onChange} />
      </ComposedView>,
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Show properties (0)" }),
    );

    rerender(
      <ComposedView value={{ ...slots }}>
        <TaskView onChange={onChange} />
      </ComposedView>,
    );

    expect(
      screen.getByRole("button", { name: "Show properties (1)" }),
    ).toBeInTheDocument();
  });
});

/* global describe, expect, it, jest -- Provided by the projects test runner. */
import type { ComponentProps, PropsWithChildren, ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { CoreOptions } from "./core-options";

// Keep the actual priority menu and its selection behavior; only its visual
// primitives and unrelated properties are replaced in this focused test.
jest.mock("ui", () => {
  const Container = ({ children }: PropsWithChildren) => <div>{children}</div>;
  return {
    Avatar: () => null,
    Box: Container,
    Button: ({ children }: PropsWithChildren) => (
      <button type="button">{children}</button>
    ),
    Command: Object.assign(Container, {
      Empty: () => null,
      Group: Container,
      Input: () => null,
      Item: ({
        active,
        children,
        onSelect,
      }: PropsWithChildren<{
        active: boolean;
        onSelect: () => void;
      }>) => (
        <button aria-pressed={active} onClick={onSelect} type="button">
          {children}
        </button>
      ),
    }),
    Divider: () => null,
    Flex: Container,
    Popover: Object.assign(Container, {
      Content: Container,
      Trigger: Container,
    }),
    Text: Container,
  };
});
jest.mock("@/components/ui/property-option", () => ({
  PropertyOption: ({ label, value }: { label: string; value: ReactNode }) =>
    label === "Priority" ? value : null,
}));
jest.mock("@/components/ui/story/assignees-menu", () => ({
  AssigneesMenu: Object.assign(() => null, {
    Trigger: () => null,
    Items: () => null,
  }),
}));
jest.mock("@/components/ui/story/auto-scheduling-menu", () => ({
  AutoSchedulingMenu: Object.assign(() => null, {
    Trigger: () => null,
    Items: () => null,
  }),
}));
jest.mock("@/components/ui/story/estimate-menu", () => ({
  EstimateMenu: Object.assign(() => null, {
    Trigger: () => null,
    Items: () => null,
  }),
}));
jest.mock("@/components/ui/story/statuses-menu", () => ({
  StatusesMenu: Object.assign(() => null, {
    Trigger: () => null,
    Items: () => null,
  }),
}));
jest.mock("@/components/ui/story/time-needed-menu", () => ({
  TimeNeededMenu: Object.assign(() => null, {
    Trigger: () => null,
    Items: () => null,
  }),
}));
jest.mock("@/components/ui/story-status-icon", () => ({
  StoryStatusIcon: () => null,
}));
jest.mock("./collaborators-option", () => ({
  CollaboratorsOption: () => null,
}));
jest.mock("./use-option-hotkey", () => ({ useOptionHotkey: () => undefined }));

describe("CoreOptions priority", () => {
  it("reflects the current priority and allows clearing it without a redundant update", () => {
    const onUpdate = jest.fn();
    const props: ComponentProps<typeof CoreOptions> = {
      canUseBackgroundMaya: false,
      disabled: false,
      isCompact: true,
      isNotifications: false,
      members: [],
      onUpdate,
      storyId: "story",
      story: {
        assigneeId: null,
        autoSchedulingEnabled: false,
        autoSchedulingLocked: false,
        collaborators: [],
        estimateScheme: "points",
        estimateValue: null,
        estimatedDurationMinutes: null,
        minimumFocusBlockMinutes: null,
        priority: "High",
        statusId: "status",
        teamId: "team",
      },
    };
    const { rerender } = render(<CoreOptions {...props} />);

    expect(screen.getByRole("button", { name: "High 3" })).toHaveAttribute(
      "aria-pressed",
      "true",
    );
    fireEvent.click(screen.getByRole("button", { name: "High 3" }));
    expect(onUpdate).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "No Priority 0" }));
    expect(onUpdate).toHaveBeenCalledTimes(1);
    expect(onUpdate).toHaveBeenCalledWith({
      priority: "No Priority",
    });

    rerender(
      <CoreOptions
        {...props}
        story={{ ...props.story, priority: "No Priority" }}
      />,
    );
    expect(
      screen.getByRole("button", { name: "No Priority 0" }),
    ).toHaveAttribute("aria-pressed", "true");
    fireEvent.click(screen.getByRole("button", { name: "No Priority 0" }));
    expect(onUpdate).toHaveBeenCalledTimes(1);
  });
});

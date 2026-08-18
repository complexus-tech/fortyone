/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { ElementType, ReactNode } from "react";
import type { CalendarScheduleBlock } from "@/lib/queries/calendar/types";
import { CalendarScheduleBlockDetailsDialog } from "./calendar-schedule-block-details-dialog";

const useStoryById = jest.fn();

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/workspace${path}`,
  }),
}));

jest.mock("@/modules/story/hooks/story", () => ({
  useStoryById: (storyId: string) => useStoryById(storyId),
}));

jest.mock("@/modules/story/components/options", () => ({
  Options: () => <div>Story properties</div>,
}));

jest.mock("@/modules/story/utils/story-url", () => ({
  getStoryPath: ({ id }: { id: string }) => `/stories/${id}`,
}));

jest.mock("icons", () => ({
  ClockIcon: () => <span aria-hidden>Clock icon</span>,
  EditIcon: () => <span aria-hidden>Edit icon</span>,
  ExternalLinkIcon: () => <span aria-hidden>External link icon</span>,
  StoryIcon: () => <span aria-hidden>Story icon</span>,
}));

jest.mock("ui", () => {
  const Container = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );
  const Dialog = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );
  Dialog.Content = Container;
  Dialog.Header = Container;
  Dialog.Description = Container;
  Dialog.Body = Container;
  Dialog.Title = Container;

  return {
    Box: Container,
    Button: ({
      children,
      href,
      onClick,
      target,
    }: {
      children: ReactNode;
      href?: string;
      onClick?: () => void;
      target?: string;
    }) =>
      href ? (
        <a href={href} target={target}>
          {children}
        </a>
      ) : (
        <button onClick={onClick} type="button">
          {children}
        </button>
      ),
    Dialog,
    Flex: Container,
    Skeleton: Container,
    Text: ({
      as: Component = "span",
      children,
    }: {
      as?: ElementType;
      children: ReactNode;
    }) => <Component>{children}</Component>,
    Tooltip: Container,
  };
});

const block: CalendarScheduleBlock = {
  blockType: "work",
  createdAt: "2026-08-18T08:00:00Z",
  endAt: "2026-08-18T10:00:00Z",
  hasConflict: false,
  id: "block-1",
  isLocked: false,
  source: "user",
  startAt: "2026-08-18T09:00:00Z",
  storyCode: "PRID-62",
  storyId: "story-62",
  title: "Schedule calendar story details",
  updatedAt: "2026-08-18T08:00:00Z",
};

describe("CalendarScheduleBlockDetailsDialog", () => {
  beforeEach(() => {
    useStoryById.mockReset();
    useStoryById.mockReturnValue({
      data: {
        description: "Show the real story context from the calendar.",
        id: "story-62",
        sequenceId: 62,
        teamCode: "PRID",
        title: "Schedule calendar story details",
      },
    });
  });

  it("shows the scheduled story and its editable property controls", () => {
    render(
      <CalendarScheduleBlockDetailsDialog
        block={block}
        onEdit={jest.fn()}
        onOpenChange={jest.fn()}
      />,
    );

    expect(useStoryById).toHaveBeenCalledWith("story-62");
    expect(screen.getByText("PRID-62")).toBeInTheDocument();
    expect(
      screen.getByText("Show the real story context from the calendar."),
    ).toBeInTheDocument();
    expect(screen.getByText("Story properties")).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "Open story in a new tab" }),
    ).toHaveAttribute("target", "_blank");
  });

  it("keeps the schedule editor available for user-managed blocks", () => {
    const onEdit = jest.fn();
    render(
      <CalendarScheduleBlockDetailsDialog
        block={block}
        onEdit={onEdit}
        onOpenChange={jest.fn()}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Edit schedule" }));

    expect(onEdit).toHaveBeenCalledWith(block);
  });
});

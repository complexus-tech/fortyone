/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { ElementType, ReactNode } from "react";
import type { CalendarScheduleBlock } from "@/lib/queries/calendar/types";
import { CalendarScheduleBlockDetailsDialog } from "./calendar-schedule-block-details-dialog";

const mockUseStoryById = jest.fn();
const mockCopyText = jest.fn();
const mockToastError = jest.fn();

jest.mock("@/hooks", () => ({
  useCopyToClipboard: () => [null, mockCopyText],
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/workspace${path}`,
  }),
}));

jest.mock("@/hooks/owner", () => ({
  useIsAdminOrOwner: () => ({ isAdminOrOwner: true }),
}));

jest.mock("@/modules/story/hooks/story", () => ({
  useStoryById: (storyId: string) => mockUseStoryById(storyId),
}));

jest.mock("@/modules/story/components/story-actions-menu", () => ({
  StoryActionsMenu: () => <button type="button">More story actions</button>,
}));

jest.mock("./calendar-block", () => ({
  isCalendarScheduleBlockEditable: () => true,
}));

jest.mock("sonner", () => ({
  toast: {
    error: (...args: unknown[]) => mockToastError(...args),
  },
}));

jest.mock(
  "next/dynamic",
  () => () =>
    function MockMainDetails({ storyId }: { storyId: string }) {
      return <div>Editable story details for {storyId}</div>;
    },
);

jest.mock("@/modules/story/utils/story-url", () => ({
  getStoryPath: ({ id }: { id: string }) => `/stories/${id}`,
  getStoryReference: ({
    sequenceId,
    teamCode,
  }: {
    sequenceId: number;
    teamCode: string;
  }) => `${teamCode}-${sequenceId}`,
}));

jest.mock("icons", () => ({
  ClockIcon: () => <span aria-hidden>Clock icon</span>,
  CopyIcon: () => <span aria-hidden>Copy icon</span>,
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
    mockUseStoryById.mockReset();
    mockCopyText.mockReset();
    mockToastError.mockReset();
    mockCopyText.mockResolvedValue(true);
    mockUseStoryById.mockReturnValue({
      data: {
        description: "Show the real story context from the calendar.",
        id: "story-62",
        sequenceId: 62,
        teamCode: "PRID",
        title: "Schedule calendar story details",
      },
    });
  });

  it("uses the editable story surface and exposes story actions in the header", () => {
    render(
      <CalendarScheduleBlockDetailsDialog
        block={block}
        onEdit={jest.fn()}
        onOpenChange={jest.fn()}
      />,
    );

    expect(mockUseStoryById).toHaveBeenCalledWith("story-62");
    expect(screen.getByText("PRID-62")).toBeInTheDocument();
    expect(
      screen.getByText("Editable story details for story-62"),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "More story actions" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "Open story in a new tab" }),
    ).toHaveAttribute("target", "_blank");
  });

  it("copies the canonical story link from the header", () => {
    render(
      <CalendarScheduleBlockDetailsDialog
        block={block}
        onEdit={jest.fn()}
        onOpenChange={jest.fn()}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Copy story link" }));

    expect(mockCopyText).toHaveBeenCalledWith(
      `${window.location.origin}/workspace/stories/story-62`,
    );
    expect(mockToastError).not.toHaveBeenCalled();
  });

  it("shows an error toast when the story link cannot be copied", async () => {
    mockCopyText.mockResolvedValue(false);
    render(
      <CalendarScheduleBlockDetailsDialog
        block={block}
        onEdit={jest.fn()}
        onOpenChange={jest.fn()}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Copy story link" }));

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith("Could not copy story link");
    });
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

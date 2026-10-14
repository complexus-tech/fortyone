/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import type { StoryActivity } from "@/modules/stories/types";
import { Activity } from "./activity";

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "viewer" } } }),
}));
jest.mock("@/hooks", () => ({
  useTerminology: () => ({ getTermDisplay: () => "story" }),
  useWorkspacePath: () => ({ withWorkspace: (path: string) => `/acme${path}` }),
}));
jest.mock("@/lib/hooks/labels", () => ({ useLabels: () => ({ data: [] }) }));
jest.mock("@/lib/hooks/members", () => ({
  useMembers: () => ({ data: [] }),
  useMayaAssignee: () => ({ data: undefined }),
}));
jest.mock("@/lib/hooks/profile", () => ({ useProfile: () => ({ data: {} }) }));
jest.mock("@/lib/hooks/statuses", () => ({
  useStatuses: () => ({ data: [] }),
}));
jest.mock("@/modules/objectives/hooks/use-objective", () => ({
  useObjective: () => ({ data: undefined }),
}));
jest.mock("@/modules/sprints/hooks/sprint-details", () => ({
  useSprint: () => ({ data: undefined }),
}));
jest.mock("@/modules/teams/hooks/use-team-settings", () => ({
  useTeamSettings: () => ({ data: undefined }),
}));
jest.mock("ui", () => {
  const Container = ({ children }: { children?: ReactNode }) => (
    <div>{children}</div>
  );
  return {
    Avatar: () => null,
    Box: Container,
    Flex: Container,
    Text: ({
      children,
      className,
      title,
    }: {
      children?: ReactNode;
      className?: string;
      title?: string;
    }) => (
      <span className={className} title={title}>
        {children}
      </span>
    ),
    TimeAgo: ({ timestamp }: { timestamp: string }) => (
      <time dateTime={timestamp}>{timestamp}</time>
    ),
    Tooltip: ({ children }: { children?: ReactNode }) => <>{children}</>,
  };
});

const activity: StoryActivity = {
  id: "activity-1",
  storyId: "story-1",
  userId: "editor",
  user: {
    id: "editor",
    username: "taylor",
    fullName: "Taylor Reed",
    avatarUrl: null,
    isActive: true,
    isSystem: false,
  },
  type: "update",
  field: "checklist_item",
  currentValue: "Publish launch notes",
  oldValue: {
    id: "deleted-item",
    title: "Publish launch notes",
    completed: false,
    position: 0,
  },
  newValue: null,
  reason: "checklist_item_removed",
  createdAt: "2026-10-14T10:30:00Z",
};

describe("checklist activity", () => {
  it("shows the recorded actor, timestamp and title even after an item is removed", () => {
    const { container } = render(<Activity {...activity} />);

    expect(screen.getByText("taylor")).toBeInTheDocument();
    expect(screen.getByText("removed checklist item")).toBeInTheDocument();
    expect(screen.getByText("Publish launch notes")).toBeInTheDocument();
    expect(container.querySelector("time")).toHaveAttribute(
      "datetime",
      activity.createdAt,
    );
    expect(container).not.toHaveTextContent(
      /checklist_item_removed|\[object Object\]/,
    );
  });

  it("retains both historical titles on hover and keeps long titles truncatable", () => {
    render(
      <Activity
        {...activity}
        oldValue={{ id: "deleted-item", title: "Review launch notes" }}
        reason="checklist_item_updated"
      />,
    );

    expect(screen.getByText("renamed checklist item")).toBeInTheDocument();
    expect(screen.getByTitle("Review launch notes")).toHaveClass("truncate");
    expect(screen.getByTitle("Publish launch notes")).toHaveClass("truncate");
  });
});

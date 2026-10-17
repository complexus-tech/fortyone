/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { render, screen } from "@testing-library/react";
import type { AppNotification } from "./types";
import { NotificationCard } from "./card";

jest.mock("next/navigation", () => ({
  usePathname: () => "/workspace/notifications",
}));
jest.mock("@/hooks", () => ({
  useTerminology: () => ({ getTermDisplay: () => "task" }),
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/workspace${path}`,
  }),
}));
jest.mock("./hooks/read-mutation", () => ({
  useReadNotificationMutation: () => ({ mutate: jest.fn() }),
}));
jest.mock("./hooks/mark-unread-mutation", () => ({
  useMarkUnreadMutation: () => ({ mutate: jest.fn() }),
}));
jest.mock("./hooks/delete-mutation", () => ({
  useDeleteMutation: () => ({ mutate: jest.fn() }),
}));

const notification: AppNotification = {
  id: "notification-1",
  recipientId: "recipient",
  workspaceId: "workspace",
  type: "story_comment",
  entityType: "story",
  entityId: "story-1",
  actorId: "actor",
  actor: {
    id: "actor",
    fullName: "Ava",
    username: "ava",
    avatarUrl: "",
    isActive: true,
    isSystem: false,
  },
  title: "History of the story",
  message: {
    template: "<p>{actor} commented on the story: {content}</p>",
    variables: {
      content: {
        value: "<p>The <strong>story</strong> is ready &amp; reviewed</p>",
        type: "text",
      },
      commentId: {
        value: "55555555-5555-4555-8555-555555555555",
        type: "comment",
      },
    },
  },
  createdAt: new Date().toISOString(),
  readAt: null,
};

describe("NotificationCard", () => {
  it("renders the correct title, actor, readable content, and comment destination together", () => {
    const { container } = render(<NotificationCard {...notification} />);
    const link = screen.getByRole("link");

    expect(screen.getByText(notification.title)).toBeInTheDocument();
    expect(link).toHaveTextContent(
      "Ava commented on the task: The story is ready & reviewed",
    );
    expect(link).not.toHaveTextContent("Ava Ava");
    expect(link).not.toHaveTextContent("<p>");
    expect(container.querySelector("strong")).not.toBeInTheDocument();
    const url = new URL(link.getAttribute("href")!, "https://fortyone.app");
    expect(url.pathname).toBe("/workspace/notifications/notification-1");
    expect(url.searchParams.get("entityId")).toBe("story-1");
    expect(url.searchParams.get("comment")).toBe(
      "55555555-5555-4555-8555-555555555555",
    );
    expect(url.searchParams.get("observedCreatedAt")).toBe(
      notification.createdAt,
    );
  });

  it("renders complete system copy with null variables without adding an actor", () => {
    render(
      <NotificationCard
        {...notification}
        entityType="strategy"
        message={{
          template: "Strategy guidance is ready to review.",
          variables: null,
        }}
        type="strategy_update"
      />,
    );

    expect(screen.getByRole("link")).toHaveTextContent(
      "Strategy guidance is ready to review.",
    );
    expect(screen.getByRole("link")).not.toHaveTextContent("Ava Strategy");
  });
});

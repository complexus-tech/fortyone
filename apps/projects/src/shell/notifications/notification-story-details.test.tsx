/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { render, screen, waitFor } from "@testing-library/react";
import { NotificationStoryDetails } from "./notification-story-details";

const readNotification = jest.fn();

jest.mock("@/modules/notifications/public/client", () => ({
  useReadNotificationMutation: () => ({ mutate: readNotification }),
}));

jest.mock("@/modules/story/public/client", () => ({
  StoryPage: ({
    isNotifications,
    storyId,
  }: {
    isNotifications?: boolean;
    storyId: string;
  }) => <output data-notifications={isNotifications} data-story-id={storyId} />,
}));

describe("NotificationStoryDetails", () => {
  it("marks the notification read once before rendering the story capability", async () => {
    render(
      <NotificationStoryDetails
        entityId="story-1"
        notificationId="notification-1"
      />,
    );

    await waitFor(() => {
      expect(readNotification).toHaveBeenCalledWith("notification-1");
    });

    expect(screen.getByRole("status")).toHaveAttribute(
      "data-story-id",
      "story-1",
    );
    expect(screen.getByRole("status")).toHaveAttribute(
      "data-notifications",
      "true",
    );
  });
});

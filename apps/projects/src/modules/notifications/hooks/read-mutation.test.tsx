/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import { notificationKeys } from "@/constants/keys";
import type { AppNotification } from "../types";
import { readNotification } from "../actions/read";
import { useReadNotificationMutation } from "./read-mutation";

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "workspace" }),
}));
jest.mock("../actions/read", () => ({ readNotification: jest.fn() }));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));

const observedCreatedAt = "2026-10-03T20:00:00.123456Z";
const notification: AppNotification = {
  id: "notification-1",
  recipientId: "recipient",
  workspaceId: "workspace",
  type: "story_update",
  entityType: "story",
  entityId: "story",
  actorId: "actor",
  title: "Updated",
  message: { template: "Updated", variables: {} },
  createdAt: observedCreatedAt,
  readAt: null,
};
const listKey = [...notificationKeys.all("workspace"), "", "infinite"];
const unreadKey = notificationKeys.unread("workspace");

const setup = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  client.setQueryData(listKey, {
    pages: [
      {
        notifications: [notification],
        pagination: { page: 1, pageSize: 25, hasMore: false, nextPage: 2 },
      },
    ],
    pageParams: [1],
  });
  client.setQueryData(unreadKey, 1);
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  return {
    client,
    ...renderHook(() => useReadNotificationMutation(), { wrapper }),
  };
};

describe("read mutation conflict handling", () => {
  beforeEach(() => jest.clearAllMocks());

  it("carries the displayed version and rolls back a 409 API error response", async () => {
    let complete: (
      value: Awaited<ReturnType<typeof readNotification>>,
    ) => void = () => undefined;
    jest.mocked(readNotification).mockImplementation(
      () =>
        new Promise((resolve) => {
          complete = resolve;
        }),
    );
    const { client, result } = setup();
    const before = client.getQueryData(listKey);
    act(() => {
      result.current.mutate({ id: notification.id, observedCreatedAt });
    });
    await waitFor(() => {
      expect(readNotification).toHaveBeenCalledWith(
        notification.id,
        "workspace",
        observedCreatedAt,
      );
    });
    expect(client.getQueryData(unreadKey)).toBe(0);
    act(() => {
      complete({
        data: null,
        error: { message: "notification was refreshed" },
      });
    });
    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
    expect(client.getQueryData(listKey)).toEqual(before);
    expect(client.getQueryData(unreadKey)).toBe(1);
    expect(toast.error).toHaveBeenCalled();
    client.clear();
  });

  it("preserves a realtime refresh received before the stale read fails", async () => {
    let complete: (
      value: Awaited<ReturnType<typeof readNotification>>,
    ) => void = () => undefined;
    jest.mocked(readNotification).mockImplementation(
      () =>
        new Promise((resolve) => {
          complete = resolve;
        }),
    );
    const { client, result } = setup();
    act(() => {
      result.current.mutate({ id: notification.id, observedCreatedAt });
    });
    await waitFor(() => {
      expect(readNotification).toHaveBeenCalled();
    });
    const refreshed = {
      pages: [
        {
          notifications: [
            {
              ...notification,
              title: "New priority",
              createdAt: "2026-10-03T20:01:00Z",
            },
          ],
        },
      ],
      pageParams: [1],
    };
    client.setQueryData(listKey, refreshed);
    client.setQueryData(unreadKey, 1);
    act(() => {
      complete({
        data: null,
        error: { message: "notification was refreshed" },
      });
    });
    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
    expect(client.getQueryData(listKey)).toEqual(refreshed);
    expect(client.getQueryData(unreadKey)).toBe(1);
    client.clear();
  });
});

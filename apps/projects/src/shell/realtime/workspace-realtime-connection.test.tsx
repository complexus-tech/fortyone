import { render } from "@testing-library/react";
import { useQueryClient } from "@tanstack/react-query";
import { usePostHog } from "posthog-js/react";
import { ServerSentEvents } from "@/app/server-sent-events";
import {
  calendarKeys,
  chatCapabilityKeys,
  notificationKeys,
  workspaceKeys,
} from "@/constants/keys";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { useSession } from "@/lib/auth/client";
import { deliverDesktopNotification } from "@/shell/desktop/notification";
import { reconcileChatQueries } from "@/modules/chat/realtime";
import { createChatAlertDelivery } from "@/modules/chat/alerts/delivery";
import { applyChatTypingEvent, clearChatTyping } from "@/modules/chat/typing";
import { storyKeys } from "@/modules/stories/constants";
import { WorkspaceRealtimeConnection } from "./workspace-realtime-connection";

jest.mock("@tanstack/react-query", () => ({
  useQueryClient: jest.fn(),
}));

jest.mock("@/modules/chat/realtime", () => ({
  reconcileChatQueries: jest.fn(),
}));
jest.mock("@/modules/chat/alerts/delivery", () => ({
  createChatAlertDelivery: jest.fn(() => ({
    receive: jest.fn(),
    dispose: jest.fn(),
  })),
}));
jest.mock("@/modules/chat/typing", () => ({
  applyChatTypingEvent: jest.fn(),
  clearChatTyping: jest.fn(),
}));

jest.mock("posthog-js/react", () => ({
  usePostHog: jest.fn(),
}));

jest.mock("@/lib/api-url", () => ({
  getApiUrl: jest.fn(() => "https://api.fortyone.test"),
}));

jest.mock("@/lib/hooks/workspaces", () => ({
  useCurrentWorkspace: jest.fn(),
}));

jest.mock("@/lib/auth/client", () => ({ useSession: jest.fn() }));
jest.mock("@/shell/desktop/notification", () => ({
  deliverDesktopNotification: jest.fn(),
}));

class MockEventSource {
  static readonly instances: MockEventSource[] = [];

  close = jest.fn();
  onerror: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onopen: ((event: Event) => void) | null = null;

  constructor(
    readonly url: string,
    readonly options?: EventSourceInit,
  ) {
    MockEventSource.instances.push(this);
  }
}

const eventSourceDescriptor = Object.getOwnPropertyDescriptor(
  globalThis,
  "EventSource",
);
const captureException = jest.fn();
const queryClient = {
  invalidateQueries: jest.fn(),
  setQueriesData: jest.fn(),
  setQueryData: jest.fn(),
};

let connectionWorkspaceSlug: string | undefined;

const mockedUseCurrentWorkspace = jest.mocked(useCurrentWorkspace);
const mockedUsePostHog = jest.mocked(usePostHog);
const mockedUseQueryClient = jest.mocked(useQueryClient);

const connection = () => {
  const source = MockEventSource.instances.at(-1);

  if (!source) {
    throw new Error("Expected an EventSource connection");
  }

  return source;
};

beforeAll(() => {
  Object.defineProperty(globalThis, "EventSource", {
    configurable: true,
    value: MockEventSource,
  });
});

beforeEach(() => {
  jest.clearAllMocks();
  MockEventSource.instances.splice(0);
  connectionWorkspaceSlug = "engineering";

  mockedUseCurrentWorkspace.mockImplementation(
    () =>
      ({
        workspace: connectionWorkspaceSlug
          ? { id: "workspace-id", slug: connectionWorkspaceSlug }
          : undefined,
      }) as ReturnType<typeof useCurrentWorkspace>,
  );
  mockedUsePostHog.mockReturnValue({
    captureException,
  } as unknown as ReturnType<typeof usePostHog>);
  mockedUseQueryClient.mockReturnValue(
    queryClient as unknown as ReturnType<typeof useQueryClient>,
  );
  jest
    .mocked(useSession)
    .mockReturnValue({ data: null, status: "unauthenticated" });
  jest.mocked(deliverDesktopNotification).mockResolvedValue(undefined);
});

afterAll(() => {
  if (eventSourceDescriptor) {
    Object.defineProperty(globalThis, "EventSource", eventSourceDescriptor);
    return;
  }

  Reflect.deleteProperty(globalThis, "EventSource");
});

describe("WorkspaceRealtimeConnection", () => {
  it("routes chat alerts separately from the work inbox and disposes their pending delivery", () => {
    jest.mocked(useSession).mockReturnValue({
      data: { user: { id: "10000000-0000-4000-8000-000000000004" } },
      status: "authenticated",
    } as ReturnType<typeof useSession>);
    const view = render(<WorkspaceRealtimeConnection />);
    const delivery = jest
      .mocked(createChatAlertDelivery)
      .mock.results.at(-1)?.value;
    const payload = {
      type: "chat.notification",
      notificationId: "10000000-0000-4000-8000-000000000001",
      eventId: "10000000-0000-4000-8000-000000000002",
      workspaceId: "10000000-0000-4000-8000-000000000003",
      userId: "10000000-0000-4000-8000-000000000004",
      actorId: "10000000-0000-4000-8000-000000000005",
      conversationId: "10000000-0000-4000-8000-000000000006",
      messageId: "10000000-0000-4000-8000-000000000007",
      threadRootId: null,
      createdAt: "2026-10-17T12:00:00Z",
      occurredAt: "2026-10-17T12:00:00Z",
      title: "New message",
      body: "You have a new message.",
    };
    connection().onmessage?.({ data: JSON.stringify(payload) } as MessageEvent);
    expect(delivery.receive).toHaveBeenCalledWith(
      expect.objectContaining({
        kind: "chat-notification",
        conversationId: payload.conversationId,
      }),
    );
    expect(deliverDesktopNotification).not.toHaveBeenCalled();
    expect(queryClient.invalidateQueries).not.toHaveBeenCalled();
    view.unmount();
    expect(delivery.dispose).toHaveBeenCalled();
  });

  it("does not put legacy generic chat events back into the work inbox", () => {
    render(<WorkspaceRealtimeConnection />);
    connection().onmessage?.({
      data: JSON.stringify({
        entity_id: "message-id",
        entity_type: "chat_message",
      }),
    } as MessageEvent);
    expect(deliverDesktopNotification).not.toHaveBeenCalled();
    expect(queryClient.invalidateQueries).not.toHaveBeenCalled();
  });

  it("forwards live notifications to native delivery without replaying on reconnect", () => {
    jest.mocked(useSession).mockReturnValue({
      data: { user: { id: "account-id" } },
      status: "authenticated",
    } as ReturnType<typeof useSession>);
    render(<WorkspaceRealtimeConnection />);
    const data = JSON.stringify({ entity_id: "story-1", entity_type: "story" });
    connection().onmessage?.({ data } as MessageEvent);
    expect(deliverDesktopNotification).toHaveBeenCalledWith(
      data,
      {
        accountId: "account-id",
        workspaceId: "workspace-id",
        workspaceSlug: "engineering",
      },
      expect.any(Function),
    );
    connection().onerror?.(new Event("error"));
    connection().onopen?.(new Event("open"));
    expect(deliverDesktopNotification).toHaveBeenCalledTimes(1);
  });

  it("closes the previous account subscription and rejects its late events", () => {
    jest.mocked(useSession).mockReturnValue({
      data: { user: { id: "account-id" } },
      status: "authenticated",
    } as ReturnType<typeof useSession>);
    const view = render(<WorkspaceRealtimeConnection />);
    const previous = connection();
    jest.mocked(useSession).mockReturnValue({
      data: { user: { id: "other-account" } },
      status: "authenticated",
    } as ReturnType<typeof useSession>);
    view.rerender(<WorkspaceRealtimeConnection />);
    previous.onmessage?.({
      data: JSON.stringify({ entity_id: "story-1", entity_type: "story" }),
    } as MessageEvent);
    expect(previous.close).toHaveBeenCalledTimes(1);
    expect(deliverDesktopNotification).not.toHaveBeenCalled();
  });
  it("keeps the existing route entrypoint as a compatibility export", () => {
    expect(ServerSentEvents).toBe(WorkspaceRealtimeConnection);
  });

  it("opens one credentialed workspace connection and closes it on unmount", () => {
    const view = render(<WorkspaceRealtimeConnection />);
    const source = connection();

    expect(source.url).toBe(
      "https://api.fortyone.test/workspaces/engineering/notifications/subscribe",
    );
    expect(source.options).toEqual({ withCredentials: true });

    view.rerender(<WorkspaceRealtimeConnection />);

    expect(MockEventSource.instances).toHaveLength(1);

    view.unmount();

    expect(source.close).toHaveBeenCalledTimes(1);
  });

  it("does not connect until an active workspace is available", () => {
    connectionWorkspaceSlug = undefined;

    render(<WorkspaceRealtimeConnection />);

    expect(MockEventSource.instances).toHaveLength(0);
  });

  it("reconnects when the active workspace changes", () => {
    const view = render(<WorkspaceRealtimeConnection />);
    const firstConnection = connection();

    connectionWorkspaceSlug = "product";
    view.rerender(<WorkspaceRealtimeConnection />);

    expect(firstConnection.close).toHaveBeenCalledTimes(1);
    expect(MockEventSource.instances).toHaveLength(2);
    expect(connection().url).toBe(
      "https://api.fortyone.test/workspaces/product/notifications/subscribe",
    );

    firstConnection.onmessage?.({
      data: JSON.stringify({ entityId: "story-1", entityType: "story" }),
    } as MessageEvent);
    expect(queryClient.invalidateQueries).not.toHaveBeenCalled();
    expect(queryClient.invalidateQueries).not.toHaveBeenCalledWith({
      queryKey: notificationKeys.all("product"),
    });
  });

  it("preserves notification and calendar cache invalidation behavior", () => {
    render(<WorkspaceRealtimeConnection />);

    connection().onmessage?.({
      data: JSON.stringify({ entityId: "story-1", entityType: "story" }),
    } as MessageEvent);
    connection().onmessage?.({
      data: JSON.stringify({ type: "calendar.updated" }),
    } as MessageEvent);

    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: notificationKeys.all("engineering"),
    });
    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: storyKeys.detail("engineering", "story-1"),
    });
    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: calendarKeys.all("engineering"),
    });
  });

  it("updates known detail data and invalidates only this workspace's collection caches", () => {
    render(<WorkspaceRealtimeConnection />);

    connection().onmessage?.({
      data: JSON.stringify({
        changes: { statusId: "done" },
        storyId: "story-1",
        type: "story.workspace_update",
      }),
    } as MessageEvent);
    expect(queryClient.setQueriesData).not.toHaveBeenCalled();
    const detailUpdater = queryClient.setQueryData.mock.calls[0]?.[1] as (
      oldData: { id: string; statusId?: string } | undefined,
    ) => { id: string; statusId?: string } | undefined;

    expect(detailUpdater({ id: "story-1", statusId: "in-progress" })).toEqual({
      id: "story-1",
      statusId: "done",
    });
    expect(queryClient.setQueryData).toHaveBeenCalledWith(
      storyKeys.detail("engineering", "story-1"),
      expect.any(Function),
    );

    const collectionFilter = queryClient.invalidateQueries.mock.calls
      .map(([filter]) => filter)
      .find(
        (filter) =>
          typeof filter?.predicate === "function" &&
          filter.predicate({
            queryKey: ["stories", "engineering", "grouped", {}],
          }),
      );
    const matchesCollection = collectionFilter?.predicate as
      | ((query: { queryKey: readonly unknown[] }) => boolean)
      | undefined;
    const nativePreview = {
      queryKey: [
        "native-link-preview",
        "user-1",
        "engineering",
        "story",
        "ENG-1",
      ],
      state: {
        data: { resource: { kind: "story", value: { id: "story-1" } } },
      },
    };
    const previewFilter = queryClient.invalidateQueries.mock.calls
      .map(([filter]) => filter)
      .find(
        (filter) =>
          typeof filter?.predicate === "function" &&
          filter.predicate(nativePreview),
      );
    expect(previewFilter).toBeDefined();
    expect(
      previewFilter?.predicate({
        ...nativePreview,
        queryKey: ["native-link-preview", "user-1", "product"],
      }),
    ).toBe(false);
    expect(
      previewFilter?.predicate({
        ...nativePreview,
        state: {
          data: { resource: { kind: "story", value: { id: "story-2" } } },
        },
      }),
    ).toBe(false);
    expect(matchesCollection).toBeDefined();
    expect(
      matchesCollection?.({
        queryKey: ["stories", "engineering", "grouped", {}],
      }),
    ).toBe(true);
    expect(
      matchesCollection?.({ queryKey: ["stories", "product", "list"] }),
    ).toBe(false);
    expect(
      matchesCollection?.({
        queryKey: ["stories", "engineering", "detail", "story-1"],
      }),
    ).toBe(false);
    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: calendarKeys.all("engineering"),
    });

    connection().onmessage?.({ data: "not-json" } as MessageEvent);
    connection().onerror?.(new Event("error"));
    expect(captureException).toHaveBeenCalledTimes(2);
  });

  it("does not apply malformed story updates and reconciles authoritative caches", () => {
    render(<WorkspaceRealtimeConnection />);

    connection().onmessage?.({
      data: JSON.stringify({
        changes: { assigneeId: { unexpected: true } },
        storyId: "story-1",
        type: "story.workspace_update",
      }),
    } as MessageEvent);

    expect(queryClient.setQueriesData).not.toHaveBeenCalled();
    expect(queryClient.setQueryData).not.toHaveBeenCalled();
    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: notificationKeys.all("engineering"),
    });
    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: storyKeys.all("engineering"),
    });
    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: calendarKeys.all("engineering"),
    });
    expect(captureException).toHaveBeenCalledTimes(1);
  });

  it("refetches checklist, activities, and associations without changing story completion", () => {
    render(<WorkspaceRealtimeConnection />);

    connection().onmessage?.({
      data: JSON.stringify({
        changes: { checklistChanged: true },
        storyId: "story-1",
        type: "story.workspace_update",
      }),
    } as MessageEvent);

    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: storyKeys.detail("engineering", "story-1"),
    });
    expect(queryClient.invalidateQueries).not.toHaveBeenCalledWith({
      queryKey: calendarKeys.all("engineering"),
    });
    expect(queryClient.invalidateQueries).not.toHaveBeenCalledWith({
      queryKey: notificationKeys.all("engineering"),
    });
    expect(queryClient.setQueryData).not.toHaveBeenCalled();
    expect(captureException).not.toHaveBeenCalled();
  });

  it("does not persist the checklist invalidation marker in story detail data", () => {
    render(<WorkspaceRealtimeConnection />);

    connection().onmessage?.({
      data: JSON.stringify({
        changes: { checklistChanged: true, title: "Updated title" },
        storyId: "story-1",
        type: "story.workspace_update",
      }),
    } as MessageEvent);

    const detailUpdater = queryClient.setQueryData.mock
      .calls[0]?.[1] as (oldData: {
      id: string;
      title: string;
      completedAt: null;
    }) => { id: string; title: string; completedAt: null };

    expect(
      detailUpdater({ id: "story-1", title: "Original", completedAt: null }),
    ).toEqual({ id: "story-1", title: "Updated title", completedAt: null });
    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: storyKeys.detail("engineering", "story-1"),
    });
  });

  it("reconciles caches after EventSource reconnects", () => {
    render(<WorkspaceRealtimeConnection />);
    const source = connection();

    source.onerror?.(new Event("error"));
    expect(queryClient.invalidateQueries).not.toHaveBeenCalled();
    source.onopen?.(new Event("open"));

    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: workspaceKeys.settings("engineering"),
    });
    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: chatCapabilityKeys.workspace("engineering"),
    });

    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: notificationKeys.all("engineering"),
    });
    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: storyKeys.all("engineering"),
    });
    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: calendarKeys.all("engineering"),
    });
  });
});

it("accepts chat hints only for the active account and workspace", () => {
  const accountId = "11111111-1111-4111-8111-111111111111";
  const workspaceId = "22222222-2222-4222-8222-222222222222";
  const conversationId = "33333333-3333-4333-8333-333333333333";
  mockedUseCurrentWorkspace.mockReturnValue({
    workspace: { id: workspaceId, slug: "engineering" },
  } as ReturnType<typeof useCurrentWorkspace>);
  jest.mocked(useSession).mockReturnValue({
    data: { user: { id: accountId } },
    status: "authenticated",
  } as ReturnType<typeof useSession>);
  render(<WorkspaceRealtimeConnection />);
  const payload = {
    type: "chat.updated",
    eventId: "44444444-4444-4444-8444-444444444444",
    userId: accountId,
    workspaceId,
    conversationId,
  };
  connection().onmessage?.({
    data: JSON.stringify({ ...payload, userId: conversationId }),
  } as MessageEvent);
  connection().onmessage?.({
    data: JSON.stringify({ ...payload, workspaceId: conversationId }),
  } as MessageEvent);
  expect(reconcileChatQueries).not.toHaveBeenCalled();
  connection().onmessage?.({ data: JSON.stringify(payload) } as MessageEvent);
  expect(reconcileChatQueries).toHaveBeenCalledWith(
    queryClient,
    "engineering",
    conversationId,
  );
  expect(deliverDesktopNotification).not.toHaveBeenCalled();
});

it("routes typing only to the current account and clears transient activity on disconnect", () => {
  const userId = "11111111-1111-4111-8111-111111111111";
  const workspaceId = "22222222-2222-4222-8222-222222222222";
  const conversationId = "33333333-3333-4333-8333-333333333333";
  mockedUseCurrentWorkspace.mockReturnValue({
    workspace: { id: workspaceId, slug: "engineering" },
  } as ReturnType<typeof useCurrentWorkspace>);
  jest.mocked(useSession).mockReturnValue({
    data: { user: { id: userId } },
    status: "authenticated",
  } as ReturnType<typeof useSession>);
  const view = render(<WorkspaceRealtimeConnection />);
  const payload = {
    type: "chat.typing",
    userId,
    workspaceId,
    conversationId,
    actorId: conversationId,
    fullName: "Sarah Chen",
    threadRootId: null,
    isTyping: true,
    expiresAt: "2026-10-16T19:00:06Z",
  };
  connection().onmessage?.({
    data: JSON.stringify({ ...payload, workspaceId: conversationId }),
  } as MessageEvent);
  connection().onmessage?.({
    data: JSON.stringify({ ...payload, userId: conversationId }),
  } as MessageEvent);
  expect(applyChatTypingEvent).not.toHaveBeenCalled();
  connection().onmessage?.({ data: JSON.stringify(payload) } as MessageEvent);
  expect(applyChatTypingEvent).toHaveBeenCalledWith(
    expect.objectContaining({
      kind: "chat-typing",
      workspaceId,
      conversationId,
    }),
  );
  expect(queryClient.invalidateQueries).not.toHaveBeenCalled();
  expect(deliverDesktopNotification).not.toHaveBeenCalled();
  connection().onerror?.(new Event("error"));
  expect(clearChatTyping).toHaveBeenCalledWith(workspaceId);
  view.unmount();
  expect(clearChatTyping).toHaveBeenCalledTimes(2);
});

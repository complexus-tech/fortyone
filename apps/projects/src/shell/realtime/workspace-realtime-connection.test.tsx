import { render } from "@testing-library/react";
import { useQueryClient } from "@tanstack/react-query";
import { usePostHog } from "posthog-js/react";
import { ServerSentEvents } from "@/app/server-sent-events";
import { calendarKeys, notificationKeys } from "@/constants/keys";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { storyKeys } from "@/modules/stories/constants";
import { WorkspaceRealtimeConnection } from "./workspace-realtime-connection";

jest.mock("@tanstack/react-query", () => ({
  useQueryClient: jest.fn(),
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
          ? { slug: connectionWorkspaceSlug }
          : undefined,
      }) as ReturnType<typeof useCurrentWorkspace>,
  );
  mockedUsePostHog.mockReturnValue({
    captureException,
  } as unknown as ReturnType<typeof usePostHog>);
  mockedUseQueryClient.mockReturnValue(
    queryClient as unknown as ReturnType<typeof useQueryClient>,
  );
});

afterAll(() => {
  if (eventSourceDescriptor) {
    Object.defineProperty(globalThis, "EventSource", eventSourceDescriptor);
    return;
  }

  Reflect.deleteProperty(globalThis, "EventSource");
});

describe("WorkspaceRealtimeConnection", () => {
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
    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: notificationKeys.all("engineering"),
    });
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
      .find((filter) => typeof filter?.predicate === "function");
    const matchesCollection = collectionFilter?.predicate as
      | ((query: { queryKey: readonly unknown[] }) => boolean)
      | undefined;
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

  it("reconciles caches after EventSource reconnects", () => {
    render(<WorkspaceRealtimeConnection />);
    const source = connection();

    source.onerror?.(new Event("error"));
    expect(queryClient.invalidateQueries).not.toHaveBeenCalled();
    source.onopen?.(new Event("open"));

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

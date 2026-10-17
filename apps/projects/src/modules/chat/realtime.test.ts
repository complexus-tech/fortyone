import { QueryClient, QueryObserver } from "@tanstack/react-query";
import { chatKeys } from "./constants";
import { reconcileChatQueries } from "./realtime";

afterEach(() => {
  jest.useRealTimers();
});

it("refreshes visible reaction copies in the conversation and thread without touching other rooms or workspaces", async () => {
  jest.useFakeTimers();
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity } },
  });
  let reactions: { emoji: string; count: number; reacted: boolean }[] = [];
  const keys = [
    chatKeys.messages("second", "product"),
    chatKeys.replies("second", "product", "root"),
    [...chatKeys.all("second"), "message", "root"],
  ];
  const reads = keys.map(() =>
    jest.fn(async () => ({ version: 1, reactions })),
  );
  const observers = keys.map(
    (queryKey, index) =>
      new QueryObserver(client, {
        queryKey,
        queryFn: reads[index],
        initialData: { version: 1, reactions: [] },
      }),
  );
  const unsubscribe = observers.map((observer) =>
    observer.subscribe(() => undefined),
  );
  const otherRoom = chatKeys.replies("second", "design", "other-root");
  const otherWorkspace = chatKeys.messages("another-workspace", "product");
  client.setQueryData(otherRoom, { reactions: [] });
  client.setQueryData(otherWorkspace, { reactions: [] });

  try {
    const refreshReactions = async (next: typeof reactions) => {
      reactions = next;
      const refresh = reconcileChatQueries(client, "SECOND", "product");
      await jest.advanceTimersByTimeAsync(150);
      await refresh;
      for (const key of keys) {
        expect(client.getQueryData(key)).toEqual({ version: 1, reactions });
      }
    };
    await refreshReactions([{ emoji: "👍", count: 1, reacted: false }]);
    await refreshReactions([]);
    for (const read of reads) expect(read).toHaveBeenCalledTimes(2);
    expect(client.getQueryState(otherRoom)?.isInvalidated).toBe(false);
    expect(client.getQueryState(otherWorkspace)?.isInvalidated).toBe(false);
  } finally {
    for (const stop of unsubscribe) stop();
    client.clear();
  }
});

import { MutationObserver } from "@tanstack/react-query";
import { getQueryClient } from "./get-query-client";

describe("query client mutation retries", () => {
  it("does not repeat a write when its response is lost", async () => {
    const queryClient = getQueryClient();
    const create = jest
      .fn()
      .mockRejectedValue(new Error("Response lost after commit"));
    const observer = new MutationObserver(queryClient, { mutationFn: create });

    await expect(observer.mutate(undefined)).rejects.toThrow(
      "Response lost after commit",
    );
    expect(create).toHaveBeenCalledTimes(1);
    expect(queryClient.getDefaultOptions().queries?.retry).toBe(1);
    queryClient.clear();
  });

  it("preserves explicit retry policies for idempotent mutations", async () => {
    const queryClient = getQueryClient();
    const create = jest
      .fn()
      .mockRejectedValueOnce(new Error("Response lost"))
      .mockResolvedValueOnce({ id: "created-once" });
    const observer = new MutationObserver(queryClient, {
      mutationFn: create,
      retry: 1,
      retryDelay: 0,
    });

    await expect(observer.mutate(undefined)).resolves.toEqual({
      id: "created-once",
    });
    expect(create).toHaveBeenCalledTimes(2);
    queryClient.clear();
  });
});

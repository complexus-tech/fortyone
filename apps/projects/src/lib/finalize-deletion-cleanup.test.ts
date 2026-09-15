import { finalizeDeletionCleanup } from "./finalize-deletion-cleanup";

it("navigates after a synchronous storage failure without hiding that failure", async () => {
  const failure = new Error("Storage unavailable");
  const navigate = jest.fn();
  await expect(
    finalizeDeletionCleanup(() => {
      throw failure;
    }, navigate),
  ).rejects.toBe(failure);
  expect(navigate).toHaveBeenCalledTimes(1);
});

it("navigates after asynchronous query cleanup failure", async () => {
  const failure = new Error("Cleanup failed");
  const navigate = jest.fn();
  await expect(
    finalizeDeletionCleanup(() => Promise.reject(failure), navigate),
  ).rejects.toBe(failure);
  expect(navigate).toHaveBeenCalledTimes(1);
});

it("waits for pending cleanup before navigating exactly once", async () => {
  let finish: (() => void) | undefined;
  const cleanup = new Promise<void>((resolve) => {
    finish = resolve;
  });
  const navigate = jest.fn();
  const result = finalizeDeletionCleanup(() => cleanup, navigate);
  await Promise.resolve();
  expect(navigate).not.toHaveBeenCalled();
  finish?.();
  await result;
  expect(navigate).toHaveBeenCalledTimes(1);
});

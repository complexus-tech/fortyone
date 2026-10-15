import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { retryFailedDrafts } from "../drafts";
import { native, getBoundAccountId } from "../native";
import { useBrowserSignIn } from "./use-browser-sign-in";
vi.mock("../drafts", () => ({ retryFailedDrafts: vi.fn() }));
vi.mock("../native", () => ({
  native: { signIn: vi.fn(), logout: vi.fn(), deleteDraft: vi.fn() },
  getBoundAccountId: vi.fn(),
  errorMessage: (error: unknown) =>
    error instanceof Error ? error.message : String(error),
}));
beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(retryFailedDrafts).mockResolvedValue(undefined);
  vi.mocked(native.signIn).mockResolvedValue(undefined);
  vi.mocked(getBoundAccountId).mockReturnValue("account");
});
describe("browser account reconnection", () => {
  it("waits for durable drafts, suppresses duplicate requests, and preserves the account and downloaded work", async () => {
    let saved!: () => void;
    vi.mocked(retryFailedDrafts).mockReturnValue(
      new Promise<void>((resolve) => {
        saved = resolve;
      }),
    );
    const { result } = renderHook(() => useBrowserSignIn("account"));
    let first!: Promise<void>;
    act(() => {
      first = result.current.begin();
    });
    expect(result.current.busy).toBe(true);
    await act(async () => {
      await result.current.begin();
    });
    expect(native.signIn).not.toHaveBeenCalled();
    await act(async () => {
      saved();
      await first;
    });
    expect(native.signIn).toHaveBeenCalledOnce();
    expect(native.logout).not.toHaveBeenCalled();
    expect(native.deleteDraft).not.toHaveBeenCalled();
    expect(result.current.opened).toBe(true);
  });
  it("retains a failed local draft and reports the error before opening browser sign-in", async () => {
    vi.mocked(retryFailedDrafts).mockRejectedValue(
      new Error("Could not save your draft."),
    );
    const { result } = renderHook(() => useBrowserSignIn("account"));
    await act(async () => {
      await result.current.begin();
    });
    expect(native.signIn).not.toHaveBeenCalled();
    expect(result.current.error).toBe("Could not save your draft.");
  });
  it("rejects a delayed reconnect after the initiating account changes", async () => {
    vi.mocked(getBoundAccountId).mockReturnValue("another-account");
    const { result } = renderHook(() => useBrowserSignIn("account"));
    await act(async () => {
      await result.current.begin();
    });
    expect(native.signIn).not.toHaveBeenCalled();
    expect(result.current.error).toMatch(/account changed/);
  });
});

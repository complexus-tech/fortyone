import { act, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { WorkspaceSettings } from "../types";
import { native } from "../native";
import { CachedSettings } from "./cached-settings";
vi.mock("../native", () => ({
  getBoundAccountId: () => "account",
  native: {
    forAccount: () => native,
    readSettings: vi.fn(),
    syncSettings: vi.fn(),
    listen: vi.fn().mockResolvedValue(() => undefined),
  },
  errorMessage: String,
}));

describe("cached settings warmup", () => {
  it("replaces the initial download notice after native warmup without navigation or another network sync", async () => {
    Object.defineProperty(navigator, "onLine", {
      configurable: true,
      value: true,
    });
    let cached: WorkspaceSettings = {
      profile: {},
      workspace: {},
      teams: [],
      members: [],
      preferences: {},
      downloadedAt: null,
    };
    vi.mocked(native.readSettings).mockImplementation(async () => cached);
    vi.mocked(native.syncSettings).mockImplementation(async () => cached);
    render(<CachedSettings workspaceId="workspace" />);
    await screen.findByText(
      /Connect once to download account and workspace settings/,
    );
    await waitFor(() => expect(native.syncSettings).toHaveBeenCalledOnce());
    cached = {
      ...cached,
      profile: { fullName: "Downloaded profile" },
      downloadedAt: "2026-10-14T12:00:00Z",
    };
    const stateListener = vi
      .mocked(native.listen)
      .mock.calls.find(([name]) => name === "desktop-state-changed")?.[1];
    expect(stateListener).toBeDefined();
    await act(async () => {
      stateListener?.({ event: "desktop-state-changed", id: 1, payload: null });
    });
    expect(await screen.findByText("Downloaded profile")).toBeInTheDocument();
    expect(
      screen.queryByText(
        /Connect once to download account and workspace settings/,
      ),
    ).not.toBeInTheDocument();
    expect(native.syncSettings).toHaveBeenCalledOnce();
  });
});

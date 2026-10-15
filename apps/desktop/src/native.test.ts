import { beforeEach, describe, expect, it, vi } from "vitest";
import { invoke } from "@tauri-apps/api/core";
import { bindNativeAccount, native } from "./native";
vi.mock("@tauri-apps/api/core", () => ({
  invoke: vi.fn().mockResolvedValue(undefined),
}));
beforeEach(() => {
  vi.clearAllMocks();
});

describe("native account boundaries", () => {
  it("retries secure-store access only when an explicit synchronization requests it", async () => {
    const client = native.forAccount("account");
    await client.sync("workspace");
    expect(invoke).toHaveBeenLastCalledWith("desktop_sync", {
      workspaceId: "workspace",
      expectedAccountId: "account",
      retryCredentials: false,
    });
    await client.sync("workspace", true);
    expect(invoke).toHaveBeenLastCalledWith("desktop_sync", {
      workspaceId: "workspace",
      expectedAccountId: "account",
      retryCredentials: true,
    });
  });
  it("retains the initiating identity when a delayed request runs after reauthentication", async () => {
    bindNativeAccount("account-a");
    const initiatingClient = native.forAccount("account-a");
    bindNativeAccount("account-b");
    await initiatingClient.saveDraft("workspace", "story", "story", {
      title: "Account A work",
    });
    expect(invoke).toHaveBeenCalledWith(
      "desktop_save_draft",
      expect.objectContaining({ expectedAccountId: "account-a" }),
    );
    await native.forAccount("account-b").readWorkspace("workspace");
    expect(invoke).toHaveBeenLastCalledWith("desktop_read_workspace", {
      workspaceId: "workspace",
      expectedAccountId: "account-b",
    });
  });
});

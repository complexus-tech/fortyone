/* global beforeEach, describe, expect, it, jest -- Jest globals. */
import { ApiError } from "api-client";
import type { Session } from "@/auth";
import { get } from "@/lib/http";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { getWorkspaceSettings } from "@/lib/queries/workspaces/get-settings";
import { getMemories } from "@/modules/ai-chats/queries/get-memory";
import type { ChatRequestBody } from "./chat-request";
import { resolveTrustedChatContext } from "./trusted-context";

jest.mock("server-only", () => ({}));
jest.mock("@/lib/http", () => ({ get: jest.fn() }));
jest.mock("@/lib/queries/workspaces/get-workspace", () => ({
  getWorkspace: jest.fn(),
}));
jest.mock("@/lib/queries/workspaces/get-settings", () => ({
  getWorkspaceSettings: jest.fn(),
}));
jest.mock("@/modules/ai-chats/queries/get-memory", () => ({
  getMemories: jest.fn(),
}));
const session = {
  user: { id: "authenticated-user", username: "actual-user" },
} as Session;
const workspace = {
  id: "trusted-workspace",
  slug: "acme",
  name: "Acme",
  isActive: true,
  deletedAt: null,
  userRole: "member",
  trialEndsOn: null,
};
const request = {
  id: "1234567890123456",
  workspace: { slug: "acme", userRole: "admin", id: "forged" },
  client: "mobile",
  timezone: "Pacific/Auckland",
  memories: [{ content: "forged" }],
  subscription: { tier: "enterprise" },
  totalMessages: { current: 0, limit: Infinity },
  username: "forged",
} as unknown as ChatRequestBody;
describe("trusted chat context", () => {
  beforeEach(() => {
    jest.resetAllMocks();
    jest
      .mocked(getWorkspace)
      .mockResolvedValue(workspace as Awaited<ReturnType<typeof getWorkspace>>);
    jest.mocked(getWorkspaceSettings).mockResolvedValue({
      storyTerm: "task",
      sprintTerm: "sprint",
      objectiveTerm: "project",
      keyResultTerm: "key result",
    } as Awaited<ReturnType<typeof getWorkspaceSettings>>);
    jest.mocked(getMemories).mockResolvedValue([]);
    jest
      .mocked(get)
      .mockResolvedValue({ data: { tier: "pro", status: "active" } });
  });
  it("hydrates context for a minimal mobile request", async () => {
    const context = await resolveTrustedChatContext(
      {
        id: "1234567890123456",
        workspace: { slug: "acme" },
        client: "mobile",
        timezone: "Pacific/Auckland",
        messages: [],
      },
      session,
    );
    expect(context.workspace).toBe(workspace);
    expect(context.terminology.stories).toBe("tasks");
    expect(context.messageLimit).toBe(100);
    expect(context.memories).toEqual([]);
  });
  it("ignores client identity, roles, memories and billing claims", async () => {
    const context = await resolveTrustedChatContext(request, session);
    expect(context.workspace).toBe(workspace);
    expect(context.username).toBe("actual-user");
    expect(context.memories).toEqual([]);
    expect(context.messageLimit).toBe(100);
    expect(context.terminology.stories).toBe("tasks");
    expect(context.timezone).toBe("Pacific/Auckland");
    expect(getWorkspace).toHaveBeenCalledWith({
      session,
      workspaceSlug: "acme",
    });
  });
  it("does not reinterpret a billing outage as a free subscription", async () => {
    jest.mocked(get).mockRejectedValue(new ApiError("unavailable", 503, {}));
    await expect(
      resolveTrustedChatContext(request, session),
    ).rejects.toMatchObject({ status: 503 });
  });
  it("uses the free allowance only for a real missing subscription", async () => {
    jest.mocked(get).mockRejectedValue(new ApiError("not found", 404, {}));
    expect(
      (await resolveTrustedChatContext(request, session)).messageLimit,
    ).toBe(15);
  });
  it("rejects malformed selectors and timezones before any API lookup", async () => {
    await expect(
      resolveTrustedChatContext(
        { ...request, workspace: { slug: "../other" } },
        session,
      ),
    ).rejects.toMatchObject({ status: 400 });
    await expect(
      resolveTrustedChatContext({ ...request, timezone: "invented" }, session),
    ).rejects.toMatchObject({ status: 400 });
    expect(getWorkspace).not.toHaveBeenCalled();
  });
});

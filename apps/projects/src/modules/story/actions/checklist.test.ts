import { auth } from "@/auth";
import { patch, post, put, remove } from "@/lib/http";
import type { ChecklistCommand } from "../types/checklist";
import { updateChecklistAction } from "./checklist";

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/http", () => ({
  patch: jest.fn(),
  post: jest.fn(),
  put: jest.fn(),
  remove: jest.fn(),
}));

beforeEach(() => {
  jest.clearAllMocks();
  jest
    .mocked(auth)
    .mockResolvedValue({ user: { id: "viewer" } } as Awaited<
      ReturnType<typeof auth>
    >);
  for (const request of [patch, post, put, remove]) {
    jest.mocked(request).mockResolvedValue({ data: { id: "item" } });
  }
});

it.each<[ChecklistCommand, typeof post, string, unknown]>([
  [
    { type: "create", title: "Write release notes" },
    post,
    "stories/story/checklist",
    { title: "Write release notes" },
  ],
  [
    { type: "update", itemId: "item", updates: { completed: true } },
    patch,
    "stories/story/checklist/item",
    { completed: true },
  ],
  [
    { type: "reorder", itemIds: ["second", "first"] },
    put,
    "stories/story/checklist/order",
    { itemIds: ["second", "first"] },
  ],
  [
    { type: "convert", itemId: "item" },
    post,
    "stories/story/checklist/item/convert",
    {},
  ],
])(
  "routes %j without leaking command metadata into the body",
  async (command, request, path, body) => {
    await updateChecklistAction("story", command, "acme");
    expect(request).toHaveBeenCalledWith(
      path,
      body,
      expect.objectContaining({ workspaceSlug: "acme" }),
      { retry: 0 },
    );
  },
);

it.each<[ChecklistCommand, string]>([
  [{ type: "delete", itemId: "item" }, "stories/story/checklist/item"],
  [{ type: "clear" }, "stories/story/checklist/items"],
])("deletes %j in the parent story scope", async (command, path) => {
  jest.mocked(remove).mockResolvedValue({ data: null });
  await expect(
    updateChecklistAction("story", command, "acme"),
  ).resolves.toBeNull();
  expect(remove).toHaveBeenCalledWith(
    path,
    expect.objectContaining({ workspaceSlug: "acme" }),
    { retry: 0 },
  );
});

it("rejects API errors before a mutation can report success", async () => {
  jest
    .mocked(patch)
    .mockResolvedValue({ error: { message: "This story is read-only" } });
  await expect(
    updateChecklistAction(
      "story",
      { type: "update", itemId: "item", updates: { completed: true } },
      "acme",
    ),
  ).rejects.toThrow("This story is read-only");
});

it("does not send a write when the session has expired", async () => {
  jest.mocked(auth).mockResolvedValue(null);
  await expect(
    updateChecklistAction("story", { type: "clear" }, "acme"),
  ).rejects.toThrow("Please sign in");
  expect(remove).not.toHaveBeenCalled();
});

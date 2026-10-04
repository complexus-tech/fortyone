import { DEFAULT_STORIES_FILTER } from "@/components/ui/stories-filter-types";
import { listPresets } from "./api";
import { resolveSavedViews } from "./resolve-views";
import type { SavedView } from "./resolve-views";

jest.mock("./api", () => ({ listPresets: jest.fn() }));
const mockList = jest.mocked(listPresets);
const TEAM = "00000000-0000-4000-8000-000000000100";
const id = (number: number) =>
  `00000000-0000-4000-8000-${number.toString().padStart(12, "0")}`;
const view = (number: number): SavedView => ({
  id: id(number),
  teamId: TEAM,
  ownerId: id(200),
  kind: "view",
  visibility: "team",
  name: `View ${number}`,
  canEdit: true,
  createdAt: "2026-10-01",
  updatedAt: "2026-10-01",
  configuration: {
    version: 1,
    layout: "kanban",
    filters: { ...DEFAULT_STORIES_FILTER, assignedToMe: true },
    viewOptions: {
      groupBy: "priority",
      orderBy: "deadline",
      orderDirection: "asc",
      displayColumns: ["Status"],
      showSubStories: true,
      showEmptyGroups: false,
    },
  },
});
const CTX = { workspaceSlug: "acme", session: null };

describe("fresh saved-view resolution", () => {
  beforeEach(() => mockList.mockReset());
  it("resolves beyond the first fifty results and retains all view settings", async () => {
    mockList.mockResolvedValueOnce({
      items: Array.from({ length: 50 }, (_, index) => view(index + 1)),
      nextCursor: "page-two",
    });
    mockList.mockResolvedValueOnce({ items: [view(51)], nextCursor: "" });
    const controller = new AbortController();
    expect(
      await resolveSavedViews(TEAM, [id(51)], CTX, controller.signal),
    ).toEqual([view(51)]);
    expect(mockList).toHaveBeenNthCalledWith(
      2,
      TEAM,
      "view",
      "page-two",
      CTX,
      controller.signal,
    );
  });
  it("returns unavailable when every current page lacks an archived view", async () => {
    mockList
      .mockResolvedValueOnce({ items: [view(1)], nextCursor: "page-two" })
      .mockResolvedValueOnce({ items: [], nextCursor: "" });
    expect(await resolveSavedViews(TEAM, [id(51)], CTX)).toEqual([]);
    expect(mockList).toHaveBeenCalledTimes(2);
  });
  it("rejects cyclic cursor responses instead of looping", async () => {
    mockList.mockResolvedValue({ items: [], nextCursor: "same-cursor" });
    await expect(resolveSavedViews(TEAM, [id(51)], CTX)).rejects.toThrow(
      "Views could not be loaded",
    );
    expect(mockList).toHaveBeenCalledTimes(2);
  });
  it("bounds automatic traversal and reports how to browse a large team", async () => {
    mockList.mockImplementation(async (_team, _kind, cursor) => ({
      items: [],
      nextCursor: String(Number(cursor) + 1),
    }));
    await expect(resolveSavedViews(TEAM, [id(51)], CTX)).rejects.toThrow(
      "Open Views to browse",
    );
    expect(mockList).toHaveBeenCalledTimes(100);
  });
  it("ignores invalid links and aborts without requesting another page", async () => {
    expect(await resolveSavedViews(TEAM, ["invalid"], CTX)).toEqual([]);
    expect(mockList).not.toHaveBeenCalled();
    const controller = new AbortController();
    controller.abort();
    await expect(
      resolveSavedViews(TEAM, [id(51)], CTX, controller.signal),
    ).rejects.toMatchObject({ name: "AbortError" });
    expect(mockList).not.toHaveBeenCalled();
  });
  it("does not fall back to stale configuration when the fresh lookup fails", async () => {
    mockList.mockRejectedValue(new Error("Access denied"));
    await expect(resolveSavedViews(TEAM, [id(51)], CTX)).rejects.toThrow(
      "Access denied",
    );
  });
});

import { DEFAULT_STORIES_FILTER } from "@/components/ui/stories-filter-types";
import { VIEW_ICON_KEYS } from "@/shared/views/metadata";
import { viewConfiguration } from "./schemas";

const snapshot = {
  version: 1,
  layout: "list",
  filters: DEFAULT_STORIES_FILTER,
  viewOptions: {
    groupBy: "status",
    orderBy: "created",
    orderDirection: "desc",
    displayColumns: ["Status", "Created"],
    showEmptyGroups: true,
    showSubStories: false,
  },
};
describe("Saved view metadata transport", () => {
  it("accepts older snapshots without adding metadata", () => {
    expect(viewConfiguration.parse(snapshot)).toEqual(snapshot);
  });
  it.each(VIEW_ICON_KEYS)(
    "accepts trusted %s icon and Unicode description at boundary",
    (icon) => {
      const value = { ...snapshot, icon, description: "🗓".repeat(2000) };
      expect(viewConfiguration.parse(value)).toEqual(value);
    },
  );
  it("supports explicit automatic icon reset", () => {
    expect(
      viewConfiguration.parse({ ...snapshot, icon: null }).icon,
    ).toBeNull();
  });
  it.each(["https://example.com/icon.svg", "<svg>", "unknown", ""])(
    "rejects untrusted icon %s",
    (icon) => {
      expect(viewConfiguration.safeParse({ ...snapshot, icon }).success).toBe(
        false,
      );
    },
  );
  it("rejects a description beyond2,000 code points", () => {
    expect(
      viewConfiguration.safeParse({
        ...snapshot,
        description: "🗓".repeat(2001),
      }).success,
    ).toBe(false);
  });
  it("preserves exact My Work scope with nullable optional filters", () => {
    const scope = {
      kind: "my-work",
      tab: "upcoming",
      category: null,
      overdue: false,
      createdAfter: "2026-10-01",
      createdBefore: null,
    };
    expect(viewConfiguration.parse({ ...snapshot, scope }).scope).toEqual(
      scope,
    );
  });
  it.each([
    { kind: "my-work", tab: "archived" },
    { kind: "my-work", tab: "all", category: "deleted" },
    { kind: "my-work", tab: "all", userId: "other" },
  ])("rejects unsupported scope %j", (scope) => {
    expect(viewConfiguration.safeParse({ ...snapshot, scope }).success).toBe(
      false,
    );
  });
});

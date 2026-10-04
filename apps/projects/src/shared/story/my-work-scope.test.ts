import { DEFAULT_STORIES_FILTER } from "@/components/ui/stories-filter-types";
import { getMyWorkScopeFilterParams, MY_WORK_TABS } from "./my-work-scope";

const now = new Date(2026, 9, 4, 12);
describe("saved My Work scope", () => {
  it("keeps all relationship flags and explicit cross-team filters without narrowing ownership", () => {
    const result = getMyWorkScopeFilterParams(
      { ...DEFAULT_STORIES_FILTER, teamIds: ["one", "two"] },
      { kind: "my-work", tab: "all" },
      now,
    );
    expect(result).toMatchObject({
      assignedToMe: true,
      createdByMe: true,
      collaboratingWithMe: true,
      teamIds: ["one", "two"],
    });
  });
  it("re-evaluates Upcoming on the day of opening", () => {
    expect(
      getMyWorkScopeFilterParams(
        DEFAULT_STORIES_FILTER,
        { kind: "my-work", tab: "upcoming" },
        now,
      ),
    ).toMatchObject({
      deadlineAfter: "2026-10-05",
      deadlineBefore: "2026-10-11",
    });
    expect(
      getMyWorkScopeFilterParams(
        DEFAULT_STORIES_FILTER,
        { kind: "my-work", tab: "upcoming" },
        new Date(2026, 9, 7, 12),
      ),
    ).toMatchObject({
      deadlineAfter: "2026-10-08",
      deadlineBefore: "2026-10-14",
    });
  });
  it("retains category and fixed created-date scope", () => {
    expect(
      getMyWorkScopeFilterParams(
        DEFAULT_STORIES_FILTER,
        {
          kind: "my-work",
          tab: "created",
          category: "paused",
          createdAfter: "2026-09-01",
          createdBefore: "2026-09-30",
        },
        now,
      ),
    ).toMatchObject({
      createdByMe: true,
      categories: ["paused"],
      createdAfter: "2026-09-01",
      createdBefore: "2026-09-30",
    });
  });
  it("keeps explicit deadline operators ahead of relative tab and overdue dates", () => {
    expect(
      getMyWorkScopeFilterParams(
        {
          ...DEFAULT_STORIES_FILTER,
          endDate: "2026-10-12",
          operators: { endDate: "isNot" },
        },
        { kind: "my-work", tab: "today", overdue: true },
        now,
      ),
    ).toMatchObject({
      deadlineAfter: undefined,
      deadlineBefore: undefined,
      deadlineNot: "2026-10-12",
      categories: ["started"],
    });
  });
  it("retains blocked semantics", () => {
    expect(
      getMyWorkScopeFilterParams(
        DEFAULT_STORIES_FILTER,
        { kind: "my-work", tab: "blocked" },
        now,
      ),
    ).toMatchObject({
      hasBlockedBy: true,
      assignedToMe: true,
      categories: ["backlog", "unstarted", "started", "paused"],
    });
  });
  it.each(MY_WORK_TABS)("retains the %s relationship query", (tab) => {
    const result = getMyWorkScopeFilterParams(
      DEFAULT_STORIES_FILTER,
      { kind: "my-work", tab },
      now,
    );
    if (tab === "created") expect(result.createdByMe).toBe(true);
    else if (tab === "collaborating")
      expect(result.collaboratingWithMe).toBe(true);
    else expect(result.assignedToMe).toBe(true);
  });
});

/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  filterAndSortDocumentSummaries,
  getDocumentListState,
  paginateDocumentSummaries,
} from "./document-list-state";
import type { DocumentListState, WorkspaceDocumentSummary } from "./types";

const summary = (
  id: string,
  overrides: Partial<WorkspaceDocumentSummary> = {},
): WorkspaceDocumentSummary => ({
  canEdit: true,
  createdAt: "2026-07-01T08:00:00.000Z",
  createdBy: "user-1",
  id,
  relatedWorkCount: 0,
  title: `Document ${id}`,
  updatedAt: "2026-08-01T08:00:00.000Z",
  updatedBy: "user-1",
  visibility: "workspace",
  workspaceId: "workspace-1",
  ...overrides,
});

const defaultState: DocumentListState = {
  access: "all",
  direction: "desc",
  owner: "all",
  page: 1,
  sort: "updated",
  updated: "all",
};

describe("document list state", () => {
  it("parses valid URL state and falls back for invalid values", () => {
    const validParams = new URLSearchParams(
      "access=private&owner=others&updated=30d&sort=title&direction=desc&page=3",
    );
    expect(getDocumentListState(validParams)).toEqual({
      access: "private",
      direction: "desc",
      owner: "others",
      page: 3,
      sort: "title",
      updated: "30d",
    });

    expect(
      getDocumentListState(
        new URLSearchParams("access=unknown&sort=unknown&page=-2"),
      ),
    ).toEqual(defaultState);

    expect(
      getDocumentListState(new URLSearchParams("sort=created&direction=asc")),
    ).toEqual({ ...defaultState, direction: "asc" });
  });

  it("filters access, owner, and updated date before sorting", () => {
    const documents = [
      summary("older-private", {
        createdBy: "user-2",
        updatedAt: "2026-06-01T08:00:00.000Z",
        visibility: "private",
      }),
      summary("recent-private", {
        createdBy: "user-2",
        title: "Alpha",
        updatedAt: "2026-08-05T08:00:00.000Z",
        visibility: "private",
      }),
      summary("recent-workspace", {
        createdBy: "user-2",
        updatedAt: "2026-08-06T08:00:00.000Z",
      }),
      summary("mine-private", {
        createdBy: "user-1",
        updatedAt: "2026-08-06T08:00:00.000Z",
        visibility: "private",
      }),
    ];

    const result = filterAndSortDocumentSummaries({
      currentUserId: "user-1",
      documents,
      now: new Date("2026-08-07T08:00:00.000Z"),
      scope: "all",
      state: {
        ...defaultState,
        access: "private",
        owner: "others",
        updated: "7d",
      },
    });

    expect(result.map(({ id }) => id)).toEqual(["recent-private"]);
  });

  it("ignores the redundant owner filter in the My documents scope", () => {
    const result = filterAndSortDocumentSummaries({
      currentUserId: "user-1",
      documents: [summary("mine")],
      scope: "mine",
      state: { ...defaultState, owner: "others" },
    });

    expect(result).toHaveLength(1);
  });

  it("sorts titles case-insensitively in either direction", () => {
    const documents = [
      summary("bravo", { title: "bravo" }),
      summary("alpha", { title: "Alpha" }),
    ];

    const ascending = filterAndSortDocumentSummaries({
      documents,
      scope: "all",
      state: { ...defaultState, direction: "asc", sort: "title" },
    });
    const descending = filterAndSortDocumentSummaries({
      documents,
      scope: "all",
      state: { ...defaultState, direction: "desc", sort: "title" },
    });

    expect(ascending.map(({ id }) => id)).toEqual(["alpha", "bravo"]);
    expect(descending.map(({ id }) => id)).toEqual(["bravo", "alpha"]);
  });

  it("clamps pages and reports one-based result bounds", () => {
    const documents = Array.from({ length: 31 }, (_, index) =>
      summary(String(index + 1)),
    );

    expect(paginateDocumentSummaries(documents, 2)).toMatchObject({
      end: 30,
      page: 2,
      pageCount: 3,
      start: 16,
      total: 31,
    });
    expect(paginateDocumentSummaries(documents, 99)).toMatchObject({
      end: 31,
      page: 3,
      start: 31,
    });
    expect(paginateDocumentSummaries([], 1)).toMatchObject({
      end: 0,
      page: 1,
      pageCount: 1,
      start: 0,
      total: 0,
    });
  });
});

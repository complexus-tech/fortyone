/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import {
  createDelimitedImportDraft,
  inferImportMapping,
  mapRowsToImportTasks,
  parseDelimitedText,
} from "./csv";

describe("work import CSV parsing", () => {
  it("accepts the bundled Jira sample as a deterministic Jira import", () => {
    const text = readFileSync(
      resolve(process.cwd(), "../../docs/samples/jira-import-sample.csv"),
      "utf8",
    );
    const draft = createDelimitedImportDraft({
      fileHash: "sample-hash",
      fileName: "jira-import-sample.csv",
      text,
    });

    expect(draft.sourceType).toBe("jira_csv");
    expect(draft.tasks).toHaveLength(10);
    expect(draft.tasks[0]).toEqual(
      expect.objectContaining({
        sourceId: "ENG-101",
        title: "Set up the migration workspace",
        priority: "Urgent",
        startDate: "2026-09-01",
        endDate: "2026-09-03",
      }),
    );
    expect(draft.tasks[3]?.description).toContain(
      'Second line includes a "must-have" phrase and a comma',
    );
    expect(draft.warnings).toEqual([]);
  });

  it("parses quoted commas, escaped quotes, and multiline descriptions", () => {
    const parsed = parseDelimitedText(
      'Summary,Description,Priority\r\n"Ship, safely","First line\nSecond ""quoted"" line",Highest',
    );

    expect(parsed.columns).toEqual(["Summary", "Description", "Priority"]);
    expect(parsed.rows).toEqual([
      {
        Summary: "Ship, safely",
        Description: 'First line\nSecond "quoted" line',
        Priority: "Highest",
      },
    ]);
  });

  it("recognizes Jira exports and maps Jira fields deterministically", () => {
    const draft = createDelimitedImportDraft({
      fileHash: "abc123",
      fileName: "jira.csv",
      text: [
        "Issue key,Summary,Description,Status,Priority,Assignee,Due date",
        "ENG-42,Import the backlog,Reviewed scope,In Progress,High,owner@example.com,2026-09-12",
      ].join("\n"),
    });

    expect(draft.sourceType).toBe("jira_csv");
    expect(draft.tasks).toEqual([
      expect.objectContaining({
        sourceId: "ENG-42",
        title: "Import the backlog",
        priority: "High",
        assigneeEmail: "owner@example.com",
        endDate: "2026-09-12",
      }),
    ]);
  });

  it("lets a reviewed mapping rebuild tasks without trusting AI row output", () => {
    const rows = [{ Name: "Migration", Owner: "person@example.com" }];
    const mapping = {
      ...inferImportMapping(["Name", "Owner"]),
      title: "Name",
      assigneeEmail: "Owner",
    };

    expect(mapRowsToImportTasks(rows, mapping)).toEqual([
      expect.objectContaining({
        title: "Migration",
        assigneeEmail: "person@example.com",
      }),
    ]);
  });

  it("does not grant Jira-wide idempotency without a real issue-key column", () => {
    const draft = createDelimitedImportDraft({
      fileHash: "abc123",
      fileName: "jira.csv",
      text: "Issue type,Summary\nTask,Migration task",
    });

    expect(draft.sourceType).toBe("csv");
    expect(draft.tasks[0]?.sourceId).toBe("row-2");
  });

  it("leaves ambiguous numeric dates blank and explains the required format", () => {
    const draft = createDelimitedImportDraft({
      fileHash: "abc123",
      fileName: "tasks.csv",
      text: "Title,Due date\nMigration,01/02/2026",
    });

    expect(draft.tasks[0]?.endDate).toBeNull();
    expect(draft.warnings).toContain(
      "1 ambiguous numeric date was left blank. Use YYYY-MM-DD or an unambiguous date format before importing it.",
    );
  });

  it("bounds titles to the story storage contract", () => {
    const [task] = mapRowsToImportTasks([{ Title: "x".repeat(300) }], {
      ...inferImportMapping(["Title"]),
      title: "Title",
    });

    expect(task.title).toHaveLength(255);
  });
});

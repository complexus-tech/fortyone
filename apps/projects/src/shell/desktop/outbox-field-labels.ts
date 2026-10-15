import type { QueryClient } from "@tanstack/react-query";
import { htmlToPlainText } from "lib/src/html-to-plain-text";

const references: Record<string, string> = {
  statusId: "statuses",
  assigneeId: "members",
  teamId: "teams",
  sprintId: "sprints",
  parentId: "stories",
  labelIds: "labels",
  labels: "labels",
  collaboratorIds: "members",
  objectiveId: "objectives",
  keyResultId: "key-results",
};

function referenceName(value: unknown, id: string): string | undefined {
  if (Array.isArray(value)) {
    for (const item of value) {
      const name = referenceName(item, id);
      if (name) return name;
    }
    return;
  }
  if (!value || typeof value !== "object") return;
  const record = value as Record<string, unknown>;
  if (record.id === id) {
    const name = record.fullName ?? record.name ?? record.title;
    if (typeof name === "string") return htmlToPlainText(name);
  }
  for (const key of ["pages", "groups", "stories", "subStories", "data"]) {
    const name = referenceName(record[key], id);
    if (name) return name;
  }
}

/** Use already downloaded workspace metadata; reviewing a conflict needs no new request. */
export function outboxFieldValue(
  client: QueryClient,
  workspaceSlug: string,
  field: string,
  value: unknown,
): string {
  if (value === null || value === undefined || value === "") return "None";
  if (Array.isArray(value))
    return value.length
      ? value
          .map((item) => outboxFieldValue(client, workspaceSlug, field, item))
          .join(", ")
      : "None";
  if (typeof value !== "string") return JSON.stringify(value);
  const family = references[field];
  if (family) {
    for (const [, data] of client.getQueriesData({
      queryKey: [family, workspaceSlug],
    })) {
      const name = referenceName(data, value);
      if (name) return name;
    }
  }
  return htmlToPlainText(value);
}

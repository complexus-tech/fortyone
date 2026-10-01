import type { z } from "zod";
import type { WorkspaceCtx } from "@/lib/http";
import { get } from "@/lib/http";
import { workExportSchema } from "./backup-format";

export type WorkExport = z.infer<typeof workExportSchema>;

export const getWorkExport = (teamId: string | null, ctx: WorkspaceCtx) => {
  const search = new URLSearchParams();
  if (teamId) search.set("teamId", teamId);
  return get(
    `exports/work${search.size ? `?${search.toString()}` : ""}`,
    ctx,
    { cache: "no-store" },
    (response: unknown) => {
      if (
        typeof response !== "object" ||
        response === null ||
        !("data" in response)
      ) {
        throw new Error("Missing export data");
      }
      return workExportSchema.parse(response.data);
    },
  );
};

const csvCell = (value: string, numeric = false) => {
  // Quoting does not stop spreadsheet formulas; neutralize text before escaping.
  const safe =
    !numeric && /^[\s]*[=+\-@\t\r]/.test(value) ? `'${value}` : value;
  return `"${safe.replaceAll('"', '""')}"`;
};

export const workExportToCSV = (backup: WorkExport): string => {
  const teams = new Map(
    backup.analysis.teams.map((team) => [team.sourceId, team.name]),
  );
  const people = new Map(
    backup.analysis.people.map((person) => [
      person.sourceId,
      person.name || person.email || person.sourceId,
    ]),
  );
  const canonical = new Map(
    backup.taskData.map((task) => [task.sourceId, task]),
  );
  const fields = backup.customFields;
  const headers = [
    "Task ID",
    "Title",
    "Team",
    "Status",
    "Priority",
    "Assignee",
    "Description",
    "Start date",
    "Due date",
    "Created at",
    "Completed at",
    ...fields.map(
      (field) =>
        `${teams.get(field.teamSourceId ?? "") || "Workspace"}: ${field.name}${field.currency ? ` (${field.currency})` : ""} [${field.sourceId.slice(0, 8)}]`,
    ),
  ];
  const lines = [headers.map((header) => csvCell(header)).join(",")];
  for (const task of backup.analysis.tasks) {
    const data = canonical.get(task.sourceId);
    const values = new Map(
      data?.customFieldValues.map((value) => [
        value.sourceFieldId,
        value.value,
      ]) ?? [],
    );
    const cells = [
      task.sourceId,
      data?.title ?? task.title,
      teams.get(task.teamSourceId ?? "") ?? "",
      task.status ?? "",
      task.priority,
      task.assigneeName || task.assigneeEmail || "",
      data?.description ?? task.description,
      task.startDate ?? "",
      task.endDate ?? "",
      data?.createdAt ?? "",
      data?.completedAt ?? "",
    ].map((value) => csvCell(value));
    for (const field of fields) {
      const raw = values.get(field.sourceId) ?? "";
      const numeric =
        (field.type === "money" || field.type === "number") &&
        /^-?\d+(?:\.\d+)?$/.test(raw);
      let value = raw;
      if (field.type === "select") {
        value =
          field.options.find((option) => option.sourceId === raw)?.name ?? raw;
      } else if (field.type === "person") {
        value = people.get(raw) ?? raw;
      }
      cells.push(csvCell(value, numeric));
    }
    lines.push(cells.join(","));
  }
  return `\uFEFF${lines.join("\r\n")}\r\n`;
};

export const downloadWorkExport = (
  backup: WorkExport,
  format: "json" | "csv",
  workspaceSlug: string,
) => {
  const content =
    format === "json"
      ? JSON.stringify(backup)
      : workExportToCSV(backup);
  const blob = new Blob([content], {
    type:
      format === "json"
        ? "application/json;charset=utf-8"
        : "text/csv;charset=utf-8",
  });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = `fortyone-${workspaceSlug}-${new Date().toISOString().slice(0, 10)}.${format}`;
  document.body.append(link);
  link.click();
  link.remove();
  // Keep the object alive through the browser's download dispatch.
  window.setTimeout(() => {
    URL.revokeObjectURL(url);
  }, 1_000);
};

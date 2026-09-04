import "server-only";

import { createHash } from "node:crypto";
import type {
  ImportDraft,
  ImportSourceType,
} from "@/modules/settings/workspace/imports/schema";

export const acceptedExtensions = new Set([
  ".csv",
  ".tsv",
  ".xls",
  ".xlsx",
  ".pdf",
  ".jpg",
  ".jpeg",
  ".png",
  ".webp",
  ".json",
]);
export const imageExtensions = new Set([".jpg", ".jpeg", ".png", ".webp"]);
export const delimitedExtensions = new Set([".csv", ".tsv"]);
export const jsonExtensions = new Set([".json"]);

export const getFileExtension = (fileName: string) => {
  const index = fileName.lastIndexOf(".");
  return index >= 0 ? fileName.slice(index).toLowerCase() : "";
};

export const cleanFileName = (fileName: string) =>
  fileName.replace(/[^A-Za-z0-9._ -]/g, "_").slice(0, 180) || "import";

export const digest = (value: string | Buffer) =>
  createHash("sha256").update(value).digest("hex");

export const getSourceType = (extension: string): ImportSourceType => {
  if (imageExtensions.has(extension)) return "image";
  if (extension === ".pdf") return "document";
  if (extension === ".xls" || extension === ".xlsx") return "spreadsheet";
  if (jsonExtensions.has(extension)) return "json";
  return "csv";
};

const isTrelloDraft = (draft: ImportDraft | null): draft is ImportDraft =>
  draft?.sourceType === "json" &&
  draft.sourceNamespace?.startsWith("trello:board:") === true;

export const createAIAnalysisFile = ({
  bytes,
  draft,
  extension,
  fileName,
  mimeType,
}: {
  bytes: Buffer;
  draft: ImportDraft | null;
  extension: string;
  fileName: string;
  mimeType: string;
}) => {
  if (!isTrelloDraft(draft)) {
    return {
      authoritativeTaskGraph: false,
      bytes,
      extension,
      fileName,
      mimeType,
    };
  }

  const normalizedGraph = {
    authoritativeTaskGraph: true,
    format: "fortyone-normalized-import-graph-v1",
    sourceKind: "trello",
    sourceMetadata: draft.sourceMetadata,
    summary: draft.summary,
    warnings: draft.warnings,
    teams: draft.teams,
    people: draft.people,
    labels: draft.labels,
    strategicPillars: draft.strategicPillars,
    objectives: draft.objectives,
    keyResults: draft.keyResults,
    sprints: draft.sprints,
    tasks: draft.tasks,
  };
  const baseName = fileName.replace(/\.[^.]+$/u, "") || "trello-import";

  return {
    authoritativeTaskGraph: true,
    bytes: Buffer.from(JSON.stringify(normalizedGraph), "utf8"),
    extension: ".json",
    fileName: cleanFileName(`${baseName}.normalized.json`),
    mimeType: "application/json",
  };
};

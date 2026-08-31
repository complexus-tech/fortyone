import { z } from "zod";

export const IMPORT_MAX_TASKS = 500;
export const IMPORT_MAX_FILE_BYTES = 20 * 1024 * 1024;
export const JIRA_ISSUE_KEY_PATTERN = /^[A-Za-z][A-Za-z0-9]+-[1-9]\d*$/;

export const importSourceTypeSchema = z.enum([
  "jira_csv",
  "csv",
  "spreadsheet",
  "document",
  "image",
]);

export const importPrioritySchema = z.enum([
  "No Priority",
  "Urgent",
  "High",
  "Medium",
  "Low",
]);

const importDateSchema = z
  .string()
  .trim()
  .regex(/^\d{4}-\d{2}-\d{2}$/, "Date must use YYYY-MM-DD format");

export const importMappingSchema = z.strictObject({
  title: z.string().max(300).nullable(),
  description: z.string().max(300).nullable(),
  status: z.string().max(300).nullable(),
  priority: z.string().max(300).nullable(),
  assigneeEmail: z.string().max(300).nullable(),
  startDate: z.string().max(300).nullable(),
  endDate: z.string().max(300).nullable(),
  sourceId: z.string().max(300).nullable(),
});

export const importTaskSchema = z.strictObject({
  sourceId: z.string().trim().min(1).max(300),
  title: z.string().trim().min(1).max(255),
  description: z.string().max(20_000),
  status: z.string().trim().max(200).nullable(),
  priority: importPrioritySchema,
  assigneeEmail: z.string().trim().max(320).nullable(),
  startDate: importDateSchema.nullable(),
  endDate: importDateSchema.nullable(),
});

export const importAnalysisSchema = z.strictObject({
  sourceType: importSourceTypeSchema,
  summary: z.string().trim().min(1).max(1_000),
  warnings: z.array(z.string().trim().min(1).max(500)).max(20),
  mapping: importMappingSchema.nullable(),
  tasks: z.array(importTaskSchema).max(IMPORT_MAX_TASKS),
});

export const importDestinationSchema = z.discriminatedUnion("kind", [
  z.strictObject({
    kind: z.literal("existing"),
    teamId: z.string().uuid(),
  }),
  z.strictObject({
    kind: z.literal("new"),
    name: z.string().trim().min(3).max(24),
    code: z
      .string()
      .trim()
      .min(2)
      .max(3)
      .regex(/^[A-Za-z0-9]+$/),
    color: z.string().regex(/^#[0-9A-Fa-f]{6}$/),
    isPrivate: z.boolean(),
  }),
]);

export type ImportAnalysis = z.infer<typeof importAnalysisSchema>;
export type ImportDestination = z.infer<typeof importDestinationSchema>;
export type ImportMapping = z.infer<typeof importMappingSchema>;
export type ImportPriority = z.infer<typeof importPrioritySchema>;
export type ImportSourceType = z.infer<typeof importSourceTypeSchema>;
export type ImportTask = z.infer<typeof importTaskSchema>;

export type ImportDraft = ImportAnalysis & {
  columns: string[];
  fileHash: string;
  fileName: string;
  rows: Record<string, string>[];
};

export type ImportAnalysisStartResponse = {
  analysis: ImportDraft | null;
  fileHash: string;
  responseId: string | null;
  status: "completed" | "queued";
};

export type ImportAnalysisPollResponse =
  | { status: "queued" | "in_progress" }
  | { analysis: ImportAnalysis; status: "completed" };

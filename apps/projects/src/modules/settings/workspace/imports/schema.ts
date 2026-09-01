import { z } from "zod";

export const IMPORT_MAX_TASKS = 500;
export const IMPORT_MAX_TEAMS = 100;
export const IMPORT_MAX_PEOPLE = 500;
export const IMPORT_MAX_LABELS = 500;
export const IMPORT_MAX_STRATEGIC_PILLARS = 250;
export const IMPORT_MAX_OBJECTIVES = 250;
export const IMPORT_MAX_KEY_RESULTS = 500;
export const IMPORT_MAX_SPRINTS = 250;
export const IMPORT_MAX_TASK_LINKS = 100;
export const IMPORT_MAX_FILE_BYTES = 20 * 1024 * 1024;
export const IMPORT_ESTIMATE_VALUES = [1, 2, 3, 5, 8] as const;
export const IMPORT_MAX_ESTIMATED_DURATION_MINUTES = 40 * 60;
export const IMPORT_MAX_LINK_TITLE_LENGTH = 255;
export const IMPORT_MAX_LINK_URL_LENGTH = 255;
export const JIRA_ISSUE_KEY_PATTERN = /^[A-Za-z][A-Za-z0-9]+-[1-9]\d*$/;

export const importSourceTypeSchema = z.enum([
  "jira_csv",
  "csv",
  "json",
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

export const importStatusCategorySchema = z.enum([
  "backlog",
  "unstarted",
  "started",
  "paused",
  "completed",
  "cancelled",
]);

export const importKeyResultMeasurementSchema = z.enum([
  "percentage",
  "number",
  "boolean",
]);

export const importTaskAssociationTypeSchema = z.enum([
  "blocked_by",
  "blocks",
  "related",
  "duplicate",
]);

const importDateSchema = z
  .string()
  .trim()
  .regex(/^\d{4}-\d{2}-\d{2}$/, "Date must use YYYY-MM-DD format");

const importSourceIdSchema = z.string().trim().min(1).max(300);
const importOptionalSourceIdSchema = importSourceIdSchema.nullable();
const importSourceIdListSchema = z.array(importSourceIdSchema).max(100);
const importOptionalTextSchema = z.string().trim().max(20_000).nullable();
const importEstimateValueSchema = z.literal(IMPORT_ESTIMATE_VALUES).nullable();
const importDurationMinutesSchema = z
  .number()
  .int()
  .min(1)
  .max(IMPORT_MAX_ESTIMATED_DURATION_MINUTES)
  .nullable();

export const normalizeImportLinkUrl = (value: string) => {
  const trimmed = value.trim();
  if (
    !trimmed ||
    trimmed.length > IMPORT_MAX_LINK_URL_LENGTH ||
    /\p{Cc}/u.test(trimmed)
  ) {
    return null;
  }

  try {
    const parsed = new URL(trimmed);
    if (
      (parsed.protocol !== "http:" && parsed.protocol !== "https:") ||
      parsed.username ||
      parsed.password
    ) {
      return null;
    }
    const normalized = parsed.href;
    return normalized.length <= IMPORT_MAX_LINK_URL_LENGTH ? normalized : null;
  } catch {
    return null;
  }
};

export type ImportEstimateValue = (typeof IMPORT_ESTIMATE_VALUES)[number];

export type ImportTaskEffort = {
  estimateValue: ImportEstimateValue | null | undefined;
  estimatedDurationMinutes: number | null | undefined;
  minimumFocusBlockMinutes: number | null | undefined;
};

export const isValidImportEstimateValue = (
  value: unknown,
): value is ImportEstimateValue =>
  typeof value === "number" &&
  IMPORT_ESTIMATE_VALUES.some((candidate) => candidate === value);

export const isValidImportDurationMinutes = (value: unknown): value is number =>
  typeof value === "number" &&
  Number.isInteger(value) &&
  value >= 1 &&
  value <= IMPORT_MAX_ESTIMATED_DURATION_MINUTES;

export const isValidImportTaskEffort = (effort: ImportTaskEffort) => {
  const { estimateValue, estimatedDurationMinutes, minimumFocusBlockMinutes } =
    effort;
  const estimateIsAbsent =
    estimateValue === null || estimateValue === undefined;
  const durationIsAbsent =
    estimatedDurationMinutes === null || estimatedDurationMinutes === undefined;
  const focusIsAbsent =
    minimumFocusBlockMinutes === null || minimumFocusBlockMinutes === undefined;
  const estimateIsValid =
    estimateIsAbsent || isValidImportEstimateValue(estimateValue);
  const durationIsValid =
    durationIsAbsent || isValidImportDurationMinutes(estimatedDurationMinutes);
  const focusIsValid =
    focusIsAbsent || isValidImportDurationMinutes(minimumFocusBlockMinutes);
  if (!estimateIsValid || !durationIsValid || !focusIsValid) return false;
  if (
    minimumFocusBlockMinutes === null ||
    minimumFocusBlockMinutes === undefined
  )
    return true;
  if (
    estimatedDurationMinutes === null ||
    estimatedDurationMinutes === undefined
  )
    return false;
  return minimumFocusBlockMinutes <= estimatedDurationMinutes;
};

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

export const importTeamSchema = z.strictObject({
  sourceId: importSourceIdSchema,
  name: z.string().trim().min(1).max(255),
  code: z.string().trim().min(1).max(32).nullable(),
  color: z.string().trim().max(100).nullable(),
  description: importOptionalTextSchema,
  isPrivate: z.boolean(),
});

export const importPersonSchema = z.strictObject({
  sourceId: importSourceIdSchema,
  name: z.string().trim().min(1).max(255).nullable(),
  email: z.string().trim().max(320).nullable(),
  teamSourceIds: importSourceIdListSchema,
});

export const importLabelSchema = z.strictObject({
  sourceId: importSourceIdSchema,
  name: z.string().trim().min(1).max(100),
  color: z.string().trim().max(100).nullable(),
  teamSourceId: importOptionalSourceIdSchema,
});

export const importStrategicPillarSchema = z.strictObject({
  sourceId: importSourceIdSchema,
  name: z.string().trim().min(1).max(255),
  description: importOptionalTextSchema,
  orderIndex: z.number().int().min(0).max(2_147_483_647),
});

export const importObjectiveSchema = z.strictObject({
  sourceId: importSourceIdSchema,
  name: z.string().trim().min(1).max(255),
  description: importOptionalTextSchema,
  shortSummary: z.string().trim().min(1).max(500).nullable(),
  color: z.string().trim().max(100).nullable(),
  isPrivate: z.boolean(),
  status: z.string().trim().max(200).nullable(),
  statusCategory: importStatusCategorySchema.nullable(),
  priority: importPrioritySchema,
  leadPersonSourceId: importOptionalSourceIdSchema,
  teamSourceId: importOptionalSourceIdSchema,
  pillarSourceId: importOptionalSourceIdSchema,
  startDate: importDateSchema.nullable(),
  endDate: importDateSchema.nullable(),
});

export const importKeyResultSchema = z.strictObject({
  sourceId: importSourceIdSchema,
  name: z.string().trim().min(1).max(255),
  objectiveSourceId: importOptionalSourceIdSchema,
  measurementType: importKeyResultMeasurementSchema.nullable(),
  startValue: z.number().finite().nullable(),
  currentValue: z.number().finite().nullable(),
  targetValue: z.number().finite().nullable(),
  leadPersonSourceId: importOptionalSourceIdSchema,
  contributorPersonSourceIds: importSourceIdListSchema,
  startDate: importDateSchema.nullable(),
  endDate: importDateSchema.nullable(),
});

export const importSprintSchema = z.strictObject({
  sourceId: importSourceIdSchema,
  name: z.string().trim().min(1).max(255),
  goal: z.string().trim().max(10_000).nullable(),
  teamSourceId: importOptionalSourceIdSchema,
  objectiveSourceId: importOptionalSourceIdSchema,
  startDate: importDateSchema.nullable(),
  endDate: importDateSchema.nullable(),
});

export const importTaskAssociationSchema = z.strictObject({
  type: importTaskAssociationTypeSchema,
  targetSourceId: importSourceIdSchema,
});

export const importTaskLinkSchema = z.strictObject({
  title: z.string().trim().min(1).max(IMPORT_MAX_LINK_TITLE_LENGTH).nullable(),
  url: z
    .string()
    .trim()
    .min(1)
    .max(IMPORT_MAX_LINK_URL_LENGTH)
    .refine((value) => normalizeImportLinkUrl(value) !== null, {
      message: "Link URL must be an absolute HTTP or HTTPS URL",
    }),
});

export const importTaskSchema = z
  .strictObject({
    sourceId: importSourceIdSchema,
    title: z.string().trim().min(1).max(255),
    description: z.string().max(20_000),
    status: z.string().trim().max(200).nullable(),
    statusCategory: importStatusCategorySchema.nullable(),
    priority: importPrioritySchema,
    estimateValue: importEstimateValueSchema,
    estimatedDurationMinutes: importDurationMinutesSchema,
    minimumFocusBlockMinutes: importDurationMinutesSchema,
    assigneeEmail: z.string().trim().max(320).nullable(),
    assigneeName: z.string().trim().min(1).max(255).nullable(),
    assigneePersonSourceId: importOptionalSourceIdSchema,
    collaboratorPersonSourceIds: importSourceIdListSchema,
    teamSourceId: importOptionalSourceIdSchema,
    parentSourceId: importOptionalSourceIdSchema,
    objectiveSourceId: importOptionalSourceIdSchema,
    keyResultSourceId: importOptionalSourceIdSchema,
    sprintSourceId: importOptionalSourceIdSchema,
    labelSourceIds: importSourceIdListSchema,
    associations: z.array(importTaskAssociationSchema).max(100),
    links: z.array(importTaskLinkSchema).max(IMPORT_MAX_TASK_LINKS),
    startDate: importDateSchema.nullable(),
    endDate: importDateSchema.nullable(),
  })
  .superRefine((task, ctx) => {
    if (
      task.minimumFocusBlockMinutes !== null &&
      task.estimatedDurationMinutes === null
    ) {
      ctx.addIssue({
        code: "custom",
        message:
          "Minimum focus block minutes require estimated duration minutes",
        path: ["minimumFocusBlockMinutes"],
      });
      return;
    }
    if (
      task.minimumFocusBlockMinutes !== null &&
      task.estimatedDurationMinutes !== null &&
      task.minimumFocusBlockMinutes > task.estimatedDurationMinutes
    ) {
      ctx.addIssue({
        code: "custom",
        message:
          "Minimum focus block minutes must not exceed estimated duration minutes",
        path: ["minimumFocusBlockMinutes"],
      });
    }
  });

export const importAnalysisSchema = z.strictObject({
  sourceType: importSourceTypeSchema,
  sourceNamespace: importOptionalSourceIdSchema,
  summary: z.string().trim().min(1).max(1_000),
  warnings: z.array(z.string().trim().min(1).max(500)).max(50),
  mapping: importMappingSchema.nullable(),
  teams: z.array(importTeamSchema).max(IMPORT_MAX_TEAMS),
  people: z.array(importPersonSchema).max(IMPORT_MAX_PEOPLE),
  labels: z.array(importLabelSchema).max(IMPORT_MAX_LABELS),
  strategicPillars: z
    .array(importStrategicPillarSchema)
    .max(IMPORT_MAX_STRATEGIC_PILLARS),
  objectives: z.array(importObjectiveSchema).max(IMPORT_MAX_OBJECTIVES),
  keyResults: z.array(importKeyResultSchema).max(IMPORT_MAX_KEY_RESULTS),
  sprints: z.array(importSprintSchema).max(IMPORT_MAX_SPRINTS),
  tasks: z.array(importTaskSchema).max(IMPORT_MAX_TASKS),
});

export const importDestinationSchema = z.discriminatedUnion("kind", [
  z.strictObject({
    kind: z.literal("existing"),
    teamId: z.uuid(),
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
export type ImportKeyResult = z.infer<typeof importKeyResultSchema>;
export type ImportLabel = z.infer<typeof importLabelSchema>;
export type ImportMapping = z.infer<typeof importMappingSchema>;
export type ImportObjective = z.infer<typeof importObjectiveSchema>;
export type ImportPerson = z.infer<typeof importPersonSchema>;
export type ImportPriority = z.infer<typeof importPrioritySchema>;
export type ImportSprint = z.infer<typeof importSprintSchema>;
export type ImportStrategicPillar = z.infer<typeof importStrategicPillarSchema>;
export type ImportSourceType = z.infer<typeof importSourceTypeSchema>;
export type ImportStatusCategory = z.infer<typeof importStatusCategorySchema>;
export type ImportTaskAssociation = z.infer<typeof importTaskAssociationSchema>;
export type ImportTaskAssociationType = z.infer<
  typeof importTaskAssociationTypeSchema
>;
export type ImportTaskLink = z.infer<typeof importTaskLinkSchema>;
export type ImportTask = z.infer<typeof importTaskSchema>;
export type ImportTeam = z.infer<typeof importTeamSchema>;

export const normalizeImportTaskLinks = (
  links: readonly ImportTaskLink[],
): ImportTaskLink[] => {
  const seenUrls = new Set<string>();
  return links.flatMap((link) => {
    if (seenUrls.size >= IMPORT_MAX_TASK_LINKS) return [];
    const url = normalizeImportLinkUrl(link.url);
    if (!url || seenUrls.has(url)) return [];
    seenUrls.add(url);
    return [{ title: link.title?.trim() || null, url }];
  });
};

export const createEmptyImportEntityCollections = (): Pick<
  ImportAnalysis,
  | "teams"
  | "people"
  | "labels"
  | "strategicPillars"
  | "objectives"
  | "keyResults"
  | "sprints"
> => ({
  teams: [],
  people: [],
  labels: [],
  strategicPillars: [],
  objectives: [],
  keyResults: [],
  sprints: [],
});

export type ImportDraftSourceMetadata = {
  archivedTaskSourceIds: string[];
  nestedChecklistItemCount: number;
  platform: "trello";
};

export type ImportDraft = ImportAnalysis & {
  columns: string[];
  fileHash: string;
  fileName: string;
  rows: Record<string, string>[];
  sourceMetadata?: ImportDraftSourceMetadata;
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

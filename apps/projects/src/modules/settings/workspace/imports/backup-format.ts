import { z } from "zod";
import { customFieldIconSchema } from "@/modules/custom-fields/public/schemas";
import type { ImportDraft, ImportTask } from "./schema";
import { importAnalysisSchema } from "./schema";
import { getCanonicalImportEffort } from "./canonical-effort";

const sourceId = z.string().trim().min(1).max(300);
const archivedAt = z.string().datetime({ offset: true }).nullable().optional();
export const importSourceFieldSchema = z.strictObject({
  sourceId,
  name: z.string().trim().min(1).max(100),
  type: z.enum(["text", "number", "money", "date", "select", "person"]),
  icon: customFieldIconSchema.nullable().optional(),
  currency: z
    .string()
    .regex(/^[A-Z]{3}$/)
    .nullable(),
  teamSourceId: sourceId.nullable(),
  options: z
    .array(
      z.strictObject({
        sourceId,
        name: z.string().trim().min(1).max(100),
        archivedAt,
      }),
    )
    .max(100),
  archivedAt,
});
export const importSourceCommentSchema = z.strictObject({
  sourceId,
  content: z.string().min(1).max(20_000),
  authorName: z.string().trim().min(1).max(255),
  createdAt: z.string().datetime({ offset: true }).nullable(),
  parentSourceId: sourceId.nullable().optional(),
  format: z.enum(["html", "text"]).optional(),
});
export const workExportSchema = z.strictObject({
  format: z.literal("fortyone-work-export"),
  version: z.literal(1),
  generatedAt: z.string().datetime({ offset: true }).optional(),
  scope: z.strictObject({ teamSourceId: sourceId.nullable() }).optional(),
  analysis: importAnalysisSchema,
  customFields: z.array(importSourceFieldSchema).max(2_500),
  taskData: z
    .array(
      z.strictObject({
        sourceId,
        comments: z.array(importSourceCommentSchema).max(5_000),
        customFieldValues: z
          .array(
            z.strictObject({
              sourceFieldId: sourceId,
              value: z.string().max(10_000).nullable(),
            }),
          )
          .max(200),
        title: z.string().min(1).max(500).optional(),
        description: z.string().max(100_000).optional(),
        descriptionHTML: z.string().max(100_000).optional(),
        estimateValue: z.number().int().min(0).max(32767).nullable().optional(),
        estimatedDurationMinutes: z
          .number()
          .int()
          .min(1)
          .max(525600)
          .nullable()
          .optional(),
        minimumFocusBlockMinutes: z
          .number()
          .int()
          .min(1)
          .max(525600)
          .nullable()
          .optional(),
        createdAt: z.string().datetime({ offset: true }).nullable().optional(),
        updatedAt: z.string().datetime({ offset: true }).nullable().optional(),
        completedAt: z
          .string()
          .datetime({ offset: true })
          .nullable()
          .optional(),
        archivedAt: z.string().datetime({ offset: true }).nullable().optional(),
      }),
    )
    .max(10_000),
});

export const readWorkExport = (
  value: unknown,
  { fileHash, fileName }: { fileHash: string; fileName: string },
): ImportDraft | null => {
  if (
    typeof value !== "object" ||
    value === null ||
    !("format" in value) ||
    value.format !== "fortyone-work-export"
  )
    return null;
  const result = workExportSchema.safeParse(value);
  if (!result.success)
    throw new Error(
      "This FortyOne backup has an unsupported version or invalid work data.",
    );
  const { analysis, customFields, taskData } = result.data;
  const dataById = new Map(taskData.map((item) => [item.sourceId, item]));
  if (dataById.size !== taskData.length)
    throw new Error("This backup contains duplicate task data IDs.");
  const taskIds = new Set(analysis.tasks.map((task) => task.sourceId));
  if (
    taskIds.size !== analysis.tasks.length ||
    dataById.size !== taskIds.size ||
    taskData.some((item) => !taskIds.has(item.sourceId))
  )
    throw new Error(
      "This backup is missing canonical task data or contains duplicate task IDs.",
    );
  const fieldIds = new Set(customFields.map((field) => field.sourceId));
  if (
    fieldIds.size !== customFields.length ||
    customFields.some(
      (field) =>
        new Set(field.options.map((option) => option.sourceId)).size !==
        field.options.length,
    )
  )
    throw new Error(
      "This backup contains duplicate custom field or option IDs.",
    );
  const tasks: ImportTask[] = analysis.tasks.map((task) => {
    const data = dataById.get(task.sourceId);
    if (!data) return task;
    const {
      comments,
      customFieldValues,
      sourceId: _sourceId,
      ...canonical
    } = data;
    return {
      ...task,
      title: canonical.title ?? task.title,
      comments,
      customFieldValues,
      canonical,
    };
  });
  const unsupportedEffort = tasks.filter(
    (task) => getCanonicalImportEffort(task).unsupported,
  ).length;
  const longFormattedComments = tasks.reduce(
    (total, task) =>
      total +
      (task.comments ?? []).filter(
        (comment) => comment.format === "html" && comment.content.length > 8500,
      ).length,
    0,
  );
  return {
    ...analysis,
    customFields,
    tasks,
    columns: [],
    rows: [],
    fileHash,
    fileName,
    warnings: [
      ...analysis.warnings,
      ...(unsupportedEffort
        ? [
            `${unsupportedEffort} tasks have source effort outside FortyOne's native estimate (1, 2, 3, 5, 8) or duration (up to 2400 minutes) rules. Those effort values remain in source metadata and the import receipt; unsupported native values are left unset.`,
          ]
        : []),
      ...(longFormattedComments
        ? [
            `${longFormattedComments} formatted comments exceed the native comment size. Their full source HTML is preserved as attributed text across replies.`,
          ]
        : []),
      "Comments retain source author and time as attribution. FortyOne records the importing member as the audit author.",
      "Imported HTML is sanitized; unsafe markup and embedded mention metadata are removed.",
      "Original lifecycle dates remain in the import receipt. Imported work uses the reviewed status and starts unarchived.",
      "Attachment URLs are retained as links. File contents are not copied by this export import.",
    ].slice(0, 50),
  };
};

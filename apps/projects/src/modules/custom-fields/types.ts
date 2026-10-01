import { z } from "zod";

export const customFieldIconKeys = [
  "text",
  "number",
  "calendar",
  "list",
  "person",
  "team",
  "workspace",
  "goal",
  "star",
  "checklist",
  "link",
  "email",
  "clock",
  "work",
  "attachment",
  "globe",
  "tag",
  "pin",
  "objective",
  "strategy",
  "roadmap",
  "home",
  "health",
  "approval",
  "chat",
  "comment",
  "share",
  "image",
  "video",
  "microphone",
  "book",
  "help",
  "info",
  "analytics",
  "dashboard",
  "workflow",
  "kanban",
  "sprint",
  "automation",
  "history",
  "lock",
  "key",
  "code",
  "warning",
  "money",
  "coins",
  "wallet",
  "credit-card",
  "bank",
  "invoice",
  "percent",
  "calculator",
  "piggy-bank",
  "target-money",
  "building",
  "megaphone",
  "store",
  "package",
  "phone",
  "handshake",
  "shopping-cart",
] as const;

export const customFieldIconSchema = z.enum(customFieldIconKeys);
export type CustomFieldIconKey = z.infer<typeof customFieldIconSchema>;

export const customFieldTypeSchema = z.enum([
  "text",
  "number",
  "money",
  "date",
  "select",
  "person",
]);

export const customFieldSchema = z
  .object({
    id: z.uuid(),
    teamId: z.uuid(),
    name: z.string().min(1),
    type: customFieldTypeSchema,
    icon: customFieldIconSchema.nullable().optional(),
    currency: z
      .string()
      .regex(/^[A-Z]{3}$/)
      .nullable(),
    options: z.array(
      z.object({
        id: z.uuid(),
        name: z.string().min(1),
        archivedAt: z.string().nullable().optional(),
      }),
    ),
    showOnCreate: z.boolean(),
    archivedAt: z.string().nullable(),
    createdAt: z.string(),
    updatedAt: z.string(),
  })
  .superRefine((field, ctx) => {
    if ((field.type === "money") !== Boolean(field.currency))
      ctx.addIssue({
        code: "custom",
        path: ["currency"],
        message: "Only money fields require a currency",
      });
    if (
      new Set(field.options.map((option) => option.id)).size !==
      field.options.length
    )
      ctx.addIssue({
        code: "custom",
        path: ["options"],
        message: "Option IDs must be unique",
      });
  });

export const customFieldValueSchema = z.object({
  fieldId: z.uuid(),
  value: z.string().nullable(),
});

export const storyCustomFieldsSchema = z.object({
  fields: z.array(customFieldSchema),
  values: z.array(customFieldValueSchema),
  version: z.number().int().nonnegative().optional(),
});

export const customFieldReportSchema = z
  .object({
    field: customFieldSchema,
    aggregation: z.enum(["sum", "average", "min", "max", "count"]),
    groupBy: z.enum(["none", "status", "assignee", "month"]),
    currency: z.string().nullable(),
    totalCount: z.number().int().nonnegative(),
    valuedCount: z.number().int().nonnegative(),
    missingCount: z.number().int().nonnegative(),
    rows: z.array(
      z.object({
        key: z.string(),
        label: z.string(),
        value: z.union([z.literal(""), z.string().regex(/^-?\d+(?:\.\d+)?$/)]),
        count: z.number().int().nonnegative(),
      }),
    ),
  })
  .superRefine((report, ctx) => {
    if (report.totalCount !== report.valuedCount + report.missingCount) {
      ctx.addIssue({
        code: "custom",
        path: ["missingCount"],
        message: "Report coverage must account for every matching item",
      });
    }
    if (
      report.rows.reduce((count, row) => count + row.count, 0) !==
      report.valuedCount
    ) {
      ctx.addIssue({
        code: "custom",
        path: ["valuedCount"],
        message: "Report groups must account for every valued item",
      });
    }
    const nullableAggregate =
      report.aggregation === "average" ||
      report.aggregation === "min" ||
      report.aggregation === "max";
    report.rows.forEach((row, index) => {
      // SQL preserves absent averages/extrema as empty strings, while an
      // actual numeric zero remains an exact decimal string.
      if (
        nullableAggregate
          ? (row.value === "") !== (row.count === 0)
          : row.value === ""
      ) {
        ctx.addIssue({
          code: "custom",
          path: ["rows", index, "value"],
          message:
            "Only an empty average, minimum, or maximum group has no value",
        });
      }
    });
  });

export type CustomField = z.infer<typeof customFieldSchema>;
export type CustomFieldType = CustomField["type"];
export type CustomFieldValue = z.infer<typeof customFieldValueSchema>;
export type StoryCustomFields = z.infer<typeof storyCustomFieldsSchema>;
export type CustomFieldReport = z.infer<typeof customFieldReportSchema>;
export const customFieldStoryValuesSchema = z.object({
  items: z.array(
    z.object({
      storyId: z.uuid(),
      values: z.array(customFieldValueSchema),
      version: z.number().int().nonnegative(),
    }),
  ),
});
export type CustomFieldStoryValues = z.infer<
  typeof customFieldStoryValuesSchema
>;
export type CustomFieldDraft = {
  name: string;
  type: CustomFieldType;
  icon?: CustomFieldIconKey | null;
  currency: string | null;
  options: { id: string; name: string }[];
  showOnCreate: boolean;
};
export type CustomFieldUpdate = Pick<
  CustomFieldDraft,
  "name" | "icon" | "options" | "showOnCreate"
>;
export type CustomFieldValueMap = Record<string, string | null>;
export type CustomFieldReportInput = {
  fieldId: string;
  aggregation: CustomFieldReport["aggregation"];
  groupBy: CustomFieldReport["groupBy"];
  statusIds: string[];
  assigneeIds: string[];
  startDate?: string;
  endDate?: string;
  dateBasis: string;
};

export const CUSTOM_FIELD_TYPE_LABELS: Record<CustomFieldType, string> = {
  text: "Text",
  number: "Number",
  money: "Money",
  date: "Date",
  select: "Select",
  person: "Person",
};

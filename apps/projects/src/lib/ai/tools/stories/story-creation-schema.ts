import { z } from "zod";
import { isEstimateValue } from "@/lib/estimate";
import { MAX_TIME_NEEDED_MINUTES } from "@/lib/time-needed";

const storyPrioritySchema = z.enum([
  "No Priority",
  "Low",
  "Medium",
  "High",
  "Urgent",
]);

const estimateValueSchema = z
  .number()
  .int()
  .refine((value) => value === 0 || isEstimateValue(value), {
    message: "Complexity must be 1, 2, 3, 5, or 8.",
  })
  .nullable()
  .optional()
  .describe(
    "Relative complexity value using the team's scale. Use 1, 2, 3, 5, or 8. This is not a time duration; use 0, null, or omit when unset.",
  );

const estimatedDurationMinutesSchema = z
  .number()
  .int()
  .positive()
  .max(MAX_TIME_NEEDED_MINUTES)
  .nullable()
  .optional()
  .describe(
    "Total time needed in minutes. Maya uses this exact amount when reserving calendar time. Never infer it from complexity. For one story, ask when it is missing unless the user says to skip planning. For multiple stories, omit it unless the user explicitly provides this story's duration or says one duration applies to every story.",
  );

const minimumFocusBlockMinutesSchema = z
  .number()
  .int()
  .positive()
  .max(MAX_TIME_NEEDED_MINUTES)
  .nullable()
  .optional()
  .describe(
    "Optional smallest schedulable focus block in minutes. It cannot exceed estimatedDurationMinutes; omit to let Maya fill the assignee's available calendar time.",
  );

const autoSchedulingEnabledSchema = z
  .boolean()
  .optional()
  .describe(
    "Set true only when the user explicitly wants Maya to reserve and maintain focus time on the assignee's calendar. A future/start/delivery date alone is not consent. Omitting this field keeps calendar scheduling off.",
  );

const storyPlanningShape = {
  assigneeId: z
    .string()
    .nullable()
    .optional()
    .describe("Assignee user ID (UUID)"),
  priority: storyPrioritySchema.optional().describe("Story priority"),
  estimateValue: estimateValueSchema,
  estimatedDurationMinutes: estimatedDurationMinutesSchema,
  minimumFocusBlockMinutes: minimumFocusBlockMinutesSchema,
  autoSchedulingEnabled: autoSchedulingEnabledSchema,
  labelIds: z
    .array(z.string())
    .nullable()
    .optional()
    .describe("Label IDs to attach to the story."),
  sprintId: z
    .string()
    .nullable()
    .optional()
    .describe(
      "Sprint ID. When no explicit endDate is supplied, the sprint end date becomes the story's delivery date.",
    ),
  objectiveId: z
    .string()
    .nullable()
    .optional()
    .describe("Objective ID to assign story"),
  keyResultId: z
    .string()
    .nullable()
    .optional()
    .describe("Key result ID to assign story"),
  parentId: z
    .string()
    .nullable()
    .optional()
    .describe("Parent story ID for sub-stories"),
  startDate: z
    .string()
    .nullable()
    .optional()
    .describe(
      "Earliest delivery/work date in YYYY-MM-DD format. This is a date-level scheduling constraint, not an exact calendar time.",
    ),
  endDate: z
    .string()
    .nullable()
    .optional()
    .describe(
      "Delivery or due date in YYYY-MM-DD format. Maya schedules focus time before this date when auto-scheduling is enabled.",
    ),
};

const storyContentShape = {
  title: z.string().describe("Story title (required)"),
  description: z.string().nullable().optional().describe("Story description"),
  descriptionHTML: z
    .string()
    .nullable()
    .optional()
    .describe(
      "Story description HTML (always provide clean formatted HTML when description is provided)",
    ),
};

type SchedulingInput = {
  assigneeId?: string | null;
  autoSchedulingEnabled?: boolean;
  endDate?: string | null;
  estimatedDurationMinutes?: number | null;
  sprintId?: string | null;
};

const addAutoSchedulingIssues = (
  input: SchedulingInput,
  ctx: z.RefinementCtx,
  pathPrefix: PropertyKey[] = [],
) => {
  if (!input.autoSchedulingEnabled) return;

  if (!input.assigneeId) {
    ctx.addIssue({
      code: "custom",
      message:
        "Choose an assignee before enabling calendar scheduling so Maya knows whose calendar to use.",
      path: [...pathPrefix, "assigneeId"],
    });
  }

  if (!input.estimatedDurationMinutes) {
    ctx.addIssue({
      code: "custom",
      message:
        "Set the time needed before enabling calendar scheduling so Maya knows how much focus time to reserve.",
      path: [...pathPrefix, "estimatedDurationMinutes"],
    });
  }

  if (!input.endDate && !input.sprintId) {
    ctx.addIssue({
      code: "custom",
      message:
        "Set a delivery date or choose a sprint with an end date before enabling calendar scheduling.",
      path: [...pathPrefix, "endDate"],
    });
  }
};

export const createStoryInputSchema = z
  .object({
    ...storyContentShape,
    teamId: z
      .string()
      .describe("Team ID where story belongs (required) (UUID)"),
    statusId: z
      .string()
      .nullable()
      .optional()
      .describe(
        "Initial status ID (UUID). Resolve it when the user specifies a status; otherwise omit it to use the team's default status.",
      ),
    ...storyPlanningShape,
    priority: storyPrioritySchema
      .default("No Priority")
      .describe("Story priority (required)"),
  })
  .superRefine((input, ctx) => {
    addAutoSchedulingIssues(input, ctx);
  });

const bulkStoryItemSchema = z.object({
  ...storyContentShape,
  teamId: z
    .string()
    .nullable()
    .optional()
    .describe(
      "Team ID for this story. Use null or omit when sharedValues supplies the team.",
    ),
  statusId: z
    .string()
    .nullable()
    .optional()
    .describe(
      "Initial status ID. Omit to use sharedValues or the team's default status.",
    ),
  ...storyPlanningShape,
});

const bulkStorySharedValuesSchema = z
  .object({
    teamId: z.string().optional().describe("Shared team ID"),
    statusId: z
      .string()
      .nullable()
      .optional()
      .describe("Shared initial status ID"),
    ...storyPlanningShape,
  })
  .describe(
    "Values explicitly chosen for every story in this request. Use this for a shared team, assignee, status, or other genuinely common metadata. Include a shared delivery date, sprint, time needed, or calendar choice only when the user explicitly says it applies to every story. Individual story values override shared values; missing bulk planning values stay unset and calendar scheduling stays off.",
  );

export type BulkStoryItemInput = z.infer<typeof bulkStoryItemSchema>;
export type BulkStorySharedValues = z.infer<typeof bulkStorySharedValuesSchema>;

const isDefinedBulkStoryValue = (value: unknown) =>
  value !== null && value !== undefined;

export const applyBulkStorySharedValues = (
  sharedValues: BulkStorySharedValues | undefined,
  story: BulkStoryItemInput,
) => {
  // OpenAI strict tool schemas encode omitted optional properties as null.
  // Treat those placeholders as inheritance so they cannot erase the shared
  // assignee, delivery date, duration, team, or calendar choice for a batch.
  const explicitStoryValues = Object.fromEntries(
    Object.entries(story).filter(([, value]) => isDefinedBulkStoryValue(value)),
  ) as BulkStoryItemInput;

  return {
    autoSchedulingEnabled: false,
    priority: "No Priority" as const,
    ...sharedValues,
    ...explicitStoryValues,
  };
};

export const bulkCreateStoriesInputSchema = z
  .object({
    sharedValues: bulkStorySharedValuesSchema.optional(),
    storiesData: z
      .array(bulkStoryItemSchema)
      .min(1, "Provide at least one story to create.")
      .max(50, "Create at most 50 stories in one request.")
      .describe("The titles and per-story values for this creation request."),
  })
  .superRefine(({ sharedValues, storiesData }, ctx) => {
    storiesData.forEach((story, index) => {
      const resolvedStory = applyBulkStorySharedValues(sharedValues, story);

      if (!resolvedStory.teamId) {
        ctx.addIssue({
          code: "custom",
          message:
            "Provide a teamId for this story or one shared teamId for the request.",
          path: ["storiesData", index, "teamId"],
        });
      }

      addAutoSchedulingIssues(resolvedStory, ctx, ["storiesData", index]);
    });
  });

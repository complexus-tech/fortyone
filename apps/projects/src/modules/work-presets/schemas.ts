import { z } from "zod";
import { customFieldValueSchema } from "@/modules/custom-fields/public/schemas";

const ids = z.array(z.uuid()).max(100).nullable();
const optionalDate = z
  .string()
  .regex(/^\d{4}-\d{2}-\d{2}$/)
  .nullable()
  .optional();
const priority = z.enum(["No Priority", "Urgent", "High", "Medium", "Low"]);
const filterFields = z.enum([
  "contentContains",
  "statusIds",
  "assigneeIds",
  "reporterIds",
  "priorities",
  "teamIds",
  "sprintIds",
  "labelIds",
  "estimateValues",
  "objectiveId",
  "startDate",
  "endDate",
  "hasNoAssignee",
]);
const operators = z.enum([
  "contains",
  "doesNotContain",
  "isAnyOf",
  "isNotAnyOf",
  "is",
  "isNot",
  "isOnOrBefore",
  "isOnOrAfter",
  "isEmpty",
  "isNotEmpty",
]);
const viewConfiguration = z.strictObject({
  version: z.literal(1),
  layout: z.enum(["list", "kanban"]),
  filters: z.strictObject({
    statusIds: ids,
    assigneeIds: ids,
    reporterIds: ids,
    teamIds: ids,
    sprintIds: ids,
    labelIds: ids,
    priorities: z.array(priority).max(5).nullable(),
    estimateValues: z.array(z.number().nonnegative()).max(100).nullable(),
    parentId: z.uuid().nullable(),
    objectiveId: z.uuid().nullable(),
    epicId: z.uuid().nullable(),
    keyResultId: z.uuid().nullable(),
    contentContains: z.string().max(200).nullable().optional(),
    startDate: optionalDate,
    endDate: optionalDate,
    completedAfter: optionalDate,
    completedBefore: optionalDate,
    hasNoAssignee: z.boolean().nullable(),
    hasBlockedBy: z.boolean().nullable(),
    assignedToMe: z.boolean(),
    createdByMe: z.boolean(),
    isCompleted: z.boolean().nullable().optional(),
    isNotCompleted: z.boolean().nullable().optional(),
    operators: z.partialRecord(filterFields, operators).optional(),
  }),
  viewOptions: z.strictObject({
    displayColumnsVersion: z.number().int().min(0).max(3).optional(),
    groupBy: z.enum(["status", "assignee", "priority", "none"]),
    orderBy: z.enum(["priority", "deadline", "created", "updated"]),
    orderDirection: z.enum(["asc", "desc"]),
    showEmptyGroups: z.boolean(),
    showSubStories: z.boolean(),
    displayColumns: z
      .array(
        z.enum([
          "ID",
          "Status",
          "Assignee",
          "Estimate",
          "Time needed",
          "Priority",
          "Deadline",
          "Created",
          "Updated",
          "Sprint",
          "Objective",
          "Key Result",
          "Epic",
          "Labels",
        ]),
      )
      .max(14),
    hiddenKanbanGroups: z
      .partialRecord(
        z.enum(["status", "assignee", "priority"]),
        z.array(z.string().max(64)).max(100),
      )
      .optional(),
    selectedCustomFieldIds: z.array(z.uuid()).max(3).optional(),
  }),
});
export const templateConfiguration = z.strictObject({
  version: z.literal(1),
  title: z.string().max(255),
  description: z.string().max(20000),
  descriptionHTML: z.string().max(40000),
  priority,
  statusId: z.uuid().optional(),
  assigneeId: z.uuid().optional(),
  labelIds: z.array(z.uuid()).max(100).optional(),
  estimateLabel: z.string().max(32).optional(),
  estimateValue: z.number().nonnegative().max(10000).optional(),
  estimatedDurationMinutes: z.number().int().min(5).max(10080).optional(),
  minimumFocusBlockMinutes: z.number().int().min(5).max(10080).optional(),
  checklist: z.array(z.string().min(1).max(500)).max(50),
  customFieldValues: z.array(customFieldValueSchema).max(50).optional(),
});
const common = {
  id: z.uuid(),
  teamId: z.uuid(),
  ownerId: z.uuid(),
  name: z.string().min(1).max(100),
  visibility: z.enum(["personal", "team"]),
  createdAt: z.string(),
  updatedAt: z.string(),
  canEdit: z.boolean(),
};
export const presetSchema = z.discriminatedUnion("kind", [
  z.object({
    ...common,
    kind: z.literal("view"),
    configuration: viewConfiguration,
  }),
  z.object({
    ...common,
    kind: z.literal("template"),
    configuration: templateConfiguration,
  }),
]);
export const presetPageSchema = z.object({
  items: z.array(presetSchema).max(100),
  nextCursor: z.string().max(8192),
});

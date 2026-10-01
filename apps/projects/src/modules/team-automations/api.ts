import { z } from "zod";
import type { WorkspaceCtx } from "@/lib/http";
import { get, post, put, remove } from "@/lib/http";
import { templateConfiguration } from "@/modules/work-presets/public/schemas";
import type { Automation, AutomationInput, AutomationRun } from "./types";

const priority = z.enum(["No Priority", "Low", "Medium", "High", "Urgent"]);
const ruleConfiguration = z.strictObject({
  version: z.literal(1),
  trigger: z.enum(["story.created", "story.updated"]),
  conditions: z.strictObject({
    statusIds: z.array(z.uuid()).max(100).optional(),
    priorities: z.array(priority).max(5).optional(),
    assigneeIds: z.array(z.uuid()).max(100).optional(),
    unassigned: z.boolean().optional(),
  }),
  actions: z.strictObject({
    statusId: z.uuid().optional(),
    priority: priority.optional(),
    assigneeId: z.uuid().optional(),
    clearAssignee: z.boolean().optional(),
  }),
});
const recurrenceConfiguration = z.strictObject({
  version: z.literal(1),
  schedule: z.strictObject({
    frequency: z.enum(["daily", "weekly", "monthly"]),
    timezone: z.string().max(100),
    startsOn: z.string(),
    localTime: z.string(),
    weekday: z.number().int().min(0).max(6),
    monthDay: z.number().int().min(1).max(31),
  }),
  draft: templateConfiguration.omit({ version: true, estimateLabel: true }),
});
const common = {
  id: z.uuid(),
  teamId: z.uuid(),
  ownerId: z.uuid(),
  name: z.string(),
  paused: z.boolean(),
  nextRunAt: z.string().nullable(),
  lastRunAt: z.string().nullable(),
  lastError: z.string(),
  createdAt: z.string(),
  canEdit: z.boolean(),
};
const automationSchema = z.discriminatedUnion("kind", [
  z.strictObject({
    ...common,
    kind: z.literal("rule"),
    configuration: ruleConfiguration,
  }),
  z.strictObject({
    ...common,
    kind: z.literal("recurrence"),
    configuration: recurrenceConfiguration,
  }),
]);
const runSchema = z
  .object({
    id: z.string().uuid(),
    automationId: z.string().uuid(),
    occurrence: z.string(),
    status: z.enum(["running", "succeeded", "skipped", "failed"]),
    storyId: z.string().uuid().nullable(),
    error: z.string(),
    startedAt: z.string(),
    finishedAt: z.string().nullable(),
  })
  .strict();
const data = (response: unknown): unknown => {
  if (!response || typeof response !== "object" || !("data" in response))
    throw new Error("Missing automation data");
  return response.data;
};
const decodeAutomation = (response: unknown): Automation =>
  automationSchema.parse(data(response));
const decodeAutomations = (response: unknown): Automation[] =>
  z.array(automationSchema).max(100).parse(data(response));
const decodeRuns = (response: unknown): AutomationRun[] =>
  z.array(runSchema).max(50).parse(data(response));
export const listAutomations = (teamId: string, ctx: WorkspaceCtx) =>
  get(
    `team-automations?teamId=${encodeURIComponent(teamId)}`,
    ctx,
    undefined,
    decodeAutomations,
  );
export const createAutomation = (input: AutomationInput, ctx: WorkspaceCtx) =>
  post("team-automations", input, ctx, undefined, decodeAutomation);
export const pauseAutomation = (
  id: string,
  paused: boolean,
  ctx: WorkspaceCtx,
) =>
  put(`team-automations/${id}`, { paused }, ctx, undefined, decodeAutomation);
export const archiveAutomation = (id: string, ctx: WorkspaceCtx) =>
  remove(`team-automations/${id}`, ctx);
export const listRuns = (id: string, ctx: WorkspaceCtx) =>
  get(`team-automations/${id}/runs`, ctx, undefined, decodeRuns);

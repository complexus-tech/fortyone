export type MayaWorkPlanActionModel = {
  assigneeId?: string;
  endAt?: string;
  id: string;
  label: string;
  reason: string;
  riskCode?: string;
  riskMessage?: string;
  startAt?: string;
  status: string;
  title?: string;
  type: string;
};

export type MayaWorkPlanModel = {
  actions: MayaWorkPlanActionModel[];
  message: string;
  runStatus: string;
  summary: string;
};

const isRecord = (value: unknown): value is Record<string, unknown> =>
  Boolean(value) && typeof value === "object" && !Array.isArray(value);

const asString = (value: unknown) =>
  typeof value === "string" ? value.trim() : "";

const actionLabel = (type: string) => {
  switch (type) {
    case "assign_story":
      return "Selected owner";
    case "schedule_work_block":
      return "Scheduled time";
    case "flag_schedule_risk":
      return "Schedule risk";
    default:
      return "Plan update";
  }
};

export const isMayaWorkPlanOutput = (
  output: unknown,
): output is Record<string, unknown> =>
  isRecord(output) && output.kind === "maya-work-plan";

export const getMayaWorkPlanModel = (
  output: unknown,
): MayaWorkPlanModel | null => {
  if (!isMayaWorkPlanOutput(output)) return null;

  const plan = isRecord(output.plan) ? output.plan : {};
  const run = isRecord(plan.run) ? plan.run : {};
  const rawActions: unknown[] = Array.isArray(plan.actions) ? plan.actions : [];
  const message = asString(output.message);
  const actions = rawActions.flatMap((rawAction, index) => {
    if (!isRecord(rawAction)) return [];

    const type = asString(rawAction.type);
    const payload = isRecord(rawAction.payload) ? rawAction.payload : {};
    const assignStory = isRecord(payload.assignStory)
      ? payload.assignStory
      : {};
    const scheduleBlock = isRecord(payload.scheduleBlock)
      ? payload.scheduleBlock
      : {};
    const risk = isRecord(payload.risk) ? payload.risk : {};

    return [
      {
        assigneeId:
          asString(assignStory.assigneeId) ||
          asString(scheduleBlock.userId) ||
          undefined,
        endAt: asString(scheduleBlock.endAt) || undefined,
        id: asString(rawAction.id) || `${type || "action"}-${index}`,
        label: actionLabel(type),
        reason:
          asString(rawAction.reason) ||
          "Maya did not provide a reason for this action.",
        riskCode: asString(risk.code) || undefined,
        riskMessage: asString(risk.message) || undefined,
        startAt: asString(scheduleBlock.startAt) || undefined,
        status: asString(rawAction.status) || "proposed",
        title: asString(scheduleBlock.title) || undefined,
        type,
      },
    ];
  });

  return {
    actions,
    message,
    runStatus: asString(run.status) || "proposed",
    summary:
      asString(run.summary) ||
      message ||
      "Maya prepared a work plan for this story.",
  };
};

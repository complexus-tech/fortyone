import type { TaskTemplateConfiguration } from "@/modules/work-presets/public/types";

export type AutomationDraft = Omit<
  TaskTemplateConfiguration,
  "version" | "estimateLabel"
>;
export type RuleConfiguration = {
  version: 1;
  trigger: "story.created" | "story.updated";
  conditions: {
    statusIds?: string[];
    priorities?: string[];
    assigneeIds?: string[];
    unassigned?: boolean;
  };
  actions: {
    statusId?: string;
    priority?: string;
    assigneeId?: string;
    clearAssignee?: boolean;
  };
};
export type RecurrenceConfiguration = {
  version: 1;
  schedule: {
    frequency: "daily" | "weekly" | "monthly";
    timezone: string;
    startsOn: string;
    localTime: string;
    weekday: number;
    monthDay: number;
  };
  draft: AutomationDraft;
};
export type AutomationConfiguration =
  | RuleConfiguration
  | RecurrenceConfiguration;
type AutomationCommon = {
  id: string;
  teamId: string;
  ownerId: string;
  name: string;
  paused: boolean;
  nextRunAt: string | null;
  lastRunAt: string | null;
  lastError: string;
  createdAt: string;
  canEdit: boolean;
};
export type Automation = AutomationCommon &
  (
    | { kind: "rule"; configuration: RuleConfiguration }
    | { kind: "recurrence"; configuration: RecurrenceConfiguration }
  );
export type AutomationInput = Pick<
  Automation,
  "teamId" | "kind" | "name" | "configuration"
>;
export type AutomationRun = {
  id: string;
  automationId: string;
  occurrence: string;
  status: "running" | "succeeded" | "skipped" | "failed";
  storyId: string | null;
  error: string;
  startedAt: string;
  finishedAt: string | null;
};

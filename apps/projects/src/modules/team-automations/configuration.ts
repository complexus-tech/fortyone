import type { TaskTemplateConfiguration } from "@/modules/work-presets/public/types";
import type { AutomationDraft, RuleConfiguration } from "./types";

export const draftFromTemplate = (
  template: TaskTemplateConfiguration,
): AutomationDraft => ({
  title: template.title,
  description: template.description,
  descriptionHTML: template.descriptionHTML,
  priority: template.priority,
  statusId: template.statusId,
  assigneeId: template.assigneeId,
  labelIds: template.labelIds,
  estimateValue: template.estimateValue,
  estimatedDurationMinutes: template.estimatedDurationMinutes,
  minimumFocusBlockMinutes: template.minimumFocusBlockMinutes,
  checklist: template.checklist,
  customFieldValues: template.customFieldValues,
});
export const buildRule = ({
  trigger,
  statusCondition,
  priorityCondition,
  statusAction,
  priorityAction,
  assigneeAction,
}: {
  trigger: RuleConfiguration["trigger"];
  statusCondition: string;
  priorityCondition: string;
  statusAction: string;
  priorityAction: string;
  assigneeAction: string;
}): RuleConfiguration => ({
  version: 1,
  trigger,
  conditions: {
    ...(statusCondition !== "any" ? { statusIds: [statusCondition] } : {}),
    ...(priorityCondition !== "any" ? { priorities: [priorityCondition] } : {}),
  },
  actions: {
    ...(statusAction !== "unchanged" ? { statusId: statusAction } : {}),
    ...(priorityAction !== "unchanged" ? { priority: priorityAction } : {}),
    ...(assigneeAction === "unassigned" ? { clearAssignee: true } : {}),
    ...(assigneeAction !== "unchanged" && assigneeAction !== "unassigned"
      ? { assigneeId: assigneeAction }
      : {}),
  },
});

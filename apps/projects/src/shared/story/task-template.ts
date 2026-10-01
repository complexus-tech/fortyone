import type { DetailedStory, NewStory, StoryPriority } from "./types";

export type TaskTemplateConfiguration = {
  version: 1;
  title: string;
  description: string;
  descriptionHTML: string;
  priority: StoryPriority;
  statusId?: string;
  assigneeId?: string;
  labelIds?: string[];
  estimateLabel?: string;
  estimateValue?: number;
  estimatedDurationMinutes?: number;
  minimumFocusBlockMinutes?: number;
  checklist: string[];
  customFieldValues?: NonNullable<NewStory["customFieldValues"]>;
};

export const buildTaskTemplate = (
  story: DetailedStory,
  customFieldValues: NonNullable<NewStory["customFieldValues"]> = [],
): TaskTemplateConfiguration => ({
  version: 1,
  title: story.title,
  description: story.description,
  descriptionHTML: story.descriptionHTML,
  priority: story.priority,
  statusId: story.statusId,
  ...(story.assigneeId ? { assigneeId: story.assigneeId } : {}),
  labelIds: story.labels ?? [],
  ...(story.estimateLabel ? { estimateLabel: story.estimateLabel } : {}),
  ...(story.estimateValue !== null
    ? { estimateValue: story.estimateValue }
    : {}),
  ...(story.estimatedDurationMinutes !== null
    ? { estimatedDurationMinutes: story.estimatedDurationMinutes }
    : {}),
  ...(story.minimumFocusBlockMinutes !== null
    ? { minimumFocusBlockMinutes: story.minimumFocusBlockMinutes }
    : {}),
  checklist: [],
  customFieldValues,
});

export const getTemplateDraft = (
  template: TaskTemplateConfiguration,
  available: {
    statusIds: string[];
    memberIds: string[];
    labelIds: string[];
  },
): Partial<NewStory> => ({
  priority: template.priority,
  ...(template.statusId && available.statusIds.includes(template.statusId)
    ? { statusId: template.statusId }
    : {}),
  assigneeId:
    template.assigneeId && available.memberIds.includes(template.assigneeId)
      ? template.assigneeId
      : null,
  labelIds: (template.labelIds ?? []).filter((id) =>
    available.labelIds.includes(id),
  ),
  estimateValue: template.estimateValue ?? null,
  estimatedDurationMinutes: template.estimatedDurationMinutes ?? null,
  minimumFocusBlockMinutes: template.minimumFocusBlockMinutes ?? null,
});

const escapeHTML = (value: string) =>
  value.replace(
    /[&<>"']/g,
    (character) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        character
      ]!,
  );

export const templateDescriptionHTML = (
  template: TaskTemplateConfiguration,
) => {
  const description = (
    template.descriptionHTML || `<p>${escapeHTML(template.description)}</p>`
  ).replace(/data-checked="true"/g, 'data-checked="false"');
  if (!template.checklist.length) return description;
  return `${description}<ul data-type="taskList">${template.checklist.map((item) => `<li data-type="taskItem" data-checked="false"><p>${escapeHTML(item)}</p></li>`).join("")}</ul>`;
};

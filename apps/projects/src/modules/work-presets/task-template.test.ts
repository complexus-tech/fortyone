import type { DetailedStory } from "@/shared/story/types";
import {
  buildTaskTemplate,
  getTemplateDraft,
  templateDescriptionHTML,
} from "@/shared/story/task-template";
import type { TaskTemplateConfiguration } from "./types";

const template: TaskTemplateConfiguration = {
  version: 1,
  title: "Release review",
  description: "Review the release",
  descriptionHTML:
    '<ul data-type="taskList"><li data-checked="true">Already done</li></ul>',
  priority: "High",
  statusId: "old-status",
  assigneeId: "removed-person",
  labelIds: ["retained-label", "archived-label"],
  checklist: ["Check <customer> & timeline"],
};

describe("task template application", () => {
  it("filters stale references while leaving live scheduling and task identity out of the draft", () => {
    const draft = getTemplateDraft(template, {
      statusIds: ["new-status"],
      memberIds: [],
      labelIds: ["retained-label"],
    });
    expect(draft).toMatchObject({
      assigneeId: null,
      priority: "High",
      labelIds: ["retained-label"],
    });
    expect(draft).not.toHaveProperty("statusId");
    expect(draft).not.toHaveProperty("teamId");
    expect(draft).not.toHaveProperty("id");
    expect(draft).not.toHaveProperty("sprintId");
    expect(draft).not.toHaveProperty("endDate");
  });
  it("resets reused checklist completion and safely inserts explicit checklist text", () => {
    const html = templateDescriptionHTML(template);
    expect(html).not.toContain('data-checked="true"');
    expect(html).toContain('data-checked="false"');
    expect(html).toContain("Check &lt;customer&gt; &amp; timeline");
  });
  it("captures only predefined reusable metadata from a live task", () => {
    const story = {
      ...template,
      id: "live-task",
      teamId: "team",
      reporterId: "owner",
      sprintId: "sprint",
      endDate: "2026-10-10",
      labels: ["retained-label"],
      estimateValue: 3,
      estimateLabel: "M",
      estimatedDurationMinutes: 60,
      minimumFocusBlockMinutes: 30,
    } as unknown as DetailedStory;
    const saved = buildTaskTemplate(story, [{ fieldId: "field", value: null }]);
    expect(saved).toMatchObject({
      estimateValue: 3,
      labelIds: ["retained-label"],
      customFieldValues: [{ fieldId: "field", value: null }],
    });
    expect(saved).not.toHaveProperty("id");
    expect(saved).not.toHaveProperty("sprintId");
    expect(saved).not.toHaveProperty("endDate");
  });
});

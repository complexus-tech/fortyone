import { buildRule, draftFromTemplate } from "./configuration";

it("persists only selected conditions and concrete actions", () => {
  const rule = buildRule({
    trigger: "story.updated",
    statusCondition: "any",
    priorityCondition: "Urgent",
    statusAction: "status-active",
    priorityAction: "unchanged",
    assigneeAction: "unassigned",
  });
  expect(rule).toEqual({
    version: 1,
    trigger: "story.updated",
    conditions: { priorities: ["Urgent"] },
    actions: { statusId: "status-active", clearAssignee: true },
  });
  expect(JSON.stringify(rule)).not.toContain("unchanged");
});

it("takes a typed draft snapshot without saved-view metadata or estimate labels", () => {
  const template = {
    version: 1 as const,
    title: "Weekly review",
    description: "Steps",
    descriptionHTML: "<p>Steps</p>",
    priority: "High" as const,
    statusId: "status-active",
    assigneeId: "member-active",
    labelIds: ["label-active"],
    estimateLabel: "Two points",
    estimateValue: 2,
    estimatedDurationMinutes: 30,
    minimumFocusBlockMinutes: 15,
    checklist: ["Review release"],
    customFieldValues: [{ fieldId: "field-id", value: null }],
  };
  const draft = draftFromTemplate(template);
  expect(draft).toMatchObject({
    title: "Weekly review",
    statusId: "status-active",
    estimateValue: 2,
    minimumFocusBlockMinutes: 15,
    customFieldValues: [{ fieldId: "field-id", value: null }],
  });
  expect(draft).not.toHaveProperty("version");
  expect(draft).not.toHaveProperty("estimateLabel");
});

/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getActivityFieldMeta,
  getLabelActivityDisplayValue,
  isAssociationActivityField,
} from "./activity-field-renderers";

const createRendererOptions = () => ({
  activityAssignees: [],
  renderEstimate: (value: string) => `estimate-${value}`,
  renderObjective: (value: string) => `objective-${value}`,
  renderPriority: (value: string) => `priority-${value}`,
  renderSprint: (value: string) => `sprint-${value}`,
  renderTimeNeeded: (value: string) => `time-needed-${value}`,
  statuses: [],
  withWorkspace: (path: string) => `/acme${path}`,
});

describe("activity field renderers", () => {
  it("keeps the field labels and fallback used by activity copy", () => {
    const options = createRendererOptions();

    expect(getActivityFieldMeta("estimate_unit", options).label).toBe(
      "Complexity",
    );
    expect(getActivityFieldMeta("collaborator_ids", options).label).toBe(
      "Collaborators",
    );
    expect(getActivityFieldMeta("custom_field", options).label).toBe(
      "custom_field",
    );
  });

  it("delegates renderers that need Activity-owned hooks", () => {
    const options = createRendererOptions();

    expect(getActivityFieldMeta("estimate_unit", options).render("3")).toBe(
      "estimate-3",
    );
    expect(getActivityFieldMeta("priority", options).render("high")).toBe(
      "priority-high",
    );
    expect(getActivityFieldMeta("sprint_id", options).render("sprint-1")).toBe(
      "sprint-sprint-1",
    );
    expect(
      getActivityFieldMeta("objective_id", options).render("objective-1"),
    ).toBe("objective-objective-1");
  });

  it("keeps association and label display semantics", () => {
    expect(isAssociationActivityField("blocking_id")).toBe(true);
    expect(isAssociationActivityField("labels")).toBe(false);
    expect(
      getLabelActivityDisplayValue([
        { color: "#e11d48", id: "label-1", name: "Design" },
      ]),
    ).toBe("Design");
    expect(
      getLabelActivityDisplayValue([
        { color: "#e11d48", id: "label-1", name: "Design" },
        { color: "#2563eb", id: "label-2", name: "Frontend" },
      ]),
    ).toBe("2 labels");
  });
});

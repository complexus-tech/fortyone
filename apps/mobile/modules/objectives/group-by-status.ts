import type { Objective, ObjectiveStatus } from "./types";

export function groupObjectivesByStatus(
  objectives: Objective[],
  statuses: ObjectiveStatus[],
) {
  const remaining = new Map<string, Objective[]>();
  for (const objective of objectives) {
    const items = remaining.get(objective.statusId) ?? [];
    items.push(objective);
    remaining.set(objective.statusId, items);
  }
  const sections: {
    key: string;
    title: string;
    status?: ObjectiveStatus;
    data: Objective[];
  }[] = [];
  for (const status of [...statuses].sort(
    (a, b) => a.orderIndex - b.orderIndex,
  )) {
    const data = remaining.get(status.id);
    if (!data) continue;
    sections.push({ key: status.id, title: status.name, status, data });
    remaining.delete(status.id);
  }
  // Keep objectives visible if their status has been removed or is unavailable.
  for (const [key, data] of remaining) {
    sections.push({ key, title: "No status", data });
  }
  return sections;
}

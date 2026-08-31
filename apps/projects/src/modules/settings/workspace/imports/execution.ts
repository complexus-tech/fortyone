import type { Member } from "@/types";
import type { State, StateCategory } from "@/types/states";
import type { ImportTask } from "./schema";
import type { ImportStoryPayload } from "./api";

const STATUS_CATEGORY_ALIASES: {
  category: StateCategory;
  matches: readonly string[];
}[] = [
  {
    category: "completed",
    matches: ["done", "closed", "complete", "completed", "resolved"],
  },
  {
    category: "cancelled",
    matches: ["cancelled", "canceled", "rejected", "won't do", "wont do"],
  },
  {
    category: "paused",
    matches: ["paused", "blocked", "on hold", "waiting"],
  },
  {
    category: "started",
    matches: ["in progress", "doing", "started", "active", "review"],
  },
  { category: "backlog", matches: ["backlog", "icebox"] },
  {
    category: "unstarted",
    matches: ["to do", "todo", "open", "new", "ready", "selected"],
  },
];

const normalize = (value: string) =>
  value.trim().toLowerCase().replace(/[_-]+/g, " ").replace(/\s+/g, " ");

const hasControlCharacter = (value: string) => /\p{Cc}/u.test(value);

const toHex = (bytes: ArrayBuffer) =>
  Array.from(new Uint8Array(bytes), (byte) =>
    byte.toString(16).padStart(2, "0"),
  ).join("");

export const getBoundedImportSourceKey = async (sourceId: string) => {
  const value = sourceId.trim();
  const encoded = new TextEncoder().encode(value);
  if (value && encoded.byteLength <= 256 && !hasControlCharacter(value)) {
    return value;
  }

  const digest = await globalThis.crypto.subtle.digest("SHA-256", encoded);
  return `source:${toHex(digest)}`;
};

const getInferredCategory = (sourceStatus: string) => {
  const value = normalize(sourceStatus);
  return STATUS_CATEGORY_ALIASES.find(({ matches }) =>
    matches.some((match) => value === match || value.startsWith(`${match} `)),
  )?.category;
};

export const resolveImportStatus = (
  sourceStatus: string | null,
  statuses: State[],
): State | undefined => {
  if (sourceStatus) {
    const exact = statuses.find(
      (status) => normalize(status.name) === normalize(sourceStatus),
    );
    if (exact) return exact;

    const category = getInferredCategory(sourceStatus);
    if (category) {
      const categoryStatuses = statuses.filter(
        (status) => status.category === category,
      );
      return (
        categoryStatuses.find((status) => status.isDefault) ??
        categoryStatuses.sort((a, b) => a.orderIndex - b.orderIndex).at(0)
      );
    }
  }

  return (
    statuses.find(
      (status) => status.isDefault && status.category === "unstarted",
    ) ??
    statuses.find((status) => status.isDefault) ??
    [...statuses].sort((a, b) => a.orderIndex - b.orderIndex).at(0)
  );
};

export const resolveImportAssignee = (
  assigneeEmail: string | null,
  members: Member[],
) => {
  if (!assigneeEmail) return undefined;
  const email = assigneeEmail.trim().toLowerCase();
  return members.find(
    (member) => member.isActive && member.email.trim().toLowerCase() === email,
  );
};

export const toImportStoryPayload = ({
  members,
  statuses,
  task,
  teamId,
}: {
  members: Member[];
  statuses: State[];
  task: ImportTask;
  teamId: string;
}): ImportStoryPayload => {
  const status = resolveImportStatus(task.status, statuses);
  const assignee = resolveImportAssignee(task.assigneeEmail, members);

  return {
    title: task.title.trim(),
    description: task.description.trim(),
    teamId,
    ...(status ? { statusId: status.id } : {}),
    ...(assignee ? { assigneeId: assignee.id } : {}),
    priority: task.priority,
    ...(task.startDate ? { startDate: task.startDate } : {}),
    ...(task.endDate ? { endDate: task.endDate } : {}),
  };
};

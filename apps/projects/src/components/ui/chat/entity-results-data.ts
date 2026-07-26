import type { StoryPriority } from "@/modules/stories/types";

export type EntityResultIcon =
  | { kind: "avatar"; name: string; src?: string }
  | { color: string; kind: "color" }
  | {
      kind: "icon";
      name:
        | "comment"
        | "feedback"
        | "key-result"
        | "link"
        | "notification"
        | "objective"
        | "sprint"
        | "team";
    }
  | { kind: "priority"; priority: StoryPriority };

export type EntityResultTrailing =
  | { kind: "status"; label: string; tone: EntityResultTone }
  | { kind: "text"; label: string };

export type EntityResultTone =
  | "danger"
  | "info"
  | "muted"
  | "success"
  | "warning";

export type EntityResultItem = {
  href?: string;
  icon: EntityResultIcon;
  id: string;
  title: string;
  trailing?: EntityResultTrailing;
};

export type EntityResultsModel = {
  emptyMessage: string;
  items: EntityResultItem[];
  title: string;
};

export const ENTITY_RESULT_TOOL_TYPES = new Set([
  "tool-listObjectivesTool",
  "tool-listTeamObjectivesTool",
  "tool-listKeyResultsTool",
  "tool-listSprints",
  "tool-listRunningSprints",
  "tool-listTeams",
  "tool-listPublicTeams",
  "tool-listTeamMembers",
  "tool-members",
  "tool-listCustomerFeedbackTool",
  "tool-listIntegrationRequestsTool",
  "tool-notifications",
  "tool-comments",
  "tool-labels",
  "tool-links",
]);

const STORY_PRIORITIES = new Set<StoryPriority>([
  "No Priority",
  "Urgent",
  "High",
  "Medium",
  "Low",
]);

const isRecord = (value: unknown): value is Record<string, unknown> =>
  Boolean(value) && typeof value === "object" && !Array.isArray(value);

const asString = (value: unknown) =>
  typeof value === "string" ? value.trim() : "";

const asNumber = (value: unknown) =>
  typeof value === "number" && Number.isFinite(value) ? value : undefined;

const asPriority = (value: unknown): StoryPriority =>
  typeof value === "string" && STORY_PRIORITIES.has(value as StoryPriority)
    ? (value as StoryPriority)
    : "No Priority";

const asRecords = (value: unknown): Record<string, unknown>[] =>
  Array.isArray(value) ? value.filter(isRecord) : [];

const uniqueItems = (items: EntityResultItem[]) => {
  const seen = new Set<string>();
  return items.filter((item) => {
    if (seen.has(item.id)) return false;
    seen.add(item.id);
    return true;
  });
};

const compactItems = (items: (EntityResultItem | null)[]) =>
  items.filter((item): item is EntityResultItem => item !== null);

const statusTone = (status: string): EntityResultTone => {
  switch (status.toLowerCase().replaceAll("_", " ")) {
    case "accepted":
    case "completed":
    case "done":
    case "on track":
      return "success";
    case "at risk":
    case "in progress":
    case "planned":
    case "reviewing":
      return "warning";
    case "closed":
    case "declined":
    case "off track":
      return "danger";
    case "pending":
    case "unread":
      return "info";
    default:
      return "muted";
  }
};

const textTrailing = (label: string): EntityResultTrailing | undefined =>
  label ? { kind: "text", label } : undefined;

const statusTrailing = (label: string): EntityResultTrailing | undefined =>
  label
    ? {
        kind: "status",
        label,
        tone: statusTone(label),
      }
    : undefined;

const percentage = (
  current: number | undefined,
  target: number | undefined,
) => {
  if (current === undefined || target === undefined || target === 0) return "";
  return `${Math.round((current / target) * 100)}%`;
};

const stripHtml = (value: string) =>
  value
    .replace(/<[^>]*>/g, " ")
    .replace(/&nbsp;/g, " ")
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&#39;/g, "'")
    .replace(/&quot;/g, '"')
    .replace(/\s+/g, " ")
    .trim();

const safeExternalUrl = (value: unknown) => {
  const url = asString(value);
  if (!url) return undefined;

  try {
    const parsedUrl = new URL(url);
    return parsedUrl.protocol === "http:" || parsedUrl.protocol === "https:"
      ? parsedUrl.href
      : undefined;
  } catch {
    return undefined;
  }
};

const urlHost = (value: unknown) => {
  const url = safeExternalUrl(value);
  if (!url) return "";

  return new URL(url).hostname.replace(/^www\./, "");
};

const toObjectives = (output: Record<string, unknown>): EntityResultItem[] =>
  compactItems(
    asRecords(output.objectives).map((objective): EntityResultItem | null => {
      const id = asString(objective.id);
      const title = asString(objective.name);
      const teamId = asString(objective.teamId);
      const health = asString(objective.health);
      if (!id || !title) return null;

      return {
        href: teamId ? `/teams/${teamId}/objectives/${id}` : undefined,
        icon: { kind: "icon", name: "objective" },
        id,
        title,
        trailing: statusTrailing(health),
      } satisfies EntityResultItem;
    }),
  );

const toKeyResults = (output: Record<string, unknown>): EntityResultItem[] =>
  compactItems(
    asRecords(output.keyResults).map((keyResult): EntityResultItem | null => {
      const id = asString(keyResult.id);
      const title = asString(keyResult.name);
      const teamId = asString(keyResult.teamId);
      const objectiveId = asString(keyResult.objectiveId);
      if (!id || !title) return null;

      return {
        href:
          teamId && objectiveId
            ? `/teams/${teamId}/objectives/${objectiveId}`
            : undefined,
        icon: { kind: "icon", name: "key-result" },
        id,
        title,
        trailing: textTrailing(
          percentage(
            asNumber(keyResult.currentValue),
            asNumber(keyResult.targetValue),
          ),
        ),
      } satisfies EntityResultItem;
    }),
  );

const toSprints = (
  output: Record<string, unknown>,
  runningOnly: boolean,
): EntityResultItem[] =>
  compactItems(
    asRecords(output.sprints).map((sprint): EntityResultItem | null => {
      const id = asString(sprint.id);
      const title = asString(sprint.name);
      const teamId = asString(sprint.teamId);
      const stats = isRecord(sprint.stats) ? sprint.stats : {};
      const total = asNumber(stats.total);
      const completed = asNumber(stats.completed);
      if (!id || !title) return null;

      const progress =
        total !== undefined && completed !== undefined
          ? `${completed}/${total}`
          : "";

      return {
        href: teamId ? `/teams/${teamId}/sprints/${id}/stories` : "/sprints",
        icon: { kind: "icon", name: "sprint" },
        id,
        title,
        trailing: runningOnly
          ? statusTrailing("In progress")
          : textTrailing(progress),
      } satisfies EntityResultItem;
    }),
  );

const toTeams = (output: Record<string, unknown>): EntityResultItem[] =>
  compactItems(
    asRecords(output.teams).map((team): EntityResultItem | null => {
      const id = asString(team.id);
      const title = asString(team.name);
      if (!id || !title) return null;

      const memberCount = asNumber(team.memberCount);
      return {
        href: `/teams/${id}/stories`,
        icon: {
          color: asString(team.color) || "var(--color-icon)",
          kind: "color",
        },
        id,
        title,
        trailing: textTrailing(
          memberCount === undefined
            ? asString(team.code)
            : `${memberCount} ${memberCount === 1 ? "member" : "members"}`,
        ),
      } satisfies EntityResultItem;
    }),
  );

const toMembers = (output: Record<string, unknown>): EntityResultItem[] =>
  compactItems(
    asRecords(output.members).map((member): EntityResultItem | null => {
      const id = asString(member.id);
      const title =
        asString(member.fullName) ||
        asString(member.name) ||
        asString(member.username);
      if (!id || !title) return null;

      return {
        href: `/profile/${id}`,
        icon: {
          kind: "avatar",
          name: title,
          src: asString(member.avatarUrl) || undefined,
        },
        id,
        title,
        trailing: textTrailing(asString(member.role)),
      } satisfies EntityResultItem;
    }),
  );

const flattenTeamItems = (
  output: Record<string, unknown>,
  key: "feedback" | "requests",
): Record<string, unknown>[] =>
  asRecords(output.teams).flatMap((team) => {
    const teamId = asString(team.teamId);
    return asRecords(team[key]).map((item) => ({
      ...item,
      teamId,
    }));
  });

const toFeedback = (output: Record<string, unknown>): EntityResultItem[] =>
  compactItems(
    flattenTeamItems(output, "feedback").map(
      (feedback): EntityResultItem | null => {
        const id = asString(feedback.id);
        const title = asString(feedback.title);
        const teamId = asString(feedback.teamId);
        const status =
          asString(feedback.statusLabel) || asString(feedback.status);
        if (!id || !title) return null;

        return {
          href: teamId ? `/teams/${teamId}/feedback/${id}` : undefined,
          icon: { kind: "icon", name: "feedback" },
          id,
          title,
          trailing: statusTrailing(status),
        } satisfies EntityResultItem;
      },
    ),
  );

const toRequests = (output: Record<string, unknown>): EntityResultItem[] =>
  compactItems(
    flattenTeamItems(output, "requests").map(
      (request): EntityResultItem | null => {
        const id = asString(request.id);
        const title = asString(request.title);
        const teamId = asString(request.teamId);
        const status = asString(request.status);
        if (!id || !title) return null;

        return {
          href: teamId ? `/teams/${teamId}/requests/${id}` : undefined,
          icon: {
            kind: "priority",
            priority: asPriority(request.priority),
          },
          id,
          title,
          trailing: statusTrailing(status),
        } satisfies EntityResultItem;
      },
    ),
  );

const toNotifications = (output: Record<string, unknown>): EntityResultItem[] =>
  compactItems(
    asRecords(output.notifications).map(
      (notification): EntityResultItem | null => {
        const id = asString(notification.id);
        const title = asString(notification.title);
        if (!id || !title) return null;

        return {
          href: `/notifications/${id}`,
          icon: { kind: "icon", name: "notification" },
          id,
          title,
          trailing:
            notification.isRead === false
              ? statusTrailing("Unread")
              : undefined,
        } satisfies EntityResultItem;
      },
    ),
  );

const toComments = (output: Record<string, unknown>): EntityResultItem[] =>
  compactItems(
    asRecords(output.comments).map((comment): EntityResultItem | null => {
      const id = asString(comment.id);
      const title = stripHtml(asString(comment.content));
      const commenter = isRecord(comment.commenter) ? comment.commenter : {};
      const commenterName =
        asString(commenter.name) || asString(commenter.username) || "Unknown";
      if (!id || !title) return null;

      const replyCount = asNumber(comment.replyCount) ?? 0;
      return {
        icon: {
          kind: "avatar",
          name: commenterName,
          src: asString(commenter.avatarUrl) || undefined,
        },
        id,
        title,
        trailing: textTrailing(
          replyCount
            ? `${replyCount} ${replyCount === 1 ? "reply" : "replies"}`
            : commenterName,
        ),
      } satisfies EntityResultItem;
    }),
  );

const toLabels = (output: Record<string, unknown>): EntityResultItem[] =>
  compactItems(
    asRecords(output.labels).map((label): EntityResultItem | null => {
      const id = asString(label.id);
      const title = asString(label.name);
      if (!id || !title) return null;

      return {
        icon: {
          color: asString(label.color) || "var(--color-icon)",
          kind: "color",
        },
        id,
        title,
        trailing: textTrailing(label.teamId ? "Team" : "Workspace"),
      } satisfies EntityResultItem;
    }),
  );

const toLinks = (output: Record<string, unknown>): EntityResultItem[] =>
  compactItems(
    asRecords(output.links).map((link): EntityResultItem | null => {
      const id = asString(link.id);
      const href = safeExternalUrl(link.url);
      const title = asString(link.title) || urlHost(link.url);
      if (!id || !title) return null;

      return {
        href,
        icon: { kind: "icon", name: "link" },
        id,
        title,
        trailing: textTrailing(urlHost(link.url)),
      } satisfies EntityResultItem;
    }),
  );

export const getEntityResultsModel = (
  toolType: string,
  value: unknown,
): EntityResultsModel | null => {
  if (!ENTITY_RESULT_TOOL_TYPES.has(toolType) || !isRecord(value)) return null;
  if (value.success !== true) return null;

  let model: EntityResultsModel | null = null;

  switch (toolType) {
    case "tool-listObjectivesTool":
    case "tool-listTeamObjectivesTool":
      if (!Array.isArray(value.objectives)) return null;
      model = {
        emptyMessage: "No objectives found.",
        items: toObjectives(value),
        title: "Objectives",
      };
      break;
    case "tool-listKeyResultsTool":
      if (!Array.isArray(value.keyResults)) return null;
      model = {
        emptyMessage: "No key results found.",
        items: toKeyResults(value),
        title: "Key results",
      };
      break;
    case "tool-listSprints":
    case "tool-listRunningSprints":
      if (!Array.isArray(value.sprints)) return null;
      model = {
        emptyMessage: "No sprints found.",
        items: toSprints(value, toolType === "tool-listRunningSprints"),
        title: "Sprints",
      };
      break;
    case "tool-listTeams":
    case "tool-listPublicTeams":
      if (!Array.isArray(value.teams)) return null;
      model = {
        emptyMessage: "No teams found.",
        items: toTeams(value),
        title: "Teams",
      };
      break;
    case "tool-listTeamMembers":
    case "tool-members":
      if (!Array.isArray(value.members)) return null;
      model = {
        emptyMessage: "No members found.",
        items: toMembers(value),
        title: "Members",
      };
      break;
    case "tool-listCustomerFeedbackTool":
      if (!Array.isArray(value.teams)) return null;
      model = {
        emptyMessage: "No customer feedback found.",
        items: toFeedback(value),
        title: "Customer feedback",
      };
      break;
    case "tool-listIntegrationRequestsTool":
      if (!Array.isArray(value.teams)) return null;
      model = {
        emptyMessage: "No integration requests found.",
        items: toRequests(value),
        title: "Integration requests",
      };
      break;
    case "tool-notifications":
      if (!Array.isArray(value.notifications)) return null;
      model = {
        emptyMessage: "No notifications found.",
        items: toNotifications(value),
        title: "Notifications",
      };
      break;
    case "tool-comments":
      if (!Array.isArray(value.comments)) return null;
      model = {
        emptyMessage: "No comments found.",
        items: toComments(value),
        title: "Comments",
      };
      break;
    case "tool-labels":
      if (!Array.isArray(value.labels)) return null;
      model = {
        emptyMessage: "No labels found.",
        items: toLabels(value),
        title: "Labels",
      };
      break;
    case "tool-links":
      if (!Array.isArray(value.links)) return null;
      model = {
        emptyMessage: "No links found.",
        items: toLinks(value),
        title: "Links",
      };
      break;
  }

  return model ? { ...model, items: uniqueItems(model.items) } : null;
};

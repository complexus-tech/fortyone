import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getGroupedStories } from "@/modules/stories/queries/get-grouped-stories";
import { getMyStories } from "@/modules/my-work/queries/get-stories";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import type { GroupedStoryParams, Story } from "@/modules/stories/types";

const ACTIVE_STORY_CATEGORIES = [
  "backlog",
  "unstarted",
  "started",
  "paused",
] as const;

type AssignmentActor = NonNullable<Story["assignedBy"]>;

export type AssignedByResolution = {
  stories: Story[];
  resolved?: AssignmentActor;
  candidates: AssignmentActor[];
};

const normalizeActorName = (value: string) =>
  value.trim().toLocaleLowerCase().split(/\s+/u).filter(Boolean).join(" ");

const actorDisplayName = (actor: AssignmentActor) =>
  actor.fullName.trim() || actor.username.trim();

const actorMatchScore = (actor: AssignmentActor, normalizedQuery: string) => {
  const username = normalizeActorName(actor.username);
  const fullName = normalizeActorName(actor.fullName);
  if (normalizedQuery === username || normalizedQuery === fullName) return 3;
  if (fullName.split(" ").includes(normalizedQuery)) return 2;
  if (
    username.includes(normalizedQuery) ||
    fullName.includes(normalizedQuery)
  ) {
    return 1;
  }
  return 0;
};

export const filterStoriesByAssignedBy = (
  stories: Story[],
  query: string,
): AssignedByResolution => {
  const normalizedQuery = normalizeActorName(query);
  const matches = new Map<string, { actor: AssignmentActor; score: number }>();
  let bestScore = 0;

  for (const story of stories) {
    if (!story.assignedBy) continue;
    const score = actorMatchScore(story.assignedBy, normalizedQuery);
    if (score === 0) continue;
    const existing = matches.get(story.assignedBy.id);
    if (!existing || score > existing.score) {
      matches.set(story.assignedBy.id, { actor: story.assignedBy, score });
    }
    bestScore = Math.max(bestScore, score);
  }

  const candidates = [...matches.values()]
    .filter(({ score }) => score === bestScore)
    .map(({ actor }) => actor)
    .sort((left, right) => {
      const byName = actorDisplayName(left).localeCompare(
        actorDisplayName(right),
      );
      return byName || left.id.localeCompare(right.id);
    });
  if (candidates.length !== 1) return { stories: [], candidates };

  const [resolved] = candidates;
  return {
    stories: stories.filter((story) => story.assignedBy?.id === resolved.id),
    resolved,
    candidates,
  };
};

export const listTeamStoriesInputSchema = z.object({
  assignedBy: z
    .string()
    .trim()
    .min(1)
    .max(100)
    .optional()
    .describe(
      "Optional assigner name or username. Use only when the user asks what a named person assigned to them.",
    ),
  teamId: z
    .string()
    .optional()
    .describe(
      "Optional team ID to get stories from. Omit for workspace-wide story questions.",
    ),
  filters: z
    .object({
      teamIds: z
        .array(z.string())
        .optional()
        .describe(
          "Filter by one or more team IDs. Ignored when teamId is provided.",
        ),
      statusIds: z
        .array(z.string())
        .optional()
        .describe("Filter by status IDs"),
      assigneeIds: z
        .array(z.string())
        .optional()
        .describe("Filter by assignee IDs"),
      reporterIds: z
        .array(z.string())
        .optional()
        .describe("Filter by reporter or creator IDs"),
      titleContains: z
        .string()
        .optional()
        .describe("Filter stories whose title or content contains this text"),
      priorities: z
        .array(z.string())
        .optional()
        .describe("Filter by priorities"),
      sprintIds: z
        .array(z.string())
        .optional()
        .describe("Filter by sprint IDs"),
      labelIds: z.array(z.string()).optional().describe("Filter by label IDs"),
      estimateValues: z
        .array(z.number().int())
        .optional()
        .describe(
          "Filter by relative complexity values for the team's complexity scale",
        ),
      objectiveId: z.string().optional().describe("Filter by objective ID"),
      parentId: z.string().optional().describe("Filter by parent story ID"),
      hasNoAssignee: z
        .boolean()
        .optional()
        .describe("Show only stories with no assignee"),
      assignedToMe: z
        .boolean()
        .optional()
        .describe("Show only stories assigned to me"),
      createdByMe: z
        .boolean()
        .optional()
        .describe("Show only stories created by me"),
      createdAfter: z
        .string()
        .optional()
        .describe(
          "Filter stories created after this date (ISO  date string e.g 2005-06-13)",
        ),
      createdBefore: z
        .string()
        .optional()
        .describe(
          "Filter stories created before this date (ISO  date string e.g 2005-06-13)",
        ),
      updatedAfter: z
        .string()
        .optional()
        .describe(
          "Filter stories updated after this date (ISO  date string e.g 2005-06-13)",
        ),
      updatedBefore: z
        .string()
        .optional()
        .describe(
          "Filter stories updated before this date (ISO  date string e.g 2005-06-13)",
        ),
      deadlineAfter: z
        .string()
        .optional()
        .describe(
          "Filter stories with deadlines after this date (ISO  date string e.g 2005-06-13)",
        ),
      deadlineBefore: z
        .string()
        .optional()
        .describe(
          "Filter stories with deadlines before this date (ISO  date string e.g 2005-06-13)",
        ),
      completedAfter: z
        .string()
        .optional()
        .describe(
          "Filter stories completed after this date (ISO  date string e.g 2005-06-13)",
        ),
      completedBefore: z
        .string()
        .optional()
        .describe(
          "Filter stories completed before this date (ISO  date string e.g 2005-06-13)",
        ),
      includeArchived: z
        .boolean()
        .optional()
        .describe("Include archived stories"),
      includeDeleted: z
        .boolean()
        .optional()
        .describe("Include deleted stories"),
      categories: z
        .array(
          z.enum([
            "backlog",
            "unstarted",
            "started",
            "paused",
            "completed",
            "cancelled",
          ]),
        )
        .optional()
        .describe("Filter stories by status categories"),
      storiesPerGroup: z
        .number()
        .min(1)
        .max(100)
        .default(20)
        .optional()
        .describe("Number of stories to return per group (default: 20)"),
    })
    .optional()
    .describe("Optional filters for story queries"),
  groupBy: z
    .enum(["status", "assignee", "priority", "none"])
    .default("status")
    .describe("Group by status, assignee, or priority"),
  orderBy: z
    .enum(["created", "updated", "deadline", "priority"])
    .optional()
    .describe("Sort field for stories inside each group"),
  orderDirection: z.enum(["asc", "desc"]).optional().describe("Sort order"),
});

export const listTeamStories = tool({
  description:
    "List stories across the workspace or within specific teams, grouped by status, assignee, priority, or not grouped. Supports board-grade filters for manager, project manager, developer, and assignee questions.",
  inputSchema: listTeamStoriesInputSchema,

  execute: async (
    { assignedBy, teamId, filters, groupBy, orderBy, orderDirection },
    { context: experimentalContext },
  ) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access stories",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };

      const workspace = await getWorkspace(ctx);
      const userRole = workspace.userRole;

      const params: GroupedStoryParams = {
        groupBy: assignedBy ? "none" : groupBy,
        orderBy,
        orderDirection,
        teamIds: teamId ? [teamId] : filters?.teamIds,
        ...filters,
      };
      if (assignedBy) {
        params.assignedToMe = true;
        params.categories = [...ACTIVE_STORY_CATEGORIES];
        params.groupBy = "none";
        params.storiesPerGroup = 100;
      }
      if (teamId) {
        params.teamIds = [teamId];
      }

      const result = await getGroupedStories(ctx, params);
      if (assignedBy) {
        const activeStories = result.groups.flatMap((group) => group.stories);
        const activeStoryIDs = new Set(activeStories.map((story) => story.id));
        const myStories = await getMyStories(ctx);
        const attributedStories = myStories.filter((story) =>
          activeStoryIDs.has(story.id),
        );
        const resolution = filterStoriesByAssignedBy(
          attributedStories,
          assignedBy,
        );
        const sourceTruncated = result.groups.some((group) => group.hasMore);
        const candidates = resolution.candidates.map((actor) => ({
          name: actorDisplayName(actor),
          username: actor.username,
        }));

        if (!resolution.resolved) {
          const ambiguous = candidates.length > 1;
          return {
            success: true,
            kind: "story-list",
            stories: [],
            returnedCount: 0,
            assignmentFilter: {
              query: assignedBy,
              status: ambiguous
                ? "ambiguous"
                : sourceTruncated
                  ? "partial"
                  : "not_found",
              candidates,
            },
            truncated: sourceTruncated,
            userRole,
            message: ambiguous
              ? `More than one assigner matches ${assignedBy}.`
              : sourceTruncated
                ? `The active assignment scan was truncated before ${assignedBy} could be resolved.`
                : `No active stories have verified assignment history from ${assignedBy}.`,
          };
        }

        return {
          success: true,
          kind: "story-list",
          stories: [
            {
              key: "none",
              loadedCount: resolution.stories.length,
              totalCount: resolution.stories.length,
              hasMore: sourceTruncated,
              nextPage: sourceTruncated ? 2 : 1,
              stories: resolution.stories,
            },
          ],
          returnedCount: resolution.stories.length,
          assignmentFilter: {
            query: assignedBy,
            status: "matched",
            resolved: candidates[0],
          },
          truncated: sourceTruncated,
          meta: result.meta,
          userRole,
          message: `Found ${resolution.stories.length} active ${resolution.stories.length === 1 ? "story" : "stories"} assigned to you by ${actorDisplayName(resolution.resolved)}${sourceTruncated ? " in the loaded results" : ""}.`,
        };
      }
      const returnedStoryCount = result.groups.reduce(
        (count, group) => count + group.stories.length,
        0,
      );

      return {
        success: true,
        kind: "story-list",
        stories: result.groups.map((group) => ({
          ...group,
          stories: group.stories,
        })),
        returnedCount: returnedStoryCount,
        meta: result.meta,
        userRole,
        message: teamId
          ? `Found ${returnedStoryCount} ${returnedStoryCount === 1 ? "story" : "stories"} in this team.`
          : `Found ${returnedStoryCount} ${returnedStoryCount === 1 ? "story" : "stories"} in this workspace.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to list team stories",
      };
    }
  },
});

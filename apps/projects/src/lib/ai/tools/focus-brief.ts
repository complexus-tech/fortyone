import { tool } from "ai";
import { z } from "zod";
import { auth } from "@/auth";
import { getPulseReport } from "@/modules/analytics/queries/get-pulse-report";
import type { AnalyticsFilters } from "@/modules/analytics/types";
import { getGroupedStories } from "@/modules/stories/queries/get-grouped-stories";
import type {
  GroupedStoryParams,
  Story,
  StoryPriority,
} from "@/modules/stories/types";

const MAX_FOCUS_CANDIDATES = 5;
const MAX_FOCUS_RISKS = 4;
const MAX_STORY_POOL_SIZE = 10;
const DUE_SOON_DAYS = 7;

const FOCUS_SIGNAL_ORDER = [
  "blocked",
  "overdue",
  "due-today",
  "urgent",
  "high-priority",
  "due-soon",
] as const;

type FocusSignal = (typeof FOCUS_SIGNAL_ORDER)[number];

const focusSubjectSchema = z.discriminatedUnion("type", [
  z.object({ type: z.literal("current-user") }),
  z.object({
    type: z.literal("member"),
    memberId: z.string().describe("Member ID returned by resolveMember."),
  }),
  z.object({ type: z.literal("workspace") }),
]);

const uniqueIds = (ids?: string[]) => {
  const values = ids?.map((id) => id.trim()).filter(Boolean);
  return values?.length ? Array.from(new Set(values)) : undefined;
};

const flattenStories = (
  groups: Awaited<ReturnType<typeof getGroupedStories>>,
) => groups.groups.flatMap((group) => group.stories);

const getDateOnly = (value: string) => value.slice(0, 10);

const addDays = (date: string, days: number) => {
  const value = new Date(`${date}T00:00:00.000Z`);
  value.setUTCDate(value.getUTCDate() + days);
  return value.toISOString().slice(0, 10);
};

const getDeadlineSignal = (
  endDate: string | null,
  reportDate: string,
): FocusSignal | undefined => {
  if (!endDate) return undefined;

  const deadline = getDateOnly(endDate);
  const today = getDateOnly(reportDate);
  if (deadline < today) return "overdue";
  if (deadline === today) return "due-today";
  if (deadline <= addDays(today, DUE_SOON_DAYS)) return "due-soon";

  return undefined;
};

const getPrioritySignal = (
  priority: StoryPriority,
): FocusSignal | undefined => {
  if (priority === "Urgent") return "urgent";
  if (priority === "High") return "high-priority";
  return undefined;
};

const signalRank = (signals: Set<FocusSignal>) =>
  FOCUS_SIGNAL_ORDER.findIndex((signal) => signals.has(signal));

const storyPriorityRank = (priority: StoryPriority) => {
  const order: StoryPriority[] = [
    "Urgent",
    "High",
    "Medium",
    "Low",
    "No Priority",
  ];
  return order.indexOf(priority);
};

const buildCandidates = ({
  blockedStories,
  deadlineStories,
  priorityStories,
  reportDate,
}: {
  blockedStories: Story[];
  deadlineStories: Story[];
  priorityStories: Story[];
  reportDate: string;
}) => {
  const blockedIds = new Set(blockedStories.map((story) => story.id));
  const candidates = new Map<
    string,
    { signals: Set<FocusSignal>; story: Story }
  >();

  for (const story of [
    ...blockedStories,
    ...deadlineStories,
    ...priorityStories,
  ]) {
    const existing = candidates.get(story.id) ?? {
      signals: new Set<FocusSignal>(),
      story,
    };
    if (blockedIds.has(story.id)) existing.signals.add("blocked");

    const deadlineSignal = getDeadlineSignal(story.endDate, reportDate);
    if (deadlineSignal) existing.signals.add(deadlineSignal);

    const prioritySignal = getPrioritySignal(story.priority);
    if (prioritySignal) existing.signals.add(prioritySignal);

    if (existing.signals.size > 0) candidates.set(story.id, existing);
  }

  return Array.from(candidates.values())
    .sort((left, right) => {
      const signalDifference =
        signalRank(left.signals) - signalRank(right.signals);
      if (signalDifference !== 0) return signalDifference;

      const priorityDifference =
        storyPriorityRank(left.story.priority) -
        storyPriorityRank(right.story.priority);
      if (priorityDifference !== 0) return priorityDifference;

      return (left.story.endDate ?? "9999").localeCompare(
        right.story.endDate ?? "9999",
      );
    })
    .slice(0, MAX_FOCUS_CANDIDATES)
    .map(({ signals, story }) => ({
      id: story.id,
      reference: story.team?.code
        ? `${story.team.code}-${story.sequenceId}`
        : String(story.sequenceId),
      title: story.title,
      priority: story.priority,
      teamName: story.team?.name ?? "Unknown team",
      assigneeName:
        story.assignee?.fullName.trim() ||
        story.assignee?.username ||
        undefined,
      sprintName: story.sprint?.name,
      objectiveName: story.objective?.name,
      startDate: story.startDate,
      endDate: story.endDate,
      estimateLabel: story.estimateLabel,
      estimatedDurationMinutes: story.estimatedDurationMinutes,
      signals: FOCUS_SIGNAL_ORDER.filter((signal) => signals.has(signal)),
    }));
};

export const focusBriefTool = tool({
  description:
    "Return compact, text-first prioritization evidence from current project data. Use when the user asks what they, a named member, their team, or the workspace should focus on, what needs attention, or what to do next. Default I/me/my to the current user. Do not use when the user explicitly asks to see a report, chart, workload breakdown, member list, or story list. This is private supporting evidence; answer with a concise prose recommendation.",
  inputSchema: z.object({
    subject: focusSubjectSchema
      .default({ type: "current-user" })
      .describe("Who the focus brief is for."),
    teamIds: z
      .array(z.string())
      .max(20)
      .optional()
      .describe("Optional team IDs to focus on."),
    sprintIds: z
      .array(z.string())
      .max(20)
      .optional()
      .describe("Optional sprint IDs to focus on."),
    objectiveId: z.string().optional().describe("Optional objective ID."),
  }),
  execute: async (
    { subject, teamIds: inputTeamIds, sprintIds: inputSprintIds, objectiveId },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to build a focus brief",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug?: string })
        .workspaceSlug;

      if (!workspaceSlug) {
        return { success: false, error: "Workspace context is required" };
      }

      const teamIds = uniqueIds(inputTeamIds);
      const sprintIds = uniqueIds(inputSprintIds);
      let assigneeIds: string[] | undefined;
      if (subject.type === "member") {
        assigneeIds = [subject.memberId];
      } else if (subject.type === "current-user") {
        assigneeIds = [session.user.id];
      }
      const filters: AnalyticsFilters = {
        teamIds,
        assigneeIds,
        sprintIds,
        objectiveIds: objectiveId ? [objectiveId] : undefined,
      };
      const baseStoryFilters: GroupedStoryParams = {
        groupBy: "none",
        categories: ["backlog", "unstarted", "started", "paused"],
        teamIds,
        sprintIds,
        objectiveId,
        assignedToMe: subject.type === "current-user" || undefined,
        assigneeIds: subject.type === "member" ? assigneeIds : undefined,
      };
      const ctx = { session, workspaceSlug };
      const [pulse, blockedPool, deadlinePool, priorityPool] =
        await Promise.all([
          getPulseReport(ctx, filters),
          getGroupedStories(ctx, {
            ...baseStoryFilters,
            hasBlockedBy: true,
            orderBy: "priority",
            orderDirection: "asc",
            storiesPerGroup: MAX_FOCUS_CANDIDATES,
          }),
          getGroupedStories(ctx, {
            ...baseStoryFilters,
            orderBy: "deadline",
            orderDirection: "asc",
            storiesPerGroup: MAX_STORY_POOL_SIZE,
          }),
          getGroupedStories(ctx, {
            ...baseStoryFilters,
            orderBy: "priority",
            orderDirection: "asc",
            storiesPerGroup: MAX_STORY_POOL_SIZE,
          }),
        ]);
      const candidates = buildCandidates({
        blockedStories: flattenStories(blockedPool),
        deadlineStories: flattenStories(deadlinePool),
        priorityStories: flattenStories(priorityPool),
        reportDate: pulse.reportDate,
      });

      return {
        success: true,
        kind: "focus-brief-data",
        asOf: pulse.reportDate,
        scope: {
          type: subject.type,
          teamIds,
          sprintIds,
          objectiveId,
        },
        signals: {
          risks: pulse.risks.slice(0, MAX_FOCUS_RISKS).map((risk) => ({
            kind: risk.kind,
            severity: risk.severity,
            count: risk.count,
            title: risk.title,
            reason: risk.description,
          })),
          workload: {
            open: pulse.stories.openStories,
            started: pulse.stories.startedStories,
            paused: pulse.stories.pausedStories,
            blocked: pulse.stories.blockedStories,
            overdue: pulse.stories.overdueStories,
            urgent: pulse.stories.urgentStories,
            highPriority: pulse.stories.highPriorityStories,
            unestimated: pulse.stories.unestimatedStories,
          },
          delivery: {
            atRiskSprints: pulse.sprints.atRiskSprints,
            overdueSprints: pulse.sprints.overdueSprints,
            atRiskObjectives: pulse.objectives.atRiskObjectives,
            offTrackObjectives: pulse.objectives.offTrackObjectives,
            overdueObjectives: pulse.objectives.overdueObjectives,
            dueSoonObjectives: pulse.objectives.objectivesDueSoon,
          },
          requests: {
            pending: pulse.requests.pendingRequests,
            urgent: pulse.requests.urgentRequests,
            highPriority: pulse.requests.highRequests,
            stale: pulse.requests.staleRequests,
          },
        },
        candidates,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to build a focus brief",
      };
    }
  },
});

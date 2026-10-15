import type { QueryClient } from "@tanstack/react-query";
import type { WorkspaceCtx } from "@/lib/http";
import type { GroupedStoryParams } from "@/modules/stories/types";
import {
  labelKeys,
  memberKeys,
  notificationKeys,
  statusKeys,
  teamKeys,
  userKeys,
  workspaceKeys,
} from "@/constants/keys";
import { getProfile } from "@/lib/queries/users/profile";
import { getWorkspaceSettings } from "@/lib/queries/workspaces/get-settings";
import { getMembers, getTeamMembers } from "@/lib/queries/members/get-members";
import { getStatuses } from "@/lib/queries/states/get-states";
import { getLabels } from "@/lib/queries/labels/get-labels";
import { getAutomationPreferences } from "@/lib/queries/users/automation-preferences";
import { getTeams, getJoinedTeams } from "@/modules/teams/queries/get-teams";
import { getTeamSettings } from "@/modules/teams/queries/get-team-settings";
import { getNotificationPreferences } from "@/modules/notifications/queries/get-preferences";
import { getGroupedStories } from "@/modules/stories/queries/get-grouped-stories";
import { storyKeys } from "@/modules/stories/constants";
import { getTeamCustomFields } from "@/modules/custom-fields/api";
import { customFieldKeys } from "@/modules/custom-fields/hooks";
import type { Team } from "@/modules/teams/types";

/** These are the same keys and query functions consumed by the existing screens. */
export async function warmDesktopSettings(
  client: QueryClient,
  ctx: WorkspaceCtx,
) {
  const slug = ctx.workspaceSlug;
  const staleTime = 10 * 60 * 1000;
  const tasks = [
    () =>
      client.fetchQuery({
        queryKey: userKeys.profile(),
        queryFn: () => getProfile(),
        staleTime,
      }),
    () =>
      client.fetchQuery({
        queryKey: workspaceKeys.settings(slug),
        queryFn: () => getWorkspaceSettings(ctx),
        staleTime,
      }),
    () =>
      client.fetchQuery({
        queryKey: teamKeys.lists(slug),
        queryFn: () => getTeams(ctx),
        staleTime,
      }),
    () =>
      client.fetchQuery({
        queryKey: teamKeys.joined(slug),
        queryFn: () => getJoinedTeams(ctx),
        staleTime,
      }),
    () =>
      client.fetchQuery({
        queryKey: [...memberKeys.lists(slug), ""],
        queryFn: () => getMembers(ctx, ""),
        staleTime,
      }),
    () =>
      client.fetchQuery({
        queryKey: statusKeys.lists(slug),
        queryFn: () => getStatuses(ctx),
        staleTime: 60_000,
      }),
    () =>
      client.fetchQuery({
        queryKey: [...labelKeys.lists(slug), "", ""],
        queryFn: () => getLabels(ctx),
        staleTime,
      }),
    () =>
      client.fetchQuery({
        queryKey: userKeys.automationPreferences(slug),
        queryFn: () => getAutomationPreferences(ctx),
        staleTime,
      }),
    () =>
      client.fetchQuery({
        queryKey: notificationKeys.preferences(slug),
        queryFn: () => getNotificationPreferences(ctx),
        staleTime,
      }),
  ];
  const failures: unknown[] = [];
  // Limit background concurrency so the visible screen keeps network priority.
  for (let offset = 0; offset < tasks.length; offset += 3) {
    // eslint-disable-next-line no-await-in-loop -- Bounded batches preserve network priority for the visible screen.
    const settled = await Promise.allSettled(
      tasks.slice(offset, offset + 3).map((task) => task()),
    );
    for (const result of settled)
      if (result.status === "rejected") failures.push(result.reason);
  }
  const teams = client.getQueryData<Team[]>(teamKeys.joined(slug)) ?? [];
  const teamTasks = teams.flatMap((team) => [
    () =>
      client.fetchQuery({
        queryKey: customFieldKeys.team(slug, team.id),
        queryFn: () => getTeamCustomFields(team.id, ctx),
        staleTime,
      }),
    () =>
      client.fetchQuery({
        queryKey: [...memberKeys.team(slug, team.id), ""],
        queryFn: () => getTeamMembers(team.id, ctx, ""),
        staleTime,
      }),
    () =>
      client.fetchQuery({
        queryKey: teamKeys.settings(slug, team.id),
        queryFn: () => getTeamSettings(team.id, ctx),
        staleTime,
      }),
  ]);
  for (let offset = 0; offset < teamTasks.length; offset += 3) {
    // eslint-disable-next-line no-await-in-loop -- Team-dependent creation metadata is fetched after joined teams, in bounded batches.
    const settled = await Promise.allSettled(
      teamTasks.slice(offset, offset + 3).map((task) => task()),
    );
    for (const result of settled)
      if (result.status === "rejected") failures.push(result.reason);
  }
  return failures;
}

export async function warmDesktopMyWork(
  client: QueryClient,
  ctx: WorkspaceCtx,
  params: GroupedStoryParams,
) {
  return client.fetchQuery({
    queryKey: storyKeys.mineGrouped(ctx.workspaceSlug, params),
    queryFn: () => getGroupedStories(ctx, params),
    staleTime: 2 * 60 * 1000,
  });
}

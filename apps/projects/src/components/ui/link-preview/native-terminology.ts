"use client";

import { useQuery } from "@tanstack/react-query";
import { teamKeys, workspaceKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { useTerminology } from "@/hooks/use-terminology-display";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import { getWorkspaceSettings } from "@/lib/queries/workspaces/get-settings";
import { getTeamDetails } from "@/modules/teams/queries/get-team";
import type { Team } from "@/modules/teams/types";
import { useCachedQueryData } from "@/shared/query-cache/use-cached-query-data";
import type { NativeLinkResource } from "./native-query";
import type { NativeLinkTarget } from "./native-target";

const displayTerm = (term: string) => ({
  term,
  label: term.charAt(0).toUpperCase() + term.slice(1),
});

/** Read terminology from the resource's workspace, never the workspace sharing it. */
export const useNativePreviewTerminology = (
  target: NativeLinkTarget,
  resource: NativeLinkResource | null,
) => {
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  const sameWorkspace = workspaceSlug.toLowerCase() === target.workspaceSlug;
  const teamId = resource?.kind === "story" ? resource.value.teamId : null;
  const { getTermDisplay } = useTerminology(sameWorkspace ? teamId : null);
  const teams = useCachedQueryData<Team[]>(
    teamKeys.lists(target.workspaceSlug),
  );
  const cachedTeam = teams?.find((team) => team.id === teamId);
  const canReadTarget = Boolean(
    !sameWorkspace &&
      session &&
      resource &&
      (target.kind === "story" || target.kind === "objective"),
  );
  const ctx = { session, workspaceSlug: target.workspaceSlug };
  const settings = useQuery({
    queryKey: workspaceKeys.settings(target.workspaceSlug),
    queryFn: () => getWorkspaceSettings(ctx),
    enabled: canReadTarget,
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
    retry: false,
  });
  const team = useQuery({
    queryKey: teamKeys.detail(target.workspaceSlug, teamId ?? ""),
    queryFn: ({ signal }) => getTeamDetails(teamId!, ctx, signal),
    enabled: canReadTarget && Boolean(teamId) && !cachedTeam,
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
    retry: false,
  });

  if (target.kind === "document") return displayTerm("document");
  if (target.kind === "unsupported") return displayTerm("item");
  const termKey = target.kind === "story" ? "storyTerm" : "objectiveTerm";
  if (sameWorkspace) return displayTerm(getTermDisplay(termKey));

  // Cached destination settings must not label an unauthorized resource.
  if (!canReadTarget || settings.isError || !settings.data)
    return displayTerm("item");
  if (target.kind === "objective")
    return displayTerm(settings.data.objectiveTerm);

  const resolvedTeam = cachedTeam ?? (!team.isError ? team.data : null);
  if (!resolvedTeam) return displayTerm("item");
  return displayTerm(resolvedTeam.storyTerm || settings.data.storyTerm);
};

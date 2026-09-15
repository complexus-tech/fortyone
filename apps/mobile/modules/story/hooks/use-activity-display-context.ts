import type { Story } from "@/modules/stories/types";
import type { ActivityDisplayContext } from "../components/activity-display";
import { useMemo } from "react";
import { useTerminology } from "@/hooks";
import { useLabels } from "@/modules/labels/hooks/use-labels";
import { useMembers } from "@/modules/members/hooks/use-members";
import { useMayaAssignee } from "@/modules/members/hooks/use-maya-assignee";
import { useTeamObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useTeamSprints } from "@/modules/sprints/hooks";
import { useStatuses } from "@/modules/statuses/hooks/use-statuses";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { useAuthStore } from "@/store/auth";
import {
  indexActivityPeople,
  indexActivityReferences,
} from "../components/activity-display";

// One set of subscriptions/indexes per feed, rather than per virtualized row.
export function useActivityDisplayContext(
  story: Pick<Story, "teamId" | "estimateScheme">,
): ActivityDisplayContext {
  const { data: members } = useMembers();
  const { data: mayaAssignee } = useMayaAssignee();
  const { data: statuses } = useStatuses();
  const { data: sprints } = useTeamSprints(story.teamId);
  const { data: objectives } = useTeamObjectives(story.teamId);
  const { data: labels } = useLabels();
  const { data: profile } = useProfile();
  const currentUserId = useAuthStore((state) => state.userId);
  const { getTermDisplay } = useTerminology();

  return useMemo(
    () => ({
      members: indexActivityPeople(members ?? [], mayaAssignee),
      statuses: indexActivityReferences(statuses ?? []),
      sprints: indexActivityReferences(sprints ?? []),
      objectives: indexActivityReferences(objectives ?? []),
      labels: indexActivityReferences(labels ?? []),
      currentUserId,
      timezone: profile?.timezone,
      storyTerm: getTermDisplay("storyTerm"),
      sprintTerm: getTermDisplay("sprintTerm"),
      objectiveTerm: getTermDisplay("objectiveTerm"),
      keyResultTerm: getTermDisplay("keyResultTerm"),
      estimateScheme: story.estimateScheme ?? "tshirt",
    }),
    [
      members,
      mayaAssignee,
      statuses,
      sprints,
      objectives,
      labels,
      currentUserId,
      profile?.timezone,
      getTermDisplay,
      story.estimateScheme,
    ],
  );
}

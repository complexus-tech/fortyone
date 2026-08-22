"use client";
import { Box, Button, Divider, Flex, Text } from "ui";
import { MoreHorizontalIcon, TeamIcon } from "icons";
import { cn } from "lib";
import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { DndContext, type DragEndEvent } from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { useSession } from "@/lib/auth/client";
import { useJoinedTeams } from "@/modules/teams/hooks/teams";
import { useLocalStorage, useUserRole, useWorkspacePath } from "@/hooks";
import { TeamsMenu } from "@/components/ui/teams-menu";
import { useRemoveMemberMutation } from "@/modules/teams/hooks/remove-member-mutation";
import { useAddMemberMutation } from "@/modules/teams/hooks/add-member-mutation";
import { useReorderTeamsMutation } from "@/modules/teams/hooks/use-reorder-teams";
import { useTeamFeedbackSummaries } from "@/modules/team-feedback/hooks/use-team-feedback-summaries";
import { ConfirmDialog } from "@/components/ui";
import type { Team as TeamType } from "@/modules/teams/types";
import { openDialogAfterMenuClose } from "@/utils/menu-dialog-state";
import { Team } from "./team";
import {
  partitionSidebarTeams,
  pinSidebarTeam,
  reorderVisibleSidebarTeams,
} from "./team-visibility";

const DEFAULT_EXPANDED_TEAM_ID = "__first-visible-team__";

export const Teams = ({ isCollapsed = false }: { isCollapsed?: boolean }) => {
  const { data: teams = [] } = useJoinedTeams();
  const { data: feedbackSummaries = [] } = useTeamFeedbackSummaries();
  const { userRole } = useUserRole();
  const { workspaceSlug } = useWorkspacePath();
  const { teamId: activeTeamId } = useParams<{ teamId?: string }>();
  const { data: session } = useSession();
  const [team, setTeam] = useState<TeamType | null>(null);
  const [isOpen, setIsOpen] = useState(false);
  const [storedExpandedTeamId, setStoredExpandedTeamId] = useLocalStorage<
    string | null
  >(`sidebar:${workspaceSlug}:expanded-team`, DEFAULT_EXPANDED_TEAM_ID);
  const reorderTeams = useReorderTeamsMutation();

  const { visibleTeams, overflowTeams } = partitionSidebarTeams(
    teams,
    activeTeamId,
  );
  const visibleTeamIds = visibleTeams.map((team) => team.id);
  const activeTeamIsVisible = Boolean(
    activeTeamId && visibleTeamIds.includes(activeTeamId),
  );
  let expandedTeamId: string | null = storedExpandedTeamId;

  if (expandedTeamId !== null && !visibleTeamIds.includes(expandedTeamId)) {
    expandedTeamId = activeTeamIsVisible
      ? activeTeamId ?? null
      : visibleTeams[0]?.id ?? null;
  }

  useEffect(() => {
    if (!activeTeamId || !activeTeamIsVisible) return;

    setStoredExpandedTeamId((currentTeamId) =>
      currentTeamId === activeTeamId ? currentTeamId : activeTeamId,
    );
  }, [activeTeamId, activeTeamIsVisible, setStoredExpandedTeamId]);

  const { mutate: removeMember, isPending } = useRemoveMemberMutation();
  const { mutate: addMember } = useAddMemberMutation();
  const feedbackSummaryByTeamId = new Map(
    feedbackSummaries.map((summary) => [summary.teamId, summary]),
  );

  const handleTeam = (teamId: string, action: "join" | "leave") => {
    if (action === "join") {
      addMember({ teamId, memberId: session?.user.id ?? "" });
    } else {
      const team = teams.find((t) => t.id === teamId);
      if (!team) return;
      setTeam(team);
      openDialogAfterMenuClose(setIsOpen);
    }
  };

  const handlePinTeam = (teamId: string) => {
    const reorderedTeams = pinSidebarTeam(teams, teamId);

    if (reorderedTeams === teams) return;

    reorderTeams.mutate({
      teamIds: reorderedTeams.map((team) => team.id),
    });
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (!over || active.id === over.id) {
      return;
    }

    const reorderedTeams = reorderVisibleSidebarTeams(
      teams,
      visibleTeams,
      String(active.id),
      String(over.id),
    );

    if (reorderedTeams === teams) return;

    reorderTeams.mutate({
      teamIds: reorderedTeams.map((team) => team.id),
    });
  };

  return (
    <Box className={cn("mt-4", isCollapsed && "mt-3")}>
      {isCollapsed ? (
        <Divider className="mx-auto mb-3 w-4/5 border-t border-dashed" />
      ) : null}
      <Flex
        align="center"
        className={cn("mb-2", isCollapsed && "justify-center")}
        justify="between"
      >
        {isCollapsed ? null : (
          <Text className="pl-2.5 font-medium" color="muted" data-teams-heading>
            Your Teams
          </Text>
        )}
        {userRole !== "guest" || overflowTeams.length > 0 ? (
          <TeamsMenu>
            <TeamsMenu.Trigger>
              <Button
                asIcon={!isCollapsed}
                className={cn(
                  isCollapsed &&
                    "min-h-14 w-full flex-col justify-center gap-1 px-1 py-2 text-center",
                )}
                color="tertiary"
                data-manage-teams-button
                fullWidth={isCollapsed}
                leftIcon={
                  isCollapsed ? (
                    <TeamIcon className="h-5.5 w-auto shrink-0" />
                  ) : (
                    <MoreHorizontalIcon />
                  )
                }
                size={isCollapsed ? undefined : "sm"}
                variant="naked"
              >
                {isCollapsed ? (
                  <span className="line-clamp-2 w-full text-xs leading-3.5 font-semibold">
                    Your Teams
                  </span>
                ) : (
                  <span className="sr-only">Manage Teams</span>
                )}
              </Button>
            </TeamsMenu.Trigger>
            <TeamsMenu.Items
              hideManageTeams={userRole !== "admin"}
              onPinTeam={handlePinTeam}
              overflowTeams={overflowTeams}
              readOnly={userRole === "guest"}
              setTeam={handleTeam}
            />
          </TeamsMenu>
        ) : null}
      </Flex>
      <DndContext onDragEnd={handleDragEnd}>
        <SortableContext
          items={visibleTeamIds}
          strategy={verticalListSortingStrategy}
        >
          <Flex
            className={cn(isCollapsed && "gap-1.5")}
            direction="column"
            gap={1}
          >
            {visibleTeams.map((team) => (
              <Team
                color={team.color}
                feedbackSummary={feedbackSummaryByTeamId.get(team.id)}
                id={team.id}
                isCollapsed={isCollapsed}
                isOpen={expandedTeamId === team.id}
                isPrivate={team.isPrivate}
                key={team.id}
                name={team.name}
                onOpenChange={(open) => {
                  setStoredExpandedTeamId(open ? team.id : null);
                }}
                sortingDisabled={false}
                totalTeams={teams.length}
              />
            ))}
          </Flex>
        </SortableContext>
      </DndContext>

      <ConfirmDialog
        description={
          team?.isPrivate
            ? "Once you leave this team, you will not be able to rejoin later, you will need to be invited again by an admin."
            : "You can rejoin the team later from the sidebar."
        }
        isLoading={isPending}
        isOpen={isOpen}
        loadingText="Leaving team..."
        onClose={() => {
          setIsOpen(false);
        }}
        onConfirm={() => {
          if (team) {
            removeMember({
              teamId: team.id,
              memberId: session?.user.id ?? "",
            });
            setIsOpen(false);
          }
        }}
        title={`Leave ${team?.name} team?`}
      />
    </Box>
  );
};

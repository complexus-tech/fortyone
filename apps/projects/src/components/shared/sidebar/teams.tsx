"use client";
import { Box, Flex, Text, Button } from "ui";
import { MoreHorizontalIcon } from "icons";
import { useSession } from "next-auth/react";
import { useState } from "react";
import { DndContext, type DragEndEvent } from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useUserRole } from "@/hooks";
import { TeamsMenu } from "@/components/ui/teams-menu";
import { useRemoveMemberMutation } from "@/modules/teams/hooks/remove-member-mutation";
import { useAddMemberMutation } from "@/modules/teams/hooks/add-member-mutation";
import { useReorderTeamsMutation } from "@/modules/teams/hooks/use-reorder-teams";
import { ConfirmDialog } from "@/components/ui";
import type { Team as TeamType } from "@/modules/teams/types";
import { Team } from "./team";

// Helper function to reorder array items
const arrayMove = <T,>(array: T[], from: number, to: number): T[] => {
  const newArray = [...array];
  const [removed] = newArray.splice(from, 1);
  newArray.splice(to, 0, removed);
  return newArray;
};

export const Teams = () => {
  const { data: teams = [] } = useTeams();
  const { userRole } = useUserRole();
  const { data: session } = useSession();
  const [team, setTeam] = useState<TeamType | null>(null);
  const [isOpen, setIsOpen] = useState(false);
  const reorderTeams = useReorderTeamsMutation();

  const { mutate: removeMember, isPending } = useRemoveMemberMutation();
  const { mutate: addMember } = useAddMemberMutation();

  const handleTeam = (teamId: string, action: "join" | "leave") => {
    if (action === "join") {
      addMember({ teamId, memberId: session?.user?.id ?? "" });
    } else {
      const team = teams.find((t) => t.id === teamId);
      if (!team) return;
      setTeam(team);
      setIsOpen(true);
    }
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (!over || active.id === over.id) {
      return;
    }

    const oldIndex = teams.findIndex((team) => team.id === active.id);
    const newIndex = teams.findIndex((team) => team.id === over.id);

    if (oldIndex === -1 || newIndex === -1 || oldIndex === newIndex) {
      return;
    }

    // Call the reorder teams action
    reorderTeams.mutate({
      teamIds: arrayMove(teams, oldIndex, newIndex).map((team) => team.id),
    });
  };

  return (
    <Box className="mt-5">
      <Flex align="center" className="mb-3" justify="between">
        <Text className="pl-2.5 font-medium" color="muted" data-teams-heading>
          Your Teams
        </Text>
        {userRole !== "guest" && (
          <TeamsMenu>
            <TeamsMenu.Trigger>
              <Button
                asIcon
                color="tertiary"
                data-manage-teams-button
                leftIcon={<MoreHorizontalIcon />}
                rounded="full"
                size="sm"
                variant="naked"
              >
                <span className="sr-only">Manage Teams</span>
              </Button>
            </TeamsMenu.Trigger>
            <TeamsMenu.Items
              hideManageTeams={userRole !== "admin"}
              setTeam={handleTeam}
            />
          </TeamsMenu>
        )}
      </Flex>
      <DndContext onDragEnd={handleDragEnd}>
        <SortableContext
          items={teams.map((team) => team.id)}
          strategy={verticalListSortingStrategy}
        >
          <Flex direction="column" gap={1}>
            {teams.map((team, idx) => (
              <Team
                {...team}
                idx={idx}
                key={team.id}
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
              memberId: session?.user?.id ?? "",
            });
            setIsOpen(false);
          }
        }}
        title={`Leave ${team?.name} team?`}
      />
    </Box>
  );
};

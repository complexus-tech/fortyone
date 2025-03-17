"use client";
import { Box, Flex, Text, Button } from "ui";
import { MoreHorizontalIcon } from "icons";
import { useSession } from "next-auth/react";
import { useState } from "react";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useUserRole } from "@/hooks";
import { TeamsMenu } from "@/components/ui/teams-menu";
import { useRemoveMemberMutation } from "@/modules/teams/hooks/remove-member-mutation";
import { useAddMemberMutation } from "@/modules/teams/hooks/add-member-mutation";
import { ConfirmDialog } from "@/components/ui";
import type { Team as TeamType } from "@/modules/teams/types";
import { Team } from "./team";

export const Teams = () => {
  const { data: teams = [] } = useTeams();
  const { userRole } = useUserRole();
  const { data: session } = useSession();
  const [team, setTeam] = useState<TeamType | null>(null);
  const [isOpen, setIsOpen] = useState(false);

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
  return (
    <Box className="mt-5">
      <Flex align="center" className="mb-2.5" justify="between">
        <Text className="pl-2.5 font-medium" color="muted">
          Your Teams
        </Text>
        {userRole !== "guest" && (
          <TeamsMenu>
            <TeamsMenu.Trigger>
              <Button
                asIcon
                color="tertiary"
                leftIcon={<MoreHorizontalIcon />}
                rounded="full"
                size="sm"
                variant="naked"
              >
                <span className="sr-only">Manage Teams</span>
              </Button>
            </TeamsMenu.Trigger>
            <TeamsMenu.Items setTeam={handleTeam} />
          </TeamsMenu>
        )}
      </Flex>
      <Flex direction="column" gap={1}>
        {teams.map((team) => (
          <Team {...team} key={team.id} totalTeams={teams.length} />
        ))}
      </Flex>

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

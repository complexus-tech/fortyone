"use client";
import { Box, Flex, Text, Button } from "ui";
import { MoreHorizontalIcon } from "icons";
import { useSession } from "next-auth/react";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useUserRole } from "@/hooks";
import { TeamsMenu } from "@/components/ui/teams-menu";
import { useRemoveMemberMutation } from "@/modules/teams/hooks/remove-member-mutation";
import { useAddMemberMutation } from "@/modules/teams/hooks/add-member-mutation";
import { Team } from "./team";

export const Teams = () => {
  const { data: teams = [] } = useTeams();
  const { userRole } = useUserRole();
  const { data: session } = useSession();
  const { mutate: removeMember } = useRemoveMemberMutation();
  const { mutate: addMember } = useAddMemberMutation();

  const handleTeam = (teamId: string, action: "join" | "leave") => {
    if (action === "join") {
      addMember({ teamId, memberId: session?.user?.id ?? "" });
    } else {
      removeMember({ teamId, memberId: session?.user?.id ?? "" });
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
        {teams.map(({ id, name, color }) => (
          <Team color={color} id={id} key={id} name={name} />
        ))}
      </Flex>
    </Box>
  );
};

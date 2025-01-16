import { Flex, Text, ProgressBar, Box, Badge } from "ui";
import Link from "next/link";
import { ObjectiveIcon } from "icons";
import { format } from "date-fns";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { useMembers } from "@/lib/hooks/members";
import { useTeams } from "@/modules/teams/hooks/teams";
import { TeamColor } from "@/components/ui";
import type { Objective } from "../../../modules/objectives/types";

export const ObjectiveCard = ({
  id,
  name,
  leadUser,
  teamId,
  endDate,
  stats: { completed, total },
  createdAt,
}: Objective) => {
  const { data: members = [] } = useMembers();
  const { data: teams = [] } = useTeams();
  const lead = members.find((member) => member.id === leadUser);
  const team = teams.find((team) => team.id === teamId);
  const progress = Math.round((completed / total) * 100) || 0;

  return (
    <RowWrapper className="px-6 py-3">
      <Box className="flex w-[300px] shrink-0 items-center gap-2">
        <Link
          className="flex w-full items-center gap-2 hover:opacity-90"
          href={`/teams/${teamId}/objectives/${id}`}
        >
          <Flex
            align="center"
            className="size-8 shrink-0 rounded-lg bg-gray-100/50 dark:bg-dark-200"
            justify="center"
          >
            <ObjectiveIcon className="h-4" />
          </Flex>
          <Text className="truncate font-medium">{name}</Text>
        </Link>
      </Box>
      <Flex align="center" gap={4}>
        <Box className="flex w-[120px] shrink-0 items-center gap-2">
          <TeamColor color={team?.color} />
          <Text className="truncate" color="muted">
            {team?.name}
          </Text>
        </Box>

        <Box className="w-[140px] shrink-0">
          <Text className="truncate" color="muted">
            {lead?.username}
          </Text>
        </Box>

        <Box className="w-[120px] shrink-0">
          <ProgressBar className="h-1.5" progress={progress} />
        </Box>

        <Box className="w-[120px] shrink-0">
          <Text color="muted">{format(new Date(endDate), "MMM dd, yyyy")}</Text>
        </Box>

        <Box className="w-[120px] shrink-0">
          <Text color="muted">
            {format(new Date(createdAt), "MMM dd, yyyy")}
          </Text>
        </Box>

        <Box className="w-[100px] shrink-0">
          <Badge color="tertiary">No Health</Badge>
        </Box>
      </Flex>
    </RowWrapper>
  );
};

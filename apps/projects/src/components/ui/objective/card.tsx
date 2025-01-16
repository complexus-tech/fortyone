import { Flex, Text, Avatar, ProgressBar, Box, Badge, Tooltip } from "ui";
import Link from "next/link";
import { CalendarIcon, ObjectiveIcon } from "icons";
import { format } from "date-fns";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { useMembers } from "@/lib/hooks/members";
import type { Objective } from "../../../modules/objectives/types";

type ObjectiveStatus = "completed" | "in progress" | "upcoming";

const statusColors = {
  completed: "tertiary",
  "in progress": "primary",
  upcoming: "tertiary",
} as const;

export const ObjectiveCard = ({
  id,
  name,
  leadUser,
  teamId,
  startDate,
  endDate,
  createdAt,
  stats: { completed, started, total, backlog, cancelled },
}: Objective) => {
  const { data: members = [] } = useMembers();
  const lead = members.find((member) => member.id === leadUser);

  let objectiveStatus: ObjectiveStatus = "completed";
  const startDateObj = new Date(startDate);
  const endDateObj = new Date(endDate);

  if (startDateObj < new Date() && endDateObj > new Date()) {
    objectiveStatus = "in progress";
  } else if (startDateObj > new Date()) {
    objectiveStatus = "upcoming";
  }
  const progress = Math.round((completed / total) * 100);

  return (
    <RowWrapper>
      <Link
        className="flex flex-1 items-center gap-4"
        href={`/teams/${teamId}/objectives/${id}`}
      >
        <Flex
          align="center"
          className="size-10 rounded-lg bg-gray-100/50 dark:bg-dark-200"
          justify="center"
        >
          <ObjectiveIcon />
        </Flex>
        <Box className="space-y-1">
          <Text className="text-[1.05rem] antialiased" fontWeight="semibold">
            {name}
          </Text>
          <Text className="flex items-center gap-1.5" color="muted">
            <CalendarIcon className="h-[1.1rem]" />
            {format(startDateObj, "MMM d")} - {format(endDateObj, "MMM d")}
          </Text>
        </Box>
      </Link>

      <Flex className="items-center" gap={4}>
        <Badge
          className="h-8 px-2 text-base capitalize tracking-wide"
          color={statusColors[objectiveStatus]}
        >
          {objectiveStatus}
        </Badge>

        <Tooltip title={`${progress}% Complete`}>
          <Flex align="center" className="w-36" gap={3}>
            <ProgressBar className="h-2 flex-1" progress={progress} />
            <Text>{progress}%</Text>
          </Flex>
        </Tooltip>

        <Box className="w-40">
          <AssigneesMenu>
            <AssigneesMenu.Trigger>
              <button className="flex items-center gap-1.5" type="button">
                <Avatar name={lead?.fullName} size="xs" src={lead?.avatarUrl} />
                <Text
                  className="relative -top-px max-w-[14ch] truncate"
                  // color="muted"
                  fontWeight="medium"
                >
                  {lead?.username || "Lead"}
                </Text>
              </button>
            </AssigneesMenu.Trigger>
            <AssigneesMenu.Items onAssigneeSelected={(_assigneeId) => {}} />
          </AssigneesMenu>
        </Box>
      </Flex>
    </RowWrapper>
  );
};

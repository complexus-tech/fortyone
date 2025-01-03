import { Flex, Text, Avatar, ProgressBar, Box } from "ui";
import Link from "next/link";
import { CalendarIcon, ObjectiveIcon } from "icons";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { Objective } from "../../../modules/objectives/types";
import { format } from "date-fns";
import { useMembers } from "@/lib/hooks/members";

export const ObjectiveCard = ({
  id,
  name,
  leadUser,
  teamId,
  startDate,
  endDate,
  createdAt,
}: Objective) => {
  const { data: members = [] } = useMembers();
  const lead = members.find((member) => member.id === leadUser);
  return (
    <RowWrapper className="py-5">
      <Link
        className="flex w-[250px] items-center gap-2 truncate hover:opacity-90"
        href={`/teams/${teamId}/objectives/${id}`}
      >
        <ObjectiveIcon className="relative -top-px h-[1.1rem]" />
        {name}
      </Link>
      <Flex align="center" gap={5}>
        <Flex align="center" className="w-40" gap={2}>
          <ProgressBar className="h-1.5 w-24" progress={20} />
          <Text color="muted" fontWeight="medium">
            20%
          </Text>
        </Flex>
        <Text className="flex w-40 items-center gap-1" color="muted">
          <CalendarIcon className="h-[1.1rem] w-auto" strokeWidth={2} />
          {format(new Date(startDate), "MMM dd, yyyy")}
        </Text>
        <Text className="flex w-40 items-center gap-1" color="muted">
          <CalendarIcon className="h-[1.1rem] w-auto" strokeWidth={2} />
          {format(new Date(endDate), "MMM dd, yyyy")}
        </Text>
        <Box className="w-40">
          <AssigneesMenu>
            <AssigneesMenu.Trigger>
              <button className="flex items-center gap-1.5" type="button">
                <Avatar name={lead?.fullName} size="xs" src={lead?.avatarUrl} />
                <Text
                  className="relative -top-px max-w-[14ch] truncate"
                  color="muted"
                  fontWeight="medium"
                >
                  {lead?.username || "Lead"}
                </Text>
              </button>
            </AssigneesMenu.Trigger>
            <AssigneesMenu.Items onAssigneeSelected={(assigneeId) => {}} />
          </AssigneesMenu>
        </Box>
        <Text className="w-40 text-left" color="muted">
          {format(new Date(createdAt), "MMM dd, yyyy")}
        </Text>
      </Flex>
    </RowWrapper>
  );
};

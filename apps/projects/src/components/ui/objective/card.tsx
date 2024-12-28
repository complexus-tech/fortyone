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
    <RowWrapper>
      <Link
        className="w-[250px] truncate hover:opacity-90"
        href={`/teams/${teamId}/objectives/${id}`}
      >
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
        {/* <Box className="w-8">
          <Menu>
            <Menu.Button>
              <Button
                color="tertiary"
                leftIcon={<MoreHorizontalIcon className="h-5 w-auto" />}
                size="sm"
                variant="naked"
              >
                <span className="sr-only">More options</span>
              </Button>
            </Menu.Button>
            <Menu.Items align="end" className="w-64">
              <Menu.Group className="mb-3 mt-1 px-4">
                <Text color="muted" textOverflow="truncate">
                  Manage objective
                </Text>
              </Menu.Group>
              <Menu.Separator className="mb-1.5" />
              <Menu.Group>
                <Menu.Item>
                  <SettingsIcon className="h-5 w-auto" />
                  Settings
                </Menu.Item>
                <Menu.Item>
                  <StoryStatusIcon className="h-[1.2rem] w-auto" />
                  Status
                </Menu.Item>
                <Menu.Item>
                  <Avatar
                    className="h-5 w-auto"
                    color="naked"
                    name="Joseph Mukorivo"
                    size="sm"
                    src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
                  />
                  Lead
                </Menu.Item>
                <Menu.Item>
                  <CalendarIcon className="h-5 w-auto" />
                  Start date
                </Menu.Item>
                <Menu.Item>
                  <CalendarIcon className="h-5 w-auto" />
                  Due date
                </Menu.Item>
                <Menu.Item>
                  <StarIcon className="h-[1.2rem] w-auto" />
                  Favourite
                </Menu.Item>
                <Menu.Item>
                  <DeleteIcon className="h-[1.2rem] w-auto" />
                  Delete
                </Menu.Item>
              </Menu.Group>
            </Menu.Items>
          </Menu>
        </Box> */}
      </Flex>
    </RowWrapper>
  );
};

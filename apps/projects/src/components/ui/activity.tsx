import { useMembers } from "@/lib/hooks/members";
import { StoryActivity } from "@/modules/stories/types";
import { format } from "date-fns";
import { Box, Flex, Text, Avatar, TimeAgo, Tooltip, Button } from "ui";
import Link from "next/link";
import { useStatuses } from "@/lib/hooks/statuses";
import { CalendarIcon } from "icons";

export const Activity = ({
  userId,
  field,
  currentValue,
  type,
  createdAt,
}: StoryActivity) => {
  const { data: members = [] } = useMembers();
  const { data: statuses = [] } = useStatuses();
  const member = members.find((member) => member.id === userId);

  const fieldMap = {
    title: {
      label: "Title",
      render: (value: string) => <span>{value}</span>,
    },
    description: {
      label: "Description",
      render: (value: string) => (
        <span>{value.length > 50 ? `${value.slice(0, 50)}...` : value}</span>
      ),
    },
    status_id: {
      label: "Status",
      render: (value: string) => (
        <span>{statuses.find((status) => status.id === value)?.name}</span>
      ),
    },
    priority: {
      label: "Priority",
      render: (value: string) => <span>{value}</span>,
    },
    assignee_id: {
      label: "Assignee",
      render: (value: string) => (
        <Link
          href={`/profile/${members?.find((member) => member.id === value)?.id}`}
        >
          {members?.find((member) => member.id === value)?.username}
        </Link>
      ),
    },
    start_date: {
      label: "Start date",
      render: (value: string) => (
        <span>{format(new Date(value.replace(/ [A-Z]+$/, "")), "PP")}</span>
      ),
    },
    end_date: {
      label: "Due date",
      render: (value: string) => (
        <span>{format(new Date(value.replace(/ [A-Z]+$/, "")), "PP")}</span>
      ),
    },
    sprint_id: {
      label: "Sprint",
      render: (value: string) => <span>{value}</span>,
    },
    epic_id: {
      label: "Epic",
      render: (value: string) => <span>{value}</span>,
    },
    objective_id: {
      label: "Objective",
      render: (value: string) => <span>{value}</span>,
    },
    blocked_by_id: {
      label: "Blocked by",
      render: (value: string) => <span>{value}</span>,
    },
    blocking_id: {
      label: "Blocking",
      render: (value: string) => <span>{value}</span>,
    },
    related_id: {
      label: "Related to",
      render: (value: string) => <span>{value}</span>,
    },
  } as {
    [key: string]: {
      label: string;
      icon?: React.ReactNode;
      render: (value: string) => React.ReactNode;
    };
  };

  return (
    <Flex align="center" className="z-[1]" gap={1}>
      <Tooltip
        className="py-2.5"
        title={
          member && (
            <Box>
              <Flex gap={2}>
                <Avatar
                  name={member?.fullName}
                  src={member?.avatarUrl}
                  className="mt-0.5"
                />
                <Box>
                  <Link
                    href={`/profile/${member?.id}`}
                    className="mb-2 flex gap-1"
                  >
                    <Text fontWeight="medium" fontSize="md">
                      {member?.fullName}
                    </Text>
                    <Text color="muted" fontSize="md">
                      ({member?.username})
                    </Text>
                  </Link>
                  <Button
                    size="xs"
                    color="tertiary"
                    className="mb-0.5 ml-px px-2"
                    href={`/profile/${member?.id}`}
                  >
                    Go to profile
                  </Button>
                </Box>
              </Flex>
            </Box>
          )
        }
      >
        <Flex gap={1} className="cursor-pointer">
          <Box className="relative top-[1px] flex aspect-square items-center rounded-full bg-white p-[0.3rem] dark:bg-dark-300">
            <Avatar name={member?.fullName} size="xs" src={member?.avatarUrl} />
          </Box>
          <Text
            className="relative top-0.5 ml-1 text-black dark:text-white"
            fontWeight="medium"
          >
            {member?.username}
          </Text>
        </Flex>
      </Tooltip>
      {(type === "update" || type === "create") && (
        <Text className="text-[0.95rem]" color="muted">
          {type === "create" ? "created the story" : "changed"}
        </Text>
      )}

      {type === "update" && (
        <>
          <Text
            className="text-[0.95rem] text-black dark:text-white"
            fontWeight="medium"
          >
            {fieldMap[field]?.label}
          </Text>
          <Text className="text-[0.95rem]" color="muted">
            to
          </Text>
          <Text
            className="text-[0.95rem] text-black dark:text-white"
            as="span"
            fontWeight="medium"
          >
            {fieldMap[field]?.render(currentValue)}
          </Text>
        </>
      )}
      {(type === "update" || type === "create") && (
        <>
          <Text className="text-[0.95rem]" color="muted">
            ·
          </Text>
          <Text className="text-[0.95rem]" color="muted">
            <TimeAgo timestamp={createdAt} />
          </Text>
        </>
      )}
    </Flex>
  );
};

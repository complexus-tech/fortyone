import { useMembers } from "@/lib/hooks/members";
import { StoryActivity, StoryPriority } from "@/modules/stories/types";
import { format } from "date-fns";
import { Box, Flex, Text, Avatar, TimeAgo, Tooltip, Button } from "ui";
import Link from "next/link";
import { useStatuses } from "@/lib/hooks/statuses";
import { cn } from "lib";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useSprints } from "@/lib/hooks/sprints";
import { CalendarIcon, ObjectiveIcon, SprintsIcon } from "icons";
import { PriorityIcon } from "./priority-icon";
import { StoryStatusIcon } from "./story-status-icon";

export const Activity = ({
  userId,
  parentId,
  field,
  currentValue,
  type,
  createdAt,
  children,
}: StoryActivity) => {
  const { data: members = [] } = useMembers();
  const { data: statuses = [] } = useStatuses();
  const { data: objectives = [] } = useObjectives();
  const { data: sprints = [] } = useSprints();
  const member = members.find((member) => member.id === userId);

  const Comment = ({
    parentId,
    currentValue,
  }: {
    parentId: string | null;
    currentValue: string;
  }) => (
    <Box
      className={cn(
        "prose prose-stone ml-9 mt-1 max-w-full rounded-lg rounded-br-none bg-gray-50/60 p-4 text-[0.95rem] leading-6 dark:prose-invert prose-headings:font-medium prose-a:text-primary prose-pre:bg-gray-50 prose-pre:text-[1.1rem] prose-pre:text-dark-200 dark:bg-dark-200/70 dark:prose-pre:bg-dark-200/80 dark:prose-pre:text-gray-200",
        {
          "ml-12": parentId,
        },
      )}
      html={currentValue}
    />
  );

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
        <span className="flex items-center gap-1">
          <StoryStatusIcon className="h-4" statusId={value} />
          {statuses.find((status) => status.id === value)?.name}
        </span>
      ),
    },
    priority: {
      label: "Priority",
      render: (value: string) => (
        <span className="flex items-center gap-1">
          <PriorityIcon className="h-4" priority={value as StoryPriority} />
          {value}
        </span>
      ),
    },
    assignee_id: {
      label: "Assignee",
      render: (value: string) => (
        <>
          {!value || value?.includes("nil") ? (
            <span>Unassigned</span>
          ) : (
            <Link
              className="flex items-center gap-1.5"
              href={`/profile/${members?.find((member) => member.id === value)?.id}`}
            >
              <Avatar
                className="relative top-px"
                name={members?.find((member) => member.id === value)?.fullName}
                src={members?.find((member) => member.id === value)?.avatarUrl}
                size="xs"
              />
              {members?.find((member) => member.id === value)?.username}
            </Link>
          )}
        </>
      ),
    },
    start_date: {
      label: "Start date",
      render: (value: string) => (
        <span className="flex items-center gap-1">
          <CalendarIcon className="h-4" />
          {format(new Date(value.replace(/ [A-Z]+$/, "")), "PP")}
        </span>
      ),
    },
    end_date: {
      label: "Due date",
      render: (value: string) => (
        <span className="flex items-center gap-1">
          <CalendarIcon className="h-4" />
          {format(new Date(value.replace(/ [A-Z]+$/, "")), "PP")}
        </span>
      ),
    },
    sprint_id: {
      label: "Sprint",
      render: (value: string) => (
        <>
          {!value || value?.includes("nil") ? (
            <span>No sprint</span>
          ) : (
            <Link
              className="flex items-center gap-1"
              href={`/teams/${sprints?.find((sprint) => sprint.id === value)?.teamId}/sprints/${sprints?.find((sprint) => sprint.id === value)?.id}/stories`}
            >
              <SprintsIcon className="h-4" />
              {sprints?.find((sprint) => sprint.id === value)?.name}
            </Link>
          )}
        </>
      ),
    },
    epic_id: {
      label: "Epic",
      render: (value: string) => <span>{value}</span>,
    },
    objective_id: {
      label: "Objective",
      render: (value: string) => (
        <>
          {!value || value?.includes("nil") ? (
            <span>No objective</span>
          ) : (
            <Link
              className="flex items-center gap-1"
              href={`/teams/${objectives?.find((objective) => objective.id === value)?.teamId}/objectives/${objectives?.find((objective) => objective.id === value)?.id}`}
            >
              <ObjectiveIcon className="h-4" />
              {objectives?.find((objective) => objective.id === value)?.name}
            </Link>
          )}
        </>
      ),
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
    <Box className="relative pb-4">
      <Box
        className={cn(
          "pointer-events-none absolute left-4 top-0 z-0 h-full border-l border-dashed border-gray-200 dark:border-dark-50",
        )}
      />
      <Flex align="center" className="z[1]" gap={1}>
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
              <Avatar
                name={member?.fullName}
                size="xs"
                src={member?.avatarUrl}
              />
            </Box>
            <Text
              className="relative top-0.5 ml-1 text-black dark:text-white"
              fontWeight="medium"
            >
              {member?.username}
            </Text>
          </Flex>
        </Tooltip>
        <Text className="text-[0.95rem]" color="muted">
          {type === "create"
            ? "created the story"
            : type === "update"
              ? "changed"
              : "commented"}
        </Text>

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
        <Text className="text-[0.95rem]" color="muted">
          ·
        </Text>
        <Text className="text-[0.95rem]" color="muted">
          <TimeAgo timestamp={createdAt} />
        </Text>
      </Flex>
      {type === "comment" && (
        <>
          <Comment parentId={parentId} currentValue={currentValue} />
          {children &&
            children?.length > 0 &&
            children.map((child) => <Comment key={child.id} {...child} />)}
        </>
      )}
    </Box>
  );
};

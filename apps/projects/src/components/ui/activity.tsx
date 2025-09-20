import type { ReactNode } from "react";
import { format } from "date-fns";
import { Box, Flex, Text, Avatar, TimeAgo, Tooltip, Button } from "ui";
import Link from "next/link";
import { cn } from "lib";
import { CalendarIcon, ObjectiveIcon, SprintsIcon } from "icons";
import { useMembers } from "@/lib/hooks/members";
import type { StoryActivity, StoryPriority } from "@/modules/stories/types";
import { useStatuses } from "@/lib/hooks/statuses";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useSprints } from "@/modules/sprints/hooks/sprints";
import { PriorityIcon } from "./priority-icon";
import { StoryStatusIcon } from "./story-status-icon";

export const Activity = ({
  userId,
  field,
  currentValue,
  type,
  createdAt,
}: StoryActivity) => {
  const { data: members = [] } = useMembers();
  const { data: statuses = [] } = useStatuses();
  const { data: objectives = [] } = useObjectives();
  const { data: sprints = [] } = useSprints();
  const member = members.find((m) => m.id === userId);

  if (field === "completed_at") {
    return null;
  }

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
          <StoryStatusIcon className="h-5" statusId={value} />
          {statuses.find((status) => status.id === value)?.name}
        </span>
      ),
    },
    priority: {
      label: "Priority",
      render: (value: string) => (
        <span className="flex items-center gap-1">
          <PriorityIcon className="h-5" priority={value as StoryPriority} />
          {value}
        </span>
      ),
    },
    assignee_id: {
      label: "Assignee",
      render: (value: string) => (
        <>
          {!value || value.includes("nil") ? (
            <span>Unassigned</span>
          ) : (
            <Link
              className="flex items-center gap-1.5 pb-0.5"
              href={`/profile/${members.find((m) => m.id === value)?.id}`}
            >
              <Avatar
                className="relative top-px"
                name={members.find((m) => m.id === value)?.fullName}
                size="xs"
                src={members.find((m) => m.id === value)?.avatarUrl}
              />
              {members.find((m) => m.id === value)?.username}
            </Link>
          )}
        </>
      ),
    },
    start_date: {
      label: "Start date",
      render: (value: string) => (
        <span className="flex items-center gap-1">
          <CalendarIcon className="h-[1.15rem]" />
          {value
            ? format(new Date(value.split(" ")[0]), "PP")
            : "No start date"}
        </span>
      ),
    },
    end_date: {
      label: "Deadline",
      render: (value: string) => (
        <span className="flex items-center gap-1">
          <CalendarIcon className="h-[1.15rem]" />
          {value ? format(new Date(value.split(" ")[0]), "PP") : "No deadline"}
        </span>
      ),
    },
    sprint_id: {
      label: "Sprint",
      render: (value: string) => (
        <>
          {!value || value.includes("nil") ? (
            <span>No sprint</span>
          ) : (
            <Link
              className="flex items-center gap-1"
              href={`/teams/${sprints.find((sprint) => sprint.id === value)?.teamId}/sprints/${sprints.find((sprint) => sprint.id === value)?.id}/stories`}
            >
              <SprintsIcon className="h-5" />
              {sprints.find((sprint) => sprint.id === value)?.name}
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
          {!value || value.includes("nil") ? (
            <span>No objective</span>
          ) : (
            <Link
              className="flex items-center gap-1"
              href={`/teams/${objectives.find((objective) => objective.id === value)?.teamId}/objectives/${objectives.find((objective) => objective.id === value)?.id}`}
            >
              <ObjectiveIcon className="h-5" />
              {objectives.find((objective) => objective.id === value)?.name}
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
  } as Record<
    string,
    {
      label: string;
      icon?: ReactNode;
      render: (value: string) => ReactNode;
    }
  >;

  return (
    <Box className="relative pb-2 last-of-type:pb-0 md:pb-4">
      <Box
        className={cn(
          "pointer-events-none absolute left-4 top-0 z-0 h-full border-l border-dashed border-gray-200 dark:border-dark-50",
        )}
      />
      <Flex align="center" className="z-[1]" gap={1}>
        <Tooltip
          className="py-2.5"
          title={
            member ? (
              <Box>
                <Flex gap={2}>
                  <Avatar
                    className="mt-0.5"
                    name={member.fullName}
                    src={member.avatarUrl}
                  />
                  <Box>
                    <Link
                      className={cn("mb-2 flex gap-1", {
                        "mb-0": member.role === "system",
                      })}
                      href={
                        member.role === "system" ? "" : `/profile/${member.id}`
                      }
                    >
                      <Text fontSize="md" fontWeight="medium">
                        {member.fullName}
                      </Text>
                      <Text color="muted" fontSize="md">
                        ({member.username})
                      </Text>
                    </Link>
                    {member.role !== "system" ? (
                      <Button
                        className="mb-0.5 ml-px px-2"
                        color="tertiary"
                        href={`/profile/${member.id}`}
                        size="xs"
                      >
                        Go to profile
                      </Button>
                    ) : (
                      <Text color="muted" fontSize="md">
                        (System Account)
                      </Text>
                    )}
                  </Box>
                </Flex>
              </Box>
            ) : null
          }
        >
          <Flex align="center" className="cursor-pointer" gap={1}>
            <Box className="relative left-px flex aspect-square items-center rounded-full bg-white p-[0.3rem] dark:bg-dark-300">
              <Avatar
                name={member?.fullName}
                size="xs"
                src={member?.avatarUrl}
              />
            </Box>
            <Text
              className="relative ml-1 text-sm text-black dark:text-white md:text-[0.95rem]"
              fontWeight="medium"
            >
              {member?.username}
            </Text>
          </Flex>
        </Tooltip>
        <Box className="line-clamp-1 flex items-center gap-1 text-sm md:text-[0.95rem]">
          <Text as="span" className="text-sm md:text-[0.95rem]" color="muted">
            {type === "create" ? "created the story" : "changed"}
          </Text>
          {type === "update" && (
            <>
              <Text
                as="span"
                className="shrink-0 text-sm text-black dark:text-white md:text-[0.95rem]"
                fontWeight="medium"
              >
                {fieldMap[field].label}
              </Text>
              {currentValue ? (
                <>
                  <Text
                    as="span"
                    className="text-sm md:text-[0.95rem]"
                    color="muted"
                  >
                    to
                  </Text>
                  <Text
                    as="span"
                    className="inline-block shrink-0 text-sm text-black dark:text-white md:text-[0.95rem]"
                    fontWeight="medium"
                  >
                    {fieldMap[field].render(currentValue)}
                  </Text>
                </>
              ) : null}
            </>
          )}
          <Text
            as="span"
            className="mx-0.5 text-sm md:text-[0.95rem]"
            color="muted"
          >
            ·
          </Text>
          <Text
            as="span"
            className="shrink-0 text-sm md:text-[0.95rem]"
            color="muted"
          >
            <TimeAgo timestamp={createdAt} />
          </Text>
        </Box>
      </Flex>
    </Box>
  );
};

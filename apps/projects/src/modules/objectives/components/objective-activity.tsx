import type { ReactNode } from "react";
import { format } from "date-fns";
import { Box, Flex, Text, Avatar, Tooltip, Button, TimeAgo } from "ui";
import Link from "next/link";
import { cn } from "lib";
import { CalendarIcon } from "icons";
import { useMembers } from "@/lib/hooks/members";
import { PriorityIcon } from "@/components/ui/priority-icon";
import { ObjectiveStatusIcon } from "@/components/ui/objective-status-icon";
import type { StoryPriority } from "@/modules/stories/types";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { useKeyResults } from "@/modules/objectives/hooks";
import { ObjectiveHealthIcon } from "@/components/ui";
import type { ObjectiveActivity, ObjectiveHealth } from "../types";

export const ObjectiveActivityComponent = ({
  userId,
  field,
  currentValue,
  type,
  updateType,
  comment,
  createdAt,
  keyResultId,
  objectiveId,
}: ObjectiveActivity) => {
  const { data: members = [] } = useMembers();
  const { data: statuses = [] } = useObjectiveStatuses();
  const member = members.find((m) => m.id === userId);
  const { data: keyResults = [] } = useKeyResults(objectiveId);
  const keyResult = keyResults.find((kr) => kr.id === keyResultId);

  if (field === "completed_at") {
    return null;
  }

  const objectiveFieldMap = {
    name: {
      label: "Name",
      render: (value: string) => <span>{value}</span>,
    },
    start_date: {
      label: "Start date",
      render: (value: string) => (
        <span className="flex items-center gap-1">
          <CalendarIcon className="h-4" />
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
          <CalendarIcon className="h-4" />
          {value ? format(new Date(value.split(" ")[0]), "PP") : "No deadline"}
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
    health: {
      label: "Health",
      render: (value: string) => {
        const health = value as ObjectiveHealth;
        return (
          <span className="flex items-center gap-1">
            <ObjectiveHealthIcon health={health} />
            {value}
          </span>
        );
      },
    },
    status_id: {
      label: "Status",
      render: (value: string) => (
        <span className="flex items-center gap-1">
          <ObjectiveStatusIcon className="h-4" statusId={value} />
          {statuses.find((status) => status.id === value)?.name}
        </span>
      ),
    },

    lead_user_id: {
      label: "Lead",
      render: (value: string) => (
        <>
          {!value || value.includes("nil") ? (
            <span>No lead</span>
          ) : (
            <Link
              className="flex items-center gap-1.5"
              href={`/profile/${members.find((m) => m.id === value)?.id}`}
            >
              <Avatar
                name={members.find((m) => m.id === value)?.fullName}
                size="xs"
                src={members.find((m) => m.id === value)?.avatarUrl}
              />
              {members.find((m) => m.id === value)?.username || "deleted user"}
            </Link>
          )}
        </>
      ),
    },
  } as Record<
    string,
    {
      label: string;
      icon?: ReactNode;
      render: (value: string) => ReactNode;
    }
  >;

  const keyResultFieldMap = {
    name: {
      label: "Name",
      render: (value: string) => <span>{value}</span>,
    },
    current_value: {
      label: "Current value",
      render: (value: string) => <span>{value}</span>,
    },
    start_value: {
      label: "Start value",
      render: (value: string) => <span>{value}</span>,
    },
    target_value: {
      label: "Target value",
      render: (value: string) => <span>{value}</span>,
    },
    start_date: {
      label: "Start date",
      render: (value: string) => (
        <span className="flex items-center gap-1">
          <CalendarIcon className="h-4" />
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
          <CalendarIcon className="h-4" />
          {value ? format(new Date(value.split(" ")[0]), "PP") : "No end date"}
        </span>
      ),
    },
    lead: {
      label: "Lead",
      render: (value: string) => (
        <>
          {!value || value.includes("nil") ? (
            <span>No lead</span>
          ) : (
            <Link
              className="flex items-center gap-1.5"
              href={`/profile/${members.find((m) => m.id === value)?.id}`}
            >
              <Avatar
                className="relative"
                name={members.find((m) => m.id === value)?.fullName}
                size="xs"
                src={members.find((m) => m.id === value)?.avatarUrl}
              />
              {members.find((m) => m.id === value)?.username || "deleted user"}
            </Link>
          )}
        </>
      ),
    },
    contributors: {
      label: "Contributors",
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

  const fieldMap =
    updateType === "objective" ? objectiveFieldMap : keyResultFieldMap;
  const entityType = updateType === "objective" ? "objective" : "key result";

  return (
    <Box className="relative pb-2 last-of-type:pb-0 md:pb-3.5">
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
                      <Text fontSize="md">{member.fullName}</Text>
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
            <Text className="relative ml-1 text-sm text-black dark:text-white md:text-[0.95rem]">
              {member?.username}
            </Text>
          </Flex>
        </Tooltip>
        <Box className="line-clamp-1 flex items-center gap-1 text-sm md:text-[0.95rem]">
          <Text as="span" className="text-sm md:text-[0.95rem]" color="muted">
            {type === "create" ? `created the ${entityType}` : "changed"}
          </Text>
          {type === "update" && (
            <>
              <Text
                as="span"
                className="flex shrink-0 items-center gap-1 text-sm md:text-[0.95rem]"
                color="muted"
              >
                {entityType}
              </Text>
              {entityType === "key result" && keyResultId ? (
                <Text
                  as="span"
                  className="max-w-[16ch] shrink-0 truncate text-sm italic md:text-[0.95rem]"
                  color="muted"
                  title={keyResult?.name}
                >
                  {keyResult?.name}
                </Text>
              ) : null}
              <Text
                as="span"
                className="shrink-0 text-sm text-black dark:text-white md:text-[0.95rem]"
              >
                {fieldMap[field].label || field}
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
                    className="inline-block max-w-[24ch] shrink-0 truncate text-sm text-black dark:text-white md:text-[0.95rem]"
                    title={currentValue}
                  >
                    {fieldMap[field].render(currentValue) || currentValue}
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
      {comment ? (
        <Flex align="start" className="ml-9 mt-2 gap-2">
          <Avatar
            className="mt-0.5"
            name={member?.fullName}
            size="xs"
            src={member?.avatarUrl}
          />
          <Box className="max-w-lg rounded-xl rounded-tl-md border border-gray-100/60 bg-gray-50/60 px-4 py-2 dark:border-dark-100/80 dark:bg-dark-300/80">
            <Text color="muted">{comment}</Text>
          </Box>
        </Flex>
      ) : null}
    </Box>
  );
};

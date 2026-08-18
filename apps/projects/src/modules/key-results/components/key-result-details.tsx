"use client";

import type { ReactNode } from "react";
import { useMemo, useState } from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { format, formatISO } from "date-fns";
import {
  CalendarIcon,
  CloseIcon,
  EditIcon,
  ExternalLinkIcon,
  ObjectiveIcon,
  OKRIcon,
  UserIcon,
} from "icons";
import {
  Line,
  LineChart,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip as ChartTooltip,
  XAxis,
  YAxis,
} from "recharts";
import {
  Avatar,
  Box,
  Button,
  DatePicker,
  Divider,
  Flex,
  ProgressBar,
  Text,
} from "ui";
import { AssigneesMenu } from "@/components/ui";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { useTerminology, useUserRole, useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { useMembers } from "@/lib/hooks/members";
import { getKeyResultActivities } from "@/modules/objectives/queries/get-key-result-activities";
import { useKeyResults } from "@/modules/objectives/hooks/use-key-results";
import { useUpdateKeyResultMutation } from "@/modules/objectives/hooks/use-update-key-result-mutation";
import type { KeyResult, Objective } from "@/modules/objectives/types";
import { UpdateKeyResultDialog } from "@/modules/objectives/stories/overview/update-key-result-dialog";
import { useKeyResultStories } from "@/modules/stories/hooks/key-result-stories";
import { getStoryPath } from "@/modules/story/utils/story-url";
import { formatKeyResultValue, getKeyResultProgress } from "../utils";

type KeyResultProgressPoint = {
  date: string;
  value: number;
};

const createProgressData = (
  keyResult: KeyResult,
  activities: Awaited<ReturnType<typeof getKeyResultActivities>>["activities"],
) => {
  const points: KeyResultProgressPoint[] = [
    { date: keyResult.startDate, value: keyResult.startValue },
  ];

  activities
    .filter(
      ({ field, updateType }) =>
        updateType === "key_result" && field === "current_value",
    )
    .toSorted(
      (first, second) =>
        new Date(first.createdAt).getTime() -
        new Date(second.createdAt).getTime(),
    )
    .forEach((activity) => {
      const value = Number(activity.currentValue);
      if (Number.isFinite(value)) {
        points.push({ date: activity.createdAt, value });
      }
    });

  const lastPoint = points.at(-1);
  if (!lastPoint || lastPoint.value !== keyResult.currentValue) {
    points.push({ date: keyResult.updatedAt, value: keyResult.currentValue });
  }

  return points;
};

const PropertyRow = ({
  children,
  icon,
  label,
}: {
  children: ReactNode;
  icon: ReactNode;
  label: string;
}) => (
  <Flex align="center" className="min-h-10 gap-4" justify="between">
    <Flex align="center" className="text-text-muted w-36 shrink-0 gap-2">
      {icon}
      <Text color="muted">{label}</Text>
    </Flex>
    <Box className="min-w-0 flex-1">{children}</Box>
  </Flex>
);

export const KeyResultDetails = ({
  initialKeyResult,
  objective,
  onClose,
}: {
  initialKeyResult: KeyResult;
  objective: Objective;
  onClose: () => void;
}) => {
  const { data: session } = useSession();
  const { userRole } = useUserRole();
  const { withWorkspace, workspaceSlug } = useWorkspacePath();
  const { getTermDisplay } = useTerminology();
  const { data: members = [] } = useMembers();
  const { data: keyResults = [] } = useKeyResults(objective.id);
  const keyResult =
    keyResults.find(({ id }) => id === initialKeyResult.id) ?? initialKeyResult;
  const { mutate: updateKeyResult } = useUpdateKeyResultMutation();
  const { data: linkedStories = [] } = useKeyResultStories(keyResult.id, true);
  const [updateMode, setUpdateMode] = useState<"other" | "progress">("other");
  const [isUpdateOpen, setIsUpdateOpen] = useState(false);
  const { data: activityData } = useQuery({
    queryKey: ["key-result-activities", workspaceSlug, keyResult.id],
    queryFn: () =>
      getKeyResultActivities(keyResult.id, 1, 100, {
        session: session!,
        workspaceSlug,
      }),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  const progress = getKeyResultProgress(keyResult);
  const lead = keyResult.lead
    ? members.find(({ id }) => id === keyResult.lead)
    : undefined;
  const progressData = useMemo(
    () => createProgressData(keyResult, activityData?.activities ?? []),
    [activityData?.activities, keyResult],
  );
  const canEdit = userRole !== "guest";

  const openEditor = (mode: "other" | "progress") => {
    setUpdateMode(mode);
    setIsUpdateOpen(true);
  };

  const updateDate = (data: { endDate?: string; startDate?: string }) => {
    updateKeyResult({
      data,
      keyResultId: keyResult.id,
      objectiveId: objective.id,
      silent: true,
    });
  };

  return (
    <>
      <Box className="border-border/70 dark:border-border dark:bg-surface absolute top-14 right-3 bottom-4 isolate z-40 w-[calc(100%-1.5rem)] overflow-y-auto rounded-xl border bg-white shadow-xl md:top-[1.625rem] md:right-6 md:bottom-[4.875rem] md:w-[34rem]">
        <Flex
          align="center"
          className="border-border/70 dark:border-border dark:bg-surface/80 sticky top-0 z-10 min-h-16 gap-3 border-b-[0.5px] bg-white/80 px-6 backdrop-blur-2xl"
          justify="between"
        >
          <Flex align="center" className="min-w-0 flex-1 gap-3">
            <OKRIcon className="text-text-muted h-5 shrink-0" strokeWidth={2} />
            <Text className="truncate" fontSize="lg" fontWeight="semibold">
              {keyResult.name}
            </Text>
          </Flex>
          <Flex align="center" gap={1}>
            <Button
              color="tertiary"
              disabled={!canEdit}
              leftIcon={<EditIcon className="h-4" />}
              onClick={() => {
                openEditor("other");
              }}
              size="sm"
              variant="naked"
            >
              Edit
            </Button>
            <Button
              aria-label={`Close ${getTermDisplay("keyResultTerm")} details`}
              className="-mr-2"
              color="tertiary"
              leftIcon={<CloseIcon className="h-4" strokeWidth={3} />}
              onClick={onClose}
              size="sm"
              variant="naked"
            />
          </Flex>
        </Flex>

        <Box className="px-6 pt-5 pb-24">
          <Box>
            <Flex align="center" justify="between">
              <Box>
                <Text color="muted">Progress</Text>
                <Text
                  className="mt-0.5 text-2xl tabular-nums"
                  fontWeight="semibold"
                >
                  {progress}%
                </Text>
              </Box>
              <Button
                disabled={!canEdit}
                onClick={() => {
                  openEditor("progress");
                }}
                size="sm"
              >
                Update progress
              </Button>
            </Flex>
            <ProgressBar className="mt-4" progress={progress} />
            <Flex align="center" className="mt-2" justify="between">
              <Text className="text-[0.95rem]" color="muted">
                {formatKeyResultValue(
                  keyResult.startValue,
                  keyResult.measurementType,
                )}
              </Text>
              <Text className="text-[0.95rem]" color="muted">
                Target{" "}
                {formatKeyResultValue(
                  keyResult.targetValue,
                  keyResult.measurementType,
                )}
              </Text>
            </Flex>
          </Box>

          <Text className="mt-6 mb-3">Progress history</Text>
          <Box className="h-56 pt-2">
            <ResponsiveContainer height="100%" width="100%">
              <LineChart data={progressData} margin={{ left: -20, right: 12 }}>
                <XAxis
                  axisLine={false}
                  dataKey="date"
                  tickFormatter={(value: string) =>
                    format(new Date(value), "MMM d")
                  }
                  tickLine={false}
                />
                <YAxis axisLine={false} tickLine={false} />
                <ChartTooltip
                  contentStyle={{
                    background: "var(--color-surface-elevated)",
                    border: "1px solid var(--color-border-strong)",
                    borderRadius: "0.75rem",
                  }}
                  formatter={(value) => [
                    formatKeyResultValue(
                      Number(value),
                      keyResult.measurementType,
                    ),
                    "Progress",
                  ]}
                  labelFormatter={(value) =>
                    format(new Date(String(value)), "MMM d, yyyy")
                  }
                />
                <ReferenceLine
                  stroke="var(--color-text-muted)"
                  strokeDasharray="4 4"
                  y={keyResult.targetValue}
                />
                <Line
                  dataKey="value"
                  dot={{ r: 3 }}
                  stroke="var(--color-info)"
                  strokeWidth={2}
                  type="monotone"
                />
              </LineChart>
            </ResponsiveContainer>
          </Box>

          <Divider className="my-6" />

          <Flex direction="column" gap={2}>
            <PropertyRow
              icon={<ObjectiveIcon className="h-4 w-4" />}
              label="Objective"
            >
              <Text className="truncate" fontWeight="medium">
                {objective.name}
              </Text>
            </PropertyRow>
            <PropertyRow
              icon={<OKRIcon className="h-4 w-4" strokeWidth={2} />}
              label="Measurement"
            >
              <Text className="capitalize" fontWeight="medium">
                {keyResult.measurementType}
              </Text>
            </PropertyRow>
            <PropertyRow icon={<UserIcon className="h-4 w-4" />} label="Lead">
              <AssigneesMenu>
                <AssigneesMenu.Trigger>
                  <Button
                    className="justify-start px-2"
                    color="tertiary"
                    disabled={!canEdit}
                    leftIcon={
                      <Avatar
                        name={lead?.fullName || lead?.username}
                        size="xs"
                        src={lead?.avatarUrl}
                      />
                    }
                    size="sm"
                    variant="naked"
                  >
                    {lead?.fullName || lead?.username || "Assign lead"}
                  </Button>
                </AssigneesMenu.Trigger>
                <AssigneesMenu.Items
                  assigneeId={keyResult.lead}
                  onAssigneeSelected={(leadUser) => {
                    updateKeyResult({
                      data: { lead: leadUser },
                      keyResultId: keyResult.id,
                      objectiveId: objective.id,
                      silent: true,
                    });
                  }}
                  placeholder="Assign lead..."
                  teamId={objective.teamId}
                />
              </AssigneesMenu>
            </PropertyRow>
            <PropertyRow
              icon={<CalendarIcon className="h-4 w-4" />}
              label="Start date"
            >
              <DatePicker>
                <DatePicker.Trigger>
                  <Button
                    color="tertiary"
                    disabled={!canEdit}
                    size="sm"
                    variant="naked"
                  >
                    {format(new Date(keyResult.startDate), "MMM d, yyyy")}
                  </Button>
                </DatePicker.Trigger>
                <DatePicker.Calendar
                  onDayClick={(day) => {
                    updateDate({
                      startDate: formatISO(day, { representation: "date" }),
                    });
                  }}
                  selected={new Date(keyResult.startDate)}
                />
              </DatePicker>
            </PropertyRow>
            <PropertyRow
              icon={<CalendarIcon className="h-4 w-4" />}
              label="Target date"
            >
              <DatePicker>
                <DatePicker.Trigger>
                  <Button
                    color="tertiary"
                    disabled={!canEdit}
                    size="sm"
                    variant="naked"
                  >
                    {format(new Date(keyResult.endDate), "MMM d, yyyy")}
                  </Button>
                </DatePicker.Trigger>
                <DatePicker.Calendar
                  onDayClick={(day) => {
                    updateDate({
                      endDate: formatISO(day, { representation: "date" }),
                    });
                  }}
                  selected={new Date(keyResult.endDate)}
                />
              </DatePicker>
            </PropertyRow>
          </Flex>

          <Divider className="my-6" />

          <Flex align="center" className="mb-2" justify="between">
            <Text>
              Linked {getTermDisplay("storyTerm", { variant: "plural" })}
            </Text>
            <Text color="muted">{linkedStories.length}</Text>
          </Flex>
          {linkedStories.length > 0 ? (
            <Flex direction="column">
              {linkedStories.map((story) => (
                <Link
                  className="border-border hover:bg-state-hover flex min-h-11 items-center justify-between gap-3 border-t first:border-t-0"
                  href={withWorkspace(
                    getStoryPath({
                      id: story.id,
                      sequenceId: story.sequenceId,
                      teamCode: story.team?.code,
                    }),
                  )}
                  key={story.id}
                >
                  <Text className="truncate">{story.title}</Text>
                  <Flex align="center" className="shrink-0 gap-2">
                    <Text className="text-[0.9rem]" color="muted">
                      {story.team?.code
                        ? `${story.team.code}-${story.sequenceId}`
                        : story.sequenceId}
                    </Text>
                    <ExternalLinkIcon className="text-text-muted h-4 w-4" />
                  </Flex>
                </Link>
              ))}
            </Flex>
          ) : (
            <Text color="muted">
              No {getTermDisplay("storyTerm", { variant: "plural" })} are linked
              yet.
            </Text>
          )}
        </Box>
      </Box>

      <UpdateKeyResultDialog
        isOpen={isUpdateOpen}
        keyResult={keyResult}
        onOpenChange={setIsUpdateOpen}
        updateMode={updateMode}
      />
    </>
  );
};

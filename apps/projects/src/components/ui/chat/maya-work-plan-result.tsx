"use client";

import { format, isSameDay } from "date-fns";
import {
  CalendarIcon,
  TimeScheduleIcon,
  UsersAddIcon,
  WarningIcon,
} from "icons";
import { cn } from "lib";
import { Box, Flex, Text } from "ui";
import { useMayaAssignee, useMembers } from "@/lib/hooks/members";
import type { MayaWorkPlanActionModel } from "./maya-work-plan-data";
import { getMayaWorkPlanModel } from "./maya-work-plan-data";

const statusLabel = (status: string) =>
  status
    .replaceAll("_", " ")
    .replace(/^./, (character) => character.toUpperCase());

const statusClasses = (status: string) => {
  switch (status) {
    case "applied":
    case "completed":
      return "bg-success";
    case "failed":
      return "bg-danger";
    default:
      return "bg-text-muted";
  }
};

const ActionIcon = ({ type }: { type: string }) => {
  if (type === "assign_story") {
    return <UsersAddIcon aria-hidden className="h-4.5" />;
  }
  if (type === "schedule_work_block") {
    return <CalendarIcon aria-hidden className="h-4.5" />;
  }
  if (type === "flag_schedule_risk") {
    return <WarningIcon aria-hidden className="text-warning h-4.5" />;
  }
  return <TimeScheduleIcon aria-hidden className="h-4.5" />;
};

const formatScheduledWindow = (action: MayaWorkPlanActionModel) => {
  if (!action.startAt || !action.endAt) return "";
  const start = new Date(action.startAt);
  const end = new Date(action.endAt);
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) return "";

  const endPattern = isSameDay(start, end) ? "h:mm a" : "MMM d, h:mm a";
  return `${format(start, "MMM d, h:mm a")} – ${format(end, endPattern)}`;
};

const getActionContext = ({
  action,
  ownerName,
  scheduledWindow,
}: {
  action: MayaWorkPlanActionModel;
  ownerName?: string;
  scheduledWindow: string;
}) => {
  if (action.type === "assign_story") {
    return ownerName || "Owner selected by Maya";
  }

  if (action.type === "schedule_work_block") {
    return (
      [action.title, scheduledWindow, ownerName].filter(Boolean).join(" · ") ||
      "Calendar time selected by Maya"
    );
  }

  return "";
};

export const MayaWorkPlanResult = ({ output }: { output: unknown }) => {
  const model = getMayaWorkPlanModel(output);
  const { data: members = [] } = useMembers();
  const { data: mayaAssignee } = useMayaAssignee();
  if (!model) return null;

  const memberNames = new Map(
    [...members, ...(mayaAssignee ? [mayaAssignee] : [])].map((member) => [
      member.id,
      member.fullName || member.username,
    ]),
  );

  return (
    <section aria-label="Maya work plan" className="w-full min-w-0">
      <Flex align="center" className="gap-3" justify="between">
        <Text className="text-base" fontWeight="semibold">
          Maya work plan
        </Text>
        <Flex align="center" className="min-w-0 shrink-0 gap-1.5">
          <span
            aria-hidden
            className={cn(
              "size-2.5 shrink-0 rounded-[2px]",
              statusClasses(model.runStatus),
            )}
          />
          <Text className="text-text-muted max-w-[16ch] truncate text-base font-medium">
            {statusLabel(model.runStatus)}
          </Text>
        </Flex>
      </Flex>
      <Text className="text-text-muted mt-1 text-base leading-6">
        {model.summary}
      </Text>

      {model.actions.length > 0 ? (
        <Box className="border-border/70 mt-3 border-t dark:border-white/[0.09]">
          {model.actions.map((action) => {
            const ownerName = action.assigneeId
              ? memberNames.get(action.assigneeId)
              : undefined;
            const scheduledWindow = formatScheduledWindow(action);
            const context = getActionContext({
              action,
              ownerName,
              scheduledWindow,
            });
            const detail = [context, action.reason]
              .filter(Boolean)
              .filter((value, index, values) => values.indexOf(value) === index)
              .join(" · ");

            return (
              <Box
                className="border-border/70 min-h-11 border-b px-px py-2 dark:border-white/[0.09]"
                key={action.id}
              >
                <Flex align="start" className="gap-3">
                  <Box className="text-text-muted mt-0.5 flex size-5 shrink-0 items-center justify-center">
                    <ActionIcon type={action.type} />
                  </Box>
                  <Box className="min-w-0 flex-1">
                    <Flex
                      align="center"
                      className="min-w-0 gap-3"
                      justify="between"
                    >
                      <Text
                        className="min-w-0 truncate text-base"
                        fontWeight="medium"
                      >
                        {action.label}
                      </Text>
                      <Flex align="center" className="min-w-0 shrink-0 gap-1.5">
                        <span
                          aria-hidden
                          className={cn(
                            "size-2.5 shrink-0 rounded-[2px]",
                            statusClasses(action.status),
                          )}
                        />
                        <Text className="text-text-muted max-w-[16ch] truncate text-base font-medium">
                          {statusLabel(action.status)}
                        </Text>
                      </Flex>
                    </Flex>

                    {detail ? (
                      <Text
                        className="text-text-muted mt-0.5 truncate text-[0.95rem] leading-5"
                        title={detail}
                      >
                        {detail}
                      </Text>
                    ) : null}

                    {action.riskMessage ? (
                      <Box className="border-warning/50 mt-2 border-l-2 pl-2.5">
                        {action.riskCode ? (
                          <Text
                            className="text-warning text-xs tracking-wide uppercase"
                            fontWeight="semibold"
                          >
                            {action.riskCode.replaceAll("_", " ")}
                          </Text>
                        ) : null}
                        <Text className="mt-0.5 text-[0.95rem] leading-5">
                          {action.riskMessage}
                        </Text>
                      </Box>
                    ) : null}
                  </Box>
                </Flex>
              </Box>
            );
          })}
        </Box>
      ) : null}
    </section>
  );
};

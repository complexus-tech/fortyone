"use client";

import { format, isSameDay } from "date-fns";
import {
  CalendarIcon,
  CheckIcon,
  ErrorIcon,
  TimeScheduleIcon,
  UsersAddIcon,
  WarningIcon,
} from "icons";
import { cn } from "lib";
import { Box, Divider, Flex, Text } from "ui";
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
      return "bg-success/10 text-success";
    case "failed":
      return "bg-danger/10 text-danger";
    default:
      return "bg-surface-muted text-text-muted";
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

const ActionStatusIcon = ({ status }: { status: string }) => {
  if (status === "failed") {
    return <ErrorIcon aria-hidden className="text-danger h-3.5" />;
  }
  if (status === "applied" || status === "completed") {
    return <CheckIcon aria-hidden className="text-success h-3.5" />;
  }
  return <TimeScheduleIcon aria-hidden className="text-text-muted h-3.5" />;
};

const formatScheduledWindow = (action: MayaWorkPlanActionModel) => {
  if (!action.startAt || !action.endAt) return "";
  const start = new Date(action.startAt);
  const end = new Date(action.endAt);
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) return "";

  const endPattern = isSameDay(start, end) ? "h:mm a" : "MMM d, h:mm a";
  return `${format(start, "MMM d, h:mm a")} – ${format(end, endPattern)}`;
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
    <Box className="border-border bg-surface-elevated/60 overflow-hidden rounded-xl border">
      <Flex align="start" className="p-3.5" gap={3}>
        <Box className="bg-surface-muted flex size-9 shrink-0 items-center justify-center rounded-lg">
          <TimeScheduleIcon aria-hidden className="text-text-muted h-5" />
        </Box>
        <Box className="min-w-0 flex-1">
          <Flex align="center" className="gap-2" wrap>
            <Text fontWeight="semibold">Maya work plan</Text>
            <span
              className={cn(
                "rounded-full px-2 py-0.5 text-xs font-medium",
                statusClasses(model.runStatus),
              )}
            >
              {statusLabel(model.runStatus)}
            </span>
          </Flex>
          <Text className="mt-1 text-sm leading-5" color="muted">
            {model.summary}
          </Text>
        </Box>
      </Flex>

      {model.actions.length > 0 ? <Divider /> : null}
      {model.actions.map((action, index) => {
        const ownerName = action.assigneeId
          ? memberNames.get(action.assigneeId)
          : undefined;
        const scheduledWindow = formatScheduledWindow(action);
        const showRiskReason =
          action.riskMessage && action.riskMessage !== action.reason;

        return (
          <Box
            className={cn("p-3.5", {
              "border-border border-t": index > 0,
            })}
            key={action.id}
          >
            <Flex align="start" className="gap-2.5">
              <Box className="text-text-muted mt-0.5 shrink-0">
                <ActionIcon type={action.type} />
              </Box>
              <Box className="min-w-0 flex-1">
                <Flex align="center" className="gap-2" wrap>
                  <Text fontSize="md" fontWeight="medium">
                    {action.label}
                  </Text>
                  <Flex align="center" className="gap-1">
                    <ActionStatusIcon status={action.status} />
                    <Text className="text-xs" color="muted">
                      {statusLabel(action.status)}
                    </Text>
                  </Flex>
                </Flex>

                {action.type === "assign_story" ? (
                  <Text className="mt-1 text-sm" color="muted">
                    {ownerName || "Owner selected by Maya"}
                  </Text>
                ) : null}
                {action.type === "schedule_work_block" ? (
                  <Text className="mt-1 text-sm" color="muted">
                    {[action.title, scheduledWindow, ownerName]
                      .filter(Boolean)
                      .join(" · ") || "Calendar time selected by Maya"}
                  </Text>
                ) : null}

                {action.type !== "flag_schedule_risk" || showRiskReason ? (
                  <Text className="mt-1.5 text-sm leading-5">
                    {action.reason}
                  </Text>
                ) : null}

                {action.riskMessage ? (
                  <Box className="border-warning/30 bg-warning/5 mt-2 rounded-lg border px-2.5 py-2">
                    <Flex align="center" className="gap-1.5" wrap>
                      {action.riskCode ? (
                        <Text
                          className="text-warning text-[0.7rem] tracking-wide uppercase"
                          fontWeight="semibold"
                        >
                          {action.riskCode.replaceAll("_", " ")}
                        </Text>
                      ) : null}
                    </Flex>
                    <Text className="mt-0.5 text-sm leading-5">
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
  );
};

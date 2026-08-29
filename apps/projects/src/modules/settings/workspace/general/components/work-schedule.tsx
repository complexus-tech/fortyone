"use client";

import { Box } from "ui";
import { useWorkspaceSettings } from "@/lib/hooks/workspace/settings";
import { useUpdateWorkspaceSettingsMutation } from "@/lib/hooks/workspace/update-settings";
import {
  SectionHeader,
  WorkingDaysSetting,
  WorkingHoursSetting,
} from "@/modules/settings/components";
import {
  DEFAULT_WORKING_DAYS,
  DEFAULT_WORKING_END_MINUTE,
  DEFAULT_WORKING_START_MINUTE,
} from "@/modules/settings/lib/work-schedule";

export const WorkspaceWorkSchedule = () => {
  const { data: settings, isPending: isSettingsPending } =
    useWorkspaceSettings();
  const updateSettings = useUpdateWorkspaceSettingsMutation();
  const isPending = isSettingsPending || updateSettings.isPending;
  const workingDays = settings?.workingDays ?? DEFAULT_WORKING_DAYS;
  const startMinute =
    settings?.workingStartMinute ?? DEFAULT_WORKING_START_MINUTE;
  const endMinute = settings?.workingEndMinute ?? DEFAULT_WORKING_END_MINUTE;

  return (
    <Box className="border-border bg-surface mb-6 rounded-2xl border">
      <SectionHeader
        description="Default availability for planning and progress. People can override it in Preferences."
        title="Work schedule"
      />
      <WorkingDaysSetting
        effectiveValue={workingDays}
        isPending={isPending}
        onSave={(value, onSuccess) => {
          if (!value) return;
          updateSettings.mutate({ workingDays: value }, { onSuccess });
        }}
      />
      <WorkingHoursSetting
        effectiveEndMinute={endMinute}
        effectiveStartMinute={startMinute}
        isPending={isPending}
        onSave={(value, onSuccess) => {
          if (!value) return;
          updateSettings.mutate(
            {
              workingEndMinute: value.endMinute,
              workingStartMinute: value.startMinute,
            },
            { onSuccess },
          );
        }}
      />
    </Box>
  );
};

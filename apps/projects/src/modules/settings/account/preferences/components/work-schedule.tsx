"use client";

import { Box } from "ui";
import { useProfile } from "@/lib/hooks/profile";
import { useUpdateProfileMutation } from "@/lib/hooks/update-profile-mutation";
import { useWorkspaceSettings } from "@/lib/hooks/workspace/settings";
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

export const PersonalWorkSchedule = () => {
  const { data: profile, isPending: isProfilePending } = useProfile();
  const { data: workspaceSettings, isPending: isWorkspacePending } =
    useWorkspaceSettings();
  const updateProfile = useUpdateProfileMutation();
  const isPending =
    isProfilePending || isWorkspacePending || updateProfile.isPending;

  const workspaceDays = workspaceSettings?.workingDays ?? DEFAULT_WORKING_DAYS;
  const workspaceStart =
    workspaceSettings?.workingStartMinute ?? DEFAULT_WORKING_START_MINUTE;
  const workspaceEnd =
    workspaceSettings?.workingEndMinute ?? DEFAULT_WORKING_END_MINUTE;
  const hasWorkingDaysOverride = Boolean(profile?.workingDays?.length);
  const personalWorkingDays = profile?.workingDays;
  const personalStartMinute = profile?.workingStartMinute;
  const personalEndMinute = profile?.workingEndMinute;
  const hasWorkingHoursOverride =
    typeof personalStartMinute === "number" &&
    typeof personalEndMinute === "number";

  const saveSchedule = (
    values: {
      workingDays: number[] | null;
      workingEndMinute: number | null;
      workingStartMinute: number | null;
    },
    onSuccess: () => void,
  ) => {
    updateProfile.mutate({ workSchedule: values }, { onSuccess });
  };

  return (
    <Box className="border-border bg-surface mt-6 rounded-2xl border">
      <SectionHeader
        description="Override the workspace schedule Maya uses for your calendar."
        title="Work schedule"
      />
      <WorkingDaysSetting
        allowInheritance
        effectiveValue={
          hasWorkingDaysOverride && personalWorkingDays
            ? personalWorkingDays
            : workspaceDays
        }
        isInherited={!hasWorkingDaysOverride}
        isPending={isPending}
        onSave={(workingDays, onSuccess) => {
          saveSchedule(
            {
              workingDays,
              workingEndMinute: profile?.workingEndMinute ?? null,
              workingStartMinute: profile?.workingStartMinute ?? null,
            },
            onSuccess,
          );
        }}
      />
      <WorkingHoursSetting
        allowInheritance
        effectiveEndMinute={
          hasWorkingHoursOverride ? personalEndMinute : workspaceEnd
        }
        effectiveStartMinute={
          hasWorkingHoursOverride ? personalStartMinute : workspaceStart
        }
        isInherited={!hasWorkingHoursOverride}
        isPending={isPending}
        onSave={(value, onSuccess) => {
          saveSchedule(
            {
              workingDays: profile?.workingDays ?? null,
              workingEndMinute: value?.endMinute ?? null,
              workingStartMinute: value?.startMinute ?? null,
            },
            onSuccess,
          );
        }}
      />
    </Box>
  );
};

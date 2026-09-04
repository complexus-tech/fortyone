"use client";
import { Box, Button, Text } from "ui";
import {
  JOIN_ONBOARDING_STEPS,
  OnboardingStepper,
} from "@/components/onboarding/onboarding-stepper";
import { Logo } from "@/components/ui/logo";
import type { User } from "@/types/user";
import type { Workspace } from "@/types/workspace";
import { ActionCard } from "./components/action-card";
import {
  CalendarActionIcon,
  ImportActionIcon,
  TaskActionIcon,
} from "./components/welcome-action-icons";
import { getWelcomeDestinations } from "./destinations";

export const Welcome = ({
  callbackUrl,
  workspaces,
  profile,
}: {
  callbackUrl?: string;
  workspaces: Workspace[];
  profile: User;
}) => {
  const { calendarUrl, importUrl, redirectUrl, taskUrl } =
    getWelcomeDestinations(
      workspaces,
      profile.lastUsedWorkspaceId,
      callbackUrl,
    );

  return (
    <Box className="w-full px-6 md:max-w-xl">
      <Logo asIcon />
      <Text as="h1" className="mt-8 mb-6 text-4xl" fontWeight="semibold">
        Welcome to FortyOne👋
      </Text>
      <OnboardingStepper currentStep={2} steps={JOIN_ONBOARDING_STEPS} />
      <Box className="grid gap-4">
        {taskUrl ? (
          <ActionCard
            description="Turn your next idea into action."
            href={taskUrl}
            icon={<TaskActionIcon />}
            title="Create your first task"
          />
        ) : null}
        {importUrl ? (
          <ActionCard
            description="Bring tasks from another tool."
            href={importUrl}
            icon={<ImportActionIcon />}
            title="Import existing work"
          />
        ) : null}
        {calendarUrl ? (
          <ActionCard
            description="See meetings beside your work."
            href={calendarUrl}
            icon={<CalendarActionIcon />}
            title="Connect your calendar"
          />
        ) : null}
      </Box>
      <Button
        align="center"
        className="mt-4"
        color="invert"
        fullWidth
        href={redirectUrl}
      >
        Start planning
      </Button>
    </Box>
  );
};

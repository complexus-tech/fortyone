import { Box, Flex, Text, Select, Button, Wrapper } from "ui";
import {
  ObjectiveIcon,
  OKRIcon,
  PlusIcon,
  SprintsIcon,
  StoryIcon,
  WarningIcon,
} from "icons";
import type { ReactNode } from "react";
import { useMemo } from "react";
import { SectionHeader } from "@/modules/settings/components";
import { FeatureGuard, RowWrapper } from "@/components/ui";
import { useWorkspaceSettings } from "@/lib/hooks/workspace/settings";
import type { WorkspaceSettings } from "@/types";
import { useUpdateWorkspaceSettingsMutation } from "@/lib/hooks/workspace/update-settings";
import { useTerminology } from "@/hooks/use-terminology-display";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useUserRole } from "@/hooks";

type TermOption = {
  label: string;
  value: string;
};

type TermEntity = {
  name: string;
  description: string;
  icon: ReactNode;
  defaultValue: string;
  options: TermOption[];
  key: keyof WorkspaceSettings;
  disabled?: boolean;
};

export const TerminologyPreferences = () => {
  const {
    data: settings = {
      storyTerm: "story",
      sprintTerm: "sprint",
      objectiveTerm: "objective",
      keyResultTerm: "key result",
      sprintEnabled: true,
      objectiveEnabled: true,
      keyResultEnabled: true,
    },
  } = useWorkspaceSettings();
  const { mutate: updateSettings } = useUpdateWorkspaceSettingsMutation();
  const { hasFeature } = useSubscriptionFeatures();
  const { getTermDisplay } = useTerminology();
  const { userRole } = useUserRole();

  const entities: TermEntity[] = useMemo(
    () => [
      {
        name: getTermDisplay("storyTerm", {
          variant: "plural",
          capitalize: true,
        }),
        description: "Small, actionable units of work in your system",
        icon: <StoryIcon className="h-4" />,
        defaultValue: "story",
        key: "storyTerm",
        options: [
          { label: "Story", value: "story" },
          { label: "Task", value: "task" },
          { label: "Issue", value: "issue" },
        ],
      },
      {
        name: getTermDisplay("sprintTerm", {
          variant: "plural",
          capitalize: true,
        }),
        description: "Time-boxed periods for completing a set of work items",
        icon: <SprintsIcon className="h-4" />,
        defaultValue: "sprint",
        key: "sprintTerm",
        options: [
          { label: "Sprint", value: "sprint" },
          { label: "Cycle", value: "cycle" },
          { label: "Iteration", value: "iteration" },
        ],
        disabled: !settings.sprintEnabled,
      },
      {
        name: getTermDisplay("objectiveTerm", {
          variant: "plural",
          capitalize: true,
        }),
        description: "High-level goals that define what you want to achieve",
        icon: <ObjectiveIcon className="h-4" />,
        defaultValue: "objective",
        key: "objectiveTerm",
        options: [
          { label: "Objective", value: "objective" },
          { label: "Goal", value: "goal" },
          { label: "Project", value: "project" },
        ],
        disabled: !settings.objectiveEnabled,
      },
      {
        name: getTermDisplay("keyResultTerm", {
          variant: "plural",
          capitalize: true,
        }),
        description:
          "Measurable outcomes that track progress toward objectives",
        icon: <OKRIcon className="h-4" />,
        defaultValue: "key result",
        key: "keyResultTerm",
        options: [
          { label: "Key Result", value: "key result" },
          { label: "Focus Area", value: "focus area" },
          { label: "Milestone", value: "milestone" },
        ],
        disabled: !settings.keyResultEnabled,
      },
    ],
    [getTermDisplay, settings],
  );

  const handleTerminologyChange = (
    key: keyof WorkspaceSettings,
    value: string,
  ) => {
    updateSettings({ [key]: value });
  };

  return (
    <>
      <FeatureGuard
        fallback={
          <Wrapper className="mb-6 flex items-center justify-between gap-2 border border-warning bg-warning/10 p-4 dark:border-warning/20 dark:bg-warning/10">
            <Flex align="center" gap={2}>
              <WarningIcon className="text-warning dark:text-warning" />
              <Text>
                {userRole === "admin" ? "Upgrade" : "Ask your admin to upgrade"}{" "}
                to a business or enterprise plan to customize terminology
              </Text>
            </Flex>
            {userRole === "admin" && (
              <Button color="warning" href="/settings/workspace/billing">
                Upgrade now
              </Button>
            )}
          </Wrapper>
        }
        feature="customTerminology"
      >
        <Box className="mb-6 rounded-2xl border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
          <SectionHeader
            description="Customize the terminology used throughout your workspace."
            title="Terminology Preferences"
          />
          <Box>
            {entities
              .filter((entity) => !entity.disabled)
              .map((entity) => (
                <RowWrapper
                  className="last-of-type:border-b-0 md:px-6"
                  key={entity.name}
                >
                  <Flex align="center" gap={2}>
                    <Flex
                      align="center"
                      className="size-8 shrink-0 rounded-[0.6rem] bg-gray-100/50 dark:bg-dark-100"
                      justify="center"
                    >
                      {entity.icon}
                    </Flex>
                    <Box>
                      <Text className="font-medium">{entity.name}</Text>
                      <Text className="line-clamp-2" color="muted">
                        {entity.description}
                      </Text>
                    </Box>
                  </Flex>
                  <Select
                    defaultValue={entity.defaultValue}
                    disabled={!hasFeature("customTerminology")}
                    onValueChange={(value) => {
                      handleTerminologyChange(entity.key, value);
                    }}
                    value={settings[entity.key] as string}
                  >
                    <Select.Trigger className="h-9 w-max text-base disabled:opacity-50 md:min-w-36">
                      <Select.Input />
                    </Select.Trigger>
                    <Select.Content align="center">
                      <Select.Group>
                        {entity.options.map((option) => (
                          <Select.Option
                            className="text-base"
                            key={option.value}
                            value={option.value}
                          >
                            <Flex align="center" gap={2}>
                              {option.label}
                            </Flex>
                          </Select.Option>
                        ))}
                      </Select.Group>
                    </Select.Content>
                  </Select>
                </RowWrapper>
              ))}
          </Box>
        </Box>
      </FeatureGuard>

      {/* Preview Section */}
      <Box className="rounded-2xl border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="See how your selected terminology will appear throughout the application."
          title="Terminology Preview"
        />

        <Box className="grid grid-cols-1 md:grid-cols-2">
          <Box className="border-r border-gray-100 dark:border-dark-100">
            <RowWrapper className="justify-start gap-2 py-3 md:px-6">
              <StoryIcon />
              My {getTermDisplay("storyTerm", { variant: "plural" })}
            </RowWrapper>
            {settings.sprintEnabled ? (
              <RowWrapper className="justify-start gap-2 py-3 md:px-6">
                <SprintsIcon />
                Active {getTermDisplay("sprintTerm", { variant: "plural" })}
              </RowWrapper>
            ) : null}

            {settings.objectiveEnabled ? (
              <RowWrapper className="justify-start gap-2 py-3 md:px-6">
                <ObjectiveIcon />
                All {getTermDisplay("objectiveTerm", { variant: "plural" })}
              </RowWrapper>
            ) : null}

            {settings.keyResultEnabled ? (
              <RowWrapper className="justify-start gap-2 py-3 md:border-b-0 md:px-6">
                <OKRIcon />
                Track {getTermDisplay("keyResultTerm", { variant: "plural" })}
              </RowWrapper>
            ) : null}
          </Box>

          <Box>
            <RowWrapper className="py-3 md:px-6">Buttons & Actions</RowWrapper>
            <Flex className="gap-2.5 px-6 py-4" wrap>
              {settings.objectiveEnabled ? (
                <Button color="tertiary" leftIcon={<ObjectiveIcon />}>
                  Create {getTermDisplay("objectiveTerm", { capitalize: true })}
                </Button>
              ) : null}
              {settings.sprintEnabled ? (
                <Button color="tertiary" leftIcon={<SprintsIcon />}>
                  Start {getTermDisplay("sprintTerm", { capitalize: true })}
                </Button>
              ) : null}
              <Button color="tertiary" leftIcon={<PlusIcon />}>
                Create {getTermDisplay("storyTerm", { capitalize: true })}
              </Button>
              {settings.keyResultEnabled ? (
                <Button color="tertiary" leftIcon={<OKRIcon />}>
                  Add {getTermDisplay("keyResultTerm", { capitalize: true })}
                </Button>
              ) : null}
            </Flex>
          </Box>
        </Box>
      </Box>
    </>
  );
};

import { Box, Flex, Text, Select, Button } from "ui";
import { ObjectiveIcon, OKRIcon, SprintsIcon, StoryIcon } from "icons";
import type { ReactNode } from "react";
import { SectionHeader } from "@/modules/settings/components";
import { RowWrapper } from "@/components/ui";

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
};

export const TerminologyPreferences = () => {
  const entities: TermEntity[] = [
    {
      name: "Stories",
      description: "Small, actionable units of work in your system",
      icon: <StoryIcon className="h-4" />,
      defaultValue: "story",
      options: [
        { label: "Story", value: "story" },
        { label: "Task", value: "task" },
        { label: "Issue", value: "issue" },
      ],
    },
    {
      name: "Sprints",
      description: "Time-boxed periods for completing a set of work items",
      icon: <SprintsIcon className="h-4" />,
      defaultValue: "sprint",
      options: [
        { label: "Sprint", value: "sprint" },
        { label: "Cycle", value: "cycle" },
        { label: "Iteration", value: "iteration" },
      ],
    },
    {
      name: "Objectives",
      description: "High-level goals that define what you want to achieve",
      icon: <ObjectiveIcon className="h-4" />,
      defaultValue: "objective",
      options: [
        { label: "Objective", value: "objective" },
        { label: "Goal", value: "goal" },
        { label: "Project", value: "project" },
      ],
    },
    {
      name: "Key Results",
      description: "Measurable outcomes that track progress toward objectives",
      icon: <OKRIcon className="h-4" />,
      defaultValue: "key result",
      options: [
        { label: "Key Result", value: "key result" },
        { label: "Focus Area", value: "focus area" },
        { label: "Strategic Priority", value: "strategic priority" },
      ],
    },
  ];

  return (
    <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
      <SectionHeader
        action={
          <Button className="shrink-0" color="tertiary" disabled>
            Save changes
          </Button>
        }
        description="Customize the terminology used in your workspace by your team."
        title="Terminology Preferences"
      />
      <Box>
        {entities.map((entity) => (
          <RowWrapper
            className="px-6 last-of-type:border-b-0"
            key={entity.name}
          >
            <Flex align="center" gap={2}>
              <Flex
                align="center"
                className="size-8 shrink-0 rounded-lg bg-gray-100/50 dark:bg-dark-100"
                justify="center"
              >
                {entity.icon}
              </Flex>
              <Box>
                <Text className="font-medium">{entity.name}</Text>
                <Text color="muted">{entity.description}</Text>
              </Box>
            </Flex>
            <Select defaultValue={entity.defaultValue}>
              <Select.Trigger className="h-9 w-max min-w-36 text-base">
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
  );
};

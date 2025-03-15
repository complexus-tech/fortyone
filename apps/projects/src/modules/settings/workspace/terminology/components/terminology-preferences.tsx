import { Box, Flex, Text, Select, Button } from "ui";
import {
  ObjectiveIcon,
  OKRIcon,
  PlusIcon,
  SprintsIcon,
  StoryIcon,
} from "icons";
import type { ReactNode } from "react";
import { useState } from "react";
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
  key: string;
};

export const TerminologyPreferences = () => {
  const entities: TermEntity[] = [
    {
      name: "Stories",
      description: "Small, actionable units of work in your system",
      icon: <StoryIcon className="h-4" />,
      defaultValue: "story",
      key: "story",
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
      key: "sprint",
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
      key: "objective",
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
      key: "keyResult",
      options: [
        { label: "Key Result", value: "key result" },
        { label: "Focus Area", value: "focus area" },
        { label: "Strategic Priority", value: "strategic priority" },
      ],
    },
  ];

  const [terminology, setTerminology] = useState<Record<string, string>>({
    story: "story",
    sprint: "sprint",
    objective: "objective",
    keyResult: "key result",
  });

  const getTermForms = (term: string) => {
    const singular = term;
    let plural = term;

    if (term.endsWith("y")) {
      plural = `${term.slice(0, -1)}ies`;
    } else if (term === "focus area") {
      plural = "focus areas";
    } else if (term === "strategic priority") {
      plural = "strategic priorities";
    } else {
      plural = `${term}s`;
    }

    return {
      singular,
      plural,
    };
  };

  const getTermLabel = (key: string) => {
    const entity = entities.find((e) => e.key === key);
    if (!entity) return "";

    const selectedOption = entity.options.find(
      (o) => o.value === terminology[key],
    );
    return selectedOption ? selectedOption.label : "";
  };

  const forms = {
    story: getTermForms(terminology.story),
    sprint: getTermForms(terminology.sprint),
    objective: getTermForms(terminology.objective),
    keyResult: getTermForms(terminology.keyResult),
  };

  const handleTerminologyChange = (key: string, value: string) => {
    setTerminology((prev) => ({
      ...prev,
      [key]: value,
    }));
  };

  return (
    <>
      {/* Terminology Selection Section */}
      <Box className="mb-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Customize the terminology used throughout your workspace."
          title="Terminology Preferences"
        />
        <Box>
          {entities.map((entity) => (
            <RowWrapper className="px-6" key={entity.name}>
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
              <Select
                defaultValue={entity.defaultValue}
                onValueChange={(value) => {
                  handleTerminologyChange(entity.key, value);
                }}
                value={terminology[entity.key]}
              >
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
        <Flex className="justify-end px-6 py-3.5">
          <Button className="shrink-0" disabled={false}>
            Save changes
          </Button>
        </Flex>
      </Box>

      {/* Preview Section */}
      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="See how your selected terminology will appear throughout the application."
          title="Terminology Preview"
        />

        <Box className="grid grid-cols-1 md:grid-cols-2">
          <Box className="border-r border-gray-100 dark:border-dark-100">
            <RowWrapper className="justify-start gap-2 px-6 py-3">
              <StoryIcon />
              My {forms.story.plural}
            </RowWrapper>
            <RowWrapper className="justify-start gap-2 px-6 py-3">
              <SprintsIcon />
              Active {forms.sprint.plural}
            </RowWrapper>
            <RowWrapper className="justify-start gap-2 px-6 py-3">
              <ObjectiveIcon />
              All {forms.objective.plural}
            </RowWrapper>
            <RowWrapper className="justify-start gap-2 border-b-0 px-6 py-3">
              <OKRIcon />
              Track {forms.keyResult.plural}
            </RowWrapper>
          </Box>

          <Box>
            <RowWrapper className="px-6 py-3">Buttons & Actions</RowWrapper>
            <Flex className="gap-2.5 px-6 py-4" wrap>
              <Button color="tertiary" leftIcon={<ObjectiveIcon />}>
                Create {getTermLabel("objective")}
              </Button>
              <Button color="tertiary" leftIcon={<SprintsIcon />}>
                Start {getTermLabel("sprint")}
              </Button>
              <Button color="tertiary" leftIcon={<PlusIcon />}>
                Create {getTermLabel("story")}
              </Button>
              <Button color="tertiary" leftIcon={<OKRIcon />}>
                Add {getTermLabel("keyResult")}
              </Button>
            </Flex>
          </Box>
        </Box>
      </Box>
    </>
  );
};

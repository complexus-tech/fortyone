import { Box, Flex, Text, Select } from "ui";
import { ObjectiveIcon } from "icons";
import { SectionHeader } from "@/modules/settings/components";
import { RowWrapper } from "@/components/ui";

export const TerminologyPreferences = () => {
  return (
    <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
      <SectionHeader
        description="Customize the terminology used in your workspace to match your team's preferences."
        title="Terminology Preferences"
      />
      <Box>
        {/* Story terminology */}
        <RowWrapper className="px-6">
          <Flex align="center" gap={2}>
            <Flex
              align="center"
              className="size-8 shrink-0 rounded-lg bg-gray-100/50 dark:bg-dark-100"
              justify="center"
            >
              <ObjectiveIcon className="h-4" />
            </Flex>
            <Box>
              <Text className="font-medium">Stories</Text>
              <Text color="muted">
                Small, actionable units of work in your system
              </Text>
            </Box>
          </Flex>
          <Select defaultValue="story">
            <Select.Trigger className="h-9 w-36 text-base">
              <Select.Input />
            </Select.Trigger>
            <Select.Content align="center">
              <Select.Group>
                <Select.Option className="text-base" value="story">
                  <Flex align="center" gap={2}>
                    Story
                  </Flex>
                </Select.Option>
                <Select.Option className="text-base" value="task">
                  <Flex align="center" gap={2}>
                    Task
                  </Flex>
                </Select.Option>
                <Select.Option className="text-base" value="issue">
                  <Flex align="center" gap={2}>
                    Issue
                  </Flex>
                </Select.Option>
              </Select.Group>
            </Select.Content>
          </Select>
        </RowWrapper>

        {/* Sprint terminology */}
        <RowWrapper className="px-6">
          <Box>
            <Text className="font-medium">Sprints</Text>
            <Text color="muted">
              Time-boxed periods for completing a set of work items
            </Text>
          </Box>
          <Select defaultValue="sprint">
            <Select.Trigger className="h-9 w-max px-2 text-base">
              <Select.Input />
            </Select.Trigger>
            <Select.Content align="center">
              <Select.Group>
                <Select.Option className="text-base" value="sprint">
                  <Flex align="center" gap={2}>
                    Sprint
                  </Flex>
                </Select.Option>
                <Select.Option className="text-base" value="cycle">
                  <Flex align="center" gap={2}>
                    Cycle
                  </Flex>
                </Select.Option>
                <Select.Option className="text-base" value="iteration">
                  <Flex align="center" gap={2}>
                    Iteration
                  </Flex>
                </Select.Option>
              </Select.Group>
            </Select.Content>
          </Select>
        </RowWrapper>

        {/* Objective terminology */}
        <RowWrapper className="px-6">
          <Box>
            <Text className="font-medium">Objectives</Text>
            <Text color="muted">
              High-level goals that define what you want to achieve
            </Text>
          </Box>
          <Select defaultValue="objective">
            <Select.Trigger className="h-9 w-max px-2 text-base">
              <Select.Input />
            </Select.Trigger>
            <Select.Content align="center">
              <Select.Group>
                <Select.Option className="text-base" value="objective">
                  <Flex align="center" gap={2}>
                    Objective
                  </Flex>
                </Select.Option>
                <Select.Option className="text-base" value="goal">
                  <Flex align="center" gap={2}>
                    Goal
                  </Flex>
                </Select.Option>
                <Select.Option className="text-base" value="project">
                  <Flex align="center" gap={2}>
                    Project
                  </Flex>
                </Select.Option>
              </Select.Group>
            </Select.Content>
          </Select>
        </RowWrapper>

        {/* Key Result terminology */}
        <RowWrapper className="px-6">
          <Box>
            <Text className="font-medium">Key Results</Text>
            <Text color="muted">
              Measurable outcomes that track progress toward objectives
            </Text>
          </Box>
          <Select defaultValue="key result">
            <Select.Trigger className="h-9 w-max px-2 text-base">
              <Select.Input />
            </Select.Trigger>
            <Select.Content align="center">
              <Select.Group>
                <Select.Option className="text-base" value="key result">
                  <Flex align="center" gap={2}>
                    Key Result
                  </Flex>
                </Select.Option>
                <Select.Option className="text-base" value="focus area">
                  <Flex align="center" gap={2}>
                    Focus Area
                  </Flex>
                </Select.Option>
                <Select.Option className="text-base" value="strategic priority">
                  <Flex align="center" gap={2}>
                    Strategic Priority
                  </Flex>
                </Select.Option>
              </Select.Group>
            </Select.Content>
          </Select>
        </RowWrapper>
      </Box>
    </Box>
  );
};

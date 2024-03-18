"use client";
import { Box, Button, Divider, Flex, Popover, Switch, Text, Select } from "ui";
import { ArrowDownIcon, PreferencesIcon } from "icons";
import { useLocalStorage } from "@/hooks";

export type StoriesFilter = {
  groupBy: string;
  orderBy: string;
  storyType: string;
  showEmptyGroups: boolean;
  showSubStories: boolean;
  displayColumns: string[];
};

export const StoriesFiltersButton = () => {
  const initialFilter: StoriesFilter = {
    groupBy: "Status",
    orderBy: "Priority",
    storyType: "Active",
    showEmptyGroups: false,
    showSubStories: false,
    displayColumns: [
      "Status",
      "Assignee",
      "Priority",
      "Due date",
      "Created",
      "Updated",
      "Sprint",
      "Labels",
    ],
  };

  const [filter, setFilter] = useLocalStorage("filters", initialFilter);
  const {
    groupBy,
    orderBy,
    storyType,
    showEmptyGroups,
    showSubStories,
    displayColumns,
  } = filter!;

  const allColumns = [
    "Status",
    "Assignee",
    "Priority",
    "Due date",
    "Created",
    "Updated",
    "Sprint",
    "Epic",
    "Labels",
  ];

  const groupByOptions = ["Status", "Assignee", "Priority"];
  const orderByOptions = ["Priority", "Due date", "Created", "Updated"];
  const storyTypes = ["Active", "Backlog"];

  const isDefaultSetup =
    JSON.stringify(filter) === JSON.stringify(initialFilter);

  return (
    <Popover>
      <Popover.Trigger asChild>
        <Button
          color="tertiary"
          leftIcon={<PreferencesIcon className="h-4 w-auto" />}
          rightIcon={<ArrowDownIcon className="h-4 w-auto" />}
          size="sm"
          variant="outline"
        >
          Display
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" className="max-w-[24rem]">
        <Flex align="center" className="my-2 px-4" gap={2} justify="between">
          <Text color="muted">Group by</Text>
          <Select
            onValueChange={(value) => {
              setFilter({
                ...filter!,
                groupBy: value,
              });
            }}
            value={groupBy}
          >
            <Select.Trigger className="w-32">
              <Select.Input />
            </Select.Trigger>
            <Select.Content>
              <Select.Group>
                {groupByOptions.map((option) => (
                  <Select.Option key={option} value={option}>
                    {option}
                  </Select.Option>
                ))}
              </Select.Group>
            </Select.Content>
          </Select>
        </Flex>
        <Flex align="center" className="mb-3 px-4" gap={2} justify="between">
          <Text color="muted">Story type</Text>
          <Select
            onValueChange={(value) => {
              setFilter({
                ...filter!,
                storyType: value,
              });
            }}
            value={storyType}
          >
            <Select.Trigger className="w-32">
              <Select.Input />
            </Select.Trigger>
            <Select.Content>
              <Select.Group>
                {storyTypes.map((option) => (
                  <Select.Option key={option} value={option}>
                    {option}
                  </Select.Option>
                ))}
              </Select.Group>
            </Select.Content>
          </Select>
        </Flex>
        <Flex align="center" className="mb-3 px-4" gap={2} justify="between">
          <Text color="muted">Order by</Text>
          <Select
            onValueChange={(value) => {
              setFilter({
                ...filter!,
                orderBy: value,
              });
            }}
            value={orderBy}
          >
            <Select.Trigger className="w-32">
              <Select.Input />
            </Select.Trigger>
            <Select.Content>
              <Select.Group>
                {orderByOptions.map((option) => (
                  <Select.Option key={option} value={option}>
                    {option}
                  </Select.Option>
                ))}
              </Select.Group>
            </Select.Content>
          </Select>
        </Flex>
        <Divider className="my-2" />
        <Box className="max-w-[27rem] px-4 py-2">
          <Text className="mb-4" fontWeight="medium">
            Display options
          </Text>
          <Text className="mb-4" color="muted">
            <label
              className="flex select-none items-center justify-between gap-2"
              htmlFor="more"
            >
              Show empty groups
              <Switch
                checked={showEmptyGroups}
                id="more"
                onCheckedChange={(checked) => {
                  setFilter({
                    ...filter!,
                    showEmptyGroups: checked,
                  });
                }}
              />
            </label>
          </Text>
          <Text className="mb-2" color="muted">
            Display columns
          </Text>

          <Flex gap={2} wrap>
            {allColumns.map((column) => {
              const isSelected = displayColumns.includes(column);
              return (
                <Button
                  color={isSelected ? "primary" : "tertiary"}
                  key={column}
                  onClick={() => {
                    if (isSelected) {
                      setFilter({
                        ...filter!,
                        displayColumns: displayColumns.filter(
                          (col) => col !== column,
                        ),
                      });
                    } else {
                      setFilter({
                        ...filter!,
                        displayColumns: [...displayColumns, column],
                      });
                    }
                  }}
                  rounded="sm"
                  size="xs"
                  variant={isSelected ? "solid" : "outline"}
                >
                  {column}
                </Button>
              );
            })}
          </Flex>
        </Box>
        <Divider className="mb-3 mt-2" />
        <Text className="mb-2 px-4" color="muted">
          <label
            className="flex select-none items-center justify-between gap-2"
            htmlFor="more"
          >
            Show empty groups <Switch id="more" />
          </label>
        </Text>
        <Text className="mb-3 px-4" color="muted">
          <label
            className="flex select-none items-center justify-between gap-2"
            htmlFor="more"
          >
            Show sub stories
            <Switch
              checked={showSubStories}
              id="more"
              onCheckedChange={(checked) => {
                setFilter({
                  ...filter!,
                  showSubStories: checked,
                });
              }}
            />
          </label>
        </Text>
        {!isDefaultSetup ? (
          <>
            <Divider className="mb-2" />
            <Flex className="px-4 pb-[0.1rem]" justify="end">
              <Button
                className="text-primary dark:text-primary"
                color="tertiary"
                onClick={() => {
                  setFilter(initialFilter);
                }}
                size="sm"
                variant="naked"
              >
                Reset to default
              </Button>
            </Flex>
          </>
        ) : null}
      </Popover.Content>
    </Popover>
  );
};

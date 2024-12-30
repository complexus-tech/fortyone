"use client";
import { Box, Button, Divider, Flex, Popover, Switch, Text, Select } from "ui";
import { ArrowDownIcon, PreferencesIcon } from "icons";

export type ViewOptionsGroupBy = "Status" | "Assignee" | "Priority" | "None";
export type DisplayColumn =
  | "ID"
  | "Status"
  | "Assignee"
  | "Priority"
  | "Due date"
  | "Created"
  | "Updated"
  | "Sprint"
  | "Objective"
  | "Epic"
  | "Labels";
export type ViewOptionsOrderBy =
  | "Priority"
  | "Due date"
  | "Created"
  | "Updated";

export type StoriesViewOptions = {
  groupBy: ViewOptionsGroupBy;
  orderBy: ViewOptionsOrderBy;
  showEmptyGroups: boolean;
  displayColumns: DisplayColumn[];
};

const initialViewOptions: StoriesViewOptions = {
  groupBy: "Status",
  orderBy: "Priority",
  showEmptyGroups: false,
  displayColumns: [
    "ID",
    "Status",
    "Assignee",
    "Priority",
    "Due date",
    "Created",
    "Updated",
    // "Sprint",
    // "Labels",
  ],
};

export const StoriesViewOptionsButton = ({
  viewOptions,
  setViewOptions,
  groupByOptions = ["Status", "Assignee", "Priority"],
  orderByOptions = ["Priority", "Due date", "Created", "Updated"],
}: {
  viewOptions: StoriesViewOptions;
  setViewOptions: (v: StoriesViewOptions) => void;
  groupByOptions?: ViewOptionsGroupBy[];
  orderByOptions?: ViewOptionsOrderBy[];
}) => {
  const { groupBy, orderBy, showEmptyGroups, displayColumns } = viewOptions;

  const allColumns: DisplayColumn[] = [
    "ID",
    "Status",
    "Assignee",
    "Priority",
    "Due date",
    "Created",
    "Updated",
    // "Sprint",
    // "Epic",
    // "Objective",
    // "Labels",
  ];

  const hasFilters =
    JSON.stringify(viewOptions) !== JSON.stringify(initialViewOptions);

  return (
    <Popover>
      <Popover.Trigger asChild>
        <Button
          className="relative"
          color="tertiary"
          leftIcon={
            <PreferencesIcon className="h-4 w-auto text-gray dark:text-gray-300" />
          }
          rightIcon={
            <ArrowDownIcon className="h-3.5 w-auto text-gray dark:text-gray-300" />
          }
          size="sm"
          variant="outline"
        >
          Options
          {hasFilters && (
            <span className="absolute -right-0.5 -top-0.5 inline-block size-2 rounded-full bg-primary" />
          )}
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" className="max-w-[24rem] rounded-[0.6rem]">
        <Flex align="center" className="my-2 px-4" gap={2} justify="between">
          <Text color="muted">Group by</Text>
          <Select
            onValueChange={(value: ViewOptionsGroupBy) => {
              setViewOptions({
                ...viewOptions,
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
          <Text color="muted">Order by</Text>
          <Select
            onValueChange={(value: ViewOptionsOrderBy) => {
              setViewOptions({
                ...viewOptions,
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
                  setViewOptions({
                    ...viewOptions,
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
                  className="pl-1.5"
                  color={isSelected ? "primary" : "tertiary"}
                  key={column}
                  onClick={() => {
                    if (isSelected) {
                      setViewOptions({
                        ...viewOptions,
                        displayColumns: displayColumns.filter(
                          (col) => col !== column,
                        ),
                      });
                    } else {
                      setViewOptions({
                        ...viewOptions,
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
        <Divider className="my-2" />
        <Flex className="px-4 pb-[0.1rem]" justify="end">
          <Button
            className="text-primary dark:text-primary"
            color="tertiary"
            onClick={() => {
              setViewOptions(initialViewOptions);
            }}
            size="sm"
            variant="naked"
          >
            Reset to default
          </Button>
        </Flex>
      </Popover.Content>
    </Popover>
  );
};

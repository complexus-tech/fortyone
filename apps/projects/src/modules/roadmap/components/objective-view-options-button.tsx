"use client";

import { ArrowDownIcon, ArrowUpDownIcon, PreferencesIcon } from "icons";
import { Box, Button, Divider, Flex, Popover, Select, Switch, Text } from "ui";
import type {
  ObjectiveGroupBy,
  ObjectiveOrderBy,
  ObjectiveOrderDirection,
  ObjectiveViewOptions,
} from "../objective-board-utils";

const GROUP_OPTIONS: { label: string; value: ObjectiveGroupBy }[] = [
  { label: "Status", value: "status" },
  { label: "Lead", value: "lead" },
  { label: "Priority", value: "priority" },
];

const ORDER_OPTIONS: { label: string; value: ObjectiveOrderBy }[] = [
  { label: "Priority", value: "priority" },
  { label: "Target date", value: "target" },
  { label: "Created", value: "created" },
  { label: "Updated", value: "updated" },
];

export const ObjectiveViewOptionsButton = ({
  viewOptions,
  setViewOptions,
}: {
  viewOptions: ObjectiveViewOptions;
  setViewOptions: (value: ObjectiveViewOptions) => void;
}) => {
  return (
    <Popover>
      <Popover.Trigger asChild>
        <Button
          color="tertiary"
          leftIcon={<PreferencesIcon className="text-text-muted h-4 w-auto" />}
          rightIcon={<ArrowDownIcon className="text-text-muted h-3.5 w-auto" />}
          size="sm"
          variant="outline"
        >
          <span className="hidden md:inline">Customise</span>
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" className="min-w-[20rem] md:max-w-[24rem]">
        <Flex align="center" className="my-2 px-4" gap={2} justify="between">
          <Text color="muted">Group by</Text>
          <Select
            onValueChange={(groupBy: ObjectiveGroupBy) => {
              setViewOptions({ ...viewOptions, groupBy });
            }}
            value={viewOptions.groupBy}
          >
            <Select.Trigger className="bg-surface-muted dark:bg-surface-prominent/70 w-32">
              <Select.Input>
                {
                  GROUP_OPTIONS.find(
                    ({ value }) => value === viewOptions.groupBy,
                  )?.label
                }
              </Select.Input>
            </Select.Trigger>
            <Select.Content className="ring-border/70 shadow-2xl ring-1">
              <Select.Group>
                {GROUP_OPTIONS.map(({ label, value }) => (
                  <Select.Option key={value} value={value}>
                    {label}
                  </Select.Option>
                ))}
              </Select.Group>
            </Select.Content>
          </Select>
        </Flex>
        <Flex align="center" className="mb-3 px-4" gap={2} justify="between">
          <Text color="muted">Order by</Text>
          <Flex align="center" gap={2}>
            <Select
              onValueChange={(orderBy: ObjectiveOrderBy) => {
                setViewOptions({ ...viewOptions, orderBy });
              }}
              value={viewOptions.orderBy}
            >
              <Select.Trigger className="bg-surface-muted dark:bg-surface-prominent/70 w-28">
                <Select.Input>
                  {
                    ORDER_OPTIONS.find(
                      ({ value }) => value === viewOptions.orderBy,
                    )?.label
                  }
                </Select.Input>
              </Select.Trigger>
              <Select.Content className="ring-border/70 shadow-2xl ring-1">
                <Select.Group>
                  {ORDER_OPTIONS.map(({ label, value }) => (
                    <Select.Option key={value} value={value}>
                      {label}
                    </Select.Option>
                  ))}
                </Select.Group>
              </Select.Content>
            </Select>
            <Select
              onValueChange={(orderDirection: ObjectiveOrderDirection) => {
                setViewOptions({ ...viewOptions, orderDirection });
              }}
              value={viewOptions.orderDirection}
            >
              <Select.Trigger
                aria-label="Order direction"
                className="bg-surface-muted dark:bg-surface-prominent/70 w-28"
              >
                <span className="flex min-w-0 items-center gap-1.5">
                  <ArrowUpDownIcon className="text-text-muted h-4 w-auto shrink-0" />
                  <Select.Input>
                    {viewOptions.orderDirection === "asc" ? "Asc" : "Desc"}
                  </Select.Input>
                </span>
              </Select.Trigger>
              <Select.Content className="ring-border/70 shadow-2xl ring-1">
                <Select.Group>
                  <Select.Option value="asc">Ascending</Select.Option>
                  <Select.Option value="desc">Descending</Select.Option>
                </Select.Group>
              </Select.Content>
            </Select>
          </Flex>
        </Flex>
        <Divider className="dark:border-border-strong/80 my-2" />
        <Box className="px-4 py-2">
          <Text className="mb-4">Display options</Text>
          <label
            className="text-text-secondary flex items-center justify-between gap-2 select-none"
            htmlFor="show-empty-objective-groups"
          >
            Show empty groups
            <Switch
              checked={viewOptions.showEmptyGroups}
              id="show-empty-objective-groups"
              onCheckedChange={(showEmptyGroups) => {
                setViewOptions({ ...viewOptions, showEmptyGroups });
              }}
            />
          </label>
        </Box>
      </Popover.Content>
    </Popover>
  );
};

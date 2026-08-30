import { Button, Flex, Menu, Popover, Select, Text } from "ui";
import { ArrowDown2Icon, ArrowUpDownIcon, CheckIcon, FilterIcon } from "icons";
import type {
  DocumentAccessFilter,
  DocumentOwnerFilter,
  DocumentSortDirection,
  DocumentSortField,
  DocumentUpdatedFilter,
} from "./types";

type FilterOption<T extends string> = {
  label: string;
  value: T;
};

const accessOptions: FilterOption<DocumentAccessFilter>[] = [
  { label: "All access", value: "all" },
  { label: "Workspace", value: "workspace" },
  { label: "Shared", value: "restricted" },
  { label: "Private", value: "private" },
];

const ownerOptions: FilterOption<DocumentOwnerFilter>[] = [
  { label: "Anyone", value: "all" },
  { label: "Owned by me", value: "mine" },
  { label: "Owned by others", value: "others" },
];

const updatedOptions: FilterOption<DocumentUpdatedFilter>[] = [
  { label: "Any time", value: "all" },
  { label: "Today", value: "today" },
  { label: "Past 7 days", value: "7d" },
  { label: "Past 30 days", value: "30d" },
  { label: "Past 90 days", value: "90d" },
];

type DocumentSortOption = {
  direction: DocumentSortDirection;
  field: DocumentSortField;
  label: string;
};

const sortOptions: DocumentSortOption[] = [
  { direction: "desc", field: "updated", label: "Newest" },
  { direction: "asc", field: "updated", label: "Oldest" },
  { direction: "asc", field: "title", label: "A to Z" },
  { direction: "desc", field: "title", label: "Z to A" },
];

const DocumentFilterSelect = <T extends string>({
  label,
  onChange,
  options,
  value,
}: {
  label: string;
  onChange: (value: T) => void;
  options: FilterOption<T>[];
  value: T;
}) => (
  <Flex align="center" className="px-4 py-2" gap={4} justify="between">
    <Text color="muted">{label}</Text>
    <Select
      onValueChange={(nextValue) => {
        onChange(nextValue as T);
      }}
      value={value}
    >
      <Select.Trigger className="bg-surface-muted dark:bg-surface-prominent/70 w-40">
        <Select.Input />
      </Select.Trigger>
      <Select.Content className="ring-border/70 shadow-2xl ring-1">
        <Select.Group>
          {options.map((option) => (
            <Select.Option key={option.value} value={option.value}>
              {option.label}
            </Select.Option>
          ))}
        </Select.Group>
      </Select.Content>
    </Select>
  </Flex>
);

type DocumentFiltersProps = {
  access: DocumentAccessFilter;
  activeCount: number;
  onAccessChange: (value: DocumentAccessFilter) => void;
  onClear: () => void;
  onOwnerChange: (value: DocumentOwnerFilter) => void;
  onUpdatedChange: (value: DocumentUpdatedFilter) => void;
  owner: DocumentOwnerFilter;
  showOwner: boolean;
  updated: DocumentUpdatedFilter;
};

export const DocumentFilters = ({
  access,
  activeCount,
  onAccessChange,
  onClear,
  onOwnerChange,
  onUpdatedChange,
  owner,
  showOwner,
  updated,
}: DocumentFiltersProps) => (
  <Popover>
    <Popover.Trigger asChild>
      <Button
        aria-label={
          activeCount > 0 ? `Filters, ${activeCount} applied` : "Filters"
        }
        className="relative"
        color="tertiary"
        leftIcon={<FilterIcon className="h-4 w-auto" />}
        rightIcon={<ArrowDown2Icon className="h-3.5 w-auto" />}
        size="sm"
        variant="outline"
      >
        {activeCount > 0 ? (
          <span
            aria-hidden="true"
            className="bg-primary absolute -top-0.5 -right-0.5 size-2.5 rounded-full"
          />
        ) : null}
        Filters
      </Button>
    </Popover.Trigger>
    <Popover.Content align="start" className="min-w-[20rem] pb-2">
      <Flex align="center" className="my-2 px-4" justify="between">
        <Text color="muted">Apply filters</Text>
        {activeCount > 0 ? (
          <Button
            className="text-primary dark:text-primary"
            color="tertiary"
            onClick={onClear}
            size="sm"
            variant="naked"
          >
            Clear filters
          </Button>
        ) : null}
      </Flex>
      <DocumentFilterSelect
        label="Access"
        onChange={onAccessChange}
        options={accessOptions}
        value={access}
      />
      {showOwner ? (
        <DocumentFilterSelect
          label="Owner"
          onChange={onOwnerChange}
          options={ownerOptions}
          value={owner}
        />
      ) : null}
      <DocumentFilterSelect
        label="Updated"
        onChange={onUpdatedChange}
        options={updatedOptions}
        value={updated}
      />
    </Popover.Content>
  </Popover>
);

type DocumentSortMenuProps = {
  direction: DocumentSortDirection;
  field: DocumentSortField;
  onChange: (option: DocumentSortOption) => void;
};

export const DocumentSortMenu = ({
  direction,
  field,
  onChange,
}: DocumentSortMenuProps) => {
  const selectedOption =
    sortOptions.find(
      (option) => option.field === field && option.direction === direction,
    ) ?? sortOptions[0];

  return (
    <Menu>
      <Menu.Button>
        <Button
          className="gap-1.5 px-1.5 whitespace-nowrap"
          color="tertiary"
          leftIcon={
            <ArrowUpDownIcon
              className="text-text-muted h-4 w-auto"
              strokeWidth={2}
            />
          }
          rightIcon={
            <ArrowDown2Icon
              className="text-text-muted h-3.5 w-auto"
              strokeWidth={2}
            />
          }
          size="sm"
          variant="naked"
        >
          {selectedOption.label}
        </Button>
      </Menu.Button>
      <Menu.Items align="end" className="min-w-44 p-1">
        {sortOptions.map((option) => {
          const isActive =
            option.field === field && option.direction === direction;
          return (
            <Menu.Item
              active={isActive}
              className="justify-between gap-3"
              key={`${option.field}:${option.direction}`}
              onSelect={() => {
                onChange(option);
              }}
            >
              <span>{option.label}</span>
              {isActive ? (
                <CheckIcon className="h-4 w-auto" strokeWidth={2} />
              ) : null}
            </Menu.Item>
          );
        })}
      </Menu.Items>
    </Menu>
  );
};

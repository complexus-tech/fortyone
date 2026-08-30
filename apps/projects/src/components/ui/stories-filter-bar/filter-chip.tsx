import { useState, type ReactNode } from "react";
import { Button, Dialog, Flex, Input, Menu, Popover } from "ui";
import { CheckIcon, CloseIcon } from "icons";
import type {
  StoriesFilter,
  StoriesFilterOperator,
} from "../stories-filter-types";
import { getEditorContentClassName } from "./filter-model";
import type {
  FilterChip,
  SetStoriesFilters,
  StoriesFilterField,
} from "./types";

export const TitleFilterDialog = ({
  filters,
  onOpenChange,
  open,
  setFilters,
}: {
  filters: StoriesFilter;
  onOpenChange: (open: boolean) => void;
  open: boolean;
  setFilters: SetStoriesFilters;
}) => {
  const [draft, setDraft] = useState(filters.contentContains ?? "");

  const applyTitleFilter = () => {
    const contentContains = draft.trim();
    setFilters({
      ...filters,
      contentContains: contentContains || null,
    });
    onOpenChange(false);
  };

  return (
    <Dialog onOpenChange={onOpenChange} open={open}>
      <Dialog.Content className="max-w-lg" hideClose>
        <Dialog.Header className="px-6 pt-6">
          <Dialog.Title className="text-lg">Filter by content</Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="pt-1">
          <Input
            autoFocus
            onChange={(event) => {
              setDraft(event.target.value);
            }}
            onKeyDown={(event) => {
              if (event.key === "Enter") applyTitleFilter();
            }}
            placeholder="Content contains..."
            value={draft}
          />
        </Dialog.Body>
        <Dialog.Footer className="justify-end gap-3 border-0 pt-2">
          <Button
            color="tertiary"
            onClick={() => {
              onOpenChange(false);
            }}
            variant="outline"
          >
            Cancel
          </Button>
          <Button onClick={applyTitleFilter}>Apply</Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

export const StoriesFilterChip = ({
  chip,
  onEditTitle,
  onOperatorChange,
  onRemove,
  renderEditor,
}: {
  chip: FilterChip;
  onEditTitle: () => void;
  onOperatorChange: (operator: StoriesFilterOperator) => void;
  onRemove: () => void;
  renderEditor: (field: StoriesFilterField) => ReactNode;
}) => {
  const isEditable =
    chip.field !== "assignedToMe" &&
    chip.field !== "createdByMe" &&
    chip.field !== "hasNoAssignee";
  const shouldUseDialog = chip.field === "contentContains";
  const valueContent = (
    <div className="flex min-w-0 items-center truncate">{chip.value}</div>
  );
  let valueControl: ReactNode = (
    <div className="flex h-full items-center px-2.5">{chip.value}</div>
  );

  if (shouldUseDialog) {
    valueControl = (
      <button
        className="hover:bg-state-hover flex h-full max-w-72 items-center truncate px-2.5 text-left transition"
        onClick={onEditTitle}
        type="button"
      >
        {valueContent}
      </button>
    );
  } else if (isEditable) {
    valueControl = (
      <Popover>
        <Popover.Trigger asChild>
          <button
            className="hover:bg-state-hover flex h-full max-w-72 items-center truncate px-2.5 text-left transition"
            type="button"
          >
            {valueContent}
          </button>
        </Popover.Trigger>
        <Popover.Content
          align="start"
          className={getEditorContentClassName(chip.field)}
        >
          {renderEditor(chip.field)}
        </Popover.Content>
      </Popover>
    );
  }

  return (
    <Flex
      align="center"
      className="border-border bg-surface h-[2.1rem] shrink-0 overflow-hidden rounded-xl border"
      gap={0}
    >
      <span className="border-border text-text-secondary flex h-full items-center gap-1.5 border-r px-2.5">
        {chip.icon}
        {chip.label}
      </span>
      {chip.operatorOptions ? (
        <Menu>
          <Menu.Button>
            <button
              aria-label={`Change ${chip.label} filter operator`}
              className="hover:bg-state-hover border-border text-text-secondary flex h-[2.1rem] items-center border-r px-2.5 transition"
              type="button"
            >
              {chip.operator}
            </button>
          </Menu.Button>
          <Menu.Items align="start" className="w-44 p-1">
            {chip.operatorOptions.map((option) => (
              <Menu.Item
                active={chip.operator === option.label}
                className="justify-between"
                key={option.value}
                onSelect={() => {
                  onOperatorChange(option.value);
                }}
              >
                <span>{option.label}</span>
                {chip.operator === option.label ? (
                  <CheckIcon className="h-4 w-auto" />
                ) : null}
              </Menu.Item>
            ))}
          </Menu.Items>
        </Menu>
      ) : (
        <span className="border-border text-text-secondary flex h-full items-center border-r px-2.5">
          {chip.operator}
        </span>
      )}
      {valueControl}
      <button
        aria-label={`Remove ${chip.label} filter`}
        className="hover:bg-state-hover border-border flex h-full w-9 items-center justify-center border-l transition"
        onClick={onRemove}
        type="button"
      >
        <CloseIcon className="text-text-secondary h-3.5 w-auto" />
      </button>
    </Flex>
  );
};

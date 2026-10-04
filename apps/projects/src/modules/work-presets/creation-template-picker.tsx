"use client";

import { useState } from "react";
import { DocsIcon } from "icons";
import { Box, Button, Command, Popover, Text } from "ui";
import type { CreationTemplatePickerProps } from "@/shared/story/creation-property-slots";
import type { TaskTemplateConfiguration } from "./types";
import { useWorkPresets } from "./hooks";

const TeamCreationTemplatePicker = ({
  teamId,
  disabled = false,
  onSelect,
}: CreationTemplatePickerProps) => {
  const [open, setOpen] = useState(false);
  const [selectedTemplateId, setSelectedTemplateId] = useState<string | null>(
    null,
  );
  const query = useWorkPresets(teamId, "template", true);
  const templates = query.data?.pages.flatMap((page) => page.items) ?? [];
  const selectedTemplate = templates.find(
    ({ id }) => id === selectedTemplateId,
  );

  if (query.isPending || !templates.length) return null;

  return (
    <Popover onOpenChange={setOpen} open={open}>
      <Popover.Trigger asChild>
        <Button
          className="bg-surface-muted dark:bg-state-hover gap-1.5 border-transparent"
          color="tertiary"
          disabled={disabled || !teamId}
          leftIcon={<DocsIcon className="h-4 w-auto shrink-0" />}
          size="sm"
          title={selectedTemplate?.name}
          type="button"
          variant="solid"
        >
          <span className="max-w-[14rem] truncate">
            {selectedTemplate?.name ?? "Template"}
          </span>
        </Button>
      </Popover.Trigger>
      <Popover.Content
        align="start"
        className="w-72 max-w-[calc(100vw-2rem)] p-0"
      >
        <Command label="Task templates">
          <Command.Input
            aria-label="Search templates"
            autoFocus
            className="py-2 pr-3 text-base"
            disabled={disabled}
            placeholder="Apply template..."
          />
          <Command.Separator className="my-0" />
          <Command.List className="mt-0 max-h-64 w-full overflow-y-auto rounded-none border-0 bg-transparent py-1.5 shadow-none backdrop-blur-none dark:bg-transparent">
            <Command.Empty className="px-3 py-3 text-base">
              No matching templates.
            </Command.Empty>
            <Command.Group>
              {templates.map((template) => (
                <Command.Item
                  disabled={disabled}
                  key={template.id}
                  keywords={[template.name]}
                  onSelect={() => {
                    if (disabled) return;
                    onSelect(
                      template.configuration as TaskTemplateConfiguration,
                    );
                    setSelectedTemplateId(template.id);
                    setOpen(false);
                  }}
                  value={template.id}
                >
                  <DocsIcon className="h-4.5 w-auto shrink-0" />
                  <span className="truncate" title={template.name}>
                    {template.name}
                  </span>
                </Command.Item>
              ))}
            </Command.Group>
          </Command.List>
        </Command>
        {query.isError ? (
          <Box className="border-border space-y-1 border-t-[0.5px] p-2">
            <Text color="muted" role="alert">
              Could not load templates.
            </Text>
            <Button
              color="tertiary"
              disabled={disabled || query.isFetching}
              onClick={() => {
                void query.refetch();
              }}
              size="sm"
              type="button"
              variant="naked"
            >
              Try again
            </Button>
          </Box>
        ) : null}
        {query.hasNextPage && !query.isError ? (
          <Button
            className="border-border w-full rounded-none border-t-[0.5px]"
            color="tertiary"
            disabled={disabled || query.isFetchingNextPage}
            loading={query.isFetchingNextPage}
            loadingText="Loading..."
            onClick={() => {
              if (!query.isFetchingNextPage) void query.fetchNextPage();
            }}
            size="sm"
            type="button"
            variant="naked"
          >
            Load more
          </Button>
        ) : null}
      </Popover.Content>
    </Popover>
  );
};

export const CreationTemplatePicker = (props: CreationTemplatePickerProps) => (
  <TeamCreationTemplatePicker key={props.teamId} {...props} />
);

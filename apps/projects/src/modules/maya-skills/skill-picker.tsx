"use client";

import { useState } from "react";
import { SettingsIcon, WorkflowIcon } from "icons";
import { Box, Button, Command, Popover, Text } from "ui";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import type { MayaSkillSlotProps } from "@/shared/maya/skill-slot";
import { useMayaSkills } from "./hooks";
import { MayaSkillsDialog } from "./skills-dialog";
import { insertMayaSkillInstructions } from "./starters";

const WorkspaceSkillPicker = ({
  open,
  onOpenChange,
  value,
  onValueChange,
  onReturnFocus,
  disabled,
}: MayaSkillSlotProps) => {
  const [managing, setManaging] = useState(false);
  const query = useMayaSkills(open || managing);
  const skills = query.data ?? [];
  return (
    <>
      <Popover
        onOpenChange={(next) => {
          if (!disabled) onOpenChange(next);
        }}
        open={Boolean(open && !disabled)}
      >
        <Popover.Trigger asChild>
          <Button
            className="bg-state-hover dark:bg-state-hover gap-1"
            color="tertiary"
            disabled={disabled}
            leftIcon={<WorkflowIcon className="h-5 w-auto" />}
            rounded="md"
            type="button"
            variant="naked"
          >
            Skills
          </Button>
        </Popover.Trigger>
        <Popover.Content
          align="start"
          className="w-80 max-w-[calc(100vw-2rem)] p-0"
          onCloseAutoFocus={(event) => {
            if (!managing) {
              event.preventDefault();
              onReturnFocus();
            }
          }}
        >
          <Command label="Maya skills">
            <Command.Input
              aria-label="Search skills"
              autoFocus
              className="py-2 pr-3 text-base"
              placeholder="Search skills..."
            />
            <Command.Separator className="my-0" />
            <Command.List className="mt-0 max-h-64 w-full overflow-y-auto rounded-none border-0 bg-transparent py-1.5 shadow-none backdrop-blur-none dark:bg-transparent">
              {query.isPending ? (
                <Text aria-live="polite" className="px-3 py-3" color="muted">
                  Loading skills...
                </Text>
              ) : null}
              {!query.isPending && !query.isError ? (
                <Command.Empty className="px-3 py-3 text-base">
                  {skills.length
                    ? "No matching skills."
                    : "No saved skills yet."}
                </Command.Empty>
              ) : null}
              <Command.Group>
                {skills.map((skill) => (
                  <Command.Item
                    disabled={disabled}
                    key={skill.id}
                    keywords={[skill.name, skill.description]}
                    onSelect={() => {
                      if (disabled) return;
                      onValueChange(
                        insertMayaSkillInstructions(value, skill.instructions),
                      );
                      onOpenChange(false);
                    }}
                    value={skill.id}
                  >
                    <WorkflowIcon className="h-4.5 w-auto shrink-0" />
                    <Box className="min-w-0">
                      <Text className="truncate">{skill.name}</Text>
                      {skill.description ? (
                        <Text className="mt-0.5 line-clamp-2" color="muted">
                          {skill.description}
                        </Text>
                      ) : null}
                    </Box>
                  </Command.Item>
                ))}
              </Command.Group>
            </Command.List>
          </Command>
          {query.isError ? (
            <Box className="space-y-2 px-3 py-2">
              <Text color="danger" role="alert">
                Skills could not be loaded.
              </Text>
              <Button
                color="tertiary"
                disabled={query.isFetching}
                onClick={() => {
                  void query.refetch();
                }}
                size="sm"
                type="button"
                variant="outline"
              >
                Try again
              </Button>
            </Box>
          ) : null}
          <Button
            className="border-border w-full justify-start rounded-none border-t-[0.5px]"
            color="tertiary"
            leftIcon={<SettingsIcon className="h-4.5 w-auto" />}
            onClick={() => {
              setManaging(true);
              onOpenChange(false);
            }}
            size="sm"
            type="button"
            variant="naked"
          >
            Manage skills
          </Button>
        </Popover.Content>
      </Popover>
      {managing ? (
        <MayaSkillsDialog
          onClose={() => {
            setManaging(false);
            onReturnFocus();
          }}
        />
      ) : null}
    </>
  );
};

export const MayaSkillPicker = (props: MayaSkillSlotProps) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return (
    <WorkspaceSkillPicker
      key={`${workspaceSlug}:${session?.user.id ?? ""}`}
      {...props}
      disabled={props.disabled || !session?.user.id || !workspaceSlug}
    />
  );
};

"use client";

import { useState } from "react";
import { Button, Dialog, Input, Text } from "ui";
import { toast } from "sonner";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { AutomationSelect } from "./select-field";
import { buildRule } from "./configuration";
import type { AutomationInput, RuleConfiguration } from "./types";

const PRIORITIES = ["No Priority", "Low", "Medium", "High", "Urgent"].map(
  (value) => ({ value, label: value }),
);
export const RuleEditor = ({
  teamId,
  onClose,
  onSave,
}: {
  teamId: string;
  onClose: () => void;
  onSave: (input: AutomationInput) => Promise<unknown>;
}) => {
  const [name, setName] = useState("");
  const [trigger, setTrigger] =
    useState<RuleConfiguration["trigger"]>("story.created");
  const [statusCondition, setStatusCondition] = useState("any");
  const [priorityCondition, setPriorityCondition] = useState("any");
  const [statusAction, setStatusAction] = useState("unchanged");
  const [priorityAction, setPriorityAction] = useState("unchanged");
  const [assigneeAction, setAssigneeAction] = useState("unchanged");
  const [pending, setPending] = useState(false);
  const statuses = useTeamStatuses(teamId);
  const members = useTeamMembers(teamId);
  const statusOptions = (statuses.data ?? []).map((status) => ({
    value: status.id,
    label: status.name,
  }));
  const assigneeOptions = (members.data ?? [])
    .filter((member) => member.isActive && !member.isSystem)
    .map((member) => ({ value: member.id, label: member.fullName }));
  const hasAction =
    statusAction !== "unchanged" ||
    priorityAction !== "unchanged" ||
    assigneeAction !== "unchanged";
  const save = async () => {
    if (pending || !name.trim() || !hasAction) return;
    setPending(true);
    try {
      await onSave({
        teamId,
        kind: "rule",
        name: name.trim(),
        configuration: buildRule({
          trigger,
          statusCondition,
          priorityCondition,
          statusAction,
          priorityAction,
          assigneeAction,
        }),
      });
      onClose();
      toast.success("Rule created");
    } catch (error) {
      toast.error("Could not create rule", {
        description:
          error instanceof Error ? error.message : "Please try again.",
      });
    }
    setPending(false);
  };
  return (
    <Dialog
      onOpenChange={(open) => {
        if (!open && !pending) onClose();
      }}
      open
    >
      <Dialog.Content
        aria-busy={pending}
        className="mt-0 flex max-h-[calc(100dvh-2rem)] flex-col md:mt-0"
        hideClose={pending}
        onEscapeKeyDown={(event) => {
          if (pending) event.preventDefault();
        }}
        onInteractOutside={(event) => {
          if (pending) event.preventDefault();
        }}
        overlayClassName="items-center py-4"
      >
        <Dialog.Header className="shrink-0 px-6 py-4">
          <Dialog.Title className="pr-8 text-lg">
            Create workflow rule
          </Dialog.Title>
          <Dialog.Description className="mt-2 px-0 text-base leading-6">
            Automatically update tasks when your conditions are met.
          </Dialog.Description>
        </Dialog.Header>
        <Dialog.Body className="max-h-none min-h-0 flex-1 space-y-5">
          <Input
            autoFocus
            disabled={pending}
            id="automation-rule-name"
            label="Rule name"
            maxLength={100}
            onChange={(event) => {
              setName(event.target.value);
            }}
            placeholder="For example, assign urgent tasks"
            value={name}
          />
          <AutomationSelect
            disabled={pending}
            label="When"
            layout="inline"
            onChange={(value) => {
              setTrigger(value as RuleConfiguration["trigger"]);
            }}
            options={[
              { value: "story.created", label: "A task is created" },
              { value: "story.updated", label: "A task is updated" },
            ]}
            value={trigger}
          />
          <fieldset className="border-border min-w-0 border-t-[0.5px] pt-4">
            <legend className="float-left mb-1 w-full font-medium">If</legend>
            <Text className="clear-both mb-4" color="muted">
              All selected conditions match.
            </Text>
            <div className="space-y-3">
              <AutomationSelect
                disabled={pending || statuses.isPending}
                label="Status"
                layout="inline"
                onChange={setStatusCondition}
                options={[
                  { value: "any", label: "Any status" },
                  ...statusOptions,
                ]}
                value={statusCondition}
              />
              <AutomationSelect
                disabled={pending}
                label="Priority"
                layout="inline"
                onChange={setPriorityCondition}
                options={[
                  { value: "any", label: "Any priority" },
                  ...PRIORITIES,
                ]}
                value={priorityCondition}
              />
            </div>
          </fieldset>
          <fieldset className="border-border min-w-0 border-t-[0.5px] pt-4">
            <legend className="float-left mb-1 w-full font-medium">Then</legend>
            <Text className="clear-both mb-4" color="muted">
              Apply these changes to the task.
            </Text>
            <div className="space-y-3">
              <AutomationSelect
                disabled={pending || statuses.isPending}
                label="Set status"
                layout="inline"
                onChange={setStatusAction}
                options={[
                  { value: "unchanged", label: "Keep current status" },
                  ...statusOptions,
                ]}
                value={statusAction}
              />
              <AutomationSelect
                disabled={pending}
                label="Set priority"
                layout="inline"
                onChange={setPriorityAction}
                options={[
                  { value: "unchanged", label: "Keep current priority" },
                  ...PRIORITIES,
                ]}
                value={priorityAction}
              />
              <AutomationSelect
                disabled={pending || members.isPending}
                label="Assign to"
                layout="inline"
                onChange={setAssigneeAction}
                options={[
                  { value: "unchanged", label: "Keep current assignee" },
                  { value: "unassigned", label: "Unassigned" },
                  ...assigneeOptions,
                ]}
                value={assigneeAction}
              />
            </div>
          </fieldset>
          {statuses.isError || members.isError ? (
            <Text color="danger" role="alert">
              Could not load task properties. Close and retry.
            </Text>
          ) : null}
          <Text color="muted">
            {hasAction
              ? "This rule runs after a task is saved."
              : "Choose at least one change to create this rule."}
          </Text>
        </Dialog.Body>
        <Dialog.Footer className="shrink-0 flex-wrap justify-end gap-2 py-4">
          <Button
            color="tertiary"
            disabled={pending}
            onClick={onClose}
            variant="outline"
          >
            Cancel
          </Button>
          <Button
            disabled={
              pending ||
              !name.trim() ||
              !hasAction ||
              statuses.isPending ||
              members.isPending ||
              statuses.isError ||
              members.isError
            }
            loading={pending}
            loadingText="Creating..."
            onClick={() => {
              void save();
            }}
          >
            Create workflow rule
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

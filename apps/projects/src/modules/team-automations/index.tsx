"use client";

import { useRef, useState } from "react";
import { ArrowDownIcon, RefreshIcon, WorkflowIcon } from "icons";
import { Box, Button, Flex, Menu, Switch, Text } from "ui";
import { toast } from "sonner";
import { useUserRole } from "@/hooks/role";
import { SectionHeader } from "@/components/ui/section-header";
import { openDialogAfterMenuClose } from "@/utils/menu-dialog-state";
import { useAutomations } from "./hooks";
import { RuleEditor } from "./rule-editor";
import { RecurrenceEditor } from "./recurrence-editor";
import { AutomationDetails } from "./details";
import { AutomationRunHistory } from "./run-history";
import type { Automation } from "./types";

export const TeamAutomations = ({ teamId }: { teamId: string }) => {
  const [editor, setEditor] = useState<"rule" | "recurrence" | null>(null);
  const createTrigger = useRef<HTMLButtonElement>(null);
  const [history, setHistory] = useState<Automation | null>(null);
  const query = useAutomations(teamId);
  const { userRole } = useUserRole();
  const canCreate = userRole === "admin" || userRole === "member";
  const creationDisabled = !canCreate || query.create.isPending;
  const closeEditor = () => {
    setEditor(null);
    // These standalone editors have no Dialog.Trigger to restore focus to.
    setTimeout(() => {
      createTrigger.current?.focus();
    }, 0);
  };
  const modify = async (
    automation: Automation,
    action: "pause" | "archive",
    paused = false,
  ) => {
    try {
      if (action === "archive") await query.archive.mutateAsync(automation.id);
      else await query.pause.mutateAsync({ id: automation.id, paused });
      let message = paused ? "Automation paused" : "Automation resumed";
      if (action === "archive") message = "Automation archived";
      toast.success(message);
    } catch (error) {
      toast.error("Could not update automation", {
        description:
          error instanceof Error ? error.message : "Please try again.",
      });
    }
  };
  return (
    <>
      <Box className="border-border bg-surface rounded-2xl border">
        <SectionHeader
          action={
            <Menu>
              <Menu.Button disabled={creationDisabled}>
                <Button
                  color="tertiary"
                  disabled={creationDisabled}
                  ref={createTrigger}
                  rightIcon={
                    <ArrowDownIcon
                      aria-hidden
                      className="h-3.5 w-auto shrink-0"
                    />
                  }
                  type="button"
                  variant="outline"
                >
                  Create automation
                </Button>
              </Menu.Button>
              <Menu.Items align="end" className="min-w-56">
                <Menu.Group>
                  <Menu.Item
                    disabled={creationDisabled}
                    onSelect={() => {
                      if (!creationDisabled)
                        openDialogAfterMenuClose(() => {
                          setEditor("rule");
                        });
                    }}
                  >
                    <WorkflowIcon aria-hidden className="h-5 w-auto" />
                    Create rule
                  </Menu.Item>
                  <Menu.Item
                    disabled={creationDisabled}
                    onSelect={() => {
                      if (!creationDisabled)
                        openDialogAfterMenuClose(() => {
                          setEditor("recurrence");
                        });
                    }}
                  >
                    <RefreshIcon aria-hidden className="h-5 w-auto" />
                    Create recurring task
                  </Menu.Item>
                </Menu.Group>
              </Menu.Items>
            </Menu>
          }
          description="Save team rules and recurring work, then monitor their runs."
          title="Rules and recurring tasks"
        />
        {query.isPending ? (
          <Text className="px-6 py-5" color="muted">
            Loading automations...
          </Text>
        ) : null}
        {query.isError ? (
          <Box className="px-6 py-5">
            <Text role="alert">Could not load automations.</Text>
            <Button
              color="tertiary"
              onClick={() => {
                void query.refetch();
              }}
              variant="naked"
            >
              Try again
            </Button>
          </Box>
        ) : null}
        {!query.isPending && !query.isError && !query.data.length ? (
          <Text className="px-6 py-5" color="muted">
            Create a rule to route matching tasks, or a recurring task for work
            that repeats.
          </Text>
        ) : null}
        <div className="divide-border divide-y">
          {query.data?.map((automation) => (
            <Box className="px-6 py-5" key={automation.id}>
              <Flex align="start" className="gap-4" justify="between">
                <Box className="min-w-0">
                  <Text fontWeight="medium">{automation.name}</Text>
                  <Text className="mt-1" color="muted">
                    {automation.kind === "rule"
                      ? "Team rule"
                      : "Recurring task"}{" "}
                    · {automation.paused ? "Paused" : "Active"}
                  </Text>
                  {automation.nextRunAt && !automation.paused ? (
                    <Text className="mt-1" color="muted">
                      Next run:{" "}
                      {new Date(automation.nextRunAt).toLocaleString()}
                    </Text>
                  ) : null}
                </Box>
                {automation.canEdit ? (
                  <Switch
                    aria-label={`Enable ${automation.name}`}
                    checked={!automation.paused}
                    disabled={query.pause.isPending || query.archive.isPending}
                    onCheckedChange={(checked) => {
                      void modify(automation, "pause", !checked);
                    }}
                  />
                ) : null}
              </Flex>
              <AutomationDetails automation={automation} />
              {automation.lastError ? (
                <Text className="mt-3" color="danger" role="alert">
                  {automation.lastError}
                </Text>
              ) : null}
              <Flex className="mt-3 gap-3">
                <Button
                  color="tertiary"
                  onClick={() => {
                    setHistory(automation);
                  }}
                  variant="naked"
                >
                  View runs
                </Button>
                {automation.canEdit ? (
                  <Button
                    color="tertiary"
                    disabled={query.archive.isPending}
                    onClick={() => {
                      void modify(automation, "archive");
                    }}
                    variant="naked"
                  >
                    Archive
                  </Button>
                ) : null}
              </Flex>
            </Box>
          ))}
        </div>
      </Box>
      {editor === "rule" ? (
        <RuleEditor
          onClose={closeEditor}
          onSave={(input) => query.create.mutateAsync(input)}
          teamId={teamId}
        />
      ) : null}
      {editor === "recurrence" ? (
        <RecurrenceEditor
          onClose={closeEditor}
          onSave={(input) => query.create.mutateAsync(input)}
          teamId={teamId}
        />
      ) : null}
      {history ? (
        <AutomationRunHistory
          id={history.id}
          name={history.name}
          onClose={() => {
            setHistory(null);
          }}
        />
      ) : null}
    </>
  );
};

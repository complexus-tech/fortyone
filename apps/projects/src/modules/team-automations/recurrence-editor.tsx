"use client";

import { useState } from "react";
import { Button, Dialog, Input, Text, TextArea } from "ui";
import { toast } from "sonner";
import { format } from "date-fns";
import {
  CreateCustomFields,
  useCreateCustomFields,
} from "@/modules/custom-fields/public/creation";
import { PresetPicker } from "@/modules/work-presets/public/template-picker";
import type { TaskTemplateConfiguration } from "@/modules/work-presets/public/types";
import { AutomationSelect } from "./select-field";
import { draftFromTemplate } from "./configuration";
import type {
  AutomationDraft,
  AutomationInput,
  RecurrenceConfiguration,
} from "./types";

const WEEKDAYS = [
  "Sunday",
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
].map((label, value) => ({ value: String(value), label }));
const descriptionHTML = (value: string) =>
  `<p>${value.replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;").replaceAll("\n", "</p><p>")}</p>`;
export const RecurrenceEditor = ({
  teamId,
  onClose,
  onSave,
}: {
  teamId: string;
  onClose: () => void;
  onSave: (input: AutomationInput) => Promise<unknown>;
}) => {
  const [name, setName] = useState("");
  const [pending, setPending] = useState(false);
  const [draft, setDraft] = useState<AutomationDraft>({
    title: "",
    description: "",
    descriptionHTML: "",
    priority: "No Priority",
    checklist: [],
  });
  const [schedule, setSchedule] = useState<RecurrenceConfiguration["schedule"]>(
    () => ({
      frequency: "weekly",
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC",
      startsOn: format(new Date(), "yyyy-MM-dd"),
      localTime: "09:00",
      weekday: new Date().getDay(),
      monthDay: new Date().getDate(),
    }),
  );
  const fields = useCreateCustomFields(teamId);
  const save = async () => {
    setPending(true);
    try {
      await onSave({
        teamId,
        kind: "recurrence",
        name: name.trim() || draft.title.slice(0, 100),
        configuration: {
          version: 1,
          schedule,
          draft: {
            ...draft,
            title: draft.title.trim(),
            customFieldValues: fields.prepareValues(),
          },
        },
      });
      onClose();
      toast.success("Recurring task created");
    } catch (error) {
      toast.error("Could not create recurring task", {
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
        className="mt-0 flex max-h-[calc(100dvh-2rem)] max-w-4xl flex-col md:mt-0"
        hideClose={pending}
        onEscapeKeyDown={(event) => {
          if (pending) event.preventDefault();
        }}
        onInteractOutside={(event) => {
          if (pending) event.preventDefault();
        }}
        overlayClassName="items-center py-4"
        size="lg"
      >
        <Dialog.Header className="shrink-0 px-6 py-5">
          <Dialog.Title className="text-lg">Create recurring task</Dialog.Title>
          <Dialog.Description className="mt-2 px-0 text-base leading-6">
            Create new tasks on a schedule.
          </Dialog.Description>
        </Dialog.Header>
        <Dialog.Body className="grid max-h-none min-h-0 flex-1 gap-6 md:grid-cols-2">
          <section
            aria-labelledby="recurrence-task-details"
            className="min-w-0 space-y-4"
          >
            <div className="flex min-h-10 flex-wrap items-center justify-between gap-3">
              <Text as="h3" fontWeight="medium" id="recurrence-task-details">
                Task details
              </Text>
              <PresetPicker
                disabled={pending}
                hideWhenEmpty
                kind="template"
                label="Use task template"
                onSelect={(preset) => {
                  if (preset.kind !== "template") return;
                  const template =
                    preset.configuration as TaskTemplateConfiguration;
                  setDraft(draftFromTemplate(template));
                  fields.setValues(
                    Object.fromEntries(
                      (template.customFieldValues ?? []).map((field) => [
                        field.fieldId,
                        field.value,
                      ]),
                    ),
                  );
                }}
                teamId={teamId}
              />
            </div>
            <Input
              autoFocus
              className="h-10 px-3 text-base leading-6"
              id="automation-task-title"
              label="Task title"
              labelClassName="mb-2"
              maxLength={255}
              onChange={(event) => {
                setDraft({ ...draft, title: event.target.value });
              }}
              placeholder="For example, weekly release review"
              value={draft.title}
            />
            <div className="space-y-2">
              <label className="block" htmlFor="automation-description">
                Description
              </label>
              <TextArea
                className="rounded-lg px-3 py-2 text-base leading-6"
                id="automation-description"
                maxLength={20000}
                onChange={(event) => {
                  setDraft({
                    ...draft,
                    description: event.target.value,
                    descriptionHTML: descriptionHTML(event.target.value),
                  });
                }}
                rows={3}
                value={draft.description}
              />
            </div>
            <CreateCustomFields
              disabled={pending}
              onChange={fields.setValues}
              teamId={teamId}
              values={fields.values}
            />
          </section>
          <section
            aria-labelledby="recurrence-schedule"
            className="min-w-0 space-y-4"
          >
            <div className="flex min-h-10 items-center">
              <Text as="h3" fontWeight="medium" id="recurrence-schedule">
                Schedule
              </Text>
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              <AutomationSelect
                label="Repeat"
                onChange={(value) => {
                  setSchedule({
                    ...schedule,
                    frequency:
                      value as RecurrenceConfiguration["schedule"]["frequency"],
                  });
                }}
                options={[
                  { value: "daily", label: "Daily" },
                  { value: "weekly", label: "Weekly" },
                  { value: "monthly", label: "Monthly" },
                ]}
                value={schedule.frequency}
              />
              <Input
                className="h-10 px-3 text-base leading-6"
                id="automation-local-time"
                label="Time"
                labelClassName="mb-2"
                onChange={(event) => {
                  setSchedule({ ...schedule, localTime: event.target.value });
                }}
                type="time"
                value={schedule.localTime}
              />
              {schedule.frequency === "weekly" ? (
                <AutomationSelect
                  label="Day of week"
                  onChange={(value) => {
                    setSchedule({ ...schedule, weekday: Number(value) });
                  }}
                  options={WEEKDAYS}
                  value={String(schedule.weekday)}
                />
              ) : null}
              {schedule.frequency === "monthly" ? (
                <Input
                  className="h-10 px-3 text-base leading-6"
                  id="automation-month-day"
                  label="Day of month"
                  labelClassName="mb-2"
                  max={31}
                  min={1}
                  onChange={(event) => {
                    setSchedule({
                      ...schedule,
                      monthDay: Number(event.target.value),
                    });
                  }}
                  type="number"
                  value={schedule.monthDay}
                />
              ) : null}
              <div
                className={
                  schedule.frequency === "daily" ? "sm:col-span-2" : undefined
                }
              >
                <Input
                  className="h-10 px-3 text-base leading-6"
                  id="automation-start-date"
                  label="Start date"
                  labelClassName="mb-2"
                  onChange={(event) => {
                    setSchedule({ ...schedule, startsOn: event.target.value });
                  }}
                  type="date"
                  value={schedule.startsOn}
                />
              </div>
            </div>
            <Input
              className="h-10 px-3 text-base leading-6"
              id="automation-timezone"
              label="Timezone"
              labelClassName="mb-2"
              maxLength={100}
              onChange={(event) => {
                setSchedule({ ...schedule, timezone: event.target.value });
              }}
              placeholder="Africa/Harare"
              value={schedule.timezone}
            />
            <Input
              className="h-10 px-3 text-base leading-6"
              id="automation-schedule-name"
              label="Schedule name (optional)"
              labelClassName="mb-2"
              maxLength={100}
              onChange={(event) => {
                setName(event.target.value);
              }}
              placeholder={draft.title || "Name this recurring task"}
              value={name}
            />
          </section>
          <Text className="md:col-span-2" color="muted">
            Each occurrence creates a new task. Short months use their final
            day. Resuming skips missed dates; daylight saving changes keep the
            selected local time.
          </Text>
        </Dialog.Body>
        <Dialog.Footer className="shrink-0 flex-wrap justify-end gap-3 py-4">
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
              !draft.title.trim() ||
              !schedule.timezone.trim() ||
              !schedule.startsOn ||
              !schedule.localTime ||
              fields.isPending ||
              fields.isError
            }
            loading={pending}
            loadingText="Creating..."
            onClick={() => {
              void save();
            }}
          >
            Create recurring task
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

"use client";

import { useMemo, useState } from "react";
import { ArrowRight2Icon, CalendarIcon } from "icons";
import { Badge, Button, Dialog, Flex, Text, TextEditor } from "ui";
import { toast } from "sonner";
import { format } from "date-fns";
import { useNewStoryDialogEditors } from "@/components/ui/use-new-story-dialog-editors";
import { StoryComposerHeader } from "@/components/ui/story-composer-header";
import { getPersistableRichTextContent } from "@/lib/tiptap/rich-text-media";
import { RichTextTableMenu } from "@/lib/tiptap/rich-text-table-menu";
import {
  CreateCustomFields,
  useCreateCustomFields,
} from "@/modules/custom-fields/public/creation";
import { CreationTemplatePicker } from "@/modules/work-presets/public/template-picker";
import type { TaskTemplateConfiguration } from "@/modules/work-presets/public/types";
import { templateDescriptionHTML } from "@/shared/story/task-template";
import { draftFromTemplate } from "./configuration";
import { RecurrenceProperties } from "./recurrence-properties";
import {
  RecurrenceSchedule,
  validateRecurrenceSchedule,
} from "./recurrence-schedule";
import type {
  AutomationDraft,
  AutomationInput,
  RecurrenceConfiguration,
} from "./types";

const validateDraft = (draft: AutomationDraft) => {
  if (draft.title.trim().length > 255) {
    return "Keep the task title to 255 characters or fewer.";
  }
  if (
    draft.description.length > 20000 ||
    draft.descriptionHTML.length > 40000
  ) {
    return "Shorten the task description before saving.";
  }
  return null;
};

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
  const { titleEditor, descriptionEditor } = useNewStoryDialogEditors({
    editable: !pending,
    onStoryTitleChange: (title) => {
      setDraft((current) => ({ ...current, title }));
    },
    onDescriptionChange: (description) => {
      setDraft((current) => ({ ...current, ...description }));
    },
    storyTerm: "Task",
  });
  const scheduleValidation = useMemo(
    () => validateRecurrenceSchedule(schedule),
    [schedule],
  );
  const { canonicalTimezone } = scheduleValidation;
  const validationError = validateDraft(draft) ?? scheduleValidation.error;
  const canSave = Boolean(
    !pending &&
      titleEditor &&
      descriptionEditor &&
      draft.title.trim() &&
      schedule.timezone.trim() &&
      schedule.startsOn &&
      schedule.localTime &&
      !fields.isPending &&
      !fields.isError &&
      !validationError,
  );

  const applyTemplate = (template: TaskTemplateConfiguration) => {
    if (pending || !titleEditor || !descriptionEditor) return;
    const descriptionHTML = templateDescriptionHTML(template);
    setDraft({
      ...draftFromTemplate(template),
      descriptionHTML,
      // The composer contains the template checklist, so creation must not append it again.
      checklist: [],
    });
    titleEditor.commands.setContent({
      type: "doc",
      content: [
        {
          type: "paragraph",
          content: template.title
            ? [{ type: "text", text: template.title }]
            : [],
        },
      ],
    });
    descriptionEditor.commands.setContent(descriptionHTML);
    fields.setValues(
      Object.fromEntries(
        (template.customFieldValues ?? []).map((field) => [
          field.fieldId,
          field.value,
        ]),
      ),
    );
  };

  const save = async () => {
    if (!canSave || !titleEditor || !descriptionEditor || !canonicalTimezone)
      return;
    setPending(true);
    const title = titleEditor.getText().trim();
    const description = getPersistableRichTextContent(descriptionEditor);
    try {
      await onSave({
        teamId,
        kind: "recurrence",
        name: name.trim() || title.slice(0, 100),
        configuration: {
          version: 1,
          schedule: { ...schedule, timezone: canonicalTimezone },
          draft: {
            ...draft,
            title,
            description: description.contentText,
            descriptionHTML: description.contentHtml,
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
        aria-label="Create recurring task"
        aria-labelledby={undefined}
        className="mt-4 flex max-h-[calc(100dvh-2rem)] flex-col overflow-visible md:mt-[10%] md:max-h-[calc(90dvh-1rem)]"
        hideClose
        onEscapeKeyDown={(event) => {
          if (pending) event.preventDefault();
        }}
        onInteractOutside={(event) => {
          if (pending) event.preventDefault();
        }}
        size="lg"
      >
        <StoryComposerHeader
          actions={
            <Flex className="shrink-0" gap={2}>
              {pending ? null : <Dialog.Close />}
            </Flex>
          }
          templatePicker={
            <CreationTemplatePicker
              disabled={pending || !titleEditor || !descriptionEditor}
              onSelect={applyTemplate}
              teamId={teamId}
            />
          }
          title={
            <>
              <Badge className="dark:bg-surface-elevated/90" color="tertiary">
                <CalendarIcon className="h-4 w-auto" />
              </Badge>
              <ArrowRight2Icon
                className="h-4.5 w-auto opacity-30"
                strokeWidth={3}
              />
              <Text className="opacity-80" color="muted">
                New recurring task
              </Text>
            </>
          }
        />
        <Dialog.Description className="sr-only">
          Compose the task and choose when to create each occurrence.
        </Dialog.Description>
        <Dialog.Body className="max-h-[75dvh] min-h-0 flex-1 space-y-5 overflow-y-auto pt-0 pb-6 md:max-h-[60dvh]">
          <section aria-label="Task details" className="min-w-0">
            <TextEditor
              asTitle
              className="text-2xl font-medium"
              editor={titleEditor}
            />
            <TextEditor
              className="rich-document-editor min-h-20"
              editor={descriptionEditor}
              hideBubbleMenu={pending}
            />
            <RichTextTableMenu editor={descriptionEditor} scrollTarget={null} />
            <RecurrenceProperties
              disabled={pending}
              draft={draft}
              onChange={(partial) => {
                setDraft((current) => ({ ...current, ...partial }));
              }}
              teamId={teamId}
            >
              <CreateCustomFields
                disabled={pending}
                onChange={fields.setValues}
                teamId={teamId}
                values={fields.values}
              />
            </RecurrenceProperties>
          </section>
          <RecurrenceSchedule
            disabled={pending}
            name={name}
            onChange={(partial) => {
              setSchedule((current) => ({ ...current, ...partial }));
            }}
            onNameChange={setName}
            schedule={schedule}
            validation={scheduleValidation}
          />
          {validationError ? (
            <Text color="danger" role="alert">
              {validationError}
            </Text>
          ) : null}
        </Dialog.Body>
        <Dialog.Footer className="flex shrink-0 items-center justify-between gap-2">
          <Button
            color="tertiary"
            disabled={pending}
            onClick={onClose}
            variant="outline"
          >
            Cancel
          </Button>
          <Button
            disabled={!canSave}
            leftIcon={<CalendarIcon className="h-4 w-auto" />}
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

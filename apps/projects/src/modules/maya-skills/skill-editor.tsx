import type { FormEvent } from "react";
import { useId, useState } from "react";
import { Button, Dialog, Input, Text, TextArea } from "ui";
import { toast } from "sonner";
import { useMayaSkillMutations } from "./hooks";
import { mayaSkillInputSchema } from "./types";
import type { MayaSkill, MayaSkillInput } from "./types";

export const MayaSkillEditor = ({
  skill,
  initial,
  onClose,
}: {
  skill: MayaSkill | null;
  initial?: MayaSkillInput;
  onClose: () => void;
}) => {
  const [draft, setDraft] = useState<MayaSkillInput>(
    skill ?? initial ?? { name: "", description: "", instructions: "" },
  );
  const [error, setError] = useState<string | null>(null);
  const errorId = useId();
  const { create, update } = useMayaSkillMutations();
  const pending = create.isPending || update.isPending;
  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (pending) return;
    const parsed = mayaSkillInputSchema.safeParse(draft);
    if (!parsed.success) {
      setError(parsed.error.issues[0].message);
      return;
    }
    setError(null);
    try {
      if (skill)
        await update.mutateAsync({
          id: skill.id,
          input: { ...parsed.data, updatedAt: skill.updatedAt },
        });
      else await create.mutateAsync(parsed.data);
      toast.success(skill ? "Skill updated" : "Skill created");
      onClose();
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "The skill could not be saved. Please try again.",
      );
    }
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
        className="flex max-h-[calc(100dvh-4rem)] flex-col"
        hideClose={pending}
        onEscapeKeyDown={(event) => {
          if (pending) event.preventDefault();
        }}
        onInteractOutside={(event) => {
          if (pending) event.preventDefault();
        }}
      >
        <Dialog.Header className="shrink-0 px-6 py-4">
          <Dialog.Title className="pr-8 text-lg">
            {skill ? "Edit skill" : "Create skill"}
          </Dialog.Title>
          <Dialog.Description className="mt-2 px-0">
            Save instructions you can reuse in Maya. Skills are private to you
            in this workspace.
          </Dialog.Description>
        </Dialog.Header>
        <form className="flex min-h-0 flex-1 flex-col" onSubmit={submit}>
          <Dialog.Body className="max-h-none min-h-0 flex-1 space-y-4">
            <Input
              aria-describedby={error ? errorId : undefined}
              autoFocus
              disabled={pending}
              label="Skill name"
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  name: event.target.value,
                }));
              }}
              placeholder="e.g. Weekly update"
              value={draft.name}
            />
            <Input
              disabled={pending}
              label="Description (optional)"
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  description: event.target.value,
                }));
              }}
              placeholder="When to use this skill"
              value={draft.description}
            />
            <TextArea
              aria-describedby={error ? errorId : undefined}
              className="min-h-48 resize-y py-3 leading-6"
              disabled={pending}
              label="Instructions"
              onChange={(event) => {
                setDraft((current) => ({
                  ...current,
                  instructions: event.target.value,
                }));
              }}
              placeholder="Tell Maya what to review, how to work, and what to return."
              rows={8}
              value={draft.instructions}
            />
            {error ? (
              <Text color="danger" id={errorId} role="alert">
                {error}
              </Text>
            ) : null}
          </Dialog.Body>
          <Dialog.Footer className="flex shrink-0 justify-end gap-2">
            <Button
              color="tertiary"
              disabled={pending}
              onClick={onClose}
              type="button"
              variant="outline"
            >
              Cancel
            </Button>
            <Button
              disabled={
                pending || !draft.name.trim() || !draft.instructions.trim()
              }
              loading={pending}
              loadingText="Saving..."
              type="submit"
            >
              {skill ? "Save skill" : "Create skill"}
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};

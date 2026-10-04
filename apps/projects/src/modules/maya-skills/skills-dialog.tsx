import { useState } from "react";
import { PlusIcon } from "icons";
import { Box, Button, Dialog, Flex, Input, Text } from "ui";
import { toast } from "sonner";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { useMayaSkillMutations, useMayaSkills } from "./hooks";
import { MayaSkillEditor } from "./skill-editor";
import { MayaSkillRow } from "./skill-row";
import { MAYA_SKILL_STARTERS } from "./starters";
import type { MayaSkill, MayaSkillInput } from "./types";

export const MayaSkillsDialog = ({ onClose }: { onClose: () => void }) => {
  const query = useMayaSkills();
  const { remove } = useMayaSkillMutations();
  const [search, setSearch] = useState("");
  const [editor, setEditor] = useState<{
    skill: MayaSkill | null;
    initial?: MayaSkillInput;
  } | null>(null);
  const [deleting, setDeleting] = useState<MayaSkill | null>(null);
  const skills = query.data ?? [];
  const filtered = skills.filter((skill) =>
    `${skill.name} ${skill.description}`
      .toLocaleLowerCase()
      .includes(search.trim().toLocaleLowerCase()),
  );
  const dismiss = () => {
    if (!editor && !deleting) onClose();
  };
  const confirmDelete = async () => {
    if (!deleting || remove.isPending) return;
    try {
      await remove.mutateAsync(deleting.id);
      setDeleting(null);
      toast.success("Skill deleted");
    } catch {
      // The confirmation remains open and shows the mutation error.
    }
  };
  return (
    <>
      <Dialog
        onOpenChange={(open) => {
          if (!open) dismiss();
        }}
        open
      >
        <Dialog.Content
          className="flex max-h-[calc(100dvh-4rem)] flex-col"
          hideClose={Boolean(editor || deleting)}
        >
          <Dialog.Header className="shrink-0 px-6 py-4">
            <Dialog.Title className="pr-8 text-lg">Maya skills</Dialog.Title>
            <Dialog.Description className="mt-2 px-0">
              Reusable instructions, private to you in this workspace. Choose a
              skill in chat, review the prompt, then send it.
            </Dialog.Description>
          </Dialog.Header>
          <Dialog.Body className="max-h-none min-h-0 flex-1 space-y-4">
            <Flex align="center" className="gap-3" justify="between">
              <Input
                aria-label="Search your skills"
                className="min-w-0 flex-1"
                onChange={(event) => {
                  setSearch(event.target.value);
                }}
                placeholder="Search skills..."
                value={search}
              />
              <Button
                color="tertiary"
                leftIcon={<PlusIcon />}
                onClick={() => {
                  setEditor({ skill: null });
                }}
                type="button"
              >
                Create skill
              </Button>
            </Flex>
            {query.isPending ? (
              <Text aria-live="polite" color="muted">
                Loading skills...
              </Text>
            ) : null}
            {query.isError ? (
              <Box className="space-y-2">
                <Text color="danger" role="alert">
                  Skills could not be loaded.
                </Text>
                <Button
                  color="tertiary"
                  disabled={query.isFetching}
                  onClick={() => {
                    void query.refetch();
                  }}
                  type="button"
                  variant="outline"
                >
                  Try again
                </Button>
              </Box>
            ) : null}
            {filtered.length > 0 ? (
              <Box className="divide-border divide-y-[0.5px]">
                {filtered.map((skill) => (
                  <MayaSkillRow
                    editDisabled={query.isFetching || query.isError}
                    key={skill.id}
                    onDelete={(current) => {
                      remove.reset();
                      setDeleting(current);
                    }}
                    onEdit={(current) => {
                      setEditor({ skill: current });
                    }}
                    skill={skill}
                  />
                ))}
              </Box>
            ) : null}
            {!filtered.length && !query.isPending && !query.isError ? (
              <Text color="muted">
                {skills.length
                  ? "No matching skills."
                  : "No saved skills yet. Create your own or start with an example below."}
              </Text>
            ) : null}
            <Box className="border-border space-y-3 border-t-[0.5px] pt-4">
              <Text fontWeight="medium">Start with an example</Text>
              {MAYA_SKILL_STARTERS.map((starter) => (
                <Flex
                  align="center"
                  className="gap-3"
                  justify="between"
                  key={starter.name}
                >
                  <Box className="min-w-0 flex-1">
                    <Text>{starter.name}</Text>
                    <Text className="mt-1" color="muted">
                      {starter.description}
                    </Text>
                  </Box>
                  <Button
                    aria-label={`Use example: ${starter.name}`}
                    color="tertiary"
                    onClick={() => {
                      setEditor({ skill: null, initial: starter });
                    }}
                    size="sm"
                    type="button"
                    variant="outline"
                  >
                    Use example
                  </Button>
                </Flex>
              ))}
            </Box>
          </Dialog.Body>
          <Dialog.Footer className="shrink-0 justify-end">
            <Button
              color="tertiary"
              onClick={dismiss}
              type="button"
              variant="outline"
            >
              Done
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
      {editor ? (
        <MayaSkillEditor
          {...editor}
          onClose={() => {
            setEditor(null);
          }}
        />
      ) : null}
      {deleting ? (
        <ConfirmDialog
          confirmText="Delete skill"
          description={`Delete “${deleting.name}”? Existing chats keep the instructions you already sent.`}
          errorMessage={remove.error?.message}
          isLoading={remove.isPending}
          isOpen
          loadingText="Deleting..."
          onClose={() => {
            if (!remove.isPending) setDeleting(null);
          }}
          onConfirm={() => {
            void confirmDelete();
          }}
          title="Delete skill"
        />
      ) : null}
    </>
  );
};

"use client";

import { useId, useState } from "react";
import { Button, Dialog, Input, Select, Text } from "ui";
import { toast } from "sonner";

export const PresetNameDialog = ({
  title,
  initialName = "",
  open,
  onOpenChange,
  onSave,
  shared = false,
  onReturnFocus,
}: {
  title: string;
  initialName?: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (name: string, visibility: "personal" | "team") => Promise<unknown>;
  shared?: boolean;
  onReturnFocus?: () => void;
}) => {
  const visibilityId = useId();
  const [name, setName] = useState(initialName);
  const [visibility, setVisibility] = useState<"personal" | "team">("personal");
  const [pending, setPending] = useState(false);
  const save = async () => {
    if (!name.trim() || pending) return;
    setPending(true);
    try {
      await onSave(name.trim(), visibility);
      onOpenChange(false);
      toast.success("Saved successfully");
    } catch (error) {
      toast.error("Could not save", {
        description:
          error instanceof Error ? error.message : "Please try again.",
      });
    }
    setPending(false);
  };
  return (
    <Dialog
      onOpenChange={(nextOpen) => {
        if (!pending) onOpenChange(nextOpen);
      }}
      open={open}
    >
      <Dialog.Content
        aria-busy={pending}
        className="flex max-h-[calc(100dvh-15vw-1rem)] flex-col md:max-h-[calc(100dvh-10vw-1rem)]"
        hideClose={pending}
        onCloseAutoFocus={(event) => {
          if (!onReturnFocus) return;
          event.preventDefault();
          onReturnFocus();
        }}
        onEscapeKeyDown={(event) => {
          if (pending) event.preventDefault();
        }}
        onInteractOutside={(event) => {
          if (pending) event.preventDefault();
        }}
        size="sm"
      >
        <Dialog.Header className="shrink-0 px-6 py-5">
          <Dialog.Title className="text-lg">{title}</Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="max-h-none min-h-0 flex-1">
          <label className="mb-2 block" htmlFor="preset-name">
            Name
          </label>
          <Input
            autoFocus
            disabled={pending}
            id="preset-name"
            maxLength={100}
            onChange={(event) => {
              setName(event.target.value);
            }}
            placeholder="Give it a clear name"
            value={name}
          />
          {shared ? (
            <>
              <label className="mt-5 mb-2 block" htmlFor={visibilityId}>
                Visibility
              </label>
              <Select
                disabled={pending}
                onValueChange={(value: "personal" | "team") => {
                  setVisibility(value);
                }}
                value={visibility}
              >
                <Select.Trigger
                  className="h-[2.1rem] w-full px-3 text-base"
                  id={visibilityId}
                >
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  <Select.Option className="py-2 text-base" value="personal">
                    Only me
                  </Select.Option>
                  <Select.Option className="py-2 text-base" value="team">
                    Everyone in this team
                  </Select.Option>
                </Select.Content>
              </Select>
              <Text className="mt-2" color="muted">
                Save a reusable copy for your team.
              </Text>
            </>
          ) : null}
        </Dialog.Body>
        <Dialog.Footer className="shrink-0 justify-end gap-3 py-4">
          <Button
            color="tertiary"
            disabled={pending}
            onClick={() => {
              onOpenChange(false);
            }}
            variant="outline"
          >
            Cancel
          </Button>
          <Button
            disabled={pending || !name.trim()}
            loading={pending}
            loadingText="Saving..."
            onClick={() => void save()}
          >
            Save
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

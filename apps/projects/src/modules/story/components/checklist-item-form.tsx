import type { FormEvent } from "react";
import { useState } from "react";
import { Button, Flex, Input } from "ui";
import { cn } from "lib";

export const ChecklistItemForm = ({
  initialTitle = "",
  isPending,
  onCancel,
  onSave,
}: {
  initialTitle?: string;
  isPending: boolean;
  onCancel: () => void;
  onSave: (title: string) => void;
}) => {
  const [title, setTitle] = useState(initialTitle);
  const isEditing = Boolean(initialTitle);

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!isPending && title.trim()) onSave(title.trim());
  };

  return (
    <form
      className={cn("min-w-0 flex-1 space-y-2", { "w-full py-2": !isEditing })}
      onSubmit={handleSubmit}
    >
      <Input
        aria-label={isEditing ? "Edit checklist item" : "New checklist item"}
        autoFocus
        className={isEditing ? "h-9 px-2 leading-9" : undefined}
        disabled={isPending}
        maxLength={255}
        onChange={(event) => {
          setTitle(event.target.value);
        }}
        onKeyDown={(event) => {
          if (event.key === "Escape") {
            event.preventDefault();
            event.stopPropagation();
            if (!isPending) onCancel();
          }
        }}
        placeholder="What needs to be done?"
        value={title}
      />
      <Flex gap={2}>
        <Button
          color="tertiary"
          disabled={isPending}
          onClick={onCancel}
          size="sm"
          type="button"
          variant="outline"
        >
          Cancel
        </Button>
        <Button
          disabled={isPending || !title.trim()}
          loading={isPending}
          loadingText={isEditing ? "Saving..." : "Creating..."}
          size="sm"
          type="submit"
        >
          {isEditing ? "Save" : "Create"}
        </Button>
      </Flex>
    </form>
  );
};

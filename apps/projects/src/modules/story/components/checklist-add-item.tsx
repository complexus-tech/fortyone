import { Button } from "ui";
import { PlusIcon } from "icons";
import { ChecklistItemForm } from "./checklist-item-form";

const MAX_CHECKLIST_ITEMS = 200;

export const ChecklistAddItem = ({
  itemCount,
  isAdding,
  isPending,
  onAdd,
  onCancel,
  onSave,
}: {
  itemCount: number;
  isAdding: boolean;
  isPending: boolean;
  onAdd: () => void;
  onCancel: () => void;
  onSave: (title: string) => void;
}) => {
  if (isAdding)
    return (
      <ChecklistItemForm
        isPending={isPending}
        onCancel={onCancel}
        onSave={onSave}
      />
    );
  if (itemCount >= MAX_CHECKLIST_ITEMS)
    return (
      <p className="text-text-muted py-2">
        This checklist has reached its limit of 200 items.
      </p>
    );
  return (
    <Button
      className="mt-2"
      color="tertiary"
      disabled={isPending}
      leftIcon={<PlusIcon />}
      onClick={onAdd}
      size="sm"
      variant="naked"
    >
      Add item
    </Button>
  );
};

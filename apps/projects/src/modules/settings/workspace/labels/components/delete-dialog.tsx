import { Button, Dialog } from "ui";

type DeleteDialogProps = {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
};

export const DeleteDialog = ({
  isOpen,
  onClose,
  onConfirm,
}: DeleteDialogProps) => {
  return (
    <Dialog onOpenChange={onClose} open={isOpen}>
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title className="px-6">Delete Label</Dialog.Title>
          <Dialog.Description>
            Are you sure you want to delete this label? This action cannot be
            undone.
          </Dialog.Description>
        </Dialog.Header>

        <Dialog.Footer>
          <Button color="tertiary" onClick={onClose} variant="outline">
            Cancel
          </Button>
          <Button color="danger" onClick={onConfirm}>
            Delete Label
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

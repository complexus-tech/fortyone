import { Button, Dialog } from "ui";

export const MessageDeleteDialog = ({
  open,
  text,
  hasReplies,
  onClose,
  onDelete,
  pending = false,
}: {
  open: boolean;
  text: string;
  hasReplies: boolean;
  onClose: () => void;
  onDelete: () => void;
  pending?: boolean;
}) => (
  <Dialog
    onOpenChange={(next) => {
      if (!next && !pending) onClose();
    }}
    open={open}
  >
    <Dialog.Content
      onCloseAutoFocus={(event) => {
        event.preventDefault();
      }}
      size="sm"
    >
      <Dialog.Header className="px-6 pt-6">
        <Dialog.Title className="text-lg">Delete this message?</Dialog.Title>
      </Dialog.Header>
      <Dialog.Body className="pt-0">
        <Dialog.Description className="text-text-muted text-base leading-6">
          This message will be removed.
          {hasReplies ? " Replies will stay in the thread." : ""}
        </Dialog.Description>
        <blockquote className="border-border text-text-muted mt-4 max-h-32 overflow-y-auto border-l-2 pl-3 text-base leading-6 break-words whitespace-pre-wrap">
          {text}
        </blockquote>
        <div className="mt-5 flex justify-end gap-2">
          <Button
            color="tertiary"
            disabled={pending}
            onClick={onClose}
            variant="outline"
          >
            Cancel
          </Button>
          <Button color="danger" disabled={pending} onClick={onDelete}>
            Delete message
          </Button>
        </div>
      </Dialog.Body>
    </Dialog.Content>
  </Dialog>
);

import { Button, Dialog, Flex, Text } from "ui";
import { DeleteIcon } from "icons";

export const CommentDeleteDialog = ({
  open,
  onOpenChange,
  onDelete,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onDelete: () => void;
}) => (
  <Dialog onOpenChange={onOpenChange} open={open}>
    <Dialog.Content>
      <Dialog.Header className="px-6 pt-6">
        <Dialog.Title className="text-lg">
          Are you sure you want to delete this comment?
        </Dialog.Title>
      </Dialog.Header>
      <Dialog.Body className="pt-0">
        <Text color="muted">
          This comment and any replies to it will be permanently deleted.
        </Text>
        <Flex align="center" className="mt-4" gap={2} justify="end">
          <Button
            color="tertiary"
            onClick={() => {
              onOpenChange(false);
            }}
          >
            Cancel
          </Button>
          <Button
            leftIcon={
              <DeleteIcon className="h-[1.15rem] text-white dark:text-gray-200" />
            }
            onClick={onDelete}
          >
            Delete
          </Button>
        </Flex>
      </Dialog.Body>
    </Dialog.Content>
  </Dialog>
);

import { Plus } from "lucide-react";
import { useState } from "react";
import { Button, Flex, Text, Dialog, Input } from "ui";

export const AddLinks = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <>
      <Flex align="center">
        <Text fontWeight="medium">Links</Text>
        <Button
          className="ml-auto"
          color="tertiary"
          leftIcon={<Plus className="h-5 w-auto" strokeWidth={2} />}
          onClick={() => {
            setIsOpen(true);
          }}
          size="sm"
          variant="outline"
        >
          Add
        </Button>
      </Flex>
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content size="md" hideClose>
          <Dialog.Header className="flex items-center justify-between px-6 pt-6">
            <Dialog.Title>
              <Text fontSize="xl">Add link to issue</Text>
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="flex flex-col gap-3 pb-6">
            <Input label="URL" placeholder="https://..." required type="url" />
            <Input label="Title" placeholder="Enter title..." required />
          </Dialog.Body>
          <Dialog.Footer className="flex items-center justify-end gap-2">
            <Button
              color="tertiary"
              onClick={() => {
                setIsOpen(false);
              }}
              size="md"
              variant="outline"
            >
              Cancel
            </Button>
            <Button leftIcon={<Plus className="h-5 w-auto" />} size="md">
              Add link
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </>
  );
};

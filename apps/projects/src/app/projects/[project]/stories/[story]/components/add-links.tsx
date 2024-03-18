import { useState } from "react";
import { Button, Flex, Text, Dialog, Input } from "ui";
import { PlusIcon } from "icons";

export const AddLinks = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <>
      <Flex align="center">
        <Text fontWeight="medium">Links</Text>
        <Button
          className="ml-auto"
          color="tertiary"
          leftIcon={<PlusIcon className="h-5 w-auto" strokeWidth={2} />}
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
        <Dialog.Content hideClose size="md">
          <Dialog.Header className="flex items-center justify-between px-6 pt-6">
            <Dialog.Title>
              <Text fontSize="xl">Add link to story</Text>
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="flex flex-col gap-5 pb-6">
            <Input label="URL" placeholder="https://..." required type="url" />
            <Input label="Title" placeholder="Enter title..." />

            <Flex align="center" className="mt-4" gap={2} justify="end">
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
              <Button leftIcon={<PlusIcon className="h-5 w-auto" />} size="md">
                Add link
              </Button>
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </>
  );
};

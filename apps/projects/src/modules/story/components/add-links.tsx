import { useState } from "react";
import { Button, Flex, Text, Dialog, Input } from "ui";
import { PlusIcon } from "icons";

export const AddLinks = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <>
      <Flex align="center" justify="end">
        <Button
          color="tertiary"
          leftIcon={
            <PlusIcon
              className="text-white dark:text-gray-200"
              strokeWidth={2}
            />
          }
          onClick={() => {
            setIsOpen(true);
          }}
          size="sm"
          variant="outline"
        >
          Add link
        </Button>
      </Flex>
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content hideClose={false}>
          <Dialog.Header className="flex items-center justify-between px-6 pb-2">
            <Dialog.Title>
              <Text fontSize="lg" fontWeight="medium">
                Add link to story
              </Text>
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="flex flex-col gap-3 pb-5">
            <Input label="URL" placeholder="https://..." required type="url" />
            <Input label="Title" placeholder="Enter title..." />

            <Flex align="center" className="mt-2" gap={2} justify="end">
              <Button
                color="tertiary"
                onClick={() => {
                  setIsOpen(false);
                }}
                variant="outline"
              >
                Cancel
              </Button>
              <Button
                leftIcon={
                  <PlusIcon className="text-white dark:text-gray-200" />
                }
              >
                Add link
              </Button>
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </>
  );
};

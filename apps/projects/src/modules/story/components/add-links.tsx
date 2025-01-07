import { useState } from "react";
import { Button, Flex } from "ui";
import { PlusIcon } from "icons";
import { AddLinkDialog } from "./add-link-dialog";

export const AddLinks = ({ storyId }: { storyId: string }) => {
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
      <AddLinkDialog isOpen={isOpen} setIsOpen={setIsOpen} storyId={storyId} />
    </>
  );
};

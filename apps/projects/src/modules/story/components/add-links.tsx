import { useState } from "react";
import { Button, Flex } from "ui";
import { PlusIcon } from "icons";
import { useMediaQuery, useUserRole } from "@/hooks";
import { AddLinkDialog } from "./add-link-dialog";

export const AddLinks = ({ storyId }: { storyId: string }) => {
  const [isOpen, setIsOpen] = useState(false);
  const isMobile = useMediaQuery("(max-width: 768px)");
  const { userRole } = useUserRole();
  return (
    <>
      <Flex align="center" className="justify-end">
        <Button
          color="tertiary"
          disabled={userRole === "guest"}
          leftIcon={
            <PlusIcon
              className="text-white dark:text-gray-200"
              strokeWidth={2}
            />
          }
          onClick={() => {
            if (userRole !== "guest") {
              setIsOpen(true);
            }
          }}
          size={isMobile ? "md" : "sm"}
          variant="outline"
        >
          Add link
        </Button>
      </Flex>
      <AddLinkDialog isOpen={isOpen} setIsOpen={setIsOpen} storyId={storyId} />
    </>
  );
};

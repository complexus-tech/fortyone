"use client";

import { AiIcon } from "icons";
import { useState } from "react";
import { Box, Button, Dialog } from "ui";

export const Chat = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <>
      <Box className="fixed bottom-8 right-8">
        <Button
          className="bg-gray-50/70 backdrop-blur dark:bg-dark-200/70 md:px-5"
          color="tertiary"
          leftIcon={<AiIcon />}
          onClick={() => {
            setIsOpen(true);
          }}
          rounded="full"
          size="lg"
        >
          Ask Maya
        </Button>
      </Box>
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content
          className="max-w-[38rem] rounded-[2rem] md:mb-[2.6vh] md:mt-auto"
          overlayClassName="justify-end pr-[1.5vh]"
        >
          <Dialog.Title className="px-6 py-4 text-lg">
            Keyboard Shortcuts
          </Dialog.Title>
          <Dialog.Body className="h-[80dvh] max-h-[80dvh] pt-2">
            Test
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </>
  );
};

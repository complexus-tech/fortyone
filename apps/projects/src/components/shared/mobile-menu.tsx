"use client";
import { useState } from "react";
import { Box, Button, Dialog } from "ui";
import { MenuIcon } from "icons";
import { Navigation } from "./sidebar/navigation";
import { Teams } from "./sidebar/teams";
import { Header } from "./sidebar/header";

export const MobileMenuButton = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <Box className="md:hidden">
      <Button
        asIcon
        color="tertiary"
        leftIcon={<MenuIcon />}
        onClick={() => {
          setIsOpen(true);
        }}
        rounded="sm"
        size="xs"
        variant="naked"
      >
        <span className="sr-only">Mobile Menu</span>
      </Button>
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content
          className="mx-0 mb-0 mt-0 h-dvh w-72 rounded-none border-y-0 border-l-0 dark:bg-dark-300"
          hideClose
          overlayClassName=" justify-start"
        >
          <Dialog.Header className="py-0">
            <Dialog.Title>
              <span className="sr-only">Mobile Menu</span>
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="max-h-dvh px-4">
            <Header />
            <Navigation />
            <Teams />
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </Box>
  );
};

"use client";
import { useState } from "react";
import { Box, Button, Dialog } from "ui";
import { MenuIcon } from "icons";
import { ViewFavoritesSlot } from "@/shared/views/favorites-slot";
import { Navigation } from "./sidebar/navigation";
import { Teams } from "./sidebar/teams";
import { Header } from "./sidebar/header";
import { WorkspaceActions } from "./workspace-actions";

export const MobileMenuButton = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <Box className="md:hidden">
      <Button
        asIcon
        className="mr-1"
        color="tertiary"
        leftIcon={<MenuIcon />}
        onClick={() => {
          setIsOpen(true);
        }}
        size="sm"
        variant="naked"
      >
        <span className="sr-only">Mobile Menu</span>
      </Button>
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content
          className="bg-surface-elevated mx-0 mt-0 mb-0 h-dvh w-72 rounded-none border-y-0 border-l-0"
          hideClose
          overlayClassName=" justify-start"
        >
          <Dialog.Header className="py-0">
            <Dialog.Title>
              <span className="sr-only">Mobile Menu</span>
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body
            className="max-h-dvh px-4"
            data-sidebar-content
            onClick={(event) => {
              if (
                event.button !== 0 ||
                event.metaKey ||
                event.ctrlKey ||
                event.shiftKey ||
                event.altKey ||
                !(event.target instanceof Element)
              )
                return;
              const link = event.target.closest("a[href]");
              if (!link || !event.currentTarget.contains(link)) return;
              setIsOpen(false);
            }}
          >
            <Header />
            <Box className="border-border mb-3 border-b pb-3 empty:hidden">
              <WorkspaceActions variant="mobile" />
            </Box>
            <Navigation />
            <ViewFavoritesSlot isCollapsed={false} />
            <Teams />
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </Box>
  );
};

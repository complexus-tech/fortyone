"use client";

import { useState } from "react";
import { useHotkeys } from "react-hotkeys-hook";
import { useRouter, usePathname } from "next/navigation";
import { useUserRole } from "@/hooks";
import { KeyboardShortcuts } from "@/components/shared/keyboard-shortcuts";
import {
  NewObjectiveDialog,
  NewStoryDialog,
  InviteMembersDialog,
} from "@/components/ui";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { CommandBar } from "./command-bar";

export const Commands = () => {
  const { userRole } = useUserRole();
  const [isStoryOpen, setIsStoryOpen] = useState(false);
  const [isSprintsOpen, setIsSprintsOpen] = useState(false);
  const [isObjectivesOpen, setIsObjectivesOpen] = useState(false);
  const [isInviteMembersOpen, setIsInviteMembersOpen] = useState(false);
  const [open, setOpen] = useState(false);
  const router = useRouter();
  const pathname = usePathname();
  const [isKeyboardShortcutsOpen, setIsKeyboardShortcutsOpen] = useState(false);

  useHotkeys("mod+k", (e) => {
    e.preventDefault();
    setOpen((prev) => !prev);
  });

  useHotkeys("g+i", () => {
    if (pathname !== "/notifications") {
      router.push("/notifications");
    }
  });
  useHotkeys("g+m", () => {
    if (pathname !== "/my-work") {
      router.push("/my-work");
    }
  });

  useHotkeys("g+s", () => {
    if (pathname !== "/summary") {
      router.push("/summary");
    }
  });
  useHotkeys("g+o", () => {
    if (pathname !== "/objectives") {
      router.push("/objectives");
    }
  });

  useHotkeys("alt+shift+s", () => {
    if (pathname !== "/settings") {
      router.push("/settings");
    }
  });

  useHotkeys("alt+shift+t", () => {});

  useHotkeys("mod+/", () => {
    setIsKeyboardShortcutsOpen((prev) => !prev);
  });
  useHotkeys("mod+i", () => {
    if (userRole === "admin") {
      setIsInviteMembersOpen((prev) => !prev);
      setOpen(false);
    }
  });

  useHotkeys("/", () => {
    if (pathname !== "search") {
      router.push("/search");
    }
  });

  return (
    <>
      <CommandBar isOpen={open} setIsOpen={setOpen} />
      <KeyboardShortcuts
        isOpen={isKeyboardShortcutsOpen}
        setIsOpen={setIsKeyboardShortcutsOpen}
      />
      <NewStoryDialog isOpen={isStoryOpen} setIsOpen={setIsStoryOpen} />
      <NewSprintDialog isOpen={isSprintsOpen} setIsOpen={setIsSprintsOpen} />
      <NewObjectiveDialog
        isOpen={isObjectivesOpen}
        setIsOpen={setIsObjectivesOpen}
      />
      {userRole === "admin" && (
        <InviteMembersDialog
          isOpen={isInviteMembersOpen}
          setIsOpen={setIsInviteMembersOpen}
        />
      )}
    </>
  );
};

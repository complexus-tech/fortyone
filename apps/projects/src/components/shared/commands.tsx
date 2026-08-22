"use client";

import { useState } from "react";
import { SearchIcon } from "icons";
import { Kbd } from "ui";
import { cn } from "lib";
import { useHotkeys } from "react-hotkeys-hook";
import { useRouter, usePathname } from "next/navigation";
import { useUserRole, useWorkspacePath } from "@/hooks";
import { KeyboardShortcuts } from "@/components/shared/keyboard-shortcuts";
import {
  NewObjectiveDialog,
  NewStoryDialog,
  InviteMembersDialog,
} from "@/components/ui";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { CommandBar } from "./command-bar";

export const Commands = ({
  className,
  showTrigger = false,
}: {
  className?: string;
  showTrigger?: boolean;
}) => {
  const { userRole } = useUserRole();
  const [isStoryOpen, setIsStoryOpen] = useState(false);
  const [isSprintsOpen, setIsSprintsOpen] = useState(false);
  const [isObjectivesOpen, setIsObjectivesOpen] = useState(false);
  const [isInviteMembersOpen, setIsInviteMembersOpen] = useState(false);
  const [open, setOpen] = useState(false);
  const router = useRouter();
  const pathname = usePathname();
  const { withWorkspace } = useWorkspacePath();
  const [isKeyboardShortcutsOpen, setIsKeyboardShortcutsOpen] = useState(false);

  useHotkeys("mod+k", (e) => {
    e.preventDefault();
    setOpen((prev) => !prev);
  });

  useHotkeys("g+i", () => {
    if (pathname !== withWorkspace("/notifications")) {
      router.push(withWorkspace("/notifications"));
    }
  });
  useHotkeys("g+m", () => {
    if (pathname !== withWorkspace("/my-work")) {
      router.push(withWorkspace("/my-work"));
    }
  });

  useHotkeys("g+s", () => {
    if (pathname !== withWorkspace("/summary")) {
      router.push(withWorkspace("/summary"));
    }
  });
  useHotkeys("g+r", () => {
    if (pathname !== withWorkspace("/roadmap")) {
      router.push(withWorkspace("/roadmap"));
    }
  });

  useHotkeys("alt+shift+s", () => {
    if (pathname !== withWorkspace("/settings")) {
      router.push(withWorkspace("/settings"));
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
    if (pathname !== withWorkspace("/search")) {
      router.push(withWorkspace("/search"));
    }
  });

  return (
    <>
      {showTrigger ? (
        <button
          aria-label="Search or jump to"
          className={cn(
            "bg-surface-muted text-text-muted hover:bg-state-hover focus-visible:ring-ring flex h-11 min-w-0 items-center gap-2 rounded-lg px-3 text-left transition-colors outline-none focus-visible:ring-2",
            className,
          )}
          onClick={() => {
            setOpen(true);
          }}
          type="button"
        >
          <SearchIcon className="h-4 shrink-0" />
          <span className="min-w-0 flex-1 truncate">Search or jump to...</span>
          <Kbd className="!bg-surface/5 !text-text-muted dark:!bg-surface/10 hidden shrink-0 sm:flex">
            ⌘ K
          </Kbd>
        </button>
      ) : null}
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

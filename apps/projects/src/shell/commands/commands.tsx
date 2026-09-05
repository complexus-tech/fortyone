"use client";

import { useState } from "react";
import { SearchIcon } from "icons";
import { Kbd } from "ui";
import { cn } from "lib";
import { useHotkeys } from "react-hotkeys-hook";
import { useRouter, usePathname } from "next/navigation";
import { useQueryState } from "nuqs";
import { useTerminology } from "@/hooks/use-terminology-display";
import { KeyboardShortcuts } from "@/components/shared/keyboard-shortcuts";
import { InviteMembersDialog } from "@/components/ui/invite-members";
import { NewObjectiveDialog } from "@/components/ui/new-objective";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { NewStoryDialog } from "@/components/ui/new-story-dialog";
import { useUserRole } from "@/hooks/role";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { CommandBar } from "./command-bar";

export const Commands = ({
  className,
  showTrigger = false,
}: {
  className?: string;
  showTrigger?: boolean;
}) => {
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm", { variant: "plural" });
  const objectiveTerm = getTermDisplay("objectiveTerm", { variant: "plural" });
  const searchLabel = `Search ${storyTerm}, ${objectiveTerm}, or commands`;
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
  const [query, setQuery] = useQueryState("query", { defaultValue: "" });
  const isSearchPage = pathname === withWorkspace("/search");

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
      {showTrigger && isSearchPage ? (
        <div
          className={cn(
            "flex min-w-0 flex-1 items-center gap-2",
            className,
            "max-w-none",
          )}
        >
          <search
            className="bg-surface-muted focus-within:ring-ring flex h-11 w-full max-w-xl min-w-0 items-center gap-2 rounded-lg px-3 focus-within:ring-2"
            key={query}
          >
            <SearchIcon className="text-text-muted h-4 shrink-0" />
            <input
              aria-label={`Search ${storyTerm} and ${objectiveTerm}`}
              className="placeholder:text-text-muted min-w-0 flex-1 bg-transparent outline-none"
              defaultValue={query}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  void setQuery(event.currentTarget.value.trim());
                }
              }}
              placeholder={`Search ${storyTerm} and ${objectiveTerm}…`}
              type="search"
            />
          </search>
          <button
            aria-label="Open command menu"
            className="border-border bg-surface-muted text-text-muted hover:bg-state-hover focus-visible:ring-ring ml-auto flex h-[2rem] shrink-0 items-center rounded-lg border-[0.5px] px-2.5 transition-colors outline-none focus-visible:ring-2"
            onClick={() => {
              setOpen(true);
            }}
            type="button"
          >
            <span className="text-[0.95rem] font-medium tracking-tight">
              ⌘ K
            </span>
          </button>
        </div>
      ) : null}
      {showTrigger && !isSearchPage ? (
        <button
          aria-label={searchLabel}
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
          <span className="min-w-0 flex-1 truncate">{searchLabel}…</span>
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

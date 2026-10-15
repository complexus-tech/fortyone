import type { Dispatch, SetStateAction } from "react";
import { useEffect } from "react";
import type { Document, Story } from "../types";
import type { Page } from "../components/sidebar";
import type { useDesktop } from "./use-desktop";
import { native, errorMessage } from "../native";
import { subscribeNative } from "../events";
import { flushDrafts } from "../drafts";
import { applyTheme } from "../theme";

export type Selection =
  | {
      accountId: string;
      workspaceId: string;
      kind: "story";
      story: Story | null;
    }
  | {
      accountId: string;
      workspaceId: string;
      kind: "document";
      document: Document | null;
    }
  | null;
export function useDesktopEvents(
  desktop: ReturnType<typeof useDesktop>,
  setPage: Dispatch<SetStateAction<Page>>,
  setSelection: Dispatch<SetStateAction<Selection>>,
) {
  const workspaceId = desktop.data?.workspaceId;
  const accountId = desktop.session?.user?.id;
  const teamCount = desktop.data?.teams.length ?? 0;
  useEffect(() => {
    void native
      .settings()
      .then((settings) => applyTheme(settings.theme))
      .catch(() => applyTheme("system"));
  }, []);
  useEffect(() => {
    const action = (name: string) => {
      if (name === "sync") {
        void desktop.sync(true);
        return;
      }
      if (desktop.switchingId || document.querySelector("dialog[open]")) return;
      if (name === "create-story" && teamCount && workspaceId && accountId)
        setSelection({ accountId, workspaceId, kind: "story", story: null });
      if (name === "search") {
        setPage((current) => (current === "documents" ? current : "stories"));
        window.requestAnimationFrame(() =>
          document.getElementById("desktop-search")?.focus(),
        );
      }
    };
    const keydown = (event: KeyboardEvent) => {
      if (!(event.metaKey || event.ctrlKey) || event.altKey || event.shiftKey)
        return;
      if (["k", "n"].includes(event.key.toLowerCase())) {
        event.preventDefault();
        action(event.key.toLowerCase() === "k" ? "search" : "create-story");
      }
    };
    window.addEventListener("keydown", keydown);
    const stop = subscribeNative<{ action: string }>(
      "desktop-action",
      (event) => action(event.payload.action),
      desktop.setError,
    );
    return () => {
      window.removeEventListener("keydown", keydown);
      stop();
    };
  }, [
    accountId,
    workspaceId,
    teamCount,
    desktop.switchingId,
    desktop.sync,
    desktop.setError,
    setSelection,
    setPage,
  ]);
  useEffect(() => {
    if (desktop.starting || !accountId) return;
    const client = native.forAccount(accountId);
    let active = true;
    let handling = false;
    const navigate = async () => {
      if (handling) return;
      handling = true;
      try {
        const target = await client.takeNavigation();
        if (!target || !active) return;
        await flushDrafts();
        const selected = await desktop.selectWorkspace(target.workspaceId);
        if (!selected || !active) return;
        if (target.kind === "notifications") {
          setSelection(null);
          setPage("notifications");
          return;
        }
        let data = await client.readWorkspace(target.workspaceId);
        if (!active) return;
        const cached =
          target.kind === "document" ? data.documents : data.stories;
        if (
          target.resourceId &&
          !cached.some((item) => item.id === target.resourceId) &&
          navigator.onLine
        ) {
          data = await client.refreshResource({
            workspaceId: target.workspaceId,
            resourceId: target.resourceId,
            kind: target.kind === "document" ? "document" : "story",
          });
          if (!active) return;
          desktop.acceptData(data);
        }
        if (target.kind === "document") {
          const found = data.documents.find(
            (item) => item.id === target.resourceId,
          );
          if (found) {
            setPage("documents");
            setSelection({
              accountId,
              workspaceId: target.workspaceId,
              kind: "document",
              document: found,
            });
          } else
            desktop.setError(
              "This document is not available on this device. Connect and open the link again to download it.",
            );
        } else {
          const found = data.stories.find(
            (item) => item.id === target.resourceId,
          );
          if (found) {
            setPage("stories");
            setSelection({
              accountId,
              workspaceId: target.workspaceId,
              kind: "story",
              story: found,
            });
          } else
            desktop.setError(
              "This story is not available on this device. Connect and open the link again to download it.",
            );
        }
      } catch (reason) {
        desktop.setError(errorMessage(reason));
      } finally {
        handling = false;
      }
    };
    const stop = subscribeNative(
      "desktop-navigation",
      () => {
        void navigate();
      },
      desktop.setError,
    );
    void navigate();
    return () => {
      active = false;
      stop();
    };
  }, [
    desktop.acceptData,
    desktop.starting,
    accountId,
    desktop.selectWorkspace,
    desktop.setError,
    setPage,
    setSelection,
  ]);
  useEffect(
    () =>
      subscribeNative(
        "desktop-before-quit",
        () => {
          void flushDrafts()
            .then(native.confirmExit)
            .catch((reason: unknown) => desktop.setError(errorMessage(reason)));
        },
        desktop.setError,
      ),
    [desktop.setError],
  );
}

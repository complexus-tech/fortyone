import type { Dispatch, SetStateAction } from "react";
import { useEffect } from "react";
import type { Selection } from "./use-desktop-events";
import type { useDesktop } from "./use-desktop";
import { native, errorMessage } from "../native";

export function useResourceRefresh(
  selection: Selection,
  desktop: ReturnType<typeof useDesktop>,
  setSelection: Dispatch<SetStateAction<Selection>>,
) {
  const accountId = desktop.session?.user?.id;
  const selectionAccount = selection?.accountId;
  const resourceId =
    selection?.kind === "story" ? selection.story?.id : selection?.document?.id;
  const workspaceId = selection?.workspaceId;
  const kind = selection?.kind;
  const pendingCreate = desktop.data?.pending.some(
    (operation) =>
      operation.resourceId === resourceId &&
      operation.kind === `${kind}.create`,
  );
  useEffect(() => {
    if (
      pendingCreate ||
      accountId !== selectionAccount ||
      !desktop.online ||
      !resourceId ||
      !workspaceId ||
      !kind
    )
      return;
    if (!accountId) return;
    const client = native.forAccount(accountId);
    let active = true;
    void client
      .refreshResource({ workspaceId, resourceId, kind })
      .then((next) => {
        if (!active) return;
        desktop.acceptData(next);
        setSelection((current) => {
          if (
            !current ||
            current.workspaceId !== workspaceId ||
            current.kind !== kind
          )
            return current;
          if (current.kind === "story" && current.story?.id === resourceId) {
            const story = next.stories.find((item) => item.id === resourceId);
            return story ? { ...current, story } : null;
          }
          if (
            current.kind === "document" &&
            current.document?.id === resourceId
          ) {
            const document = next.documents.find(
              (item) => item.id === resourceId,
            );
            return document ? { ...current, document } : null;
          }
          return current;
        });
      })
      .catch(async (reason: unknown) => {
        if (!active) return;
        await desktop.reload();
        if (active) desktop.setError(errorMessage(reason));
      });
    return () => {
      active = false;
    };
  }, [
    accountId,
    selectionAccount,
    pendingCreate,
    desktop.online,
    desktop.acceptData,
    desktop.setError,
    desktop.reload,
    kind,
    resourceId,
    workspaceId,
    setSelection,
  ]);
}

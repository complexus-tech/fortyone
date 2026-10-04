import type { FavoriteViewRef } from "./favorites-store";

const SELECT_VIEW_EVENT = "fortyone:select-saved-view";
type ViewSelection = FavoriteViewRef & {
  workspaceSlug: string;
  userId: string;
};
let focusRequest: { selection: ViewSelection; expiresAt: number } | null = null;

export const requestCreatedViewFocus = (selection: ViewSelection) => {
  focusRequest = { selection, expiresAt: Date.now() + 10_000 };
};

export const consumeCreatedViewFocus = (selection: ViewSelection) => {
  if (!focusRequest || focusRequest.expiresAt < Date.now()) {
    focusRequest = null;
    return false;
  }
  const requested = focusRequest.selection;
  if (
    requested.workspaceSlug !== selection.workspaceSlug ||
    requested.userId !== selection.userId ||
    requested.teamId !== selection.teamId ||
    requested.id !== selection.id
  )
    return false;
  focusRequest = null;
  return true;
};

/** A deliberate click on the current view should reload its saved definition. */
export const selectCurrentView = (selection: ViewSelection) => {
  window.dispatchEvent(
    new CustomEvent(SELECT_VIEW_EVENT, { detail: selection }),
  );
};

export const subscribeViewSelection = (
  selection: ViewSelection,
  listener: () => void,
) => {
  const onSelect = (event: Event) => {
    const detail = (event as CustomEvent<ViewSelection>).detail;
    if (
      detail.workspaceSlug === selection.workspaceSlug &&
      detail.userId === selection.userId &&
      detail.teamId === selection.teamId &&
      detail.id === selection.id
    )
      listener();
  };
  window.addEventListener(SELECT_VIEW_EVENT, onSelect);
  return () => {
    window.removeEventListener(SELECT_VIEW_EVENT, onSelect);
  };
};

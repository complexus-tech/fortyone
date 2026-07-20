type SetDialogOpen = (open: boolean) => void;
type ScheduleDialogOpen = (
  callback: () => void,
  delay: number,
) => ReturnType<typeof setTimeout>;

/**
 * Opens a modal after the current menu selection lifecycle has completed.
 *
 * Radix menus and dialogs both manage focus and outside pointer events. Opening
 * a dialog synchronously from a menu item's `onSelect` can make those cleanup
 * cycles overlap and leave the document inert after the dialog closes.
 */
export const openDialogAfterMenuClose = (
  setOpen: SetDialogOpen,
  schedule: ScheduleDialogOpen = setTimeout,
) => {
  schedule(() => {
    setOpen(true);
  }, 0);
};

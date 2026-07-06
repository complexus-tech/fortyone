type SetDialogOpen = (open: boolean) => void;
type ScheduleDialogOpen = (
  callback: () => void,
  delay: number,
) => ReturnType<typeof setTimeout>;

export const openDialogAfterMenuClose = (
  setOpen: SetDialogOpen,
  schedule: ScheduleDialogOpen = setTimeout,
) => {
  schedule(() => {
    setOpen(true);
  }, 0);
};

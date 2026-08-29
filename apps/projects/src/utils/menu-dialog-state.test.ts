/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { openDialogAfterMenuClose } from "./menu-dialog-state";

describe("openDialogAfterMenuClose", () => {
  it("defers dialog opening until the menu select event has finished closing the menu", () => {
    const setOpen = jest.fn();
    const schedule = jest.fn();

    openDialogAfterMenuClose(setOpen, schedule);

    expect(setOpen).not.toHaveBeenCalled();
    expect(schedule).toHaveBeenCalledWith(expect.any(Function), 0);

    const [openDialog] = schedule.mock.calls[0] as [() => void, number];
    openDialog();

    expect(setOpen).toHaveBeenCalledWith(true);
  });
});

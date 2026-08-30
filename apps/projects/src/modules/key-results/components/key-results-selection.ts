export const isKeyResultGroupSelected = (
  keyResultIds: readonly string[],
  selectedKeyResultIds: ReadonlySet<string>,
) =>
  keyResultIds.length > 0 &&
  keyResultIds.every((keyResultId) => selectedKeyResultIds.has(keyResultId));

export const setKeyResultSelection = (
  selectedKeyResultIds: ReadonlySet<string>,
  keyResultId: string,
  selected: boolean,
) => {
  const nextSelected = new Set(selectedKeyResultIds);

  if (selected) {
    nextSelected.add(keyResultId);
  } else {
    nextSelected.delete(keyResultId);
  }

  return nextSelected;
};

export const setKeyResultGroupSelection = (
  selectedKeyResultIds: ReadonlySet<string>,
  keyResultIds: readonly string[],
  selected: boolean,
) => {
  const nextSelected = new Set(selectedKeyResultIds);

  for (const keyResultId of keyResultIds) {
    if (selected) {
      nextSelected.add(keyResultId);
    } else {
      nextSelected.delete(keyResultId);
    }
  }

  return nextSelected;
};

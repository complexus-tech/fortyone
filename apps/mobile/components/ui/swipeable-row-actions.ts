export type SwipeActionIdentity = {
  id?: string;
  accessibilityLabel: string;
  disabled?: boolean;
};

export type SwipeActionToken = {
  id: string;
  label: string;
  scope?: string;
};

export function captureSwipeAction(
  action: SwipeActionIdentity,
  scope?: string,
): SwipeActionToken {
  return {
    id: action.id ?? action.accessibilityLabel,
    label: action.accessibilityLabel,
    scope,
  };
}

/** Native menus can remain open after their row, action or session has changed. */
export function resolveSwipeAction<T extends SwipeActionIdentity>(
  current: {
    actions: readonly T[];
    active: boolean;
    disabled?: boolean | null;
    scope?: string;
  },
  expected: SwipeActionToken,
): T | undefined {
  if (!current.active || current.disabled || current.scope !== expected.scope)
    return undefined;
  return current.actions.find(
    (action) =>
      !action.disabled &&
      (action.id ?? action.accessibilityLabel) === expected.id &&
      action.accessibilityLabel === expected.label,
  );
}

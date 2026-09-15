export const filterPropertyOptions = <
  T extends { label: string; description?: string },
>(
  options: T[],
  search: string,
): T[] => {
  const query = search.trim().toLocaleLowerCase();
  return query
    ? options.filter((option) =>
        `${option.label} ${option.description ?? ""}`
          .toLocaleLowerCase()
          .includes(query),
      )
    : options;
};

export const memberDisplayName = (member: {
  fullName?: string | null;
  username?: string | null;
  email?: string | null;
}) =>
  [member.fullName, member.username, member.email]
    .find((value) => typeof value === "string" && value.trim())
    ?.trim() || "Unnamed member";

export const togglePropertyId = (ids: string[], id: string) =>
  ids.includes(id) ? ids.filter((value) => value !== id) : [...ids, id];

/** Dismissal is a success effect; rejected mutations stay recoverable in the picker. */
export const performPropertyChange = (
  action: () => Promise<void>,
  handlers: {
    onSuccess: () => void;
    onError: (message: string) => void;
    onSettled: () => void;
  },
) =>
  Promise.resolve()
    .then(action)
    .then(handlers.onSuccess)
    .catch((cause: unknown) => {
      handlers.onError(
        cause instanceof Error
          ? cause.message
          : "Could not update this property. Try again.",
      );
    })
    .finally(handlers.onSettled);

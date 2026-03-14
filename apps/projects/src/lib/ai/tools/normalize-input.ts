export const normalizeOptionalString = (value?: string | null) => {
  if (value == null) {
    return undefined;
  }

  const trimmedValue = value.trim();
  return trimmedValue === "" ? undefined : trimmedValue;
};

export const normalizeStringArray = (values?: string[] | null) => {
  if (values == null) {
    return undefined;
  }

  return values
    .map((value) => value.trim())
    .filter((value) => value.length > 0);
};

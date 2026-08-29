export type StoryStatus = {
  id: string;
  category: string;
  isDefault: boolean;
};

export const selectStoryStatusId = (
  statuses: StoryStatus[],
  requestedStatusId?: string | null,
) => {
  const normalizedStatusId = requestedStatusId?.trim();
  if (normalizedStatusId) {
    const requestedStatus = statuses.find(
      (status) => status.id === normalizedStatusId,
    );
    if (!requestedStatus) {
      throw new Error(
        "The selected status is not available for the target team.",
      );
    }
    return requestedStatus.id;
  }

  if (statuses.length === 0) {
    throw new Error(
      "The target team has no workflow statuses. Create a status before creating stories.",
    );
  }

  const defaultStatus =
    statuses.find((status) => status.isDefault) ?? statuses[0];
  return defaultStatus.id;
};

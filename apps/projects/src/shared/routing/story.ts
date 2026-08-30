const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

type StoryRouteIdentity = {
  id: string;
  sequenceId?: number | null;
  teamCode?: string | null;
};

export const isStoryUuid = (value: string) => UUID_PATTERN.test(value);

export const getStoryReference = ({
  id,
  sequenceId,
  teamCode,
}: StoryRouteIdentity) => {
  const normalizedTeamCode = teamCode?.trim().toUpperCase();

  if (normalizedTeamCode && Number.isInteger(sequenceId)) {
    return `${normalizedTeamCode}-${sequenceId}`;
  }

  return id;
};

export const getStoryPath = (story: StoryRouteIdentity) =>
  `/work/${encodeURIComponent(getStoryReference(story))}`;

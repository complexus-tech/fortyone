export const reorderCachedTeams = <T extends { id: string }>(
  teams: T[] | undefined,
  orderedTeamIds: string[],
) => {
  if (!teams) return teams;

  const teamsById = new Map(teams.map((team) => [team.id, team]));
  const orderedTeamIdSet = new Set(orderedTeamIds);
  const orderedTeams = orderedTeamIds.flatMap((teamId) => {
    const team = teamsById.get(teamId);
    return team ? [team] : [];
  });

  return [
    ...orderedTeams,
    ...teams.filter((team) => !orderedTeamIdSet.has(team.id)),
  ];
};

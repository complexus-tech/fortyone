export const MAX_VISIBLE_SIDEBAR_TEAMS = 4;

export const partitionSidebarTeams = <T extends { id: string }>(
  teams: T[],
  activeTeamId?: string,
) => {
  const defaultVisibleTeams = teams.slice(0, MAX_VISIBLE_SIDEBAR_TEAMS);
  const activeTeam = activeTeamId
    ? teams.find((team) => team.id === activeTeamId)
    : undefined;

  if (
    !activeTeam ||
    defaultVisibleTeams.some((team) => team.id === activeTeam.id)
  ) {
    return {
      hasPromotedActiveTeam: false,
      visibleTeams: defaultVisibleTeams,
      overflowTeams: teams.slice(MAX_VISIBLE_SIDEBAR_TEAMS),
    };
  }

  const visibleTeams = [
    ...teams.slice(0, MAX_VISIBLE_SIDEBAR_TEAMS - 1),
    activeTeam,
  ];
  const visibleTeamIds = new Set(visibleTeams.map((team) => team.id));

  return {
    hasPromotedActiveTeam: true,
    visibleTeams,
    overflowTeams: teams.filter((team) => !visibleTeamIds.has(team.id)),
  };
};

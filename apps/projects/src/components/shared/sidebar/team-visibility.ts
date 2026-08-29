export const MAX_VISIBLE_SIDEBAR_TEAMS = 4;

export const pinSidebarTeam = <T extends { id: string }>(
  teams: T[],
  teamId: string,
) => {
  const team = teams.find((item) => item.id === teamId);

  if (!team || teams[0]?.id === teamId) return teams;

  return [team, ...teams.filter((item) => item.id !== teamId)];
};

export const reorderVisibleSidebarTeams = <T extends { id: string }>(
  teams: T[],
  visibleTeams: T[],
  activeTeamId: string,
  overTeamId: string,
) => {
  const oldIndex = visibleTeams.findIndex((team) => team.id === activeTeamId);
  const newIndex = visibleTeams.findIndex((team) => team.id === overTeamId);

  if (oldIndex === -1 || newIndex === -1 || oldIndex === newIndex) {
    return teams;
  }

  const reorderedVisibleTeams = [...visibleTeams];
  const [movedTeam] = reorderedVisibleTeams.splice(oldIndex, 1);
  reorderedVisibleTeams.splice(newIndex, 0, movedTeam);

  const visibleTeamIds = new Set(visibleTeams.map((team) => team.id));

  return [
    ...reorderedVisibleTeams,
    ...teams.filter((team) => !visibleTeamIds.has(team.id)),
  ];
};

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

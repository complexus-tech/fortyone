import type { Team } from "@/modules/teams/types";

export type TeamSwitcherProps = {
  isOpened: boolean;
  setIsOpened: (open: boolean) => void;
  teams: Team[];
  currentTeamId?: string;
};

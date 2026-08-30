import { ArrowRightIcon, CheckIcon, MaximizeIcon, MinimizeIcon } from "icons";
import { Button, Dialog, Flex, Menu, Text } from "ui";
import { TeamColor } from "../team-color";

type TeamOption = {
  code: string;
  color: string;
  id: string;
  name: string;
};

export const NewObjectiveDialogHeader = ({
  currentTeam,
  isExpanded,
  objectiveTerm,
  onTeamSelect,
  onToggleExpanded,
  teams,
}: {
  currentTeam?: TeamOption | null;
  isExpanded: boolean;
  objectiveTerm: string;
  onTeamSelect: (teamId: string) => void;
  onToggleExpanded: () => void;
  teams: TeamOption[];
}) => (
  <Dialog.Header className="flex items-center justify-between px-6 pt-6">
    <Dialog.Title className="flex items-center gap-1 text-lg">
      <Menu>
        <Menu.Button>
          <Button
            className="gap-1.5 font-semibold tracking-wide"
            color="tertiary"
            leftIcon={<TeamColor color={currentTeam?.color} />}
            size="xs"
          >
            {currentTeam?.code}
          </Button>
        </Menu.Button>
        <Menu.Items align="start" className="w-52">
          <Menu.Group>
            {teams.map((team) => (
              <Menu.Item
                active={team.id === currentTeam?.id}
                className="justify-between gap-3"
                key={team.id}
                onClick={() => {
                  onTeamSelect(team.id);
                }}
              >
                <span className="flex items-center gap-1.5">
                  <TeamColor className="shrink-0" color={team.color} />
                  <span className="block truncate">{team.name}</span>
                </span>
                {team.id === currentTeam?.id ? (
                  <CheckIcon className="h-[1.1rem] w-auto" />
                ) : null}
              </Menu.Item>
            ))}
          </Menu.Group>
        </Menu.Items>
      </Menu>
      <ArrowRightIcon className="h-4 w-auto opacity-30" strokeWidth={3} />
      <Text className="opacity-80" color="muted">
        New {objectiveTerm}
      </Text>
    </Dialog.Title>
    <Flex gap={3}>
      <Button
        asIcon
        className="hover:bg-state-hover"
        color="tertiary"
        onClick={onToggleExpanded}
        size="sm"
        variant="naked"
      >
        {isExpanded ? (
          <MinimizeIcon className="h-[1.2rem] w-auto" />
        ) : (
          <MaximizeIcon className="h-[1.2rem] w-auto" />
        )}
        <span className="sr-only">
          {isExpanded ? "Minimize" : "Expand"} dialog
        </span>
      </Button>
      <Dialog.Close />
    </Flex>
  </Dialog.Header>
);

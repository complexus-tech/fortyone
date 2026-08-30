import { ArrowRight2Icon, CheckIcon, MaximizeIcon, MinimizeIcon } from "icons";
import { Button, Dialog, Flex, Menu, Text, Tooltip } from "ui";
import { TeamColor } from "./team-color";

type TeamOption = {
  code: string;
  color: string;
  id: string;
  name: string;
};

export const NewStoryDialogHeader = ({
  activeTeamId,
  currentTeam,
  isExpanded,
  onTeamSelect,
  onToggleExpanded,
  storyTerm,
  teams,
}: {
  activeTeamId?: string;
  currentTeam?: TeamOption | null;
  isExpanded: boolean;
  onTeamSelect: (teamId: string) => void;
  onToggleExpanded: () => void;
  storyTerm: string;
  teams: TeamOption[];
}) => (
  <Dialog.Header className="flex items-center justify-between px-6 pt-6">
    <Dialog.Title className="flex items-center gap-1 text-lg">
      <Menu>
        <Menu.Button>
          <Button
            className="dark:bg-surface-elevated/90 gap-1.5 text-[0.95rem] font-semibold tracking-wide"
            color="tertiary"
            leftIcon={<TeamColor color={currentTeam?.color} />}
            size="sm"
          >
            {currentTeam?.code}
          </Button>
        </Menu.Button>
        <Menu.Items align="start" className="w-52">
          <Menu.Group>
            {teams.map((team) => (
              <Menu.Item
                active={team.id === activeTeamId}
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
                {team.id === activeTeamId ? (
                  <CheckIcon className="h-[1.1rem] w-auto" />
                ) : null}
              </Menu.Item>
            ))}
          </Menu.Group>
        </Menu.Items>
      </Menu>
      <ArrowRight2Icon className="h-4.5 w-auto opacity-30" strokeWidth={3} />
      <Text className="opacity-80" color="muted">
        New {storyTerm}
      </Text>
    </Dialog.Title>
    <Flex gap={2}>
      <Tooltip title={isExpanded ? "Minimize dialog" : "Expand dialog"}>
        <Button
          className="hover:bg-state-hover px-[0.35rem]"
          color="tertiary"
          onClick={onToggleExpanded}
          size="xs"
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
      </Tooltip>
      <Dialog.Close />
    </Flex>
  </Dialog.Header>
);

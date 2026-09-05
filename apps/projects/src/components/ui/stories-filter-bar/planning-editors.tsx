import type { UIEvent } from "react";
import { Command, Divider, Flex, Text } from "ui";
import { CheckIcon, ObjectiveIcon, OKRIcon, SprintsIcon } from "icons";
import { MenuLoadingSkeleton } from "../menu-loading-skeleton";
import { TeamColor } from "../team-color";
import { normalizeArrayFilter, shouldFetchNextPage } from "./filter-model";
import type { StoriesFilterEditorProps } from "./types";

type PaginatedEditorProps = {
  hasNextPage: boolean;
  isFetchingNextPage: boolean;
  isLoading: boolean;
  onFetchNextPage: () => void;
  onQueryChange: (query: string) => void;
  query: string;
};

export type FilterTeamOption = {
  color: string;
  id: string;
  name: string;
};

export type FilterSprintOption = {
  id: string;
  name: string;
};

export type FilterObjectiveOption = {
  id: string;
  name: string;
};

export type FilterKeyResultOption = {
  id: string;
  name: string;
};

export const TeamEditor = ({
  filters,
  hasNextPage,
  isFetchingNextPage,
  isLoading,
  onFetchNextPage,
  onQueryChange,
  query,
  setFilters,
  teams,
}: StoriesFilterEditorProps &
  PaginatedEditorProps & { teams: FilterTeamOption[] }) => {
  const handleScroll = (event: UIEvent<HTMLDivElement>) => {
    if (
      shouldFetchNextPage(
        event.currentTarget,
        Boolean(hasNextPage),
        isFetchingNextPage,
      )
    ) {
      onFetchNextPage();
    }
  };

  const toggleTeam = (teamId: string) => {
    const selected = filters.teamIds ?? [];
    const teamIds = selected.includes(teamId)
      ? selected.filter((id) => id !== teamId)
      : [...selected, teamId];
    setFilters({ ...filters, teamIds: normalizeArrayFilter(teamIds) });
  };

  return (
    <Command>
      <Command.Input
        autoFocus
        onValueChange={onQueryChange}
        placeholder="Search teams..."
        value={query}
      />
      <Divider className="my-2" />
      <Command.List className="w-full border-0 bg-transparent p-0 shadow-none backdrop-blur-none dark:bg-transparent">
        {!isLoading ? (
          <Command.Empty className="px-3 py-2 text-left text-base">
            <Text color="muted">No teams found.</Text>
          </Command.Empty>
        ) : null}
        <Command.Group
          className="max-h-80 overflow-y-auto"
          onScroll={handleScroll}
        >
          {isLoading ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={5} />
            </Command.Loading>
          ) : null}
          {teams.map((team, idx) => (
            <Command.Item
              active={Boolean(filters.teamIds?.includes(team.id))}
              className="justify-between gap-4"
              key={team.id}
              onSelect={() => {
                toggleTeam(team.id);
              }}
              value={team.name}
            >
              <Flex align="center" className="min-w-0 flex-1" gap={2}>
                <TeamColor color={team.color} />
                <Text className="max-w-48 truncate">{team.name}</Text>
              </Flex>
              <Flex align="center" className="shrink-0" gap={2}>
                {filters.teamIds?.includes(team.id) ? (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                ) : null}
                <Text color="muted">{idx}</Text>
              </Flex>
            </Command.Item>
          ))}
          {isFetchingNextPage ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={2} />
            </Command.Loading>
          ) : null}
        </Command.Group>
      </Command.List>
    </Command>
  );
};

export const SprintEditor = ({
  filters,
  hasNextPage,
  isFetchingNextPage,
  isLoading,
  needsSingleTeam,
  onFetchNextPage,
  onQueryChange,
  query,
  setFilters,
  sprints,
}: StoriesFilterEditorProps &
  PaginatedEditorProps & {
    needsSingleTeam: boolean;
    sprints: FilterSprintOption[];
  }) => {
  const toggleSprint = (sprintId: string) => {
    const selected = filters.sprintIds ?? [];
    const sprintIds = selected.includes(sprintId)
      ? selected.filter((id) => id !== sprintId)
      : [...selected, sprintId];
    setFilters({ ...filters, sprintIds: normalizeArrayFilter(sprintIds) });
  };

  const handleScroll = (event: UIEvent<HTMLDivElement>) => {
    if (
      shouldFetchNextPage(
        event.currentTarget,
        Boolean(hasNextPage),
        isFetchingNextPage,
      )
    ) {
      onFetchNextPage();
    }
  };

  return (
    <Command>
      <Command.Input
        autoFocus
        onValueChange={onQueryChange}
        placeholder="Search sprints..."
        value={query}
      />
      <Divider className="my-2" />
      <Command.List className="w-full border-0 bg-transparent p-0 shadow-none backdrop-blur-none dark:bg-transparent">
        {!isLoading ? (
          <Command.Empty className="px-3 py-2 text-left text-base">
            <Text color="muted">
              {needsSingleTeam ? "Select one team first." : "No sprints found."}
            </Text>
          </Command.Empty>
        ) : null}
        <Command.Group
          className="max-h-80 overflow-y-auto"
          onScroll={handleScroll}
        >
          {isLoading ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={5} />
            </Command.Loading>
          ) : null}
          {!needsSingleTeam
            ? sprints.map((sprint, idx) => (
                <Command.Item
                  active={Boolean(filters.sprintIds?.includes(sprint.id))}
                  className="justify-between gap-4"
                  key={sprint.id}
                  onSelect={() => {
                    toggleSprint(sprint.id);
                  }}
                  value={sprint.name}
                >
                  <Flex align="center" className="min-w-0 flex-1" gap={2}>
                    <SprintsIcon className="text-text-secondary h-4 w-auto" />
                    <Text className="max-w-48 truncate">{sprint.name}</Text>
                  </Flex>
                  <Flex align="center" className="shrink-0" gap={2}>
                    {filters.sprintIds?.includes(sprint.id) ? (
                      <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                    ) : null}
                    <Text color="muted">{idx}</Text>
                  </Flex>
                </Command.Item>
              ))
            : null}
          {isFetchingNextPage ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={2} />
            </Command.Loading>
          ) : null}
        </Command.Group>
      </Command.List>
    </Command>
  );
};

export const ObjectiveEditor = ({
  filters,
  hasNextPage,
  isFetchingNextPage,
  isLoading,
  needsSingleTeam,
  objectives,
  onFetchNextPage,
  onQueryChange,
  query,
  setFilters,
}: StoriesFilterEditorProps &
  PaginatedEditorProps & {
    needsSingleTeam: boolean;
    objectives: FilterObjectiveOption[];
  }) => {
  const handleScroll = (event: UIEvent<HTMLDivElement>) => {
    if (
      shouldFetchNextPage(
        event.currentTarget,
        Boolean(hasNextPage),
        isFetchingNextPage,
      )
    ) {
      onFetchNextPage();
    }
  };

  return (
    <Command>
      <Command.Input
        autoFocus
        onValueChange={onQueryChange}
        placeholder="Search objectives..."
        value={query}
      />
      <Divider className="my-2" />
      <Command.List className="w-full border-0 bg-transparent p-0 shadow-none backdrop-blur-none dark:bg-transparent">
        {!isLoading ? (
          <Command.Empty className="px-3 py-2 text-left text-base">
            <Text color="muted">
              {needsSingleTeam
                ? "Select one team first."
                : "No objectives found."}
            </Text>
          </Command.Empty>
        ) : null}
        <Command.Group
          className="max-h-80 overflow-y-auto"
          onScroll={handleScroll}
        >
          {isLoading ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={5} />
            </Command.Loading>
          ) : null}
          {!needsSingleTeam
            ? objectives.map((objective, idx) => (
                <Command.Item
                  active={filters.objectiveId === objective.id}
                  className="justify-between gap-4"
                  key={objective.id}
                  onSelect={() => {
                    const objectiveId =
                      filters.objectiveId === objective.id
                        ? null
                        : objective.id;
                    setFilters({
                      ...filters,
                      objectiveId,
                      keyResultId:
                        objectiveId === filters.objectiveId
                          ? filters.keyResultId
                          : null,
                    });
                  }}
                  value={objective.name}
                >
                  <Flex align="center" className="min-w-0 flex-1" gap={2}>
                    <ObjectiveIcon className="text-text-secondary h-4 w-auto" />
                    <Text className="max-w-64 truncate">{objective.name}</Text>
                  </Flex>
                  <Flex align="center" className="shrink-0" gap={2}>
                    {filters.objectiveId === objective.id ? (
                      <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                    ) : null}
                    <Text color="muted">{idx}</Text>
                  </Flex>
                </Command.Item>
              ))
            : null}
          {isFetchingNextPage ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={2} />
            </Command.Loading>
          ) : null}
        </Command.Group>
      </Command.List>
    </Command>
  );
};

export const KeyResultEditor = ({
  filters,
  isPending,
  keyResultPluralLabel,
  keyResults,
  objectiveLabel,
  setFilters,
}: StoriesFilterEditorProps & {
  isPending: boolean;
  keyResultPluralLabel: string;
  keyResults: FilterKeyResultOption[];
  objectiveLabel: string;
}) => {
  return (
    <Command>
      <Command.Input
        autoFocus
        placeholder={`Search ${keyResultPluralLabel}...`}
      />
      <Divider className="my-2" />
      <Command.List className="w-full border-0 bg-transparent p-0 shadow-none backdrop-blur-none dark:bg-transparent">
        <Command.Group className="max-h-80 overflow-y-auto">
          {!filters.objectiveId ? (
            <Text className="px-3 py-2 text-left" color="muted" fontSize="md">
              Select an {objectiveLabel} filter first.
            </Text>
          ) : null}
          {isPending && filters.objectiveId ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={4} />
            </Command.Loading>
          ) : null}
          {!isPending && filters.objectiveId && keyResults.length === 0 ? (
            <Command.Empty className="px-3 py-2 text-left text-base">
              <Text color="muted">No {keyResultPluralLabel} found.</Text>
            </Command.Empty>
          ) : null}
          {keyResults.map((keyResult) => (
            <Command.Item
              active={filters.keyResultId === keyResult.id}
              className="justify-between gap-4"
              key={keyResult.id}
              onSelect={() => {
                setFilters({
                  ...filters,
                  keyResultId:
                    filters.keyResultId === keyResult.id ? null : keyResult.id,
                });
              }}
              value={keyResult.name}
            >
              <Flex align="center" className="min-w-0 flex-1" gap={2}>
                <OKRIcon className="text-text-secondary h-4 w-auto" />
                <Text className="max-w-72 truncate">{keyResult.name}</Text>
              </Flex>
              {filters.keyResultId === keyResult.id ? (
                <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
              ) : null}
            </Command.Item>
          ))}
        </Command.Group>
      </Command.List>
    </Command>
  );
};

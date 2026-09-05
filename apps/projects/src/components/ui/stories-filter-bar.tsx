"use client";
import type { ComponentProps } from "react";
import { useDeferredValue, useState } from "react";
import { useParams } from "next/navigation";
import { Button, Flex } from "ui";
import { PlusIcon } from "icons";
import type { EstimateScheme } from "@/lib/estimate";
import {
  OBJECTIVE_MENU_PAGE_SIZE,
  useTeamObjectivesInfinite,
  useTeamObjectives,
} from "@/modules/objectives/hooks/use-objectives";
import {
  SPRINT_MENU_PAGE_SIZE,
  useTeamSprintsInfinite,
  useTeamSprints,
} from "@/modules/sprints/hooks/team-sprints";
import {
  TEAM_MENU_PAGE_SIZE,
  useTeamsInfinite,
  useTeams,
} from "@/modules/teams/hooks/teams";
import {
  useTeamMembersInfinite,
  useTeamMembers,
} from "@/lib/hooks/team-members";
import {
  MEMBER_MENU_PAGE_SIZE,
  useMembersInfinite,
  useMembers,
} from "@/lib/hooks/members";
import { useStatuses } from "@/lib/hooks/statuses";
import { useTeamSettings } from "@/modules/teams/hooks/use-team-settings";
import { useKeyResults } from "@/modules/objectives/hooks";
import { useTerminology } from "@/hooks/use-terminology-display";
import { useLabels } from "@/lib/hooks/labels";
import { DEFAULT_ESTIMATE_SCHEME } from "@/lib/estimate";
import {
  KeyResultEditor,
  ObjectiveEditor,
  SprintEditor,
  TeamEditor,
} from "./stories-filter-bar/planning-editors";
import {
  PeopleEditor,
  PriorityEditor,
  StatusEditor,
  type FilterMemberOption,
  type FilterStatusOption,
} from "./stories-filter-bar/people-editors";
import {
  DateEditor,
  EstimateEditor,
  LabelEditor,
} from "./stories-filter-bar/attribute-editors";
import { getScopedStoriesFilterTeamId } from "./stories-filter-query";
import type { StoriesFilterOperator } from "./stories-filter-types";
import { getStoriesFilterOperator } from "./stories-filter-types";
import { hasActiveStoriesFilters } from "./stories-filter-utils";
import { buildFilterChips } from "./stories-filter-bar/filter-chips";
import {
  StoriesFilterChip,
  TitleFilterDialog,
} from "./stories-filter-bar/filter-chip";
import {
  EMPTY_FILTER_FIELDS,
  isFilterOperatorField,
  removeStoriesFilterField,
} from "./stories-filter-bar/filter-model";
import { StoriesFilterMenu as FilterMenu } from "./stories-filter-menu";
import type {
  StoriesFilterEditorProps,
  StoriesFilterBarProps,
  StoriesFilterField,
} from "./stories-filter-bar/types";

export type { StoriesFilterField } from "./stories-filter-bar/types";

type EditorContainerProps = StoriesFilterEditorProps & { teamId?: string };

const PeopleEditorContainer = ({
  field,
  filters,
  setFilters,
  teamId,
}: EditorContainerProps & { field: "assigneeIds" | "reporterIds" }) => {
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const workspaceMembersQuery = useMembersInfinite(
    deferredQuery,
    MEMBER_MENU_PAGE_SIZE,
    !teamId,
  );
  const teamMembersQuery = useTeamMembersInfinite(
    teamId,
    deferredQuery,
    MEMBER_MENU_PAGE_SIZE,
    Boolean(teamId),
  );
  const membersQuery = teamId ? teamMembersQuery : workspaceMembersQuery;
  const members =
    membersQuery.data?.pages.flatMap((page) => page.members) ?? [];

  return (
    <PeopleEditor
      field={field}
      filters={filters}
      hasNextPage={Boolean(membersQuery.hasNextPage)}
      isFetchingNextPage={membersQuery.isFetchingNextPage}
      isLoading={Boolean(
        membersQuery.isFetching && !membersQuery.isFetchingNextPage,
      )}
      members={members satisfies FilterMemberOption[]}
      onFetchNextPage={() => {
        void membersQuery.fetchNextPage();
      }}
      onQueryChange={setQuery}
      query={query}
      setFilters={setFilters}
    />
  );
};

const TeamEditorContainer = ({
  filters,
  setFilters,
}: StoriesFilterEditorProps) => {
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const teamsQuery = useTeamsInfinite(deferredQuery, TEAM_MENU_PAGE_SIZE);
  const teams = teamsQuery.data?.pages.flatMap((page) => page.teams) ?? [];

  return (
    <TeamEditor
      filters={filters}
      hasNextPage={Boolean(teamsQuery.hasNextPage)}
      isFetchingNextPage={teamsQuery.isFetchingNextPage}
      isLoading={Boolean(
        teamsQuery.isFetching && !teamsQuery.isFetchingNextPage,
      )}
      onFetchNextPage={() => {
        void teamsQuery.fetchNextPage();
      }}
      onQueryChange={setQuery}
      query={query}
      setFilters={setFilters}
      teams={teams}
    />
  );
};

const SprintEditorContainer = ({
  filters,
  setFilters,
  teamId,
}: EditorContainerProps) => {
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const scopedTeamId = getScopedStoriesFilterTeamId(
    teamId,
    filters.teamIds,
    getStoriesFilterOperator(filters, "teamIds"),
  );
  const sprintsQuery = useTeamSprintsInfinite(
    scopedTeamId ?? "",
    deferredQuery,
    SPRINT_MENU_PAGE_SIZE,
  );
  const sprints =
    sprintsQuery.data?.pages.flatMap((page) => page.sprints) ?? [];

  return (
    <SprintEditor
      filters={filters}
      hasNextPage={Boolean(sprintsQuery.hasNextPage)}
      isFetchingNextPage={sprintsQuery.isFetchingNextPage}
      isLoading={Boolean(
        scopedTeamId &&
          sprintsQuery.isFetching &&
          !sprintsQuery.isFetchingNextPage,
      )}
      needsSingleTeam={!scopedTeamId}
      onFetchNextPage={() => {
        void sprintsQuery.fetchNextPage();
      }}
      onQueryChange={setQuery}
      query={query}
      setFilters={setFilters}
      sprints={sprints}
    />
  );
};

const ObjectiveEditorContainer = ({
  filters,
  setFilters,
  teamId,
}: EditorContainerProps) => {
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const scopedTeamId = getScopedStoriesFilterTeamId(
    teamId,
    filters.teamIds,
    getStoriesFilterOperator(filters, "teamIds"),
  );
  const objectivesQuery = useTeamObjectivesInfinite(
    scopedTeamId ?? "",
    deferredQuery,
    OBJECTIVE_MENU_PAGE_SIZE,
  );
  const objectives =
    objectivesQuery.data?.pages.flatMap((page) => page.objectives) ?? [];

  return (
    <ObjectiveEditor
      filters={filters}
      hasNextPage={Boolean(objectivesQuery.hasNextPage)}
      isFetchingNextPage={objectivesQuery.isFetchingNextPage}
      isLoading={Boolean(
        scopedTeamId &&
          objectivesQuery.isFetching &&
          !objectivesQuery.isFetchingNextPage,
      )}
      needsSingleTeam={!scopedTeamId}
      objectives={objectives}
      onFetchNextPage={() => {
        void objectivesQuery.fetchNextPage();
      }}
      onQueryChange={setQuery}
      query={query}
      setFilters={setFilters}
    />
  );
};

const KeyResultEditorContainer = ({
  filters,
  setFilters,
}: StoriesFilterEditorProps) => {
  const { getTermDisplay } = useTerminology();
  const keyResultsQuery = useKeyResults(
    filters.objectiveId ?? "",
    Boolean(filters.objectiveId),
  );

  return (
    <KeyResultEditor
      filters={filters}
      isPending={keyResultsQuery.isPending}
      keyResultPluralLabel={getTermDisplay("keyResultTerm", {
        variant: "plural",
      })}
      keyResults={keyResultsQuery.data ?? []}
      objectiveLabel={getTermDisplay("objectiveTerm")}
      setFilters={setFilters}
    />
  );
};

const FilterValueEditor = ({
  allStatuses,
  estimateScheme,
  field,
  filters,
  setFilters,
  teamId,
}: StoriesFilterEditorProps & {
  allStatuses: (FilterStatusOption & { teamId: string })[];
  estimateScheme: EstimateScheme;
  field: StoriesFilterField;
  teamId?: string;
}) => {
  if (field === "statusIds") {
    const statuses = teamId
      ? allStatuses.filter((status) => status.teamId === teamId)
      : allStatuses;
    return (
      <StatusEditor
        filters={filters}
        setFilters={setFilters}
        statuses={statuses}
      />
    );
  }
  if (field === "assigneeIds" || field === "reporterIds") {
    return (
      <PeopleEditorContainer
        field={field}
        filters={filters}
        setFilters={setFilters}
        teamId={teamId}
      />
    );
  }
  if (field === "priorities") {
    return <PriorityEditor filters={filters} setFilters={setFilters} />;
  }
  if (field === "teamIds") {
    return <TeamEditorContainer filters={filters} setFilters={setFilters} />;
  }
  if (field === "sprintIds") {
    return (
      <SprintEditorContainer
        filters={filters}
        setFilters={setFilters}
        teamId={teamId}
      />
    );
  }
  if (field === "objectiveId") {
    return (
      <ObjectiveEditorContainer
        filters={filters}
        setFilters={setFilters}
        teamId={teamId}
      />
    );
  }
  if (field === "keyResultId") {
    return (
      <KeyResultEditorContainer filters={filters} setFilters={setFilters} />
    );
  }
  if (field === "labelIds") {
    return <LabelEditor filters={filters} setFilters={setFilters} />;
  }
  if (field === "estimateValues") {
    return (
      <EstimateEditor
        estimateScheme={estimateScheme}
        filters={filters}
        setFilters={setFilters}
      />
    );
  }
  if (field === "startDate" || field === "endDate") {
    return (
      <DateEditor field={field} filters={filters} setFilters={setFilters} />
    );
  }
  return null;
};

export const StoriesFilterMenu = (
  props: Omit<ComponentProps<typeof FilterMenu>, "renderEditor">,
) => {
  const { teamId } = useParams<{ teamId?: string }>();
  const { data: allStatuses = [] } = useStatuses();
  const scopedTeamId = getScopedStoriesFilterTeamId(
    teamId,
    props.filters.teamIds,
    getStoriesFilterOperator(props.filters, "teamIds"),
  );
  const { data: teamSettings } = useTeamSettings(scopedTeamId);
  const estimateScheme =
    teamSettings?.estimationSettings.scheme ?? DEFAULT_ESTIMATE_SCHEME;
  return (
    <FilterMenu
      {...props}
      renderEditor={(field) => (
        <FilterValueEditor
          allStatuses={allStatuses}
          estimateScheme={estimateScheme}
          field={field}
          filters={props.filters}
          setFilters={props.setFilters}
          teamId={teamId}
        />
      )}
    />
  );
};

export const StoriesFilterBar = ({
  filters,
  setFilters,
  resetFilters,
  hiddenFields = EMPTY_FILTER_FIELDS,
  showWhenEmpty = false,
}: StoriesFilterBarProps) => {
  const { getTermDisplay } = useTerminology();
  const { teamId } = useParams<{ teamId?: string }>();
  const [titleDialogOpen, setTitleDialogOpen] = useState(false);
  const scopedTeamId = getScopedStoriesFilterTeamId(
    teamId,
    filters.teamIds,
    getStoriesFilterOperator(filters, "teamIds"),
  );
  const { data: allStatuses = [] } = useStatuses();
  const { data: allUsers = [] } = useMembers();
  const resolvedTeamId = scopedTeamId ?? "";
  const { data: teamMembers = [] } = useTeamMembers(resolvedTeamId);
  const { data: teams = [] } = useTeams();
  const { data: sprints = [] } = useTeamSprints(resolvedTeamId);
  const { data: objectives = [] } = useTeamObjectives(resolvedTeamId);
  const { data: keyResults = [] } = useKeyResults(
    filters.objectiveId ?? "",
    Boolean(filters.objectiveId),
  );
  const { data: allLabels = [] } = useLabels();
  const { data: teamSettings } = useTeamSettings(scopedTeamId);
  const estimateScheme =
    teamSettings?.estimationSettings.scheme ?? DEFAULT_ESTIMATE_SCHEME;
  const hiddenFieldSet = new Set(hiddenFields);
  const users = scopedTeamId ? teamMembers : allUsers;
  const statuses = scopedTeamId
    ? allStatuses.filter((status) => status.teamId === scopedTeamId)
    : allStatuses;

  const statusById = new Map(
    statuses.map((status) => [status.id, status.name]),
  );
  const userById = new Map(
    users.map((user) => {
      const username = user.username || user.email || "Unknown user";
      return [
        user.id,
        {
          avatarUrl: user.avatarUrl ?? null,
          id: user.id,
          name: user.fullName || username,
          username,
        },
      ];
    }),
  );
  const teamById = new Map(teams.map((team) => [team.id, team.name]));
  const teamColorById = new Map(teams.map((team) => [team.id, team.color]));
  const sprintById = new Map(sprints.map((sprint) => [sprint.id, sprint.name]));
  const objectiveById = new Map(
    objectives.map((objective) => [objective.id, objective.name]),
  );
  const keyResultById = new Map(
    keyResults.map((keyResult) => [keyResult.id, keyResult.name]),
  );
  const labelById = new Map(
    allLabels.map((label) => [
      label.id,
      { color: label.color, id: label.id, name: label.name },
    ]),
  );
  const chips = buildFilterChips({
    estimateScheme,
    filters,
    getTermDisplay,
    hiddenFields: hiddenFieldSet,
    keyResults: keyResultById,
    labels: labelById,
    objectives: objectiveById,
    sprints: sprintById,
    statuses: statusById,
    teamColors: teamColorById,
    teams: teamById,
    users: userById,
  });

  const renderEditor = (field: StoriesFilterField) => (
    <FilterValueEditor
      allStatuses={allStatuses}
      estimateScheme={estimateScheme}
      field={field}
      filters={filters}
      setFilters={setFilters}
      teamId={teamId}
    />
  );

  if (!showWhenEmpty && chips.length === 0) return null;

  return (
    <Flex
      align="center"
      className="border-border bg-background h-(--app-filter-bar-height) shrink-0 border-b-[0.5px] px-4"
      gap={3}
      justify="between"
    >
      <Flex align="center" className="min-w-0 flex-1 overflow-x-auto" gap={2}>
        {chips.map((chip) => (
          <StoriesFilterChip
            chip={chip}
            key={chip.field}
            onEditTitle={() => {
              setTitleDialogOpen(true);
            }}
            onOperatorChange={(operator: StoriesFilterOperator) => {
              if (!isFilterOperatorField(chip.field)) return;
              setFilters({
                ...filters,
                operators: { ...filters.operators, [chip.field]: operator },
              });
            }}
            onRemove={() => {
              setFilters(removeStoriesFilterField(filters, chip.field));
            }}
            renderEditor={renderEditor}
          />
        ))}
        <StoriesFilterMenu
          filters={filters}
          hiddenFields={hiddenFields}
          setFilters={setFilters}
        >
          <Button
            aria-label="Add filter"
            color="tertiary"
            leftIcon={<PlusIcon className="h-4 w-auto" />}
            size="sm"
            variant="outline"
          />
        </StoriesFilterMenu>
        {titleDialogOpen ? (
          <TitleFilterDialog
            filters={filters}
            key={filters.contentContains ?? ""}
            onOpenChange={setTitleDialogOpen}
            open={titleDialogOpen}
            setFilters={setFilters}
          />
        ) : null}
      </Flex>
      {hasActiveStoriesFilters(filters) ? (
        <Flex align="center" className="shrink-0" gap={2}>
          <Button
            color="tertiary"
            onClick={resetFilters}
            size="sm"
            variant="outline"
          >
            Clear all
          </Button>
        </Flex>
      ) : null}
    </Flex>
  );
};

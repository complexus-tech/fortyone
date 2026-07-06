import type {
  CreatedStory,
  CreateStoryFromSlackInput,
  RuntimeLogInput,
  RuntimeOption,
  SlackActor,
  StoryRuntime,
} from "@/lib/runtime";

interface FixtureTeam {
  id: string;
  code: string;
  name: string;
}

export const DEFAULT_TEAM_OPTION: RuntimeOption = {
  label: "Product (PROD)",
  value: "team-product",
};

const teams: FixtureTeam[] = [
  { id: "team-product", code: "PROD", name: "Product" },
  { id: "team-engineering", code: "ENG", name: "Engineering" },
  { id: "team-growth", code: "GROW", name: "Growth" },
];

const statuses: Record<string, RuntimeOption[]> = {
  "team-engineering": [
    { label: "Backlog", value: "eng-backlog" },
    { label: "In Progress", value: "eng-progress" },
    { label: "Ready for Review", value: "eng-review" },
  ],
  "team-growth": [
    { label: "Ideas", value: "growth-ideas" },
    { label: "Active", value: "growth-active" },
    { label: "Shipped", value: "growth-shipped" },
  ],
  "team-product": [
    { label: "Discovery", value: "product-discovery" },
    { label: "Planned", value: "product-planned" },
    { label: "Done", value: "product-done" },
  ],
};

const members: Record<string, RuntimeOption[]> = {
  "team-engineering": [
    { label: "Joseph Mukorivo", value: "user-joseph" },
    { label: "Maya Test Engineer", value: "user-maya-eng" },
  ],
  "team-growth": [
    { label: "Growth Lead", value: "user-growth-lead" },
    { label: "Lifecycle Marketer", value: "user-lifecycle" },
  ],
  "team-product": [
    { label: "Product Manager", value: "user-product-manager" },
    { label: "Designer", value: "user-designer" },
  ],
};

const objectives: Record<string, RuntimeOption[]> = {
  "team-engineering": [
    {
      label: "Improve Slack integration reliability",
      value: "obj-slack-reliability",
    },
    { label: "Reduce story creation friction", value: "obj-story-friction" },
  ],
  "team-growth": [
    { label: "Increase workspace activation", value: "obj-activation" },
    { label: "Improve onboarding completion", value: "obj-onboarding" },
  ],
  "team-product": [
    { label: "Validate Maya workflows", value: "obj-maya-workflows" },
    { label: "Ship core planning loop", value: "obj-planning-loop" },
  ],
};

const labels: Record<string, RuntimeOption[]> = {
  "team-engineering": [
    { label: "Bug", value: "label-bug" },
    { label: "Integration", value: "label-integration" },
  ],
  "team-growth": [
    { label: "Experiment", value: "label-experiment" },
    { label: "Lifecycle", value: "label-lifecycle" },
  ],
  "team-product": [
    { label: "Research", value: "label-research" },
    { label: "UX", value: "label-ux" },
  ],
};

const includesQuery = (option: RuntimeOption, query: string) =>
  option.label.toLowerCase().includes(query.trim().toLowerCase());

const search = (options: RuntimeOption[], query: string) => {
  const trimmed = query.trim();
  return trimmed
    ? options.filter((option) => includesQuery(option, trimmed))
    : options;
};

const optionsForTeam = (
  optionsByTeam: Record<string, RuntimeOption[]>,
  teamId: string | undefined,
): RuntimeOption[] => {
  return teamId ? optionsByTeam[teamId] ?? [] : [];
};

const teamOptions = (): RuntimeOption[] =>
  teams.map((team) => ({
    label: `${team.name} (${team.code})`,
    value: team.id,
  }));

const createStoryRef = () => `SLACK-${Math.floor(1000 + Math.random() * 9000)}`;

export const localRuntime: StoryRuntime = {
  createStoryFromSlackForm(input: CreateStoryFromSlackInput): CreatedStory {
    const story = {
      id: `local-${Date.now()}`,
      ref: createStoryRef(),
      title: input.title.trim(),
      url: "https://app.fortyone.local/stories/local",
    };

    // eslint-disable-next-line no-console -- Local Slack story creation is intentionally logged until the Go API endpoint is wired in.
    console.info(
      JSON.stringify({
        level: "info",
        message: "Slack story created locally",
        story,
        input,
      }),
    );

    return story;
  },

  listStoryOptions(_actor: SlackActor) {
    return {
      labels,
      members,
      objectives,
      statuses,
      teams: teamOptions(),
    };
  },

  async recordSlackRuntimeLog(_input: RuntimeLogInput): Promise<void> {
    await Promise.resolve();
  },

  async recordSlackThreadComment(_input: {
    actor: SlackActor;
    messageText: string;
    storyId: string;
  }): Promise<void> {
    await Promise.resolve();
  },

  async resolveSlackIdentity(actor: SlackActor) {
    await Promise.resolve();
    return {
      userId: actor.userId,
      workspaceId: "local-workspace",
      workspaceSlug: "local-workspace",
    };
  },

  searchLabels(
    _actor: SlackActor,
    teamId: string | undefined,
    query: string,
  ): RuntimeOption[] {
    return search(optionsForTeam(labels, teamId), query);
  },

  searchMembers(
    _actor: SlackActor,
    teamId: string | undefined,
    query: string,
  ): RuntimeOption[] {
    return search(optionsForTeam(members, teamId), query);
  },

  searchObjectives(
    _actor: SlackActor,
    teamId: string | undefined,
    query: string,
  ): RuntimeOption[] {
    return search(optionsForTeam(objectives, teamId), query);
  },

  searchStatuses(
    _actor: SlackActor,
    teamId: string | undefined,
    query: string,
  ): RuntimeOption[] {
    return search(optionsForTeam(statuses, teamId), query);
  },

  searchTeams(_actor: SlackActor, query: string): RuntimeOption[] {
    return search(teamOptions(), query);
  },
};

export type LocalRuntime = StoryRuntime;

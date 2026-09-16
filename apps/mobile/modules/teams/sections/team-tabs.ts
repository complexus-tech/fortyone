export type TeamSection = "all" | "objectives" | "feedback" | "intake";
export type TeamSectionAvailability = {
  objectives: boolean;
  feedback: boolean;
  intake: boolean;
};

export function getTeamTabs(
  storyLabel: string,
  objectiveLabel: string,
  available: TeamSectionAvailability,
): { value: TeamSection; label: string }[] {
  return [
    { value: "all", label: `All ${storyLabel}` },
    ...(available.objectives
      ? [{ value: "objectives" as const, label: objectiveLabel }]
      : []),
    ...(available.feedback
      ? [{ value: "feedback" as const, label: "Feedback" }]
      : []),
    ...(available.intake
      ? [{ value: "intake" as const, label: "Intake" }]
      : []),
  ];
}

export function resolveTeamTab(
  requested: string,
  tabs: readonly { value: TeamSection }[],
): TeamSection {
  return tabs.find((tab) => tab.value === requested)?.value ?? "all";
}

export function nextTeamPage(pagination: {
  page: number;
  nextPage: number;
  hasMore: boolean;
}) {
  return pagination.hasMore && pagination.nextPage > pagination.page
    ? pagination.nextPage
    : undefined;
}

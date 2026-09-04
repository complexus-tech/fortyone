import {
  normalizeImportTaskLinks,
  type ImportAnalysis,
} from "@/modules/settings/workspace/imports/schema";

const normalizeDate = (value: string | null) => {
  const trimmed = value?.trim();
  if (!trimmed || !/^\d{4}-\d{2}-\d{2}$/.test(trimmed)) return null;
  const parsed = new Date(`${trimmed}T00:00:00.000Z`);
  return !Number.isNaN(parsed.getTime()) &&
    parsed.toISOString().slice(0, 10) === trimmed
    ? trimmed
    : null;
};

const normalizeOptionalText = (value: string | null) => value?.trim() || null;

export const normalizeSourceNamespace = (value: string | null) => {
  const normalized = normalizeOptionalText(value);
  if (
    !normalized ||
    /\p{Cc}/u.test(normalized) ||
    Buffer.byteLength(normalized, "utf8") > 300
  ) {
    return null;
  }
  return normalized;
};

const normalizeSourceIdList = (values: string[]) => {
  const normalized = new Set<string>();
  for (const value of values) {
    const trimmed = value.trim();
    if (trimmed) normalized.add(trimmed);
  }
  return [...normalized];
};

const omitDuplicateSourceIds = <T extends { sourceId: string }>(
  entities: T[],
) => {
  const counts = entities.reduce<Map<string, number>>((result, entity) => {
    result.set(entity.sourceId, (result.get(entity.sourceId) ?? 0) + 1);
    return result;
  }, new Map());
  const duplicateSourceIds = new Set(
    [...counts].flatMap(([sourceId, count]) => (count > 1 ? [sourceId] : [])),
  );

  return {
    duplicateSourceIdCount: duplicateSourceIds.size,
    entities: entities.filter(
      (entity) => !duplicateSourceIds.has(entity.sourceId),
    ),
    omittedObjectCount: entities.filter((entity) =>
      duplicateSourceIds.has(entity.sourceId),
    ).length,
  };
};

export const normalizeAnalysis = (analysis: ImportAnalysis): ImportAnalysis => {
  const sourceDateRanges = [
    ...analysis.objectives,
    ...analysis.keyResults,
    ...analysis.sprints,
    ...analysis.tasks,
  ].map(({ startDate, endDate }) => ({ startDate, endDate }));
  const invalidCalendarDateCount = sourceDateRanges.reduce(
    (total, { startDate, endDate }) =>
      total +
      [startDate, endDate].filter(
        (value) => value !== null && normalizeDate(value) === null,
      ).length,
    0,
  );
  const reversedDateRangeCount = sourceDateRanges.filter(
    ({ startDate, endDate }) => {
      const normalizedStartDate = normalizeDate(startDate);
      const normalizedEndDate = normalizeDate(endDate);
      return Boolean(
        normalizedStartDate &&
          normalizedEndDate &&
          normalizedEndDate < normalizedStartDate,
      );
    },
  ).length;
  const normalizedWithDuplicates: ImportAnalysis = {
    ...analysis,
    sourceNamespace: normalizeSourceNamespace(analysis.sourceNamespace),
    teams: analysis.teams.map((team) => ({
      ...team,
      sourceId: team.sourceId.trim(),
      name: team.name.trim(),
      code: normalizeOptionalText(team.code),
      color: normalizeOptionalText(team.color),
      description: normalizeOptionalText(team.description),
    })),
    people: analysis.people.map((person) => ({
      ...person,
      sourceId: person.sourceId.trim(),
      name: normalizeOptionalText(person.name),
      email: person.email?.trim().toLowerCase() || null,
      teamSourceIds: normalizeSourceIdList(person.teamSourceIds),
    })),
    labels: analysis.labels.map((label) => ({
      ...label,
      sourceId: label.sourceId.trim(),
      name: label.name.trim(),
      color: normalizeOptionalText(label.color),
      teamSourceId: normalizeOptionalText(label.teamSourceId),
    })),
    strategicPillars: analysis.strategicPillars.map((pillar) => ({
      ...pillar,
      sourceId: pillar.sourceId.trim(),
      name: pillar.name.trim(),
      description: normalizeOptionalText(pillar.description),
    })),
    objectives: analysis.objectives.map((objective) => ({
      ...objective,
      sourceId: objective.sourceId.trim(),
      name: objective.name.trim(),
      description: normalizeOptionalText(objective.description),
      shortSummary: normalizeOptionalText(objective.shortSummary),
      color: normalizeOptionalText(objective.color),
      status: normalizeOptionalText(objective.status),
      leadPersonSourceId: normalizeOptionalText(objective.leadPersonSourceId),
      teamSourceId: normalizeOptionalText(objective.teamSourceId),
      pillarSourceId: normalizeOptionalText(objective.pillarSourceId),
      startDate: normalizeDate(objective.startDate),
      endDate: normalizeDate(objective.endDate),
    })),
    keyResults: analysis.keyResults.map((keyResult) => ({
      ...keyResult,
      sourceId: keyResult.sourceId.trim(),
      name: keyResult.name.trim(),
      objectiveSourceId: normalizeOptionalText(keyResult.objectiveSourceId),
      leadPersonSourceId: normalizeOptionalText(keyResult.leadPersonSourceId),
      contributorPersonSourceIds: normalizeSourceIdList(
        keyResult.contributorPersonSourceIds,
      ),
      startDate: normalizeDate(keyResult.startDate),
      endDate: normalizeDate(keyResult.endDate),
    })),
    sprints: analysis.sprints.map((sprint) => ({
      ...sprint,
      sourceId: sprint.sourceId.trim(),
      name: sprint.name.trim(),
      goal: normalizeOptionalText(sprint.goal),
      teamSourceId: normalizeOptionalText(sprint.teamSourceId),
      objectiveSourceId: normalizeOptionalText(sprint.objectiveSourceId),
      startDate: normalizeDate(sprint.startDate),
      endDate: normalizeDate(sprint.endDate),
    })),
    tasks: analysis.tasks.map((task, index) => {
      const sourceId = task.sourceId.trim() || `row-${index + 2}`;
      const assigneePersonSourceId = normalizeOptionalText(
        task.assigneePersonSourceId,
      );
      return {
        ...task,
        sourceId,
        title: task.title.trim(),
        description: task.description.trim(),
        status: normalizeOptionalText(task.status),
        assigneeEmail: task.assigneeEmail?.trim().toLowerCase() || null,
        assigneeName: normalizeOptionalText(task.assigneeName),
        assigneePersonSourceId,
        collaboratorPersonSourceIds: normalizeSourceIdList(
          task.collaboratorPersonSourceIds,
        ).filter((sourceId) => sourceId !== assigneePersonSourceId),
        teamSourceId: normalizeOptionalText(task.teamSourceId),
        parentSourceId: normalizeOptionalText(task.parentSourceId),
        objectiveSourceId: normalizeOptionalText(task.objectiveSourceId),
        keyResultSourceId: normalizeOptionalText(task.keyResultSourceId),
        sprintSourceId: normalizeOptionalText(task.sprintSourceId),
        labelSourceIds: normalizeSourceIdList(task.labelSourceIds),
        associations: task.associations.map((association) => ({
          ...association,
          targetSourceId: association.targetSourceId.trim(),
        })),
        links: normalizeImportTaskLinks(task.links),
        startDate: normalizeDate(task.startDate),
        endDate: normalizeDate(task.endDate),
      };
    }),
  };

  const teams = omitDuplicateSourceIds(normalizedWithDuplicates.teams);
  const people = omitDuplicateSourceIds(normalizedWithDuplicates.people);
  const labels = omitDuplicateSourceIds(normalizedWithDuplicates.labels);
  const strategicPillars = omitDuplicateSourceIds(
    normalizedWithDuplicates.strategicPillars,
  );
  const objectives = omitDuplicateSourceIds(
    normalizedWithDuplicates.objectives,
  );
  const keyResults = omitDuplicateSourceIds(
    normalizedWithDuplicates.keyResults,
  );
  const sprints = omitDuplicateSourceIds(normalizedWithDuplicates.sprints);
  const tasks = omitDuplicateSourceIds(normalizedWithDuplicates.tasks);
  const duplicateCollections = [
    teams,
    people,
    labels,
    strategicPillars,
    objectives,
    keyResults,
    sprints,
    tasks,
  ];
  const duplicateSourceIdCount = duplicateCollections.reduce(
    (total, collection) => total + collection.duplicateSourceIdCount,
    0,
  );
  const omittedDuplicateObjectCount = duplicateCollections.reduce(
    (total, collection) => total + collection.omittedObjectCount,
    0,
  );
  const taskIds = new Set(tasks.entities.map((task) => task.sourceId));
  let duplicateTaskAssociationCount = 0;
  let selfTaskAssociationCount = 0;
  let danglingTaskAssociationCount = 0;
  const normalizedTasks = tasks.entities.map((task) => {
    const associationKeys = new Set<string>();
    const associations = task.associations.flatMap((association) => {
      if (association.targetSourceId === task.sourceId) {
        selfTaskAssociationCount += 1;
        return [];
      }
      const associationKey = `${association.type}\u0000${association.targetSourceId}`;
      if (associationKeys.has(associationKey)) {
        duplicateTaskAssociationCount += 1;
        return [];
      }
      associationKeys.add(associationKey);
      if (!taskIds.has(association.targetSourceId)) {
        danglingTaskAssociationCount += 1;
        return [];
      }
      return [association];
    });
    return { ...task, associations };
  });
  const normalized: ImportAnalysis = {
    ...normalizedWithDuplicates,
    teams: teams.entities,
    people: people.entities,
    labels: labels.entities,
    strategicPillars: strategicPillars.entities,
    objectives: objectives.entities,
    keyResults: keyResults.entities,
    sprints: sprints.entities,
    tasks: normalizedTasks,
  };
  const teamIds = new Set(normalized.teams.map((team) => team.sourceId));
  const personIds = new Set(normalized.people.map((person) => person.sourceId));
  const labelIds = new Set(normalized.labels.map((label) => label.sourceId));
  const pillarIds = new Set(
    normalized.strategicPillars.map((pillar) => pillar.sourceId),
  );
  const objectiveIds = new Set(
    normalized.objectives.map((objective) => objective.sourceId),
  );
  const keyResultIds = new Set(
    normalized.keyResults.map((keyResult) => keyResult.sourceId),
  );
  const keyResultsById = new Map(
    normalized.keyResults.map((keyResult) => [keyResult.sourceId, keyResult]),
  );
  const sprintIds = new Set(
    normalized.sprints.map((sprint) => sprint.sourceId),
  );
  const isDangling = (value: string | null, sourceIds: Set<string>) =>
    Boolean(value && !sourceIds.has(value));
  let danglingReferenceCount = 0;
  for (const person of normalized.people) {
    danglingReferenceCount += person.teamSourceIds.filter(
      (sourceId) => !teamIds.has(sourceId),
    ).length;
  }
  for (const label of normalized.labels) {
    if (isDangling(label.teamSourceId, teamIds)) danglingReferenceCount += 1;
  }
  for (const objective of normalized.objectives) {
    if (isDangling(objective.pillarSourceId, pillarIds)) {
      danglingReferenceCount += 1;
    }
    if (isDangling(objective.teamSourceId, teamIds)) {
      danglingReferenceCount += 1;
    }
    if (isDangling(objective.leadPersonSourceId, personIds)) {
      danglingReferenceCount += 1;
    }
  }
  for (const keyResult of normalized.keyResults) {
    if (isDangling(keyResult.objectiveSourceId, objectiveIds)) {
      danglingReferenceCount += 1;
    }
    if (isDangling(keyResult.leadPersonSourceId, personIds)) {
      danglingReferenceCount += 1;
    }
    danglingReferenceCount += keyResult.contributorPersonSourceIds.filter(
      (sourceId) => !personIds.has(sourceId),
    ).length;
  }
  for (const sprint of normalized.sprints) {
    if (isDangling(sprint.teamSourceId, teamIds)) {
      danglingReferenceCount += 1;
    }
    if (isDangling(sprint.objectiveSourceId, objectiveIds)) {
      danglingReferenceCount += 1;
    }
  }
  for (const task of normalized.tasks) {
    if (isDangling(task.teamSourceId, teamIds)) danglingReferenceCount += 1;
    if (isDangling(task.parentSourceId, taskIds)) danglingReferenceCount += 1;
    if (isDangling(task.objectiveSourceId, objectiveIds)) {
      danglingReferenceCount += 1;
    }
    if (isDangling(task.keyResultSourceId, keyResultIds)) {
      danglingReferenceCount += 1;
    }
    if (isDangling(task.sprintSourceId, sprintIds)) {
      danglingReferenceCount += 1;
    }
    if (isDangling(task.assigneePersonSourceId, personIds)) {
      danglingReferenceCount += 1;
    }
    danglingReferenceCount += task.collaboratorPersonSourceIds.filter(
      (sourceId) => !personIds.has(sourceId),
    ).length;
    danglingReferenceCount += task.labelSourceIds.filter(
      (sourceId) => !labelIds.has(sourceId),
    ).length;
  }

  const taskObjectiveKeyResultMismatchCount = normalized.tasks.filter(
    (task) => {
      if (!task.objectiveSourceId || !task.keyResultSourceId) return false;
      const keyResult = keyResultsById.get(task.keyResultSourceId);
      return Boolean(
        keyResult?.objectiveSourceId &&
          keyResult.objectiveSourceId !== task.objectiveSourceId,
      );
    },
  ).length;
  const unsupportedTeamDescriptionCount = normalized.teams.filter(
    (team) => team.description,
  ).length;

  const validationWarnings = [
    ...(duplicateSourceIdCount
      ? [
          `${omittedDuplicateObjectCount} source ${omittedDuplicateObjectCount === 1 ? "object was" : "objects were"} omitted because ${duplicateSourceIdCount} source ${duplicateSourceIdCount === 1 ? "ID was" : "IDs were"} duplicated and could not be related safely.`,
        ]
      : []),
    ...(duplicateTaskAssociationCount
      ? [
          `${duplicateTaskAssociationCount} duplicate task ${duplicateTaskAssociationCount === 1 ? "association was" : "associations were"} deduplicated.`,
        ]
      : []),
    ...(selfTaskAssociationCount
      ? [
          `${selfTaskAssociationCount} self-referential task ${selfTaskAssociationCount === 1 ? "association was" : "associations were"} removed.`,
        ]
      : []),
    ...(danglingTaskAssociationCount
      ? [
          `${danglingTaskAssociationCount} task ${danglingTaskAssociationCount === 1 ? "association targeting an unreturned task was" : "associations targeting unreturned tasks were"} removed.`,
        ]
      : []),
    ...(danglingReferenceCount
      ? [
          `${danglingReferenceCount} source ${danglingReferenceCount === 1 ? "relationship points" : "relationships point"} to objects that were not returned and will use safe fallbacks.`,
        ]
      : []),
    ...(taskObjectiveKeyResultMismatchCount
      ? [
          `${taskObjectiveKeyResultMismatchCount} task ${taskObjectiveKeyResultMismatchCount === 1 ? "relationship has conflicting objective and key-result references and needs" : "relationships have conflicting objective and key-result references and need"} review.`,
        ]
      : []),
    ...(unsupportedTeamDescriptionCount
      ? [
          `${unsupportedTeamDescriptionCount} source team ${unsupportedTeamDescriptionCount === 1 ? "description remains" : "descriptions remain"} visible for review but cannot be applied by FortyOne's team creation contract.`,
        ]
      : []),
    ...(invalidCalendarDateCount
      ? [
          `${invalidCalendarDateCount} invalid calendar ${invalidCalendarDateCount === 1 ? "date was" : "dates were"} omitted instead of being guessed.`,
        ]
      : []),
    ...(reversedDateRangeCount
      ? [
          `${reversedDateRangeCount} reversed date ${reversedDateRangeCount === 1 ? "range needs" : "ranges need"} review and will be skipped or imported without dates.`,
        ]
      : []),
  ];

  return {
    ...normalized,
    warnings: [
      ...new Set([...validationWarnings, ...normalized.warnings]),
    ].slice(0, 50),
  };
};

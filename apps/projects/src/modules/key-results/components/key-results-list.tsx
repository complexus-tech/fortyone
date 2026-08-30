"use client";

import { Box } from "ui";
import type { ObjectiveKeyResultGroup } from "../utils";
import type { KeyResultsMember } from "./key-results-member";
import { KeyResultsObjectiveGroup } from "./key-results-objective-group";

export const KeyResultsList = ({
  groups,
  memberById,
  selectedKeyResultIds,
  setSelectedKeyResultIds,
  teamColorById,
}: {
  groups: ObjectiveKeyResultGroup[];
  memberById: ReadonlyMap<string, KeyResultsMember>;
  selectedKeyResultIds: ReadonlySet<string>;
  setSelectedKeyResultIds: (ids: Set<string>) => void;
  teamColorById: ReadonlyMap<string, string>;
}) => (
  <Box>
    {groups.map((group) => (
      <KeyResultsObjectiveGroup
        group={group}
        key={group.objectiveId}
        memberById={memberById}
        selectedKeyResultIds={selectedKeyResultIds}
        setSelectedKeyResultIds={setSelectedKeyResultIds}
        teamColorById={teamColorById}
      />
    ))}
  </Box>
);

import { Avatar, Box, Text, Flex, Button, Tooltip, TimeAgo } from "ui";
import { cn } from "lib";
import type { ChatStatus } from "ai";
import Link from "next/link";
import { useState, type ComponentProps, type ReactNode } from "react";
import { CheckIcon, CopyIcon, PlusIcon, ReloadIcon } from "icons";
import { usePathname } from "next/navigation";
import { Streamdown, type StreamdownProps } from "streamdown";
import type { User } from "@/types";
import { BurndownChart } from "@/modules/sprints/stories/burndown";
import { useCopyToClipboard, useTerminology, useWorkspacePath } from "@/hooks";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { NewStoryDialog } from "../new-story-dialog";
import { AttachmentsDisplay } from "./attachments-display";
import { AnalyticsReport } from "./analytics-report";
import { Dot, ObjectiveStatusIcon, PriorityIcon, RowWrapper, StoryStatusIcon } from "@/components/ui";
import { TeamColor } from "@/components/ui/team-color";
import { slugify } from "@/utils";

type ToolOutputListItem = {
  id: string;
  title: string;
  href?: string;
  isExternal?: boolean;
  leftIcon?: ReactNode;
  rightIcon?: ReactNode;
  meta?: ReactNode;
};

type ToolOutputSection = {
  title: string;
  items: ToolOutputListItem[];
  emptyText?: string;
};

type ToolOutputRecord = Record<string, unknown>;

const EXTERNAL_LINK_PATTERNS = /^(https?:|mailto:|tel:)/i;
const MAX_TOOL_OUTPUT_ITEMS = 10;

const STORY_PRIORITIES = ["No Priority", "Low", "Medium", "High", "Urgent"] as const;
type StoryPriority = (typeof STORY_PRIORITIES)[number];

type ToolRecord = Record<string, unknown>;

const truncateText = (value: string, maxLength = 160) => {
  if (value.length <= maxLength) {
    return value;
  }

  return `${value.slice(0, maxLength - 3).trimEnd()}...`;
};

const asRecord = (value: unknown): ToolOutputRecord =>
  value && typeof value === "object" && !Array.isArray(value)
    ? (value as ToolOutputRecord)
    : {};

const asArray = (value: unknown): unknown[] =>
  Array.isArray(value) ? value : [];

const asString = (value: unknown): string => {
  if (typeof value === "string") {
    return value;
  }

  if (typeof value === "number") {
    return String(value);
  }

  return "";
};

const asStoryPriority = (value: unknown): StoryPriority | undefined => {
  const priority = asString(value);
  if (STORY_PRIORITIES.includes(priority as StoryPriority)) {
    return priority as StoryPriority;
  }

  return undefined;
};

const getDisplayStatusText = (status: unknown): string => {
  const record = asRecord(status);
  return asString(record.name) || asString(status);
};

const isExternalUrl = (value: string) => EXTERNAL_LINK_PATTERNS.test(value);

const getTeamRef = (team: ToolRecord): string => {
  const code = asString(team.code);
  if (code) {
    return code;
  }

  return asString(team.id).slice(0, 6);
};

const uniqueById = (items: ToolRecord[]) => {
  const seen = new Set<string>();
  return items.filter((item) => {
    const id = asString(item.id);
    if (!id || seen.has(id)) {
      return false;
    }

    seen.add(id);
    return true;
  });
};

const toStoryRows = (
  stories: ToolRecord[],
  withWorkspace: (path: string) => string,
): ToolOutputListItem[] =>
  uniqueById(stories).map((story) => {
    const storyId = asString(story.id);
    const title = asString(story.title) || "Untitled story";
    const team = asRecord(story.team);
    const teamRef = getTeamRef(team);
    const sequenceId = asString(story.sequenceId);
    const statusRecord = asRecord(story.status);
    const statusId = asString(story.statusId);
    const statusName = getDisplayStatusText(statusRecord) || statusId;

    return {
      id: storyId,
      title,
      href: withWorkspace(`/story/${storyId}/${slugify(title)}`),
      leftIcon: <PriorityIcon priority={asStoryPriority(story.priority)} />, 
      rightIcon: statusName ? (
        <Flex align="center" className="gap-1.5" gap={0}>
          <StoryStatusIcon statusId={statusId || undefined} />
          <Text color="muted" className="max-w-[13ch] truncate text-xs md:text-sm">
            {statusName}
          </Text>
        </Flex>
      ) : (
        <StoryStatusIcon statusId={statusId || undefined} />
      ),
      meta:
        teamRef && sequenceId
          ? (
              <Text className="truncate" color="muted">
                {teamRef}-{sequenceId}
              </Text>
            )
          : undefined,
    };
  });

const getSectionForStories = (
  output: ToolOutputRecord,
  withWorkspace: (path: string) => string,
): ToolOutputSection | undefined => {
  const fromGroups = asArray(output.groups).flatMap((group) => {
    const stories = asArray(asRecord(group).stories);
    return stories.map(asRecord);
  });

  const stories = toStoryRows(
    [...asArray(output.stories).map(asRecord), ...fromGroups],
    withWorkspace,
  );

  if (stories.length === 0) {
    const emptyText = asString(output.message);
    if (!emptyText) {
      return undefined;
    }
    return {
      title: "Stories",
      items: [],
      emptyText,
    };
  }

  return {
    title: "Stories",
    items: stories.slice(0, MAX_TOOL_OUTPUT_ITEMS),
  };
};

const getSectionForSprints = (
  sprints: unknown,
  withWorkspace: (path: string) => string,
): ToolOutputSection | undefined => {
  const items = uniqueById(asArray(sprints).map(asRecord)).map((sprint) => {
    const sprintId = asString(sprint.id);
    const title = asString(sprint.name) || "Untitled sprint";
    const teamId = asString(sprint.teamId);
    const startDate = asString(sprint.startDate);
    const endDate = asString(sprint.endDate);
    const hasRange = startDate && endDate;

    return {
      id: sprintId,
      title,
      href: withWorkspace(`/teams/${teamId}/sprints/${sprintId}/stories`),
      leftIcon: <Dot className="size-2" color="currentColor" />,
      rightIcon:
        hasRange && (
          <Text color="muted" className="text-xs">
            {startDate} → {endDate}
          </Text>
        ),
    };
  });

  if (items.length === 0) {
    return undefined;
  }

  return {
    title: "Sprints",
    items: items.slice(0, MAX_TOOL_OUTPUT_ITEMS),
  };
};

const getSectionForObjectives = (
  objectives: unknown,
  withWorkspace: (path: string) => string,
): ToolOutputSection | undefined => {
  const items = uniqueById(asArray(objectives).map(asRecord)).map((objective) => {
    const objectiveId = asString(objective.id);
    const title = asString(objective.name) || "Untitled objective";
    const teamId = asString(objective.teamId);
    const statusId = asString(objective.statusId);
    const status = asString(objective.statusName);

    return {
      id: objectiveId,
      title,
      href: withWorkspace(`/teams/${teamId}/objectives/${objectiveId}`),
      leftIcon: <PriorityIcon priority={asStoryPriority(objective.priority)} />,
      rightIcon: statusId || status ? (
        <Flex align="center" className="gap-1.5">
          <ObjectiveStatusIcon statusId={statusId || undefined} />
          <Text color="muted" className="max-w-[11ch] truncate text-xs md:text-sm">
            {status || "Status"}
          </Text>
        </Flex>
      ) : null,
      meta:
        asString(asRecord(objective.team).name) ? (
          <Text color="muted">{asString(asRecord(objective.team).name)}</Text>
        ) : undefined,
    };
  });

  if (items.length === 0) {
    return undefined;
  }

  return {
    title: "Objectives",
    items: items.slice(0, MAX_TOOL_OUTPUT_ITEMS),
  };
};

const getSectionForKeyResults = (
  keyResults: unknown,
  withWorkspace: (path: string) => string,
): ToolOutputSection | undefined => {
  const items = uniqueById(asArray(keyResults).map(asRecord)).map((keyResult) => {
    const keyResultId = asString(keyResult.id);
    const title = asString(keyResult.name) || "Untitled key result";
    const objectiveId = asString(keyResult.objectiveId);
    const teamId = asString(keyResult.teamId);
    const measurementType = asString(keyResult.measurementType);

    return {
      id: keyResultId,
      title,
      href: teamId && objectiveId ? withWorkspace(`/teams/${teamId}/objectives/${objectiveId}`) : undefined,
      leftIcon: <Dot className="size-2" />,
      rightIcon: measurementType ? <Text color="muted">{measurementType}</Text> : null,
      meta: asString(keyResult.objectiveName) ? (
        <Text color="muted" className="truncate">
          {asString(keyResult.objectiveName)}
        </Text>
      ) : undefined,
    };
  });

  if (items.length === 0) {
    return undefined;
  }

  return {
    title: "Key Results",
    items: items.slice(0, MAX_TOOL_OUTPUT_ITEMS),
  };
};

const getSectionForTeams = (
  teams: unknown,
): ToolOutputSection | undefined => {
  const items = uniqueById(asArray(teams).map(asRecord)).map((team) => {
    const teamId = asString(team.id);
    const title =
      asString(team.name) || asString(team.code) || `Team ${teamId.slice(0, 6)}`;
    const memberCount = asString(team.memberCount);
    const color = asString(team.color);

    return {
      id: teamId,
      title,
      leftIcon: <TeamColor className="shrink-0" color={color || undefined} />,
      rightIcon: memberCount ? <Text color="muted">{memberCount}</Text> : null,
      meta: asString(team.code) ? (
        <Text color="muted">{asString(team.code)}</Text>
      ) : undefined,
    };
  });

  if (items.length === 0) {
    return undefined;
  }

  return {
    title: "Teams",
    items: items.slice(0, MAX_TOOL_OUTPUT_ITEMS),
  };
};

const getSectionForMembers = (
  members: unknown,
  teamLabel?: string,
): ToolOutputSection | undefined => {
  const items = uniqueById(asArray(members).map(asRecord)).map((member) => {
    const memberId = asString(member.id);
    const name = asString(member.name);
    const username = asString(member.username);
    const title = name || username || memberId;
    const role = asString(member.role);

    return {
      id: memberId,
      title: title || "Member",
      leftIcon: (
        <Avatar
          name={name || username}
          size="xs"
          src={asString(member.avatarUrl)}
        />
      ),
      rightIcon: role ? <Text color="muted">{role}</Text> : undefined,
      meta: username ? <Text color="muted">@{username}</Text> : undefined,
    };
  });

  if (items.length === 0) {
    return undefined;
  }

  return {
    title: teamLabel ? `Members (${teamLabel})` : "Members",
    items,
  };
};

const getSectionForNotifications = (
  notifications: unknown,
  withWorkspace: (path: string) => string,
): ToolOutputSection | undefined => {
  const items = uniqueById(asArray(notifications).map(asRecord)).map((notification) => {
    const id = asString(notification.id);
    const title = asString(notification.title) || "Notification";
    const entityId = asString(notification.entityId);
    const entityType = asString(notification.entityType);

    return {
      id,
      title,
      href: entityId && entityType ? withWorkspace(
        `/notifications/${id}?entityId=${entityId}&entityType=${entityType}`,
      ) : undefined,
      leftIcon: <Dot className="size-2" color="var(--muted-foreground)" />,
      rightIcon: asString(notification.createdAt) ? (
        <TimeAgo timestamp={asString(notification.createdAt)} />
      ) : null,
      meta: asString(notification.message) ? (
        <Text color="muted">{truncateText(asString(notification.message), 120)}</Text>
      ) : undefined,
    };
  });

  if (items.length === 0) {
    return undefined;
  }

  return {
    title: "Notifications",
    items,
  };
};

const getSectionForComments = (
  comments: unknown,
): ToolOutputSection | undefined => {
  const items = uniqueById(asArray(comments).map(asRecord)).map((comment) => {
    const id = asString(comment.id);
    const content = asString(comment.content);
    const commenter = asRecord(comment.commenter);

    return {
      id,
      title: asString(commenter.name) || asString(commenter.username) || "Comment",
      leftIcon: (
        <Avatar
          name={asString(commenter.name) || asString(commenter.username)}
          size="xs"
          src={asString(commenter.avatarUrl)}
        />
      ),
      rightIcon: asString(comment.createdAt) ? (
        <TimeAgo timestamp={asString(comment.createdAt)} />
      ) : null,
      meta: content ? <Text color="muted">{truncateText(content, 150)}</Text> : null,
    };
  });

  if (items.length === 0) {
    return undefined;
  }

  return {
    title: "Comments",
    items,
  };
};

const getSectionForLinks = (links: unknown): ToolOutputSection | undefined => {
  const items = uniqueById(asArray(links).map(asRecord)).map((link) => {
    const id = asString(link.id);
    const title = asString(link.title) || asString(link.url) || "Link";
    const url = asString(link.url);

    return {
      id,
      title,
      href: url,
      isExternal: !!url,
      leftIcon: <Dot className="size-2" />,
      rightIcon:
        asString(link.createdAt) ? <TimeAgo timestamp={asString(link.createdAt)} /> : null,
      meta: url ? <Text color="muted">{url}</Text> : undefined,
    };
  });

  if (items.length === 0) {
    return undefined;
  }

  return {
    title: "Links",
    items,
  };
};

const getSectionForLabels = (
  labels: unknown,
): ToolOutputSection | undefined => {
  const items = uniqueById(asArray(labels).map(asRecord)).map((label) => {
    const id = asString(label.id);
    const name = asString(label.name);

    return {
      id,
      title: name || `Label ${id.slice(0, 6)}`,
      leftIcon: <Dot className="size-2" color={asString(label.color)} />,
      meta: asString(label.category) ? <Text color="muted">{asString(label.category)}</Text> : undefined,
    };
  });

  if (items.length === 0) {
    return undefined;
  }

  return {
    title: "Labels",
    items: items.slice(0, MAX_TOOL_OUTPUT_ITEMS),
  };
};

const getSectionForStoryLabels = (
  labels: unknown,
): ToolOutputSection | undefined => {
  const ids = uniqueById(asArray(labels).map((id) => ({ id: asString(id) } as ToolRecord)));

  if (ids.length === 0) {
    return undefined;
  }

  return {
    title: "Story Labels",
    items: ids
      .map((label) => ({
        id: asString(label.id),
        title: asString(label.id),
        leftIcon: <Dot className="size-2" />,
      }))
      .slice(0, MAX_TOOL_OUTPUT_ITEMS),
  };
};

const getSectionForStatuses = (
  statuses: unknown,
): ToolOutputSection | undefined => {
  const items = uniqueById(asArray(statuses).map(asRecord)).map((status) => {
    const statusId = asString(status.id);
    const name = asString(status.name);

    return {
      id: statusId,
      title: name || `Status ${statusId.slice(0, 6)}`,
      leftIcon: <Dot className="size-2" color={asString(status.color)} />,
      rightIcon:
        asString(status.category) ? (
          <Text className="text-xs" color="muted">
            {asString(status.category)}
          </Text>
        ) : null,
    };
  });

  if (items.length === 0) {
    return undefined;
  }

  return {
    title: "Statuses",
    items,
  };
};

const getSectionForObjectiveStatuses = (
  statuses: unknown,
): ToolOutputSection | undefined => {
  const items = uniqueById(asArray(statuses).map(asRecord)).map((status) => {
    const statusId = asString(status.id);
    const name = asString(status.name);

    return {
      id: statusId,
      title: name || `Status ${statusId.slice(0, 6)}`,
      leftIcon: <Dot className="size-2" color={asString(status.color)} />,
      rightIcon:
        asString(status.category) ? (
          <Text className="text-xs" color="muted">
            {asString(status.category)}
          </Text>
        ) : null,
    };
  });

  if (items.length === 0) {
    return undefined;
  }

  return {
    title: "Objective statuses",
    items,
  };
};

const getSectionForAttachments = (
  attachments: unknown,
): ToolOutputSection | undefined => {
  const items = uniqueById(asArray(attachments).map(asRecord)).map((attachment) => {
    const id = asString(attachment.id);
    const filename = asString(attachment.filename);
    const url = asString(attachment.url);

    return {
      id,
      title: filename || `Attachment ${id.slice(0, 6)}`,
      href: url,
      isExternal: !!url,
      leftIcon: <Dot className="size-2" />,
      rightIcon:
        asString(attachment.createdAt) ? <TimeAgo timestamp={asString(attachment.createdAt)} /> : null,
      meta: (
        <Text color="muted">
          {asString(attachment.formattedSize) || asString(attachment.size)}
          {url ? ` · ${url}` : ""}
        </Text>
      ),
    };
  });

  if (items.length === 0) {
    return undefined;
  }

  return {
    title: "Attachments",
    items: items.slice(0, MAX_TOOL_OUTPUT_ITEMS),
  };
};

const getSectionForActivities = (
  activities: unknown,
): ToolOutputSection | undefined => {
  const items = uniqueById(asArray(activities).map(asRecord)).map((activity) => {
    const id = asString(activity.id);
    const field = asString(activity.field);
    const type = asString(activity.type);

    return {
      id,
      title: field || type || `Activity ${id.slice(0, 6)}`,
      leftIcon: <Dot className="size-2" />,
      meta: asString(activity.currentValue) ? (
        <Text color="muted">{truncateText(asString(activity.currentValue), 140)}</Text>
      ) : undefined,
      rightIcon:
        asString(activity.createdAt) ? <TimeAgo timestamp={asString(activity.createdAt)} /> : null,
    };
  });

  if (items.length === 0) {
    return undefined;
  }

  return {
    title: "Activities",
    items,
  };
};

const getSectionForMemories = (
  memories: unknown,
): ToolOutputSection | undefined => {
  const items = uniqueById(asArray(memories).map(asRecord)).map((memory) => {
    const id = asString(memory.id);
    const content = asString(memory.content);

    return {
      id,
      title: content ? truncateText(content, 140) : `Memory ${id.slice(0, 6)}`,
      leftIcon: <Dot className="size-2" color="var(--muted-foreground)" />,
    };
  });

  if (items.length === 0) {
    return undefined;
  }

  return {
    title: "Memories",
    items,
  };
};

const getSectionForFeedback = (
  feedback: unknown,
  withWorkspace: (path: string) => string,
  fallbackName = "Feedback",
): ToolOutputSection | undefined => {
  const items = uniqueById(asArray(feedback).map(asRecord)).map((item) => {
    const id = asString(item.id);
    const teamId = asString(item.teamId);
    const title = asString(item.title) || "Feedback item";
    const statusLabel = asString(item.statusLabel);
    const boardName = asString(asRecord(item.board).name);
    const author = asString(item.author);

    return {
      id,
      title,
      href:
        teamId && id
          ? withWorkspace(`/teams/${teamId}/feedback/${id}`)
          : undefined,
      leftIcon: <Dot className="size-2" />,
      rightIcon:
        statusLabel ? (
          <Text className="text-xs" color="muted">
            {statusLabel}
          </Text>
        ) : null,
      meta: (
        <Text color="muted">
          {boardName || author ? `${boardName}${boardName && author ? " · " : ""}${author}` : ""}
        </Text>
      ),
    };
  });

  if (items.length === 0) {
    return undefined;
  }

  return {
    title: fallbackName,
    items,
  };
};

const getSectionForIntegrationRequests = (
  requests: unknown,
  withWorkspace: (path: string) => string,
): ToolOutputSection | undefined => {
  const items = uniqueById(asArray(requests).map(asRecord)).map((request) => {
    const requestId = asString(request.id);
    const teamId = asString(request.teamId);
    const title = asString(request.title) || "Integration request";
    const provider = asString(request.provider);
    const sourceType = asString(request.sourceType);
    const sourceNumber = asString(request.sourceNumber);

    return {
      id: requestId,
      title,
      href:
        teamId && requestId
          ? withWorkspace(`/teams/${teamId}/requests/${requestId}`)
          : undefined,
      leftIcon: (
        <PriorityIcon priority={asStoryPriority(request.priority)} className="shrink-0" />
      ),
      rightIcon: asString(request.status) ? (
        <Text className="text-xs" color="muted">
          {asString(request.status)}
        </Text>
      ) : null,
      meta: (
        <Text color="muted">
          {provider} {sourceType} {sourceNumber ? `#${sourceNumber}` : ""}
        </Text>
      ),
    };
  });

  if (items.length === 0) {
    return undefined;
  }

  return {
    title: "Integration requests",
    items,
  };
};

const getSectionsFromToolOutput = (
  part: ToolMessagePart,
  withWorkspace: (path: string) => string,
): ToolOutputSection[] => {
  if (part.state !== "output-available") {
    return [];
  }

  const output = asRecord(part.output);
  const sections: ToolOutputSection[] = [];

  switch (part.type) {
    case "tool-listTeamStories":
    case "tool-searchStories": {
      const storiesSection = getSectionForStories(output, withWorkspace);
      if (storiesSection) {
        sections.push(storiesSection);
      }
      break;
    }

    case "tool-getStoryDetails": {
      const story = asRecord(output.story);
      const items = story.id
        ? [
            {
              id: asString(story.id),
              title: asString(story.title) || "Untitled story",
              href: asString(story.id)
                ? withWorkspace(
                    `/story/${asString(story.id)}/${slugify(asString(story.title) || "story")}`,
                  )
                : undefined,
              leftIcon: (
                <PriorityIcon priority={asStoryPriority(story.priority)} />
              ),
              rightIcon: (
                <Flex align="center" className="gap-1.5">
                  <StoryStatusIcon statusId={asString(story.statusId)} />
                  <Text className="text-xs" color="muted">
                    {getDisplayStatusText(story.status) || asString(story.statusId)}
                  </Text>
                </Flex>
              ),
            },
          ]
        : [];

      sections.push({
        title: "Story",
        items,
        emptyText: items.length ? undefined : asString(output.message),
      });
      break;
    }

    case "tool-listSprints":
    case "tool-listRunningSprints": {
      const sprintsSection = getSectionForSprints(output.sprints, withWorkspace);
      if (sprintsSection) {
        sections.push(sprintsSection);
      }
      break;
    }

    case "tool-getSprintDetailsTool": {
      const sprint = asRecord(output.sprint);
      const sprintId = asString(sprint.id);

      if (sprintId) {
        sections.push({
          title: "Sprint",
          items: [
            {
              id: sprintId,
              title: asString(sprint.name) || "Sprint details",
              href: withWorkspace(
                `/teams/${asString(sprint.teamId)}/sprints/${sprintId}/stories`,
              ),
              leftIcon: <Dot className="size-2" />,
              rightIcon: (
                <Text className="text-xs" color="muted">
                  {asString(sprint.startDate)}
                  {asString(sprint.endDate)
                    ? ` → ${asString(sprint.endDate)}`
                    : ""}
                </Text>
              ),
            },
          ],
        });
      }

      break;
    }

    case "tool-listTeamObjectivesTool":
    case "tool-listObjectivesTool": {
      const objectivesSection = getSectionForObjectives(
        output.objectives,
        withWorkspace,
      );
      if (objectivesSection) {
        sections.push(objectivesSection);
      }
      break;
    }

    case "tool-getObjectiveDetailsTool": {
      const objective = asRecord(output.objective);
      const objectiveId = asString(objective.id);

      if (objectiveId) {
        sections.push({
          title: "Objective",
          items: [
            {
              id: objectiveId,
              title: asString(objective.name) || "Objective",
              href: withWorkspace(
                `/teams/${asString(objective.teamId)}/objectives/${objectiveId}`,
              ),
              leftIcon: <PriorityIcon priority={asStoryPriority(objective.priority)} />,
              rightIcon: asString(objective.health) ? (
                <Text className="text-xs" color="muted">
                  {asString(objective.health)}
                </Text>
              ) : null,
            },
          ],
        });
      }

      break;
    }

    case "tool-listKeyResultsTool": {
      const keyResultsSection = getSectionForKeyResults(
        output.keyResults,
        withWorkspace,
      );
      if (keyResultsSection) {
        sections.push(keyResultsSection);
      }
      break;
    }

    case "tool-listTeams":
    case "tool-listPublicTeams": {
      const teamsSection = getSectionForTeams(output.teams);
      if (teamsSection) {
        sections.push(teamsSection);
      }
      break;
    }

    case "tool-getTeamDetails": {
      const team = asRecord(output.team);
      const teamId = asString(team.id);

      if (teamId) {
        sections.push({
          title: "Team",
          items: [
            {
              id: teamId,
              title: asString(team.name) || `Team ${teamId.slice(0, 6)}`,
              href: withWorkspace(`/teams/${teamId}/stories`),
              leftIcon: <TeamColor className="shrink-0" color={asString(team.color)} />,
              rightIcon: asString(team.memberCount) ? (
                <Text className="text-xs" color="muted">
                  {asString(team.memberCount)}
                </Text>
              ) : null,
            },
          ],
        });
      }

      const membersSection = getSectionForMembers(output.members);
      if (membersSection) {
        sections.push(membersSection);
      }

      break;
    }

    case "tool-listTeamMembers": {
      const membersSection = getSectionForMembers(
        output.members,
        asString(output.teamName),
      );
      if (membersSection) {
        sections.push(membersSection);
      }

      break;
    }

    case "tool-members": {
      const membersSection = getSectionForMembers(output.members);
      if (membersSection) {
        sections.push(membersSection);
      }

      break;
    }

    case "tool-notifications": {
      const notificationsSection = getSectionForNotifications(
        output.notifications,
        withWorkspace,
      );
      if (notificationsSection) {
        sections.push(notificationsSection);
      }

      break;
    }

    case "tool-comments": {
      const commentsSection = getSectionForComments(output.comments);
      if (commentsSection) {
        sections.push(commentsSection);
      }

      break;
    }

    case "tool-links": {
      const linksSection = getSectionForLinks(output.links);
      if (linksSection) {
        sections.push(linksSection);
      }

      break;
    }

    case "tool-labels": {
      const labelsSection = getSectionForLabels(output.labels);
      if (labelsSection) {
        sections.push(labelsSection);
      }

      break;
    }

    case "tool-storyLabels": {
      const storyLabelsSection = getSectionForStoryLabels(output.labels);
      if (storyLabelsSection) {
        sections.push(storyLabelsSection);
      }

      break;
    }

    case "tool-statuses": {
      const statusesSection = getSectionForStatuses(output.data);
      if (statusesSection) {
        sections.push(statusesSection);
      }

      break;
    }

    case "tool-objectiveStatuses": {
      const statusesSection = getSectionForObjectiveStatuses(output.data);
      if (statusesSection) {
        sections.push(statusesSection);
      }

      break;
    }

    case "tool-storyActivities": {
      const activitiesSection = getSectionForActivities(output.activities);
      if (activitiesSection) {
        sections.push(activitiesSection);
      }

      break;
    }

    case "tool-listAttachments": {
      const attachmentsSection = getSectionForAttachments(output.attachments);
      if (attachmentsSection) {
        sections.push(attachmentsSection);
      }

      break;
    }

    case "tool-listMemories": {
      const memoriesSection = getSectionForMemories(output.memories);
      if (memoriesSection) {
        sections.push(memoriesSection);
      }

      break;
    }

    case "tool-listCustomerFeedbackTool": {
      const teams = asArray(output.teams);
      const grouped = teams.flatMap((teamBucket) => {
        const team = asRecord(teamBucket);
        const feedback = asArray(team.feedback);

        return getSectionForFeedback(
          feedback,
          withWorkspace,
          `Feedback (${asString(team.teamId) || "all"})`,
        )?.items ?? [];
      });

      if (grouped.length > 0) {
        sections.push({
          title: "Customer feedback",
          items: grouped.slice(0, MAX_TOOL_OUTPUT_ITEMS),
        });
      }

      break;
    }

    case "tool-listIntegrationRequestsTool": {
      const teams = asArray(output.teams);
      const grouped = teams.flatMap((teamBucket) => {
        const team = asRecord(teamBucket);
        const requests = asArray(team.requests);
        return getSectionForIntegrationRequests(requests, withWorkspace)?.items ?? [];
      });

      if (grouped.length > 0) {
        sections.push({
          title: "Integration requests",
          items: grouped.slice(0, MAX_TOOL_OUTPUT_ITEMS),
        });
      }

      break;
    }

    case "tool-search": {
      const storySection = getSectionForStories(output, withWorkspace);
      if (storySection) {
        sections.push(storySection);
      }

      const objectiveSection = getSectionForObjectives(
        output.objectives,
        withWorkspace,
      );
      if (objectiveSection) {
        sections.push(objectiveSection);
      }

      break;
    }

    case "tool-getCustomerFeedbackTool": {
      const section = getSectionForFeedback(output.feedback, withWorkspace);
      if (section) {
        sections.push(section);
      }

      break;
    }

    case "tool-getIntegrationRequestTool": {
      const request = asRecord(output.request);
      const requestId = asString(request.id);

      if (requestId) {
        sections.push({
          title: "Integration request",
          items: [
            {
              id: requestId,
              title: asString(request.title) || "Request",
              href: request.teamId
                ? withWorkspace(
                    `/teams/${asString(request.teamId)}/requests/${requestId}`,
                  )
                : undefined,
              leftIcon: <PriorityIcon priority={asStoryPriority(request.priority)} />,
              rightIcon: asString(request.status) ? (
                <Text className="text-xs" color="muted">
                  {asString(request.status)}
                </Text>
              ) : null,
            },
          ],
        });
      }

      break;
    }

    default:
      break;
  }

  return sections.filter((section) => section.items.length > 0 || section.emptyText);
};

const ToolOutputSectionView = ({ section }: { section: ToolOutputSection }) => (
  <Box className="mt-4">
    <Text as="h3" className="mb-2 text-sm font-semibold" color="muted">
      {section.title}
    </Text>
    {section.items.length === 0 ? (
      <Text className="px-4 text-sm" color="muted">
        {section.emptyText || "No results"}
      </Text>
    ) : (
      section.items.map((item, index) => (
        <ToolOutputItem
          item={item}
          key={`${section.title}-${item.id}-${index}`}
        />
      ))
    )}
  </Box>
);

const ToolOutputItem = ({ item }: { item: ToolOutputListItem }) => {
  const row = (
    <RowWrapper className="gap-3 px-0 md:px-0">
      <Flex align="center" className="min-w-0 flex-1 gap-2" gap={2}>
        {item.leftIcon ? <span className="shrink-0">{item.leftIcon}</span> : null}
        <Flex align="center" className="min-w-0 flex-1 flex-col items-start">
          <Text className="min-w-0 truncate font-medium">{item.title}</Text>
          {item.meta ? item.meta : null}
        </Flex>
      </Flex>
      <Flex align="center" className="shrink-0 gap-2" gap={2}>
        {item.rightIcon}
      </Flex>
    </RowWrapper>
  );

  if (!item.href) {
    return row;
  }

  if (item.isExternal || isExternalUrl(item.href)) {
    return (
      <a
        href={item.href}
        rel="noreferrer"
        target="_blank"
      >
        {row}
      </a>
    );
  }

  return <Link href={item.href}>{row}</Link>;
};

type ChatMessageProps = {
  isAnimating?: boolean;
  isLast: boolean;
  message: MayaUIMessage;
  profile: User | undefined;
  status: ChatStatus;
  deferToolOutputs?: boolean;
  regenerate: (messageId?: string) => void;
  onPromptSelect: (prompt: string) => void;
};

/** Maps tool part types to the single progress label shown below the chat. */
const TOOL_THINKING_LABELS: Record<string, string> = {
  // Stories
  "tool-listTeamStories": "Fetching stories",
  "tool-searchStories": "Searching stories",
  "tool-getStoryDetails": "Getting story details",
  "tool-createStory": "Creating story",
  "tool-updateStory": "Updating story",
  "tool-deleteStory": "Deleting story",
  "tool-bulkCreateStories": "Creating stories",
  "tool-bulkUpdateStories": "Updating stories",
  "tool-bulkDeleteStories": "Deleting stories",
  "tool-assignStoriesToUser": "Assigning stories",
  "tool-duplicateStory": "Duplicating story",
  "tool-restoreStory": "Restoring story",
  "tool-addStoryAssociation": "Linking stories",
  "tool-removeStoryAssociation": "Unlinking stories",
  // Sprints
  "tool-listSprints": "Loading sprints",
  "tool-listRunningSprints": "Getting active sprints",
  "tool-getSprintDetailsTool": "Getting sprint details",
  "tool-getSprintAnalyticsTool": "Analyzing sprint data",
  "tool-updateSprintSettings": "Updating sprint settings",
  // Teams
  "tool-listTeams": "Loading teams",
  "tool-listPublicTeams": "Loading public teams",
  "tool-getTeamDetails": "Getting team details",
  "tool-listTeamMembers": "Loading team members",
  "tool-createTeamTool": "Creating team",
  "tool-updateTeam": "Updating team",
  "tool-joinTeam": "Joining team",
  "tool-leaveTeam": "Leaving team",
  "tool-deleteTeam": "Deleting team",
  "tool-getTeamSettingsTool": "Loading team settings",
  // Objectives & Key Results
  "tool-listObjectivesTool": "Loading objectives",
  "tool-listTeamObjectivesTool": "Loading team objectives",
  "tool-createObjectiveTool": "Creating objective",
  "tool-updateObjectiveTool": "Updating objective",
  "tool-deleteObjectiveTool": "Deleting objective",
  "tool-getObjectiveDetailsTool": "Getting objective details",
  "tool-objectiveAnalyticsTool": "Analyzing objective data",
  "tool-getObjectiveActivitiesTool": "Loading objective activity",
  "tool-listKeyResultsTool": "Loading key results",
  "tool-createKeyResultTool": "Creating key result",
  "tool-updateKeyResultTool": "Updating key result",
  "tool-deleteKeyResultTool": "Deleting key result",
  "tool-getKeyResultActivitiesTool": "Loading key result activity",
  // Other
  "tool-navigation": "Navigating",
  "tool-search": "Searching",
  "tool-members": "Loading members",
  "tool-comments": "Loading comments",
  "tool-notifications": "Checking notifications",
  "tool-workspacePerformanceReportTool": "Building workspace report",
  "tool-workspaceCommandCenterReportTool": "Building command center",
  "tool-storyPerformanceReportTool": "Building story report",
  "tool-objectiveProgressReportTool": "Building objective report",
  "tool-teamPerformanceReportTool": "Building team report",
  "tool-sprintPerformanceReportTool": "Building sprint report",
  "tool-timelineTrendsReportTool": "Building trends report",
  "tool-mayaWorkPlanTool": "Planning work",
  "tool-getGitHubIntegrationTool": "Checking GitHub integration",
  "tool-createGitHubInstallSessionTool": "Creating GitHub install link",
  "tool-resyncGitHubRepositoriesTool": "Resyncing GitHub repositories",
  "tool-createGitHubIssueSyncLinkTool": "Linking GitHub repository",
  "tool-deleteGitHubIssueSyncLinkTool": "Removing GitHub sync link",
  "tool-updateGitHubWorkspaceSettingsTool": "Updating GitHub settings",
  "tool-getGitHubTeamSettingsTool": "Checking GitHub automation",
  "tool-updateGitHubTeamSettingsTool": "Updating GitHub automation",
  "tool-getStoryGitHubLinksTool": "Checking story GitHub links",
  "tool-getStoryGitHubCommentsTool": "Reading GitHub comments",
  "tool-postStoryGitHubCommentTool": "Posting GitHub comment",
  "tool-deleteStoryGitHubLinkTool": "Removing story GitHub link",
  "tool-statuses": "Loading statuses",
  "tool-objectiveStatuses": "Loading objective statuses",
  "tool-links": "Loading links",
  "tool-labels": "Loading labels",
  "tool-storyLabels": "Managing labels",
  "tool-storyActivities": "Loading activity",
  "tool-listAttachments": "Loading attachments",
  "tool-deleteAttachment": "Deleting attachment",
  "tool-listMemories": "Checking memory",
  "tool-createMemory": "Saving to memory",
  "tool-updateMemory": "Updating memory",
  "tool-deleteMemory": "Removing memory",
  "tool-listCustomerFeedbackTool": "Reading customer feedback",
  "tool-getCustomerFeedbackTool": "Getting feedback details",
  "tool-theme": "Changing theme",
};

const DEFAULT_PROGRESS_LABEL = "Working on it";

const isToolPart = (type: string): boolean => type.startsWith("tool-");

const LinkText = ({ children }: ComponentProps<"a">) => <>{children}</>;

const STREAMDOWN_COMPONENTS: NonNullable<StreamdownProps["components"]> = {
  a: LinkText,
};

type ToolMessagePart = MayaUIMessage["parts"][number] & {
  state: string;
  output?: unknown;
};

const isToolMessagePart = (
  part: MayaUIMessage["parts"][number],
): part is ToolMessagePart => isToolPart(part.type) && "state" in part;

const isAnalyticsReportOutput = (
  output: unknown,
): output is Record<string, unknown> => {
  if (!output || typeof output !== "object" || !("kind" in output)) {
    return false;
  }

  const kind = (output as { kind?: unknown }).kind;
  return typeof kind === "string" && kind.endsWith("-report");
};

const getSprintBurndownData = (output: unknown) => {
  const outputRecord = asRecord(output);
  const analyticsReport = asRecord(outputRecord.analyticsReport);
  return Array.isArray(analyticsReport.burndown)
    ? analyticsReport.burndown
    : [];
};

const getSuggestions = (output: unknown) => {
  const outputRecord = asRecord(output);
  return Array.isArray(outputRecord.suggestions)
    ? outputRecord.suggestions.filter(
        (suggestion): suggestion is string => typeof suggestion === "string",
      )
    : [];
};

export const getMessageProgressLabel = (message: MayaUIMessage) => {
  const lastPart = message.parts.at(-1);

  if (lastPart?.type === "text" && lastPart.text.trim()) {
    return undefined;
  }

  const latestToolPart = message.parts.filter(isToolMessagePart).at(-1);

  if (!latestToolPart) {
    return "Thinking";
  }

  return TOOL_THINKING_LABELS[latestToolPart.type] ?? DEFAULT_PROGRESS_LABEL;
};

export const getMessageText = (message: MayaUIMessage) =>
  message.parts
    .filter((part) => part.type === "text")
    .map((part) => part.text)
    .join("");

export const hasVisibleMessageContent = (
  message: MayaUIMessage,
  deferToolOutputs = false,
) => {
  if (message.role === "user") {
    return true;
  }

  if (getMessageText(message).trim()) {
    return true;
  }

  if (message.parts.some((part) => part.type === "file")) {
    return true;
  }

  if (deferToolOutputs) {
    return false;
  }

  return message.parts.some(
    (part) => isToolMessagePart(part) && part.state === "output-available",
  );
};

const RenderMessage = ({
  isAnimating,
  message,
  onPromptSelect,
  deferToolOutputs = false,
}: {
  isAnimating: boolean;
  message: MayaUIMessage;
  deferToolOutputs?: boolean;
  onPromptSelect: (prompt: string) => void;
}) => {
  const pathname = usePathname();
  const { withWorkspace } = useWorkspacePath();

  const textParts = message.parts.filter((part) => part.type === "text");
  const toolParts = message.parts.filter(isToolMessagePart);
  const toolOutputParts = deferToolOutputs
    ? []
    : toolParts.filter(
        (part) =>
          part.state === "output-available" &&
          part.type !== "tool-suggestions" &&
          part.type !== "tool-getSprintAnalyticsTool" &&
          !isAnalyticsReportOutput(part.output),
      );
  const reportParts = deferToolOutputs
    ? []
    : toolParts.filter(
        (part) =>
          part.state === "output-available" &&
          isAnalyticsReportOutput(part.output),
      );
  const sprintAnalyticsParts = deferToolOutputs
    ? []
    : toolParts.filter(
        (part) =>
          part.type === "tool-getSprintAnalyticsTool" &&
          part.state === "output-available",
      );
  const suggestionParts = deferToolOutputs
    ? []
    : toolParts.filter(
        (part) =>
          part.type === "tool-suggestions" && part.state === "output-available",
      );
  const toolOutputSections = toolOutputParts.flatMap((part) =>
    getSectionsFromToolOutput(part, withWorkspace),
  );

  return (
    <>
      {textParts.map((part, index) => (
        <Streamdown
          className="chat-tables"
          components={STREAMDOWN_COMPONENTS}
          controls={{
            table: true,
            code: true,
            mermaid: {
              download: true,
              copy: true,
              fullscreen: true,
              panZoom: true,
            },
          }}
          isAnimating={isAnimating}
          key={`${message.id}-text-${index}`}
        >
          {part.text}
        </Streamdown>
      ))}

      {toolOutputSections.map((section, sectionIndex) => (
        <ToolOutputSectionView
          key={`${message.id}-tool-output-section-${sectionIndex}`}
          section={section}
        />
      ))}

      {reportParts.map((part, index) => (
        <AnalyticsReport
          key={`${message.id}-report-${index}`}
          output={asRecord(part.output)}
        />
      ))}

      {sprintAnalyticsParts.map((part, index) => (
        <Box className="mb-3" key={`${message.id}-sprint-${index}`}>
          <Text as="h3" className="mt-4 mb-1 text-xl font-semibold antialiased">
            Burndown graph
          </Text>
          <BurndownChart
            burndownData={getSprintBurndownData(part.output)}
            className={cn("h-72", {
              "h-80": pathname.includes("/maya"),
            })}
          />
        </Box>
      ))}

      {suggestionParts.map((part, index) => (
        <Flex
          className="mt-4"
          gap={2}
          key={`${message.id}-suggestions-${index}`}
          wrap
        >
          {getSuggestions(part.output).map((suggestion, i) => (
            <Button
              className="truncate"
              color="tertiary"
              key={i}
              onClick={() => {
                onPromptSelect(suggestion);
              }}
              size="sm"
            >
              {suggestion}
            </Button>
          ))}
        </Flex>
      ))}
    </>
  );
};

export const ChatMessage = ({
  isAnimating = false,
  isLast,
  message,
  profile,
  status,
  deferToolOutputs = false,
  regenerate,
  onPromptSelect,
}: ChatMessageProps) => {
  const [_, copy] = useCopyToClipboard();
  const [hasCopied, setHasCopied] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const { getTermDisplay } = useTerminology();
  const content = getMessageText(message);
  return (
    <>
      <Flex
        className={cn({
          "flex-row-reverse": message.role === "user",
        })}
        gap={message.role === "user" ? 3 : 0}
      >
        {message.role === "user" ? (
          <Avatar
            color="tertiary"
            name={profile?.fullName || profile?.username}
            src={profile?.avatarUrl}
          />
        ) : null}
        <Flex
          className={cn("flex-1", {
            "items-end": message.role === "user",
            "max-w-[80%]": message.role === "user",
            "max-w-full": message.role === "assistant",
          })}
          direction="column"
        >
          <Box
            className={cn("mb-2 rounded-2xl px-4 py-3", {
              "bg-state-hover/80 rounded-tr-md dark:bg-white/[0.08]":
                message.role === "user",
              "bg-transparent p-0": message.role === "assistant",
            })}
          >
            <RenderMessage
              deferToolOutputs={deferToolOutputs}
              isAnimating={isAnimating}
              message={message}
              onPromptSelect={onPromptSelect}
            />
          </Box>
          <AttachmentsDisplay message={message} />
          <Flex className="mt-2 px-0.5" justify="between">
            {message.role === "assistant" &&
            status !== "streaming" &&
            !isAnimating ? (
              <Flex gap={2} justify="end">
                <Tooltip title={`Create ${getTermDisplay("storyTerm")}`}>
                  <Button
                    asIcon
                    color="tertiary"
                    onClick={() => {
                      setIsOpen(true);
                    }}
                    size="sm"
                    variant="naked"
                  >
                    <PlusIcon />
                  </Button>
                </Tooltip>
                <Tooltip title="Copy">
                  <Button
                    asIcon
                    color="tertiary"
                    onClick={() => {
                      copy(content).then(() => {
                        setHasCopied(true);
                        setTimeout(() => {
                          setHasCopied(false);
                        }, 1500);
                      });
                    }}
                    size="sm"
                    variant="naked"
                  >
                    {hasCopied ? <CheckIcon /> : <CopyIcon />}
                  </Button>
                </Tooltip>
                {isLast && message.metadata?.source !== "voice" ? (
                  <Tooltip title="Retry">
                    <Button
                      asIcon
                      color="tertiary"
                      onClick={() => {
                        regenerate();
                      }}
                      size="sm"
                      variant="naked"
                    >
                      <ReloadIcon strokeWidth={2.8} />
                    </Button>
                  </Tooltip>
                ) : null}
              </Flex>
            ) : null}
          </Flex>
        </Flex>
      </Flex>
      <NewStoryDialog
        description={content}
        isOpen={isOpen}
        setIsOpen={setIsOpen}
      />
    </>
  );
};

"use client";

import {
  CommentIcon,
  IntakeIcon,
  LinkIcon,
  Notification02Icon,
  ObjectiveIcon,
  OKRIcon,
  SprintsIcon,
  TeamIcon,
} from "icons";
import { Avatar, Tooltip } from "ui";
import { useWorkspacePath } from "@/hooks";
import { PriorityIcon } from "../priority-icon";
import type {
  EntityResultIcon,
  EntityResultTone,
  EntityResultTrailing,
} from "./entity-results-data";
import { getEntityResultsModel } from "./entity-results-data";
import { GenerativeList, GenerativeListItem } from "./generative-list";

const EntityIcon = ({ icon }: { icon: EntityResultIcon }) => {
  if (icon.kind === "avatar") {
    return <Avatar name={icon.name} size="xs" src={icon.src} />;
  }

  if (icon.kind === "color") {
    return (
      <span
        className="size-3 rounded-[2px]"
        style={{ backgroundColor: icon.color }}
      />
    );
  }

  if (icon.kind === "priority") {
    return (
      <Tooltip title={`Priority: ${icon.priority}`}>
        <span className="flex size-5 items-center justify-center">
          <PriorityIcon className="max-h-4 max-w-4" priority={icon.priority} />
        </span>
      </Tooltip>
    );
  }

  const className = "size-4 text-icon";
  switch (icon.name) {
    case "comment":
      return <CommentIcon className={className} />;
    case "feedback":
      return <IntakeIcon className={className} />;
    case "key-result":
      return <OKRIcon className={className} />;
    case "link":
      return <LinkIcon className={className} />;
    case "notification":
      return <Notification02Icon className={className} />;
    case "objective":
      return <ObjectiveIcon className={className} />;
    case "sprint":
      return <SprintsIcon className={className} />;
    case "team":
      return <TeamIcon className={className} />;
  }
};

const toneClassNames: Record<EntityResultTone, string> = {
  danger: "bg-danger",
  info: "bg-info",
  muted: "bg-text-muted",
  success: "bg-success",
  warning: "bg-warning",
};

const EntityTrailing = ({ trailing }: { trailing: EntityResultTrailing }) => {
  if (trailing.kind === "text") {
    return (
      <span className="max-w-32 truncate" title={trailing.label}>
        {trailing.label}
      </span>
    );
  }

  return (
    <span
      aria-label={trailing.label}
      className="text-text-muted inline-flex max-w-[16ch] min-w-0 items-center gap-1.5 text-base font-medium"
      title={trailing.label}
    >
      <span
        className={`size-2.5 shrink-0 rounded-[2px] ${toneClassNames[trailing.tone]}`}
      />
      <span className="truncate">{trailing.label}</span>
    </span>
  );
};

export const EntityResults = ({
  output,
  toolType,
}: {
  output: unknown;
  toolType: string;
}) => {
  const model = getEntityResultsModel(toolType, output);
  const { withWorkspace } = useWorkspacePath();

  if (!model) return null;

  return (
    <GenerativeList emptyMessage={model.emptyMessage} title={model.title}>
      {model.items.map((item) => (
        <GenerativeListItem
          href={
            item.href?.startsWith("/") ? withWorkspace(item.href) : item.href
          }
          key={item.id}
          leading={<EntityIcon icon={item.icon} />}
          title={item.title}
          trailing={
            item.trailing ? <EntityTrailing trailing={item.trailing} /> : null
          }
        />
      ))}
    </GenerativeList>
  );
};

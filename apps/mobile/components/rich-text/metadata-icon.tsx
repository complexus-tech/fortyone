import { useState } from "react";
import type { StoryPriority } from "../../modules/stories/types";
import type { StatusCategory } from "../../types/statuses";
import {
  assigneeIconGeometry,
  priorityIconGeometry,
  statusIconGeometry,
  type TaskIconGeometry,
} from "../icons/task-icon-geometry";
import { getAvatarInitials } from "../ui/avatar-initials";
import { themeColors } from "../../constants/colors";
import { isSafeMediaUrl } from "./content";

export type EditorMetadata = {
  key: string;
  label: string;
  value: string;
  accessibilityValue?: string;
  kind: "status" | "priority" | "assignee" | "labels";
  muted?: boolean;
  color?: string;
  category?: StatusCategory;
  priority?: StoryPriority;
  avatar?: { name: string; src?: string | null };
};

function TaskIconDOM({
  geometry,
  size,
}: {
  geometry: TaskIconGeometry;
  size: number;
}) {
  return (
    <svg
      width={size}
      height={size}
      viewBox={geometry.viewBox}
      fill="none"
      aria-hidden="true"
    >
      {geometry.circles?.map((circle) => <circle key={circle.r} {...circle} />)}
      {geometry.rects?.map((rect) => <rect key={rect.x} {...rect} />)}
      {geometry.paths?.map((path) => <path key={path.d} {...path} />)}
    </svg>
  );
}

function AssigneeAvatar({
  avatar,
  dark,
}: {
  avatar: NonNullable<EditorMetadata["avatar"]>;
  dark: boolean;
}) {
  const [failedSrc, setFailedSrc] = useState<string | null>(null);
  const src = avatar.src && isSafeMediaUrl(avatar.src) ? avatar.src : null;
  const theme = themeColors[dark ? "dark" : "light"];

  return (
    <span
      className="metadata-avatar"
      aria-hidden="true"
      style={{ backgroundColor: theme.accent, color: theme.foreground }}
    >
      {src && src !== failedSrc ? (
        <img src={src} alt="" onError={() => setFailedSrc(src)} />
      ) : (
        getAvatarInitials(avatar.name)
      )}
    </span>
  );
}

export function MetadataIcon({
  item,
  dark,
}: {
  item: EditorMetadata;
  dark: boolean;
}) {
  const theme = themeColors[dark ? "dark" : "light"];
  if (item.kind === "status")
    return (
      <TaskIconDOM
        size={18}
        geometry={statusIconGeometry(item.category, item.color ?? theme.icon)}
      />
    );
  if (item.kind === "priority")
    return (
      <TaskIconDOM
        size={16}
        geometry={priorityIconGeometry(
          item.priority ?? "No Priority",
          theme.textMuted,
        )}
      />
    );
  if (item.kind === "assignee")
    return item.avatar ? (
      <AssigneeAvatar avatar={item.avatar} dark={dark} />
    ) : (
      <TaskIconDOM size={18} geometry={assigneeIconGeometry(theme.icon)} />
    );

  return (
    <svg
      width={17}
      height={17}
      viewBox="0 0 24 24"
      fill="none"
      stroke={item.color ?? "currentColor"}
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M20 13 11 22 2 13V3h10l8 8a1.5 1.5 0 0 1 0 2Z" />
      <circle cx="7" cy="8" r="1" />
    </svg>
  );
}

import { EditIcon } from "icons";
import { Avatar, Tooltip } from "ui";

type EditableAvatarProps = {
  label: string;
  name?: string | null;
  onClick: () => void;
  src?: string | null;
};

export const EditableAvatar = ({
  label,
  name,
  onClick,
  src,
}: EditableAvatarProps) => (
  <Tooltip title={label}>
    <button
      aria-label={label}
      className="border-border hover:border-primary/60 focus-visible:border-primary focus-visible:ring-primary/30 group relative rounded-full border border-dashed p-1.5 transition-colors outline-none focus-visible:ring-2 focus-visible:ring-offset-2"
      onClick={onClick}
      type="button"
    >
      <Avatar
        className="h-12"
        name={name ?? undefined}
        src={src ?? undefined}
      />
      <span
        aria-hidden="true"
        className="text-foreground absolute right-0 bottom-0 flex transition-transform group-hover:scale-105"
      >
        <EditIcon
          className="!text-foreground !h-4 !w-auto"
          strokeWidth={2.25}
        />
      </span>
    </button>
  </Tooltip>
);

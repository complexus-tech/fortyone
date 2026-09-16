import { Avatar } from "ui";
import { cn } from "lib";
import type { UserSummary } from "@/types";

type Person = Pick<UserSummary, "id" | "fullName" | "username" | "avatarUrl">;

type PeopleAvatarsProps = {
  people: Person[];
  totalCount?: number;
  size?: "xs" | "sm";
};

export const PeopleAvatars = ({
  people,
  totalCount = people.length,
  size = "sm",
}: PeopleAvatarsProps) => {
  const visiblePeople = people.slice(0, 3);
  const remainingCount = Math.max(0, totalCount - visiblePeople.length);

  return (
    <span className="isolate flex items-center -space-x-1.5">
      {visiblePeople.map((person) => (
        <Avatar
          className="ring-surface relative ring-2"
          key={person.id}
          name={person.fullName || person.username}
          size={size}
          src={person.avatarUrl}
          title={person.fullName || person.username}
        />
      ))}
      {remainingCount > 0 ? (
        <span
          aria-label={`${remainingCount} more people`}
          className={cn(
            "bg-primary text-primary-foreground ring-surface relative flex shrink-0 items-center justify-center rounded-full font-medium ring-2",
            size === "sm" ? "size-7 text-[0.65rem]" : "size-5 text-[0.6rem]",
          )}
          title={`${remainingCount} more people`}
        >
          +{remainingCount}
        </span>
      ) : null}
      {totalCount === 0 ? <Avatar size={size} /> : null}
    </span>
  );
};

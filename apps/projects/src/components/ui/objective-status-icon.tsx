import { cn } from "lib";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";

export const ObjectiveStatusIcon = ({
  statusId,
  className,
}: {
  statusId?: string;
  className?: string;
}) => {
  const { data: statuses = [] } = useObjectiveStatuses();
  if (!statuses.length) return null;
  const state =
    statuses.find((state) => state.id === statusId) || statuses.at(0);
  const { category } = state!;
  return (
    <span style={{ color: state?.color }}>
      {category === "backlog" && (
        <svg
          className={cn("h-[1.15rem] w-auto", className)}
          fill="none"
          height="24"
          stroke="currentColor"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2.5"
          viewBox="0 0 24 24"
          width="24"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path d="M10.1 2.182a10 10 0 0 1 3.8 0" />
          <path d="M13.9 21.818a10 10 0 0 1-3.8 0" />
          <path d="M17.609 3.721a10 10 0 0 1 2.69 2.7" />
          <path d="M2.182 13.9a10 10 0 0 1 0-3.8" />
          <path d="M20.279 17.609a10 10 0 0 1-2.7 2.69" />
          <path d="M21.818 10.1a10 10 0 0 1 0 3.8" />
          <path d="M3.721 6.391a10 10 0 0 1 2.7-2.69" />
          <path d="M6.391 20.279a10 10 0 0 1-2.69-2.7" />
        </svg>
      )}
      {category === "unstarted" && (
        <svg
          className={cn("h-[1.15rem] w-auto", className)}
          fill="none"
          height="24"
          stroke="currentColor"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2.5"
          viewBox="0 0 24 24"
          width="24"
          xmlns="http://www.w3.org/2000/svg"
        >
          <circle cx="12" cy="12" r="10" />
        </svg>
      )}
      {category === "started" && (
        <svg
          className={cn("h-[1.15rem] w-auto", className)}
          fill="none"
          height="24"
          stroke="currentColor"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2.5"
          viewBox="0 0 24 24"
          width="24"
          xmlns="http://www.w3.org/2000/svg"
        >
          <circle cx="12" cy="12" r="10" />
        </svg>
      )}
      {category === "paused" && (
        <svg
          className={cn("h-[1.15rem] w-auto", className)}
          fill="none"
          height="24"
          stroke="currentColor"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2.5"
          viewBox="0 0 24 24"
          width="24"
          xmlns="http://www.w3.org/2000/svg"
        >
          <circle cx="12" cy="12" r="10" />
          <line x1="10" x2="10" y1="15" y2="9" />
          <line x1="14" x2="14" y1="15" y2="9" />
        </svg>
      )}
      {category === "completed" && (
        <svg
          className={cn("h-[1.15rem] w-auto", className)}
          fill="none"
          height="24"
          stroke="currentColor"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2.5"
          viewBox="0 0 24 24"
          width="24"
          xmlns="http://www.w3.org/2000/svg"
        >
          <circle cx="12" cy="12" r="10" />
          <path d="m9 12 2 2 4-4" />
        </svg>
      )}
      {category === "cancelled" && (
        <svg
          className={cn("h-[1.15rem] w-auto", className)}
          fill="none"
          height="24"
          stroke="currentColor"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2.5"
          viewBox="0 0 24 24"
          width="24"
          xmlns="http://www.w3.org/2000/svg"
        >
          <circle cx="12" cy="12" r="10" />
          <path d="m15 9-6 6" />
          <path d="m9 9 6 6" />
        </svg>
      )}
    </span>
  );
};

import { cn } from "lib";
import type { StoryPriority } from "@/modules/stories/types";

export const PriorityIcon = ({
  priority = "No Priority",
  className,
}: {
  priority?: StoryPriority;
  className?: string;
}) => {
  return (
    <>
      {(priority === "No Priority" || priority === null) && (
        <svg
          className={cn("text-gray dark:text-gray-300", className)}
          fill="currentColor"
          focusable="false"
          height="16"
          viewBox="0 0 16 16"
          width="16"
        >
          <rect height="1.5" opacity="0.9" rx="0.5" width="3" x="1" y="7.25" />
          <rect height="1.5" opacity="0.9" rx="0.5" width="3" x="6" y="7.25" />
          <rect height="1.5" opacity="0.9" rx="0.5" width="3" x="11" y="7.25" />
        </svg>
      )}

      {priority === "Urgent" && (
        <svg
          className={cn("h-5 w-auto text-primary", className)}
          fill="currentColor"
          height="24"
          viewBox="0 0 24 24"
          width="24"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M2.5 12C2.5 7.52166 2.5 5.28249 3.89124 3.89124C5.28249 2.5 7.52166 2.5 12 2.5C16.4783 2.5 18.7175 2.5 20.1088 3.89124C21.5 5.28249 21.5 7.52166 21.5 12C21.5 16.4783 21.5 18.7175 20.1088 20.1088C18.7175 21.5 16.4783 21.5 12 21.5C7.52166 21.5 5.28249 21.5 3.89124 20.1088C2.5 18.7175 2.5 16.4783 2.5 12Z"
            stroke="currentColor"
            strokeWidth="2.5"
          />
          <path
            d="M11.9998 16H12.0088"
            stroke="white"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="2.5"
          />
          <path
            d="M12 13L12 7"
            stroke="white"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="2.5"
          />
        </svg>
      )}
      {priority === "High" && (
        <svg
          className={cn("text-gray dark:text-gray-300", className)}
          fill="currentColor"
          focusable="false"
          height="16"
          viewBox="0 0 16 16"
          width="16"
        >
          <rect height="6" rx="1" width="3" x="1" y="8" />
          <rect height="9" rx="1" width="3" x="6" y="5" />
          <rect height="12" rx="1" width="3" x="11" y="2" />
        </svg>
      )}
      {priority === "Medium" && (
        <svg
          className={cn("text-gray dark:text-gray-300", className)}
          fill="currentColor"
          focusable="false"
          height="16"
          viewBox="0 0 16 16"
          width="16"
        >
          <rect height="6" rx="1" width="3" x="1" y="8" />
          <rect height="9" rx="1" width="3" x="6" y="5" />
          <rect fillOpacity="0.4" height="12" rx="1" width="3" x="11" y="2" />
        </svg>
      )}
      {priority === "Low" && (
        <svg
          className={cn("text-gray dark:text-gray-300", className)}
          fill="currentColor"
          focusable="false"
          height="16"
          viewBox="0 0 16 16"
          width="16"
        >
          <rect height="6" rx="1" width="3" x="1" y="8" />
          <rect fillOpacity="0.4" height="9" rx="1" width="3" x="6" y="5" />
          <rect fillOpacity="0.4" height="12" rx="1" width="3" x="11" y="2" />
        </svg>
      )}
    </>
  );
};

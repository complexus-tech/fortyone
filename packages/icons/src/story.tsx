import { cn } from "lib";
import type { Icon } from "./types";

export const StoryIcon = (props: Icon) => {
  const { strokeWidth = 2, className, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-icon", className)}
      fill="none"
      height="24"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={strokeWidth}
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path d="M21 13.9563V10.0437C21 8.33384 21 7.47888 20.5856 6.77939C20.1712 6.0799 19.4188 5.6647 17.9139 4.83429L14.9139 3.1789C13.4895 2.39297 12.7774 2 12 2C11.2226 2 10.5105 2.39297 9.08614 3.1789L6.08614 4.83429C4.58123 5.6647 3.82877 6.0799 3.41439 6.77939C3 7.47888 3 8.33384 3 10.0437V13.9563C3 15.6662 3 16.5211 3.41439 17.2206C3.82877 17.9201 4.58123 18.3353 6.08614 19.1657L9.08614 20.8211C10.5105 21.607 11.2226 22 12 22C12.7774 22 13.4895 21.607 14.9139 20.8211L17.9139 19.1657C19.4188 18.3353 20.1712 17.9201 20.5856 17.2206C21 16.5211 21 15.6662 21 13.9563Z" />
      <path d="M12 16C14.2091 16 16 14.2091 16 12C16 9.79086 14.2091 8 12 8C9.79086 8 8 9.79086 8 12C8 14.2091 9.79086 16 12 16Z" />
    </svg>
  );
};

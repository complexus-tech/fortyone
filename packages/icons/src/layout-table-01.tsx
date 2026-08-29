import { cn } from "lib";
import type { Icon } from "./types";

export const LayoutTable01Icon = (props: Icon) => {
  const { className, strokeWidth = 2, ...rest } = props;
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
      <path d="M3.89124 20.1088C2.5 18.7175 2.5 16.4783 2.5 12C2.5 7.52166 2.5 5.28249 3.89124 3.89124C5.28248 2.5 7.52166 2.5 12 2.5C16.4783 2.5 18.7175 2.5 20.1088 3.89124C21.5 5.28248 21.5 7.52166 21.5 12C21.5 16.4783 21.5 18.7175 20.1088 20.1088C18.7175 21.5 16.4783 21.5 12 21.5C7.52166 21.5 5.28249 21.5 3.89124 20.1088Z" />
      <path d="M2.5 9L21.5 9" />
      <path d="M16 21.5L16 2.5" />
      <path d="M8 21.4998L8 2.49976" />
    </svg>
  );
};

import { cn } from "lib";
import type { Icon } from "./types";

export const ObjectiveIcon = (props: Icon) => {
  const { className, strokeWidth = 2.2, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-icon", className)}
      fill="none"
      strokeWidth={strokeWidth}
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M21 12.0002C21 16.9708 16.9706 21.0002 12 21.0002C7.02944 21.0002 3 16.9708 3 12.0002C3 7.02963 7.02944 3.00019 12 3.00019M16.7712 13.4996C16.1341 15.5283 14.2389 16.9998 12 16.9998C9.23858 16.9998 7 14.7612 7 11.9998C7 9.76099 8.47145 7.86584 10.5 7.22871M14 10.0002H17.1716C17.702 10.0002 18.2107 9.78948 18.5858 9.41441L21.214 6.78622C21.3454 6.65481 21.2926 6.43105 21.1162 6.37228L18.5 5.50019L17.6279 2.88394C17.5691 2.70763 17.3454 2.65481 17.214 2.78622L14.5858 5.41441C14.2107 5.78948 14 6.29819 14 6.82862V10.0002ZM14 10.0002L12 12.0002"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};

import { cn } from "lib";
import type { Icon } from "./types";

export const AiIcon = (props: Icon) => {
  const { strokeWidth = 2, className, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-gray dark:text-gray-300", className)}
      fill="currentColor"
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M4.25 22C4.25 17.7198 7.71979 14.25 12 14.25C16.2802 14.25 19.75 17.7198 19.75 22C19.75 22.4142 19.4142 22.75 19 22.75H5C4.58579 22.75 4.25 22.4142 4.25 22Z"
        fill="currentColor"
      />
      <path
        fillRule="evenodd"
        clipRule="evenodd"
        d="M12 1.25C12.4142 1.25 12.75 1.58579 12.75 2V3.25H16.5C18.0188 3.25 19.25 4.48122 19.25 6V8C19.25 10.6234 17.1234 12.75 14.5 12.75H9.5C6.87665 12.75 4.75 10.6234 4.75 8V6C4.75 4.48122 5.98122 3.25 7.5 3.25H11.25V2C11.25 1.58579 11.5858 1.25 12 1.25ZM9.5 6.5C8.94772 6.5 8.5 6.94772 8.5 7.5C8.5 8.05228 8.94772 8.5 9.5 8.5H9.50998C10.0623 8.5 10.51 8.05228 10.51 7.5C10.51 6.94772 10.0623 6.5 9.50998 6.5H9.5ZM14.49 6.5C13.9377 6.5 13.49 6.94772 13.49 7.5C13.49 8.05228 13.9377 8.5 14.49 8.5H14.5C15.0523 8.5 15.5 8.05228 15.5 7.5C15.5 6.94772 15.0523 6.5 14.5 6.5H14.49Z"
        fill="currentColor"
      />
    </svg>
  );
};

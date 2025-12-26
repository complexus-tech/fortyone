import { cn } from "lib";

export const Divider = ({ className = "" }) => {
  return (
    <div
      className={cn(
        "dark:border-dark-50 border-gray-100 border-t-[0.5px]",
        className
      )}
    />
  );
};

import { cn } from "lib";

export const Divider = ({ className = "" }) => {
  return (
    <div
      className={cn("dark:border-dark-200 border-gray-100 border-t", className)}
    />
  );
};

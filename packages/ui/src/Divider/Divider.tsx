import { cn } from "lib";

export const Divider = ({ className = "" }) => {
  return (
    <div
      className={cn("dark:border-dark-100 border-gray-50 border-t", className)}
    />
  );
};

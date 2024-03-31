import { cn } from "lib";
import { HTMLAttributes, ReactNode } from "react";
import { Box } from "../Box/Box";

interface Props extends HTMLAttributes<HTMLDivElement> {
  className?: string;
  children?: ReactNode;
}

export const Wrapper = ({ children, className, ...rest }: Props) => {
  return (
    <Box
      className={cn(
        "rounded-lg border bg-gray-50/10 border-gray-100/60 px-4 py-4 shadow-sm dark:border-dark-200 dark:bg-dark-300/40",
        className
      )}
      {...rest}
    >
      {children}
    </Box>
  );
};

import { cn } from "lib";
import { HTMLAttributes, JSXElementConstructor, ReactNode } from "react";
import { Box } from "../Box/Box";

interface Props extends HTMLAttributes<HTMLDivElement> {
  className?: string;
  children?: ReactNode;
}

export const Wrapper = ({ children, className, ...rest }: Props) => {
  return (
    <Box
      className={cn(
        "rounded-lg border border-gray-50 px-4 py-4 shadow dark:border-dark-200 dark:bg-dark-300/40",
        className
      )}
      {...rest}
    >
      {children}
    </Box>
  );
};

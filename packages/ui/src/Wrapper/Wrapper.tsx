import { cn } from "lib";
import { HTMLAttributes, JSXElementConstructor, ReactNode } from "react";
import { Box } from "../Box/Box";

interface Props extends HTMLAttributes<HTMLDivElement> {
  className?: string;
  as?: "div" | "form" | "section" | "article" | JSXElementConstructor<any>;
  children?: ReactNode;
}

export const Wrapper = ({
  children,
  className,
  as = "div",
  ...rest
}: Props) => {
  return (
    <Box
      as={as}
      className={cn(
        "rounded-lg border border-gray-50 px-4 py-4 shadow dark:border-dark-200 dark:bg-dark-200/30",
        className
      )}
      {...rest}
    >
      {children}
    </Box>
  );
};

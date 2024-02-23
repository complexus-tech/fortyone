import { cn } from "lib";
import type { LinkProps } from "next/link";
import Link from "next/link";
import type { ReactNode } from "react";

type NavLinkProps = LinkProps & {
  active?: boolean;
  className?: string;
  children: ReactNode;
};

export const NavLink = ({
  href,
  active,
  className,
  children,
  ...rest
}: NavLinkProps) => {
  return (
    <Link
      className={cn(
        "group flex items-center gap-2 rounded-lg px-2 py-2 text-gray-300 outline-none transition-colors duration-200 hover:bg-gray-50/70 dark:text-gray-200 dark:hover:bg-dark-200/60 dark:hover:text-white",
        {
          "bg-gray-50 font-medium dark:bg-dark-200 dark:text-white": active,
        },
        className,
      )}
      href={href}
      {...rest}
    >
      {children}
    </Link>
  );
};

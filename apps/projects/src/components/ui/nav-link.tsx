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
        "group flex items-center gap-2 rounded-[0.6rem] px-2 py-[0.4rem] outline-none transition-colors duration-200 hover:bg-gray-100/80 dark:hover:bg-dark-200/90 dark:hover:text-white",
        {
          "bg-gray-100/80 font-medium dark:bg-dark-200 dark:text-white": active,
        },
        className,
      )}
      href={href}
      prefetch
      {...rest}
    >
      {children}
    </Link>
  );
};

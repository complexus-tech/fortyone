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
        "group flex items-center gap-2 rounded-[0.6rem] px-2 py-[0.4rem] text-foreground outline-none transition-colors duration-200 hover:bg-accent",
        {
          "bg-accent": active,
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

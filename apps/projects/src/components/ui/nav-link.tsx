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
        "group text-foreground hover:bg-accent flex items-center gap-2 rounded-lg px-2 py-[0.4rem] transition-colors duration-200 outline-none",
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

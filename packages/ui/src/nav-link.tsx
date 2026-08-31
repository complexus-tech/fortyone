import { cn } from "lib";
import Link, { LinkProps } from "next/link";
import type { AnchorHTMLAttributes, FC, ReactNode } from "react";

interface Props
  extends LinkProps,
    Omit<AnchorHTMLAttributes<HTMLAnchorElement>, keyof LinkProps> {
  children: ReactNode;
  className?: string;
  active?: boolean;
}

export const NavLink: FC<Props> = ({
  children,
  className,
  active,
  href,
  ...props
}) => {
  const classes = cn(
    "transition duration-300 text-[0.95rem] ease-linear px-2",
    {
      "text-primary": active,
    },
    className,
  );
  return (
    <Link
      {...props}
      href={href}
      className={classes}
      target={href.toString().startsWith("http") ? "_blank" : undefined}
    >
      {children}
    </Link>
  );
};

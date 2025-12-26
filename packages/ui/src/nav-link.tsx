import { cn } from "lib";
import Link, { LinkProps } from "next/link";
import { FC, ReactNode } from "react";

interface Props extends LinkProps {
  children: ReactNode;
  className?: string;
  active?: boolean;
}

export const NavLink: FC<Props> = ({ children, className, active, href }) => {
  const classes = cn(
    "transition duration-300 text-[0.95rem] ease-linear px-2",
    {
      "text-primary": active,
    },
    className
  );
  return (
    <Link
      href={href}
      className={classes}
      target={href.toString().startsWith("http") ? "_blank" : undefined}
    >
      {children}
    </Link>
  );
};

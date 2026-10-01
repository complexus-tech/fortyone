import type { ComponentPropsWithoutRef } from "react";
import { cn } from "lib";
import { NavLink } from "ui";
import styles from "./navigation.module.css";

type RequestDemoProps = Pick<
  ComponentPropsWithoutRef<typeof NavLink>,
  "className" | "onClick"
>;

export const RequestDemo = ({ className, onClick }: RequestDemoProps) => {
  return (
    <NavLink
      className={cn(
        "hover:bg-state-hover flex items-center rounded-md px-3 py-1.5 whitespace-nowrap transition",
        styles.topLevelLink,
        className,
      )}
      href="https://cal.com/fortyoneapp/15min"
      onClick={onClick}
      prefetch={false}
      rel="noreferrer"
    >
      Book a meeting
    </NavLink>
  );
};

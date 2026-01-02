import Link from "next/link";

import { cn } from "lib";
import { Flex } from "./flex";
import { ReactNode } from "react";
import { ArrowRight2Icon } from "icons";
interface BreadCrumb {
  name: string;
  url?: string;
  icon?: ReactNode;
  className?: string;
}

export type BreadCrumbsProps = {
  breadCrumbs: BreadCrumb[];
  className?: string;
};

export const BreadCrumbs = ({ breadCrumbs, className }: BreadCrumbsProps) => {
  return (
    <Flex align="center" gap={2} className={className}>
      {breadCrumbs.map(({ name, icon, url = "", className = "" }, idx) => (
        <Link
          key={idx}
          href={url}
          className={cn(
            "flex items-center gap-1.5 font-medium group first-letter:uppercase transition",
            {
              "text-text-muted":
                idx + 1 === breadCrumbs.length,
              "pointer-events-none": !url,
            },
            className
          )}
        >
          <>
            {icon && <span className="group-hover:text-primary">{icon}</span>}
            <span className="group-hover:text-primary line-clamp-1">
              {name}
            </span>
            <ArrowRight2Icon
              strokeWidth={3}
              className={cn("h-4 opacity-80", {
                hidden: idx + 1 === breadCrumbs.length,
              })}
            />
          </>
        </Link>
      ))}
    </Flex>
  );
};

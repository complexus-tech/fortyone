import Link from "next/link";

import { cn } from "lib";
import { Flex } from "../Flex/Flex";
import { ReactNode } from "react";
import { ArrowRightIcon } from "icons";
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
              "text-gray/80 dark:text-gray-300/80":
                idx + 1 === breadCrumbs.length,
              "pointer-events-none": !url,
            },
            className
          )}
        >
          <>
            {icon && <span className="group-hover:text-primary">{icon}</span>}
            <span className="group-hover:text-primary">{name}</span>
            <ArrowRightIcon
              className={cn("h-[0.8rem] w-auto opacity-80", {
                hidden: idx + 1 === breadCrumbs.length,
              })}
            />
          </>
        </Link>
      ))}
    </Flex>
  );
};

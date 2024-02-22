import Link from "next/link";

import { cn } from "lib";
import { Flex } from "../Flex/Flex";
import { ReactNode } from "react";
import { ChevronRight } from "lucide-react";
interface BreadCrumb {
  name: string;
  url?: string;
  icon?: ReactNode;
}

export type BreadCrumbsProps = {
  breadCrumbs: BreadCrumb[];
  className?: string;
};

export const BreadCrumbs = ({ breadCrumbs, className }: BreadCrumbsProps) => {
  return (
    <Flex align="center" gap={2} className={className}>
      {breadCrumbs.map(({ name, icon, url = "" }, idx) => (
        <Link
          key={idx}
          href={url}
          className={cn(
            "flex items-center gap-2 font-medium group capitalize text-gray-300 transition dark:text-gray-200",
            {
              "text-gray-250 dark:text-gray": idx + 1 === breadCrumbs.length,
              "pointer-events-none": !url,
            }
          )}
        >
          {icon && <span className="group-hover:text-primary">{icon}</span>}
          <span className="group-hover:text-primary">{name}</span>
          <ChevronRight
            className={cn("h-[0.8rem] w-auto", {
              hidden: idx + 1 === breadCrumbs.length,
            })}
          />
        </Link>
      ))}
    </Flex>
  );
};

import Link from "next/link";

import { cn } from "lib";
import { Flex } from "./flex";
import { ReactNode } from "react";
import { Tooltip } from "./tooltip";

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
      {breadCrumbs.map(({ name, icon, url = "", className = "" }, idx) => {
        if (!url) {
          return (
            <span
              key={idx}
              className={cn(
                "flex items-center gap-1.5 font-medium group first-letter:uppercase transition",
                {
                  "text-text-muted": idx + 1 === breadCrumbs.length,
                },
                className
              )}
            >
              <>
                {icon && <span>{icon}</span>}
                <Tooltip
                  title={name?.length > 24 ? name : null}
                  delayDuration={1000}
                  className="max-w-72"
                >
                  <span className="line-clamp-1 max-w-[24ch] truncate">
                    {name}
                  </span>
                </Tooltip>
                <span
                  className={cn("opacity-80", {
                    hidden: idx + 1 === breadCrumbs.length,
                  })}
                >
                  /
                </span>
              </>
            </span>
          );
        }

        return (
          <Link
            key={idx}
            href={url}
            className={cn(
              "flex items-center gap-1.5 font-medium group first-letter:uppercase transition",
              {
                "text-text-muted": idx + 1 === breadCrumbs.length,
              },
              className
            )}
          >
            <>
              {icon && <span className="group-hover:text-primary">{icon}</span>}
              <Tooltip
                title={name?.length > 24 ? name : null}
                delayDuration={1000}
                className="max-w-72"
              >
                <span className="group-hover:text-primary line-clamp-1 max-w-[24ch] truncate">
                  {name}
                </span>
              </Tooltip>
              <span
                className={cn("opacity-80", {
                  hidden: idx + 1 === breadCrumbs.length,
                })}
              >
                /
              </span>
            </>
          </Link>
        );
      })}
    </Flex>
  );
};

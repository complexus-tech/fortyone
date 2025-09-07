import type { Icon } from "icons/src/types";
import { cn } from "lib";

export const FortyOneLogo = ({ className }: Icon) => {
  return (
    <div
      className={cn(
        "font-heading text-[1.5rem] font-semibold text-black dark:text-white",
        className,
      )}
    >
      forty
      <span className="ml-0.5 inline-block bg-[#000000] px-0.5 pb-0.5 leading-none text-white dark:bg-white dark:text-black">
        one
      </span>
    </div>
  );
};

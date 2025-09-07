import Link from "next/link";
import { cn } from "lib";

export const Logo = ({
  className,
  // asIcon,
  // fill,
}: {
  className?: string;
  asIcon?: boolean;
  fill?: string;
}) => {
  return (
    <Link
      className={cn(
        "inline-block font-heading text-[1.6rem] font-semibold tracking-tight text-dark dark:text-white",
        className,
      )}
      href="/"
    >
      forty
      <span className="ml-0.5 inline-block bg-dark px-0.5 pb-0.5 leading-none text-white dark:bg-white dark:text-black">
        one
      </span>
    </Link>
  );
};

import { cn } from "lib";
import Link from "next/link";

export const Logo = ({ className }: { className?: string }) => {
  return (
    <Link href="/">
      <img
        alt="Africa Giving logo"
        className={cn("relative top-[2px] h-[5rem] w-auto", className)}
        src="/images/logo.png"
      />
    </Link>
  );
};

import { cn } from "lib";

export const MenuButton = ({ open }: { open: boolean }) => {
  return (
    <span
      aria-label="Menu button"
      className={cn("flex aspect-square h-10 items-center justify-center")}
    >
      <span>
        <span
          className={cn(
            "mb-[0.4rem] block h-px w-5 bg-dark dark:bg-white transition duration-300 ease-in-out",
            {
              "mb-0 rotate-45": open,
            },
          )}
        />
        <span
          className={cn(
            "block h-px w-5 bg-dark dark:bg-white transition duration-300 ease-in-out",
            {
              "-translate-y-[0.05rem] -rotate-45": open,
            },
          )}
        />
      </span>
    </span>
  );
};

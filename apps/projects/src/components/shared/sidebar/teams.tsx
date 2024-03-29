"use client";
import { cn } from "lib";
import { Box, Button, Flex } from "ui";
import { ArrowDownIcon, PlusIcon, TeamIcon } from "icons";
import { useLocalStorage } from "@/hooks";
import { Team } from "./team";

export const Teams = () => {
  const [isOpen, setIsOpen] = useLocalStorage<boolean>("teams-dropdown", true);

  const teams = [
    {
      id: 1,
      icon: "🚀",
      name: "Engineering",
    },
    {
      id: 2,
      icon: "🇦🇫",
      name: "Marketing",
    },
    {
      id: 3,
      icon: "🏀",
      name: "Product",
    },
  ];

  return (
    <Box className="mt-4">
      <Flex
        align="center"
        className="group mb-1 h-[2.5rem] select-none rounded-lg pl-2.5 pr-1 outline-none transition hover:bg-gray-250/5 focus:bg-gray-250/5 hover:dark:bg-dark-50/20 focus:dark:bg-dark-50/20"
        justify="between"
        onClick={() => {
          setIsOpen(!isOpen);
        }}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            setIsOpen(!isOpen);
          }
        }}
        role="button"
        tabIndex={0}
      >
        <span className="flex items-center gap-2 font-medium">
          <TeamIcon className="h-[1.3rem] w-auto" />
          Workspace Teams
          <ArrowDownIcon
            className={cn(
              "relative top-[0.2px] h-4 w-auto -rotate-90 transition-transform dark:text-gray",
              {
                "rotate-0": isOpen,
              },
            )}
            strokeWidth={3.5}
          />
        </span>
        <Button
          className="hidden group-hover:inline"
          color="tertiary"
          leftIcon={<PlusIcon className="h-5 w-auto" />}
          size="sm"
          type="button"
          variant="naked"
        >
          <span className="sr-only">Add new team</span>
        </Button>
      </Flex>
      <Flex
        className={cn(
          "ml-5 h-0 max-h-[54vh] overflow-hidden overflow-y-auto border-l border-dotted border-gray-250/15 pl-2 transition-all duration-300 dark:border-dark-50",
          {
            "h-max": isOpen,
          },
        )}
        direction="column"
        gap={1}
      >
        {teams.map(({ id, icon, name }) => (
          <Team icon={icon} id={id} key={id} name={name} />
        ))}
      </Flex>
    </Box>
  );
};

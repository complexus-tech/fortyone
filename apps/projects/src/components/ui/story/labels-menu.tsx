"use client";
import { createContext, useContext, useState, type ReactNode } from "react";
import { Button, Command, Divider, Flex, Popover, Text } from "ui";
import { CheckIcon, PlusIcon } from "icons";
import { generateRandomColor } from "lib";
import { useLabels } from "@/lib/hooks/labels";
import { useCreateLabelMutation } from "@/lib/hooks/create-label-mutation";
import { Dot } from "../dot";

const LabelsContext = createContext<{
  open: boolean;
  setOpen: (open: boolean) => void;
}>({
  open: false,
  setOpen: () => {},
});

export const useLabelsMenu = () => {
  const { open, setOpen } = useContext(LabelsContext);
  return { open, setOpen };
};

const Menu = ({ children }: { children: ReactNode }) => {
  const { open, setOpen } = useLabelsMenu();
  return (
    <Popover onOpenChange={setOpen} open={open}>
      {children}
    </Popover>
  );
};

export const LabelsMenu = ({ children }: { children: ReactNode }) => {
  const [open, setOpen] = useState(false);
  return (
    <LabelsContext.Provider value={{ open, setOpen }}>
      <Menu>{children}</Menu>
    </LabelsContext.Provider>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Popover.Trigger asChild>{children}</Popover.Trigger>
);

const Items = ({
  align = "center",
  labelIds,
  setLabelIds,
  teamId,
}: {
  labelIds: string[];
  teamId: string;
  setLabelIds: (labelIds: string[]) => void;
  align?: "center" | "start" | "end" | undefined;
}) => {
  const { mutateAsync: createLabel } = useCreateLabelMutation();
  const { data: allLabels = [] } = useLabels();
  const labels = allLabels.filter(
    (label) => label.teamId === teamId || label.teamId === null,
  );
  const [query, setQuery] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const handleCreateLabel = async () => {
    const usedColors = labels.map((label) => label.color);
    const color = generateRandomColor({ exclude: usedColors });
    try {
      setIsLoading(true);
      await createLabel(
        { name: query, color, teamId },
        {
          onSuccess(data) {
            setLabelIds([...labelIds, data.id]);
          },
        },
      );
      setQuery("");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Popover.Content align={align}>
      <Command>
        <Command.Input
          autoFocus
          onValueChange={(value) => {
            setQuery(value);
          }}
          placeholder="Update labels..."
          value={query}
        />
        <Divider className="my-2" />
        <Command.Empty className="justify-center px-1 py-0">
          <Button
            className="mx-0 border-0 text-base font-medium"
            color="tertiary"
            fullWidth
            loading={isLoading}
            loadingText="Creating label..."
            onClick={handleCreateLabel}
          >
            <PlusIcon className="h-4" strokeWidth={2.7} /> Create new label:{" "}
            <span className="font-medium text-gray-300">
              &ldquo;{query}&rdquo;
            </span>
          </Button>
        </Command.Empty>
        <Command.Group>
          {labels.map(({ id, name, color }) => (
            <Command.Item
              className="justify-between gap-4"
              key={id}
              onSelect={() => {
                if (!labelIds.includes(id)) {
                  setLabelIds([...labelIds, id]);
                } else {
                  setLabelIds(labelIds.filter((labelId) => labelId !== id));
                }
              }}
              value={name}
            >
              <Flex align="center" gap={2}>
                <Dot color={color} />
                <Text className="mr-4 flex items-center gap-3">{name}</Text>
              </Flex>
              <Flex align="center" gap={1}>
                {labelIds.includes(id) ? (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                ) : null}
              </Flex>
            </Command.Item>
          ))}
        </Command.Group>
      </Command>
    </Popover.Content>
  );
};

LabelsMenu.Trigger = Trigger;
LabelsMenu.Items = Items;

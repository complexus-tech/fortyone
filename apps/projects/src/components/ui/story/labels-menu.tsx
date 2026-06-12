"use client";
import {
  createContext,
  use,
  useDeferredValue,
  useState,
  type ReactNode,
  type UIEvent,
} from "react";
import { Button, Command, Divider, Flex, Popover, Text } from "ui";
import { CheckIcon, LoadingIcon, PlusIcon } from "icons";
import { generateRandomColor } from "lib";
import {
  LABEL_MENU_PAGE_SIZE,
  useLabelsInfinite,
} from "@/lib/hooks/labels";
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
  const { open, setOpen } = use(LabelsContext);
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
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const [isLoading, setIsLoading] = useState(false);
  const { open } = useLabelsMenu();
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useLabelsInfinite(
    { search: deferredQuery, teamId },
    LABEL_MENU_PAGE_SIZE,
    open,
  );
  const labels = data?.pages.flatMap((page) => page.labels) ?? [];

  const handleCreateLabel = async () => {
    const usedColors = labels.map((label) => label.color);
    const color = generateRandomColor({ exclude: usedColors });
    setIsLoading(true);
    await createLabel(
      { name: query, color, teamId },
      {
        onSuccess(res) {
          if (res.data) {
            setLabelIds([...labelIds, res.data.id]);
          }
        },
      },
    )
      .then(() => {
        setQuery("");
      })
      .finally(() => {
        setIsLoading(false);
      });
  };

  const handleScroll = (event: UIEvent<HTMLDivElement>) => {
    const target = event.currentTarget;
    const distanceToBottom =
      target.scrollHeight - target.scrollTop - target.clientHeight;

    if (distanceToBottom <= 80 && hasNextPage && !isFetchingNextPage) {
      void fetchNextPage();
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
            <span className="text-text-muted font-medium">
              &ldquo;{query}&rdquo;
            </span>
          </Button>
        </Command.Empty>
        <Command.Group
          className="max-h-80 overflow-y-auto"
          onScroll={handleScroll}
        >
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
          {isFetchingNextPage ? (
            <Command.Loading className="p-2">
              <Text className="flex items-center gap-2" color="muted">
                <LoadingIcon className="animate-spin" />
                Loading more labels...
              </Text>
            </Command.Loading>
          ) : null}
        </Command.Group>
      </Command>
    </Popover.Content>
  );
};

LabelsMenu.Trigger = Trigger;
LabelsMenu.Items = Items;

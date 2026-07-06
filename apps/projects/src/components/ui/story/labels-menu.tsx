"use client";
import {
  createContext,
  use,
  useDeferredValue,
  useRef,
  useState,
  type ReactNode,
  type UIEvent,
} from "react";
import { Command, Divider, Flex, Popover, Text } from "ui";
import { CheckIcon, PlusIcon } from "icons";
import { generateRandomColor } from "lib";
import { LABEL_MENU_PAGE_SIZE, useLabelsInfinite } from "@/lib/hooks/labels";
import { useCreateLabelMutation } from "@/lib/hooks/create-label-mutation";
import { Dot } from "../dot";
import { MenuLoadingSkeleton } from "../menu-loading-skeleton";
import { canCreateLabelFromQuery } from "./labels-menu-utils";

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
  setLabelIds: (labelIds: string[]) => Promise<void> | void;
  align?: "center" | "start" | "end" | undefined;
}) => {
  const { mutateAsync: createLabel } = useCreateLabelMutation();
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const [pendingLabelIds, setPendingLabelIds] = useState<string[] | null>(null);
  const selectionRequestId = useRef(0);
  const [isLoading, setIsLoading] = useState(false);
  const { open } = useLabelsMenu();
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetching,
    isFetchingNextPage,
    isPending,
  } = useLabelsInfinite(
    { search: deferredQuery, teamId },
    LABEL_MENU_PAGE_SIZE,
    open,
  );
  const labels = data?.pages.flatMap((page) => page.labels) ?? [];
  const selectedLabelIds = pendingLabelIds ?? labelIds;
  const isLoadingLabels =
    (isPending || isFetching || query !== deferredQuery) && !isFetchingNextPage;
  const canCreateLabel =
    !isLoadingLabels && !isLoading && canCreateLabelFromQuery(query, labels);

  const updateSelectedLabels = (nextLabelIds: string[]) => {
    const requestId = selectionRequestId.current + 1;
    selectionRequestId.current = requestId;
    setPendingLabelIds(nextLabelIds);

    void Promise.resolve(setLabelIds(nextLabelIds)).then(
      () => {
        if (selectionRequestId.current === requestId) {
          setPendingLabelIds(null);
        }
      },
      () => {
        if (selectionRequestId.current === requestId) {
          setPendingLabelIds(null);
        }
      },
    );
  };

  const handleCreateLabel = async () => {
    const name = query.trim();
    if (!name || isLoading) {
      return;
    }

    const usedColors = labels.map((label) => label.color);
    const color = generateRandomColor({ exclude: usedColors });
    setIsLoading(true);
    await createLabel(
      { name, color, teamId },
      {
        onSuccess(res) {
          if (res.data) {
            updateSelectedLabels([...selectedLabelIds, res.data.id]);
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
      <Command shouldFilter={false}>
        <Command.Input
          autoFocus
          onKeyDown={(event) => {
            if (event.key === "Enter" && canCreateLabel) {
              event.preventDefault();
              void handleCreateLabel();
            }
          }}
          onValueChange={(value) => {
            setQuery(value);
          }}
          placeholder="Update labels..."
          value={query}
        />
        <Divider className="my-2" />
        {!isLoadingLabels && labels.length === 0 && !canCreateLabel ? (
          <Command.Empty className="py-2">
            <Text color="muted">No labels found.</Text>
          </Command.Empty>
        ) : null}
        <Command.Group
          className="max-h-80 overflow-y-auto"
          onScroll={handleScroll}
        >
          {canCreateLabel ? (
            <Command.Item
              className="justify-between gap-4"
              onSelect={() => {
                void handleCreateLabel();
              }}
              value={`create-label-${query.trim()}`}
            >
              <Flex align="center" gap={2}>
                <PlusIcon className="h-4" strokeWidth={2.7} />
                <Text className="mr-4 flex items-center gap-2 font-medium">
                  Create label
                  <span className="text-text-muted font-medium">
                    &ldquo;{query.trim()}&rdquo;
                  </span>
                </Text>
              </Flex>
            </Command.Item>
          ) : null}
          {isLoading ? (
            <Command.Item
              className="justify-between gap-4 opacity-70"
              disabled
              value="creating-label"
            >
              <Flex align="center" gap={2}>
                <PlusIcon className="h-4" strokeWidth={2.7} />
                <Text className="mr-4 flex items-center gap-2 font-medium">
                  Creating label...
                </Text>
              </Flex>
            </Command.Item>
          ) : null}
          {isLoadingLabels ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={5} />
            </Command.Loading>
          ) : null}
          {labels.map(({ id, name, color }) => (
            <Command.Item
              className="justify-between gap-4"
              key={id}
              onSelect={() => {
                if (!selectedLabelIds.includes(id)) {
                  updateSelectedLabels([...selectedLabelIds, id]);
                } else {
                  updateSelectedLabels(
                    selectedLabelIds.filter((labelId) => labelId !== id),
                  );
                }
              }}
              value={name}
            >
              <Flex align="center" gap={2}>
                <Dot color={color} />
                <Text className="mr-4 flex items-center gap-3">{name}</Text>
              </Flex>
              <Flex align="center" gap={1}>
                {selectedLabelIds.includes(id) ? (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                ) : null}
              </Flex>
            </Command.Item>
          ))}
          {isFetchingNextPage ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={2} />
            </Command.Loading>
          ) : null}
        </Command.Group>
      </Command>
    </Popover.Content>
  );
};

LabelsMenu.Trigger = Trigger;
LabelsMenu.Items = Items;

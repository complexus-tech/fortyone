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
import { Button, Command, Divider, Flex, Popover, Text } from "ui";
import { CheckIcon, PlusIcon } from "icons";
import { generateRandomColor } from "lib";
import { LABEL_MENU_PAGE_SIZE, useLabelsInfinite } from "@/lib/hooks/labels";
import { useCreateLabelMutation } from "@/lib/hooks/create-label-mutation";
import { Dot } from "../dot";
import { MenuLoadingSkeleton } from "../menu-loading-skeleton";

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
    (isPending || isFetching || query !== deferredQuery) &&
    !isFetchingNextPage;

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
    if (!name) {
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
          onValueChange={(value) => {
            setQuery(value);
          }}
          placeholder="Update labels..."
          value={query}
        />
        <Divider className="my-2" />
        {!isLoadingLabels ? (
          <Command.Empty className="justify-center px-1 py-0">
            <Button
              className="mx-0 border-0 text-base font-medium"
              color="tertiary"
              disabled={query.trim().length === 0}
              fullWidth
              loading={isLoading}
              loadingText="Creating label..."
              onClick={handleCreateLabel}
            >
              <PlusIcon className="h-4" strokeWidth={2.7} /> Create new label:{" "}
              <span className="text-text-muted font-medium">
                &ldquo;{query.trim()}&rdquo;
              </span>
            </Button>
          </Command.Empty>
        ) : null}
        <Command.Group
          className="max-h-80 overflow-y-auto"
          onScroll={handleScroll}
        >
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

import { useSession } from "next-auth/react";
import { forwardRef, useEffect, useImperativeHandle, useState } from "react";
import { Avatar, Box, Text } from "ui";

export type MentionItem = {
  id: string;
  label: string;
  username: string;
  avatar?: string | null;
};

type MentionListProps = {
  items: MentionItem[];
  command: (item: MentionItem) => void;
};

export type MentionListRef = {
  onKeyDown: (event: KeyboardEvent) => boolean;
};

export const MentionList = forwardRef<MentionListRef, MentionListProps>(
  (props, ref) => {
    const [selectedIndex, setSelectedIndex] = useState(0);
    const { data: session } = useSession();

    const selectItem = (index: number) => {
      const item = props.items[index];
      props.command(item);
    };

    const upHandler = () => {
      setSelectedIndex(
        (selectedIndex + props.items.length - 1) % props.items.length,
      );
    };

    const downHandler = () => {
      setSelectedIndex((selectedIndex + 1) % props.items.length);
    };

    const enterHandler = () => {
      selectItem(selectedIndex);
    };

    useEffect(() => {
      setSelectedIndex(0);
    }, [props.items]);

    useImperativeHandle(ref, () => ({
      onKeyDown: ({ key }: KeyboardEvent) => {
        if (key === "ArrowUp") {
          upHandler();
          return true;
        }

        if (key === "ArrowDown") {
          downHandler();
          return true;
        }

        if (key === "Enter") {
          enterHandler();
          return true;
        }

        return false;
      },
    }));

    if (props.items.length === 0) {
      return (
        <Box className="border-border bg-surface-elevated w-56 rounded-xl border p-2">
          <Text color="muted">No users found</Text>
        </Box>
      );
    }

    return (
      <Box className="border-border bg-surface-elevated dark:shadow-dark/20 pointer-events-auto z-50 w-max min-w-64 space-y-1 rounded-xl border p-2">
        {props.items.map((item, index) => (
          <button
            className="hover:bg-state-hover focus:bg-state-active flex w-full cursor-pointer items-center gap-2 rounded-[0.6rem] px-2 py-1 outline-none select-none"
            key={item.id}
            onClick={() => {
              selectItem(index);
            }}
            onMouseEnter={() => {
              setSelectedIndex(index);
            }}
            type="button"
          >
            <Avatar name={item.label} size="sm" src={item.avatar} />
            <Text className="max-w-48 truncate">
              {item.label}
              {item.id === session?.user?.id && (
                <Text as="span" color="muted">
                  (You)
                </Text>
              )}
            </Text>
          </button>
        ))}
      </Box>
    );
  },
);

MentionList.displayName = "MentionList";

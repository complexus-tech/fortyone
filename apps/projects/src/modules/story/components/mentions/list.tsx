import { forwardRef, useEffect, useImperativeHandle, useState } from "react";
import { Avatar, Box, Text } from "ui";

export type MentionItem = {
  id: string;
  label: string;
  username: string;
  avatar?: string;
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
        <Box className="w-56 rounded-xl border border-gray-200 bg-white px-3 py-2.5 text-[0.95rem] shadow-lg dark:border-dark-100 dark:bg-dark-200">
          <Text color="muted">No users found</Text>
        </Box>
      );
    }

    return (
      <Box className="z-50 mt-1 w-max rounded-lg border border-gray-50 bg-white py-1.5 shadow shadow-gray-100 backdrop-blur dark:border-dark-50 dark:bg-dark-200 dark:shadow-dark/20">
        {props.items.map((item, index) => (
          <button
            className="flex w-full cursor-pointer select-none items-center justify-between gap-2 rounded-lg px-2 py-1.5 outline-none hover:bg-gray-100/70 focus:bg-gray-50 hover:dark:bg-dark-50 focus:dark:bg-dark-100/70"
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
            <Text className="max-w-[12rem] truncate">
              {item.label}{" "}
              {/* <Text as="span" color="muted">
                (You)
              </Text> */}
            </Text>
          </button>
        ))}
      </Box>
    );
  },
);

MentionList.displayName = "MentionList";

import { forwardRef, useEffect, useImperativeHandle, useState } from "react";

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
        <div className="rounded-lg border border-gray-200 bg-white p-3 shadow-lg dark:border-dark-200 dark:bg-dark-100">
          <div className="px-3 py-2 text-sm text-gray-250 dark:text-gray-300">
            No users found
          </div>
        </div>
      );
    }

    return (
      <div className="max-h-48 overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-dark-200 dark:bg-dark-100">
        {props.items.map((item, index) => (
          <button
            className={`flex w-full items-center gap-3 px-4 py-3 text-left text-sm transition-all duration-200 hover:bg-info/5 dark:hover:bg-info/10 ${
              index === selectedIndex
                ? "border-l-2 border-info bg-info/10 dark:bg-info/15"
                : "border-l-2 border-transparent"
            }`}
            key={item.id}
            onClick={() => {
              selectItem(index);
            }}
            onMouseEnter={() => {
              setSelectedIndex(index);
            }}
          >
            {item.avatar ? (
              <img
                alt={item.label}
                className="h-8 w-8 rounded-full ring-2 ring-gray-200 dark:ring-dark-200"
                src={item.avatar}
              />
            ) : (
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-info/20 text-sm font-medium text-info dark:bg-info/30 dark:text-info">
                {item.label.charAt(0).toUpperCase()}
              </div>
            )}
            <div className="flex flex-col">
              <span className="text-gray-800 font-medium dark:text-gray-100">
                {item.label}
              </span>
              <span className="text-xs text-gray-250 dark:text-gray-300">
                @{item.username}
              </span>
            </div>
          </button>
        ))}
      </div>
    );
  },
);

MentionList.displayName = "MentionList";

import { Button, Menu } from "ui";
import type { Editor } from "@tiptap/react";
import { cn } from "lib";
import { ArrowDownIcon, CheckIcon } from "icons";

const headings = [
  {
    name: "Paragraph",
    level: 0,
    classes: "",
  },
  {
    name: "Heading 1",
    level: 1,
    classes: "text-4xl",
  },
  {
    name: "Heading 2",
    level: 2,
    classes: "text-3xl",
  },
  {
    name: "Heading 3",
    level: 3,
    classes: "text-2xl",
  },
  {
    name: "Heading 4",
    level: 4,
    classes: "text-xl",
  },
  {
    name: "Heading 5",
    level: 5,
    classes: "text-lg",
  },
  {
    name: "Heading 6",
    level: 6,
    classes: "text-base font-medium text-gray-250 dark:text-gray-200",
  },
];

export const ToggleNode = ({ editor }: { editor: Editor | null }) => {
  return (
    <Menu>
      <Menu.Button>
        <Button
          className="font-medium"
          color="tertiary"
          rightIcon={<ArrowDownIcon className="h-4 w-auto" />}
          size="md"
          variant="naked"
        >
          {headings.find(({ level }) => editor?.isActive("heading", { level }))
            ?.name || "Paragraph"}
        </Button>
      </Menu.Button>
      <Menu.Items align="center" className="w-64">
        <Menu.Group>
          {headings.map(({ name, level, classes }) => {
            const isActive =
              editor?.isActive("heading", { level }) ||
              (level === 0 && !editor?.isActive("heading"));
            return (
              <Menu.Item
                active={isActive}
                className={cn("justify-between px-3", classes)}
                key={name}
                onClick={() => {
                  if (level === 0) {
                    editor?.chain().focus().setParagraph().run();
                  } else {
                    editor?.chain().focus().setNode("heading", { level }).run();
                  }
                }}
              >
                {name}
                {isActive ? (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                ) : null}
              </Menu.Item>
            );
          })}
        </Menu.Group>
      </Menu.Items>
    </Menu>
  );
};

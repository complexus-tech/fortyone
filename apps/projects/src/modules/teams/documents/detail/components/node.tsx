import { Button, Menu } from "ui";
import type { Editor } from "@tiptap/react";
import { ArrowDownIcon, CheckIcon } from "icons";

const headings = [
  {
    name: "Paragraph",
    level: 0,
  },
  {
    name: "Heading 1",
    level: 1,
  },
  {
    name: "Heading 2",
    level: 2,
  },
  {
    name: "Heading 3",
    level: 3,
  },
  {
    name: "Heading 4",
    level: 4,
  },
  {
    name: "Heading 5",
    level: 5,
  },
  {
    name: "Heading 6",
    level: 6,
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
          {headings.map(({ name, level }) => {
            const isActive =
              editor?.isActive("heading", { level }) ||
              (level === 0 && !editor?.isActive("heading"));
            return (
              <Menu.Item
                active={isActive}
                className="justify-between px-3"
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

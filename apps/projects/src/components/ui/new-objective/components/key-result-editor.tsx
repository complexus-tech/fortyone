import { useEditor } from "@tiptap/react";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import TextExt from "@tiptap/extension-text";
import Placeholder from "@tiptap/extension-placeholder";
import { Box, Button, Flex, Input, Select, Text, TextEditor } from "ui";
import { CloseIcon } from "icons";
import { toast } from "sonner";
import type { KeyResult, MeasureType } from "../types";

type KeyResultEditorProps = {
  keyResult: KeyResult;
  onUpdate: (id: string, updates: Partial<KeyResult>) => void;
  onCancel: () => void;
  onSave: () => void;
};

export const KeyResultEditor = ({
  keyResult,
  onUpdate,
  onCancel,
  onSave,
}: KeyResultEditorProps) => {
  const editor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExt,
      Placeholder.configure({
        placeholder: "Example: Increase user adoption from 100 to 150",
      }),
    ],
    content: keyResult.name,
    editable: true,
    onUpdate: ({ editor }) => {
      onUpdate(keyResult.id, { name: editor.getText() });
    },
  });

  const handleSave = () => {
    if (editor?.isEmpty) {
      toast.warning("Validation error", {
        description: "Please enter a name for the key result",
      });
      return;
    }
    if (
      keyResult.measureType !== "Boolean (Complete/Incomplete)" &&
      !keyResult.targetValue
    ) {
      toast.warning("Validation error", {
        description: "Please enter a target value for the key result",
      });
      return;
    }
    onSave();
  };

  return (
    <Box className="mb-6 rounded-lg border border-gray-200 px-4 pb-3.5 dark:border-dark-100">
      <Flex align="center" className="relative -top-2" justify="between">
        <TextEditor className="prose-base" editor={editor} />
        <Button
          asIcon
          color="tertiary"
          leftIcon={<CloseIcon />}
          onClick={onCancel}
          size="sm"
          variant="naked"
        >
          <span className="sr-only">Cancel</span>
        </Button>
      </Flex>
      <Box className="relative -top-4">
        <Box className="mb-4">
          <Text className="mb-1.5 font-medium">Measure as</Text>
          <Select
            defaultValue={keyResult.measureType}
            onValueChange={(value) => {
              const measureType = value as MeasureType;
              onUpdate(keyResult.id, {
                measureType,
                startValue:
                  measureType === "Boolean (Complete/Incomplete)"
                    ? 0
                    : undefined,
                targetValue:
                  measureType === "Boolean (Complete/Incomplete)"
                    ? 1
                    : undefined,
                currentValue:
                  measureType === "Boolean (Complete/Incomplete)"
                    ? 0
                    : undefined,
              });
            }}
          >
            <Select.Trigger className="h-[2.7rem] text-base">
              <Select.Input />
            </Select.Trigger>
            <Select.Content>
              <Select.Group>
                {["Number", "Percent (%)", "Boolean (Complete/Incomplete)"].map(
                  (option) => (
                    <Select.Option key={option} value={option}>
                      {option}
                    </Select.Option>
                  ),
                )}
              </Select.Group>
            </Select.Content>
          </Select>
        </Box>

        {keyResult.measureType === "Boolean (Complete/Incomplete)" ? (
          <Flex className="mb-4" gap={2}>
            <Button
              color={keyResult.currentValue === 0 ? "primary" : "tertiary"}
              onClick={() => {
                onUpdate(keyResult.id, { currentValue: 0 });
              }}
              variant={keyResult.currentValue === 0 ? "solid" : "outline"}
            >
              Incomplete
            </Button>
            <Button
              color={keyResult.currentValue === 1 ? "primary" : "tertiary"}
              onClick={() => {
                onUpdate(keyResult.id, { currentValue: 1 });
              }}
              variant={keyResult.currentValue === 1 ? "solid" : "outline"}
            >
              Complete
            </Button>
          </Flex>
        ) : (
          <Box className="grid grid-cols-2 gap-4">
            <Input
              autoFocus
              className="h-[2.7rem] bg-gray-50/30 dark:bg-dark-100/40"
              label="Starting Value"
              onChange={(e) => {
                onUpdate(keyResult.id, { startValue: Number(e.target.value) });
              }}
              placeholder="0"
              required
              type="number"
              value={keyResult.startValue?.toString() || ""}
            />
            <Input
              className="h-[2.7rem] bg-gray-50/30 dark:bg-dark-100/40"
              label="Target Value"
              onChange={(e) => {
                onUpdate(keyResult.id, { targetValue: Number(e.target.value) });
              }}
              placeholder="0"
              required
              type="number"
              value={keyResult.targetValue?.toString() || ""}
            />
          </Box>
        )}
      </Box>
      <Button
        className="relative -top-0.5"
        color="tertiary"
        onClick={handleSave}
      >
        Add Key Result
      </Button>
    </Box>
  );
};

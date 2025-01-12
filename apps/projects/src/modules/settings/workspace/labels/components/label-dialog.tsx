import { Box, Flex, Text, Button, Input, Dialog, ColorPicker } from "ui";
import { useState } from "react";
import type { Label } from "@/types";

type LabelFormData = {
  name: string;
  color: string;
};

type LabelDialogProps = {
  isOpen: boolean;
  onClose: () => void;
  mode: "create" | "edit";
  selectedLabel?: Label | null;
  onSubmit: (data: LabelFormData) => void;
};

export const LabelDialog = ({
  isOpen,
  onClose,
  mode,
  selectedLabel,
  onSubmit,
}: LabelDialogProps) => {
  const isEdit = mode === "edit";
  const [name, setName] = useState(
    isEdit && selectedLabel ? selectedLabel.name : "",
  );
  const [color, setColor] = useState(
    isEdit && selectedLabel ? selectedLabel.color : "#F8F9FA",
  );

  return (
    <Dialog onOpenChange={onClose} open={isOpen}>
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title className="px-6">
            {isEdit ? "Edit Label" : "Create Label"}
          </Dialog.Title>
          <Dialog.Description>
            {isEdit
              ? "Edit the label name and color."
              : "Create a new label to categorize stories."}
          </Dialog.Description>
        </Dialog.Header>

        <Dialog.Body>
          <Flex direction="column" gap={6}>
            <Input
              label="Label name"
              name="name"
              onChange={(e) => {
                setName(e.target.value);
              }}
              placeholder="Enter label name"
              required
              value={name}
            />

            <Box>
              <Text as="label" className="mb-2 block font-medium">
                Color
              </Text>
              <ColorPicker onChange={setColor} value={color} />
            </Box>
          </Flex>
        </Dialog.Body>

        <Dialog.Footer>
          <Button color="tertiary" onClick={onClose} variant="outline">
            Cancel
          </Button>
          <Button
            onClick={() => {
              onSubmit({ name, color });
              onClose();
            }}
          >
            {isEdit ? "Save Changes" : "Create Label"}
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

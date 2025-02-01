import { useParams } from "next/navigation";
import React, { useState } from "react";
import type { ButtonProps } from "ui";
import { Button, Dialog, Input } from "ui";
import { useCreateKeyResultMutation } from "../../hooks";

export const NewKeyResultButton = ({
  color = "tertiary",
  ...rest
}: ButtonProps) => {
  const { objectiveId } = useParams<{ objectiveId: string }>();
  const keyResultMutation = useCreateKeyResultMutation();
  const [isOpen, setIsOpen] = useState(false);

  const handleCreate = () => {
    keyResultMutation.mutate({
      objectiveId,
      name: "",
      startValue: 0,
      targetValue: 0,
      measurementType: "number",
    });
  };

  return (
    <>
      <Button
        color={color}
        onClick={() => {
          setIsOpen(true);
        }}
        {...rest}
      >
        Add Key Result
      </Button>
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content>
          <Dialog.Header className="px-6">
            <Dialog.Title className="text-lg">Create Key Result</Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Input label="Name" placeholder="Enter a name for the key result" />
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-2">
            <Button
              className="px-6"
              color="tertiary"
              onClick={() => {
                setIsOpen(false);
              }}
            >
              Cancel
            </Button>
            <Button onClick={handleCreate}>Create Key Result</Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </>
  );
};

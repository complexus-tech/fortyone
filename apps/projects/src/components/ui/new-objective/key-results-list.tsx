import { Button, Flex, Text } from "ui";
import { DeleteIcon, EditIcon, OKRIcon } from "icons";
import type { NewKeyResult } from "@/modules/objectives/types";
import { RowWrapper } from "../row-wrapper";

type KeyResultsListProps = {
  keyResults: NewKeyResult[];
  onEdit: (index: number) => void;
  onRemove: (index: number) => void;
};

export const KeyResultsList = ({
  keyResults,
  onEdit,
  onRemove,
}: KeyResultsListProps) => {
  return (
    <>
      {keyResults.map((kr, index) => (
        <RowWrapper className="px-0 py-1.5" key={kr.name}>
          <Flex align="center" gap={2}>
            <OKRIcon />
            <Text>{kr.name}</Text>
          </Flex>
          <Flex className="gap-2">
            <Button
              asIcon
              color="tertiary"
              onClick={() => {
                onEdit(index);
              }}
              size="sm"
              variant="naked"
            >
              <EditIcon />
            </Button>
            <Button
              asIcon
              color="tertiary"
              onClick={() => {
                onRemove(index);
              }}
              size="sm"
              variant="naked"
            >
              <DeleteIcon />
            </Button>
          </Flex>
        </RowWrapper>
      ))}
    </>
  );
};

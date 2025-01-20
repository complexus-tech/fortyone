import { Button, Flex, Text } from "ui";
import { DeleteIcon, EditIcon, OKRIcon } from "icons";
import type { KeyResult } from "../types";
import { RowWrapper } from "../../row-wrapper";

type KeyResultsListProps = {
  keyResults: KeyResult[];
  onEdit: (id: string) => void;
  onRemove: (id: string) => void;
};

export const KeyResultsList = ({
  keyResults,
  onEdit,
  onRemove,
}: KeyResultsListProps) => {
  return (
    <>
      {keyResults.map((kr) => (
        <RowWrapper className="px-0 py-1.5" key={kr.id}>
          <Flex align="center" gap={2}>
            <OKRIcon />
            <Text>{kr.name}</Text>
          </Flex>
          <Flex className="gap-2">
            <Button
              asIcon
              color="tertiary"
              onClick={() => {
                onEdit(kr.id);
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
                onRemove(kr.id);
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

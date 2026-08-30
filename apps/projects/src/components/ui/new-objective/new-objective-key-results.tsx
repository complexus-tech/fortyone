import type { ComponentProps } from "react";
import { PlusIcon } from "icons";
import { Box, Button, Text } from "ui";
import { KeyResultEditor } from "./key-result-editor";
import { KeyResultsList } from "./key-results-list";

type KeyResult = NonNullable<
  ComponentProps<typeof KeyResultEditor>["keyResult"]
>;
type KeyResultUpdates = Parameters<
  ComponentProps<typeof KeyResultEditor>["onUpdate"]
>[1];

export const NewObjectiveKeyResults = ({
  editingKeyResult,
  existingKeyResults,
  isEditing,
  keyResultTerm,
  keyResults,
  keyResultsTerm,
  objectiveEndDate,
  objectiveName,
  objectiveStartDate,
  onAdd,
  onCancel,
  onEdit,
  onRemove,
  onSave,
  onUpdate,
}: {
  editingKeyResult: KeyResult | null;
  existingKeyResults: KeyResult[];
  isEditing: boolean;
  keyResultTerm: string;
  keyResults: KeyResult[];
  keyResultsTerm: string;
  objectiveEndDate: string | null;
  objectiveName: string;
  objectiveStartDate: string | null;
  onAdd: () => void;
  onCancel: () => void;
  onEdit: (index: number) => void;
  onRemove: (index: number) => void;
  onSave: () => void;
  onUpdate: (updates: KeyResultUpdates) => void;
}) => (
  <Box className="mt-3">
    <Text className="mb-2 text-lg font-semibold antialiased">
      {keyResultsTerm}
    </Text>
    {!isEditing ? (
      <>
        <KeyResultsList
          keyResults={keyResults}
          onEdit={onEdit}
          onRemove={onRemove}
        />
        <Button
          color="tertiary"
          leftIcon={<PlusIcon />}
          onClick={onAdd}
          variant="outline"
        >
          Add {keyResultTerm}
        </Button>
      </>
    ) : (
      <KeyResultEditor
        keyResult={editingKeyResult}
        onCancel={onCancel}
        onSave={onSave}
        onUpdate={(_index, updates) => {
          onUpdate(updates);
        }}
        qualityContext={{
          existingKeyResults,
          objectiveEndDate,
          objectiveName,
          objectiveStartDate,
        }}
      />
    )}
  </Box>
);

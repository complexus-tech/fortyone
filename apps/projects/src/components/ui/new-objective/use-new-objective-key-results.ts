import { useState } from "react";
import type { ComponentProps } from "react";
import type { KeyResultEditor } from "./key-result-editor";

type KeyResult = NonNullable<
  ComponentProps<typeof KeyResultEditor>["keyResult"]
>;
type KeyResultUpdate = Parameters<
  ComponentProps<typeof KeyResultEditor>["onUpdate"]
>[1];
type KeyResultFormMode = "add" | "edit" | null;

export const useNewObjectiveKeyResults = ({
  keyResults,
  onKeyResultsChange,
}: {
  keyResults: KeyResult[];
  onKeyResultsChange: (
    updateKeyResults: (keyResults: KeyResult[]) => KeyResult[],
  ) => void;
}) => {
  const [keyResultMode, setKeyResultMode] = useState<KeyResultFormMode>(null);
  const [editingKeyResult, setEditingKeyResult] = useState<KeyResult | null>(
    null,
  );
  const [editingIndex, setEditingIndex] = useState<number | null>(null);

  const resetKeyResultEditor = () => {
    setKeyResultMode(null);
    setEditingKeyResult(null);
    setEditingIndex(null);
  };

  const handleAddKeyResult = () => {
    setEditingKeyResult({
      name: "",
      measurementType: "number",
      currentValue: 0,
      startValue: 0,
      targetValue: 0,
      startDate: "",
      endDate: "",
    });
    setEditingIndex(null);
    setKeyResultMode("add");
  };

  const handleEditKeyResult = (index: number) => {
    const keyResult = keyResults.at(index);
    if (!keyResult) return;

    setEditingKeyResult(keyResult);
    setEditingIndex(index);
    setKeyResultMode("edit");
  };

  const handleRemoveKeyResult = (index: number) => {
    onKeyResultsChange((currentKeyResults) =>
      currentKeyResults.filter((_, itemIndex) => itemIndex !== index),
    );
  };

  const handleSaveKeyResult = () => {
    if (keyResultMode === "add") {
      onKeyResultsChange((currentKeyResults) => [
        ...currentKeyResults,
        {
          ...editingKeyResult!,
          currentValue: editingKeyResult!.startValue,
        },
      ]);
    } else if (editingIndex !== null) {
      onKeyResultsChange((currentKeyResults) =>
        currentKeyResults.map((keyResult, index) =>
          index === editingIndex ? editingKeyResult! : keyResult,
        ),
      );
    }
    resetKeyResultEditor();
  };

  const handleKeyResultUpdate = (updates: KeyResultUpdate) => {
    setEditingKeyResult((current) =>
      current ? { ...current, ...updates } : null,
    );
  };

  return {
    editingIndex,
    editingKeyResult,
    handleAddKeyResult,
    handleEditKeyResult,
    handleKeyResultUpdate,
    handleRemoveKeyResult,
    handleSaveKeyResult,
    isEditingKeyResult: keyResultMode !== null,
    resetKeyResultEditor,
  };
};

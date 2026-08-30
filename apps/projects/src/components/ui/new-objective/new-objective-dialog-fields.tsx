import type { Editor } from "@tiptap/core";
import type { ReactNode } from "react";
import { cn } from "lib";
import { Divider, TextArea, TextEditor } from "ui";

export const NewObjectiveDialogFields = ({
  children,
  descriptionEditor,
  isExpanded,
  onShortSummaryChange,
  shortSummary,
  titleEditor,
}: {
  children?: ReactNode;
  descriptionEditor: Editor | null;
  isExpanded: boolean;
  onShortSummaryChange: (shortSummary: string) => void;
  shortSummary?: string;
  titleEditor: Editor | null;
}) => (
  <>
    <TextEditor asTitle className="text-2xl font-medium" editor={titleEditor} />
    <TextArea
      aria-label="Objective short summary"
      className="mt-3 min-h-14 resize-none border-0 bg-transparent px-0 py-1.5 text-[1.125rem] leading-6 shadow-none focus-visible:ring-0 dark:bg-transparent"
      maxLength={500}
      onChange={(event) => {
        onShortSummaryChange(event.target.value);
      }}
      placeholder="Add short summary..."
      rows={2}
      value={shortSummary}
    />
    {children}
    <Divider className="my-3 opacity-60" />
    <TextEditor
      className={cn("min-h-20", {
        "min-h-96": isExpanded,
      })}
      editor={descriptionEditor}
    />
  </>
);

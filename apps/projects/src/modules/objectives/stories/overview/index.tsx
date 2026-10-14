import { Box, Container, Divider, TextEditor } from "ui";
import { useParams } from "next/navigation";
import { useEditor } from "@tiptap/react";
import Underline from "@tiptap/extension-underline";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import { TaskItem, TaskList } from "@tiptap/extension-list";
import { useState } from "react";
import { BoardDividedPanel } from "@/components/ui";
import { ExpandableDescription } from "@/components/shared/expandable-description";
import { useMediaQuery, useTerminology } from "@/hooks";
import { useDebouncedCallback } from "@/hooks/debounce";
import { createRichTextStarterKit } from "@/lib/tiptap/starter-kit";
import { RelatedDocuments } from "@/modules/documents/related-documents";
import { GoogleDriveFileSection } from "@/modules/google-drive/public/files";
import { ObjectiveForecastRiskBanner } from "@/modules/objectives/components/objective-forecast-risk";
import {
  useCanUpdateObjective,
  useObjective,
  useUpdateObjectiveMutation,
} from "../../hooks";
import { Sidebar } from "../sidebar";
import type { ObjectiveUpdate } from "../../types";
import { Activity } from "./activity";
import { Properties } from "./properties";
import { ShortSummary } from "./short-summary";

const DEBOUNCE_DELAY = 1000;

export const Overview = () => {
  const { objectiveId } = useParams<{ objectiveId: string }>();
  const { getTermDisplay } = useTerminology();
  const { data: objective } = useObjective(objectiveId);
  const canUpdate = useCanUpdateObjective();
  const isDesktop = useMediaQuery("(min-width: 768px)");
  const updateMutation = useUpdateObjectiveMutation();
  const [summaryDraft, setSummaryDraft] = useState<string | null>(null);

  const handleUpdate = (data: ObjectiveUpdate) => {
    updateMutation.mutate({ objectiveId, data });
  };

  // Each field keeps its own pending save when focus moves between editors or tabs.
  const { callback: updateName, flush: flushName } = useDebouncedCallback(
    handleUpdate,
    DEBOUNCE_DELAY,
    { flushOnUnmount: true },
  );
  const { callback: updateSummary, flush: flushSummary } = useDebouncedCallback(
    handleUpdate,
    DEBOUNCE_DELAY,
    { flushOnUnmount: true },
  );
  const { callback: updateDescription, flush: flushDescription } =
    useDebouncedCallback(handleUpdate, DEBOUNCE_DELAY, {
      flushOnUnmount: true,
    });

  const nameEditor = useEditor({
    extensions: [
      createRichTextStarterKit(),
      Underline,
      TaskList,
      TaskItem.configure({ nested: true }),
      Link.configure({ autolink: true }),
      Placeholder.configure({
        placeholder: `${getTermDisplay("objectiveTerm", { capitalize: true })} name...`,
      }),
    ],
    content: objective?.name || "",
    editable: canUpdate,
    immediatelyRender: false,
    onUpdate: ({ editor }) => {
      updateName({ name: editor.getText() });
    },
    onBlur: flushName,
  });

  const descriptionEditor = useEditor({
    extensions: [
      createRichTextStarterKit(),
      Underline,
      Link.configure({ autolink: true }),
      Placeholder.configure({
        placeholder: `${getTermDisplay("objectiveTerm", { capitalize: true })} description...`,
      }),
    ],
    content: objective?.description || "",
    editable: canUpdate,
    immediatelyRender: false,
    onUpdate: ({ editor }) => {
      updateDescription({ description: editor.getHTML() });
    },
    onBlur: flushDescription,
  });

  const content = (
    <Container className="h-full min-h-0 overflow-y-auto pt-6">
      <Box>
        <TextEditor
          asTitle
          className="mb-0 text-2xl font-medium antialiased md:text-4xl md:font-semibold"
          editor={nameEditor}
        />
        {objective ? (
          <ObjectiveForecastRiskBanner className="mt-4" objective={objective} />
        ) : null}
        <ShortSummary
          onBlur={flushSummary}
          onChange={(shortSummary) => {
            setSummaryDraft(shortSummary);
            updateSummary({ shortSummary });
          }}
          readOnly={!canUpdate}
          value={summaryDraft ?? objective?.shortSummary ?? ""}
        />
        <Divider className="my-4 opacity-60" />
        <ExpandableDescription key={objectiveId}>
          <TextEditor className="text-foreground" editor={descriptionEditor} />
        </ExpandableDescription>
      </Box>
      <Properties />
      <RelatedDocuments entityId={objectiveId} entityType="objective" />
      <GoogleDriveFileSection
        canEdit={canUpdate}
        suggestedTitle={objective?.name}
        target={{ id: objectiveId, type: "objective" }}
      />
      <Divider className="my-6 md:my-8" />
      <Activity viewport={isDesktop ? "desktop" : "mobile"} />
    </Container>
  );

  return (
    <Box className="h-full min-h-0 overflow-hidden">
      {isDesktop ? (
        <BoardDividedPanel autoSaveId="teams:objectives:stories:divided-panel">
          <BoardDividedPanel.MainPanel>{content}</BoardDividedPanel.MainPanel>
          <BoardDividedPanel.SideBar className="h-full min-h-0" isExpanded>
            <Sidebar className="h-full min-h-0 overflow-y-auto" />
          </BoardDividedPanel.SideBar>
        </BoardDividedPanel>
      ) : (
        content
      )}
    </Box>
  );
};

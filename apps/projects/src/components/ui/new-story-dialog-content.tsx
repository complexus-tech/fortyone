import type { ComponentProps, RefObject, ReactNode } from "react";
import type { Editor } from "@tiptap/core";
import { PlusIcon } from "icons";
import { cn } from "lib";
import { Button, Dialog, Switch, Text, TextEditor } from "ui";
import { RICH_TEXT_MEDIA_ACCEPT } from "@/lib/tiptap/rich-text-media";
import { RichTextTableMenu } from "@/lib/tiptap/rich-text-table-menu";
import { NewStoryFigmaSource } from "./new-story-figma-source";
import { NewStoryDialogFields } from "./new-story-dialog-fields";
import { insertFigmaDescription } from "./new-story-dialog-figma";
import { NewStoryDialogHeader } from "./new-story-dialog-header";
import { NewStoryDialogSimilarStories } from "./new-story-dialog-similar-stories";

type DialogFieldsProps = Omit<
  ComponentProps<typeof NewStoryDialogFields>,
  "children"
>;
type FigmaSourceProps = ComponentProps<typeof NewStoryFigmaSource>;
type HeaderTeam = ComponentProps<typeof NewStoryDialogHeader>["teams"][number];
type SimilarStoriesProps = ComponentProps<typeof NewStoryDialogSimilarStories>;

export const NewStoryDialogContent = <TTeam extends HeaderTeam>({
  activeTeamId,
  canUseBackgroundMaya,
  createMore,
  currentTeam,
  currentTeamId,
  customFields,
  deadlineSourceRef,
  descriptionEditor,
  dispatch,
  estimateScheme,
  figmaArtifacts,
  isCreating,
  isExpanded,
  isMayaAssigned,
  isMayaAssigneeLoading,
  isOpen,
  mayaAssigneeId,
  mediaInputRef,
  member,
  members,
  objectiveTerm,
  onActiveTeamChange,
  onCreate,
  onCreateMoreChange,
  onFigmaArtifactsChange,
  onMediaFiles,
  onOpenChange,
  onSimilarStorySelect,
  onToggleExpanded,
  selectedLabels,
  showObjectives,
  showSprints,
  similarStories,
  sprintName,
  sprintTerm,
  statuses,
  storyForm,
  storyTerm,
  storyTermPlural,
  strategyLinkLabel,
  teamStatuses,
  teams,
  titleEditor,
  templatePicker,
}: {
  activeTeamId?: string;
  canUseBackgroundMaya: DialogFieldsProps["canUseBackgroundMaya"];
  createMore: boolean;
  currentTeam: TTeam | null;
  currentTeamId: DialogFieldsProps["currentTeamId"];
  customFields?: ReactNode;
  deadlineSourceRef: DialogFieldsProps["deadlineSourceRef"];
  descriptionEditor: Editor | null;
  dispatch: DialogFieldsProps["dispatch"];
  estimateScheme: DialogFieldsProps["estimateScheme"];
  figmaArtifacts: FigmaSourceProps["artifacts"];
  isCreating: boolean;
  isExpanded: boolean;
  isMayaAssigned: DialogFieldsProps["isMayaAssigned"];
  isMayaAssigneeLoading: boolean;
  isOpen: boolean;
  mayaAssigneeId: DialogFieldsProps["mayaAssigneeId"];
  mediaInputRef: RefObject<HTMLInputElement | null>;
  member: DialogFieldsProps["member"];
  members: SimilarStoriesProps["members"];
  objectiveTerm: DialogFieldsProps["objectiveTerm"];
  onActiveTeamChange: (team: TTeam) => void;
  onCreate: () => void;
  onCreateMoreChange: (checked: boolean) => void;
  onFigmaArtifactsChange: FigmaSourceProps["onArtifactsChange"];
  onMediaFiles: (files: File[]) => void;
  onOpenChange: (open: boolean) => void;
  onSimilarStorySelect: SimilarStoriesProps["onSelect"];
  onToggleExpanded: () => void;
  selectedLabels: DialogFieldsProps["selectedLabels"];
  showObjectives: DialogFieldsProps["showObjectives"];
  showSprints: DialogFieldsProps["showSprints"];
  similarStories: SimilarStoriesProps["stories"];
  sprintName: DialogFieldsProps["sprintName"];
  sprintTerm: DialogFieldsProps["sprintTerm"];
  statuses: SimilarStoriesProps["statuses"];
  storyForm: DialogFieldsProps["storyForm"];
  storyTerm: string;
  storyTermPlural: string;
  strategyLinkLabel: DialogFieldsProps["strategyLinkLabel"];
  teamStatuses: DialogFieldsProps["teamStatuses"];
  teams: TTeam[];
  titleEditor: Editor | null;
  templatePicker?: ReactNode;
}) => (
  <Dialog onOpenChange={onOpenChange} open={isOpen}>
    <Dialog.Content
      className="mt-4 flex max-h-[calc(100dvh-2rem)] flex-col overflow-visible md:mt-[10%] md:max-h-[calc(90dvh-1rem)]"
      hideClose
      size={isExpanded ? "xl" : "lg"}
    >
      <NewStoryDialogHeader
        activeTeamId={activeTeamId}
        currentTeam={currentTeam}
        isExpanded={isExpanded}
        onTeamSelect={(selectedTeamId) => {
          const selectedTeam = teams.find((team) => team.id === selectedTeamId);
          if (selectedTeam) onActiveTeamChange(selectedTeam);
        }}
        onToggleExpanded={onToggleExpanded}
        storyTerm={storyTerm}
        teams={teams}
        templatePicker={templatePicker}
      />
      <Dialog.Body className="max-h-[60dvh] min-h-0 flex-1 overflow-y-auto pt-0 pb-6">
        <TextEditor
          asTitle
          className="text-2xl font-medium"
          editor={titleEditor}
        />
        <TextEditor
          className={cn("rich-document-editor min-h-20", {
            "min-h-96": isExpanded,
          })}
          editor={descriptionEditor}
        />
        <input
          accept={RICH_TEXT_MEDIA_ACCEPT}
          aria-label="Upload story description media"
          className="sr-only"
          multiple
          onChange={(event) => {
            const files = Array.from(event.target.files ?? []);
            event.target.value = "";
            if (files.length > 0) onMediaFiles(files);
          }}
          ref={mediaInputRef}
          type="file"
        />
        <RichTextTableMenu editor={descriptionEditor} scrollTarget={null} />
        <NewStoryDialogFields
          canUseBackgroundMaya={canUseBackgroundMaya}
          currentTeamId={currentTeamId}
          deadlineSourceRef={deadlineSourceRef}
          dispatch={dispatch}
          estimateScheme={estimateScheme}
          isMayaAssigned={isMayaAssigned}
          mayaAssigneeId={mayaAssigneeId}
          member={member}
          objectiveTerm={objectiveTerm}
          selectedLabels={selectedLabels}
          showObjectives={showObjectives}
          showSprints={showSprints}
          sprintName={sprintName}
          sprintTerm={sprintTerm}
          storyForm={storyForm}
          strategyLinkLabel={strategyLinkLabel}
          teamStatuses={teamStatuses}
        >
          {customFields}
          <NewStoryFigmaSource
            artifacts={figmaArtifacts}
            enabled={isOpen}
            onAddDescription={(figmaDescription) => {
              insertFigmaDescription(descriptionEditor, figmaDescription);
            }}
            onArtifactsChange={onFigmaArtifactsChange}
            onTitleSuggestion={(title) => {
              if (titleEditor && !titleEditor.getText().trim()) {
                titleEditor.commands.setContent(title);
              }
            }}
          />
        </NewStoryDialogFields>
      </Dialog.Body>
      <Dialog.Footer className="flex shrink-0 items-center justify-between gap-2">
        <Text color="muted">
          <label className="flex items-center gap-2" htmlFor="more">
            Create more
            <Switch
              checked={createMore}
              id="more"
              onCheckedChange={onCreateMoreChange}
            />
          </label>
        </Text>
        <Button
          leftIcon={<PlusIcon className="text-current" />}
          loading={isCreating || isMayaAssigneeLoading}
          loadingText={
            isMayaAssigneeLoading
              ? "Preparing Maya..."
              : `Creating ${storyTerm}...`
          }
          onClick={onCreate}
          size="md"
        >
          Create {storyTerm}
        </Button>
      </Dialog.Footer>
      <NewStoryDialogSimilarStories
        currentTeamCode={currentTeam?.code}
        heading={`Similar ${storyTermPlural}`}
        members={members}
        onSelect={onSimilarStorySelect}
        statuses={statuses}
        stories={similarStories}
        teamCodes={teams}
      />
    </Dialog.Content>
  </Dialog>
);

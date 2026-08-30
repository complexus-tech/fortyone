import type { ComponentProps, ReactNode } from "react";
import { PlusIcon } from "icons";
import { Button, Dialog } from "ui";
import { NewObjectiveDialogControls } from "./new-objective-dialog-controls";
import { NewObjectiveDialogFields } from "./new-objective-dialog-fields";
import { NewObjectiveDialogHeader } from "./new-objective-dialog-header";
import { NewObjectiveKeyResults } from "./new-objective-key-results";

type HeaderProps = ComponentProps<typeof NewObjectiveDialogHeader>;
type FieldsProps = ComponentProps<typeof NewObjectiveDialogFields>;
type ControlsProps = ComponentProps<typeof NewObjectiveDialogControls>;
type KeyResultsProps = ComponentProps<typeof NewObjectiveKeyResults>;

type ObjectiveForm = {
  color?: ControlsProps["color"];
  endDate?: ControlsProps["endDate"];
  leadUser?: ControlsProps["leadUserId"];
  priority?: ControlsProps["priority"];
  shortSummary?: FieldsProps["shortSummary"];
  startDate?: ControlsProps["startDate"];
  statusId: NonNullable<ControlsProps["statusId"]>;
};

export const NewObjectiveDialogContent = ({
  currentTeam,
  descriptionEditor,
  form,
  isCreating,
  isExpanded,
  isOpen,
  keyResults,
  lead,
  newObjectiveTerm,
  objectiveTerm,
  onCreate,
  onDiscard,
  onObjectiveFormChange,
  onOpenChange,
  onTeamSelect,
  onToggleExpanded,
  qualityBanner,
  statuses,
  teams,
  titleEditor,
}: {
  currentTeam: HeaderProps["currentTeam"];
  descriptionEditor: FieldsProps["descriptionEditor"];
  form: ObjectiveForm;
  isCreating: boolean;
  isExpanded: boolean;
  isOpen: boolean;
  keyResults?: KeyResultsProps;
  lead: ControlsProps["lead"];
  newObjectiveTerm: HeaderProps["objectiveTerm"];
  objectiveTerm: string;
  onCreate: () => void;
  onDiscard: () => void;
  onObjectiveFormChange: (updates: Partial<ObjectiveForm>) => void;
  onOpenChange: (open: boolean) => void;
  onTeamSelect: HeaderProps["onTeamSelect"];
  onToggleExpanded: HeaderProps["onToggleExpanded"];
  qualityBanner: ReactNode;
  statuses: ControlsProps["statuses"];
  teams: HeaderProps["teams"];
  titleEditor: FieldsProps["titleEditor"];
}) => (
  <Dialog onOpenChange={onOpenChange} open={isOpen}>
    <Dialog.Content hideClose size="lg">
      <NewObjectiveDialogHeader
        currentTeam={currentTeam}
        isExpanded={isExpanded}
        objectiveTerm={newObjectiveTerm}
        onTeamSelect={onTeamSelect}
        onToggleExpanded={onToggleExpanded}
        teams={teams}
      />
      <Dialog.Body className="max-h-[60dvh] pt-0">
        <NewObjectiveDialogFields
          descriptionEditor={descriptionEditor}
          isExpanded={isExpanded}
          onShortSummaryChange={(shortSummary) => {
            onObjectiveFormChange({ shortSummary });
          }}
          shortSummary={form.shortSummary}
          titleEditor={titleEditor}
        >
          {qualityBanner}
        </NewObjectiveDialogFields>
        <NewObjectiveDialogControls
          color={form.color}
          endDate={form.endDate}
          lead={lead}
          leadUserId={form.leadUser}
          onColorChange={(color) => {
            onObjectiveFormChange({ color });
          }}
          onEndDateChange={(endDate) => {
            onObjectiveFormChange({ endDate });
          }}
          onLeadChange={(leadUserId) => {
            onObjectiveFormChange({ leadUser: leadUserId || undefined });
          }}
          onPriorityChange={(priority) => {
            onObjectiveFormChange({ priority });
          }}
          onStartDateChange={(startDate) => {
            onObjectiveFormChange({ startDate });
          }}
          onStatusChange={(statusId) => {
            onObjectiveFormChange({ statusId });
          }}
          priority={form.priority}
          startDate={form.startDate}
          statusId={form.statusId}
          statuses={statuses}
        />
        {keyResults ? <NewObjectiveKeyResults {...keyResults} /> : null}
      </Dialog.Body>
      <Dialog.Footer className="flex items-center justify-end gap-2">
        <Button
          className="px-5"
          color="tertiary"
          onClick={onDiscard}
          variant="outline"
        >
          Discard
        </Button>
        <Button
          leftIcon={<PlusIcon className="text-current" />}
          loading={isCreating}
          loadingText="Creating..."
          onClick={onCreate}
          size="md"
        >
          Create {objectiveTerm}
        </Button>
      </Dialog.Footer>
    </Dialog.Content>
  </Dialog>
);

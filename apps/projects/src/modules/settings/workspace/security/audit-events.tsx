import { useRef, useState } from "react";
import { MoreHorizontalIcon } from "icons";
import { Box, Button, Dialog, Menu, Table, Text } from "ui";
import type { AuditEvent } from "./types";
import { securityDate, securityLabel } from "./display";
import { SecurityTimestamp } from "./security-timestamp";

const EVENT_LABELS: Record<string, string> = {
  "workspace.security_policy_updated": "Security policy updated",
  "workspace.session_revoked": "Session revoked",
  "workspace.member_sessions_revoked": "Member sessions revoked",
  "workspace.audit_exported": "Audit log exported",
  "workspace.data_exported": "Workspace data exported",
  "custom_field.value_changed": "Custom field value changed",
};

const eventLabel = (event: AuditEvent) =>
  EVENT_LABELS[event.operation] ?? securityLabel(event.operation);

const actorLabel = (event: AuditEvent, names: Map<string, string>) => {
  const member = event.actorId ? names.get(event.actorId) : undefined;
  if (member) return member;
  if (event.actorType === "human_user") return "Member";
  if (event.actorType === "scim_credential") return "SCIM token";
  return securityLabel(event.actorType) || "System";
};

const AuditEventActions = ({
  event,
  onDetails,
}: {
  event: AuditEvent;
  onDetails: (event: AuditEvent, trigger: HTMLButtonElement | null) => void;
}) => {
  const triggerRef = useRef<HTMLButtonElement>(null);
  const openingDetails = useRef(false);
  return (
    <Menu>
      <Menu.Button>
        <Button
          aria-label={`Event actions for ${eventLabel(event)} (${event.id.slice(0, 8)})`}
          asIcon
          color="tertiary"
          leftIcon={<MoreHorizontalIcon aria-hidden />}
          ref={triggerRef}
          size="sm"
          variant="naked"
        />
      </Menu.Button>
      <Menu.Items
        align="end"
        onCloseAutoFocus={(closeEvent) => {
          if (!openingDetails.current) return;
          closeEvent.preventDefault();
          openingDetails.current = false;
        }}
      >
        <Menu.Group>
          <Menu.Item
            onSelect={() => {
              openingDetails.current = true;
              onDetails(event, triggerRef.current);
            }}
          >
            Field Details
          </Menu.Item>
        </Menu.Group>
      </Menu.Items>
    </Menu>
  );
};

export const AuditEvents = ({
  events,
  names,
}: {
  events: AuditEvent[];
  names: Map<string, string>;
}) => {
  const [selected, setSelected] = useState<AuditEvent | null>(null);
  const returnFocusRef = useRef<HTMLButtonElement | null>(null);
  return (
    <>
      <Box
        aria-label="Workspace audit events table"
        className="border-border focus-visible:ring-ring overflow-x-auto border-t-[0.5px] focus-visible:ring-2 focus-visible:outline-none focus-visible:ring-inset"
        role="region"
        tabIndex={0}
      >
        <Table
          align={null}
          className="min-w-[60rem] text-base [&_td]:px-6 [&_td]:py-3 [&_td]:whitespace-nowrap [&_td:first-child]:font-medium [&_td:last-child]:pr-6 [&_tr]:border-b-[0.5px]"
          color="light"
          size="md"
          variant="bordered"
        >
          <caption className="sr-only">
            Workspace audit events, newest first
          </caption>
          <Table.Head>
            <Table.Tr>
              {["Event", "Actor", "Resource", "Recorded (UTC)", "Details"].map(
                (label) => (
                  <Table.Th
                    className="bg-surface static px-6 text-left text-base font-medium whitespace-nowrap normal-case"
                    key={label}
                    scope="col"
                  >
                    {label}
                  </Table.Th>
                ),
              )}
            </Table.Tr>
          </Table.Head>
          <Table.Body>
            {events.map((event) => (
              <Table.Tr key={`${event.source}:${event.id}`}>
                <Table.Td>
                  <Text
                    className="max-w-64 truncate"
                    fontWeight="medium"
                    title={eventLabel(event)}
                  >
                    {eventLabel(event)}
                  </Text>
                </Table.Td>
                <Table.Td>
                  <Text
                    className="max-w-48 truncate"
                    title={actorLabel(event, names)}
                  >
                    {actorLabel(event, names)}
                  </Text>
                </Table.Td>
                <Table.Td>
                  <Text
                    className="max-w-64 truncate"
                    title={event.resourceId || undefined}
                  >
                    {securityLabel(event.resourceType)}
                    {event.resourceId ? (
                      <span className="text-text-muted font-mono">
                        {" · "}
                        {event.resourceId.slice(0, 8)}…
                      </span>
                    ) : null}
                  </Text>
                </Table.Td>
                <Table.Td>
                  <SecurityTimestamp value={event.createdAt} />
                </Table.Td>
                <Table.Td>
                  <AuditEventActions
                    event={event}
                    onDetails={(selectedEvent, trigger) => {
                      returnFocusRef.current = trigger;
                      setSelected(selectedEvent);
                    }}
                  />
                </Table.Td>
              </Table.Tr>
            ))}
          </Table.Body>
        </Table>
      </Box>
      <Dialog
        onOpenChange={(open) => {
          if (!open) setSelected(null);
        }}
        open={Boolean(selected)}
      >
        <Dialog.Content
          className="flex max-h-[calc(100dvh-2rem)] flex-col"
          onCloseAutoFocus={(event) => {
            const trigger = returnFocusRef.current;
            if (trigger?.isConnected) {
              event.preventDefault();
              trigger.focus();
            }
            returnFocusRef.current = null;
          }}
          size="lg"
        >
          <Dialog.Header className="shrink-0 px-6 py-5">
            <Dialog.Title className="text-lg">Audit event details</Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="max-h-none min-h-0 flex-1 space-y-5">
            <Dialog.Description className="px-0 text-base">
              {selected ? eventLabel(selected) : "Review this audit event."}
            </Dialog.Description>
            {selected ? (
              <>
                <dl className="grid gap-4 sm:grid-cols-2">
                  {[
                    ["Actor", actorLabel(selected, names)],
                    ["Actor type", securityLabel(selected.actorType)],
                    ["Actor ID", selected.actorId || "—"],
                    ["Recorded", securityDate(selected.createdAt)],
                    ["Resource", securityLabel(selected.resourceType)],
                    ["Resource ID", selected.resourceId || "—"],
                    ["Source", selected.source],
                    ["Operation", selected.operation],
                    ["Event ID", selected.id],
                  ].map(([label, value]) => (
                    <Box className="min-w-0" key={label}>
                      <Text as="dt" color="muted">
                        {label}
                      </Text>
                      <Text as="dd" className="mt-1 break-all select-all">
                        {value}
                      </Text>
                    </Box>
                  ))}
                </dl>
                <Box>
                  <Text className="mb-2" fontWeight="medium">
                    Metadata
                  </Text>
                  <pre className="bg-surface-elevated max-h-80 overflow-auto rounded-lg p-4 font-mono text-base break-words whitespace-pre-wrap">
                    {JSON.stringify(selected.metadata, null, 2)}
                  </pre>
                </Box>
              </>
            ) : null}
          </Dialog.Body>
          <Dialog.Footer className="shrink-0 py-4" justify="end">
            <Button
              color="tertiary"
              onClick={() => {
                setSelected(null);
              }}
              variant="outline"
            >
              Done
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </>
  );
};

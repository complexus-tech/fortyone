import type { FormEvent } from "react";
import { useId, useRef, useState } from "react";
import { Box, Button, Dialog, Flex, Input, Select, Switch, Text } from "ui";
import { PlusIcon } from "icons";
import { CUSTOM_FIELD_TYPE_LABELS } from "./types";
import type {
  CustomField,
  CustomFieldDraft,
  CustomFieldIconKey,
  CustomFieldType,
} from "./types";
import { useCustomFieldMutations } from "./hooks";
import { CustomFieldIconPicker } from "./icon-picker";

const FIELD_TYPES = Object.keys(CUSTOM_FIELD_TYPE_LABELS) as CustomFieldType[];
const FIELD_TYPE_DESCRIPTIONS: Record<CustomFieldType, string> = {
  text: "Free-form text.",
  number: "Numbers or scores.",
  money: "An amount in one currency.",
  date: "A date or deadline.",
  select: "One choice from a list.",
  person: "A team member.",
};

export const CustomFieldEditor = ({
  teamId,
  field,
  onClose,
  onReturnFocus,
}: {
  teamId: string;
  field: CustomField | null;
  onClose: () => void;
  onReturnFocus?: () => void;
}) => {
  const controlsId = useId();
  const nameInputRef = useRef<HTMLInputElement>(null);
  const [name, setName] = useState(field?.name ?? "");
  const [type, setType] = useState<CustomFieldType>(field?.type ?? "text");
  const [icon, setIcon] = useState<CustomFieldIconKey | null>(
    field?.icon ?? null,
  );
  const [currency, setCurrency] = useState(field?.currency ?? "USD");
  const [options, setOptions] = useState(
    field?.options
      .filter((option) => !option.archivedAt)
      .map(({ id, name: optionName }) => ({ id, name: optionName })) ?? [],
  );
  const [showOnCreate, setShowOnCreate] = useState(
    field?.showOnCreate ?? false,
  );
  const [error, setError] = useState<string | null>(null);
  const { create, update } = useCustomFieldMutations(teamId);
  const isPending = create.isPending || update.isPending;
  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (isPending) return;
    setError(null);
    const trimmedOptions = options.map((option) => ({
      ...option,
      name: option.name.trim(),
    }));
    if (!name.trim()) {
      setError("Give this field a name.");
      return;
    }
    if (
      type === "select" &&
      (!trimmedOptions.length || trimmedOptions.some((option) => !option.name))
    ) {
      setError("Add at least one named option.");
      return;
    }
    if (
      type === "select" &&
      new Set(trimmedOptions.map((option) => option.name.toLowerCase()))
        .size !== trimmedOptions.length
    ) {
      setError("Give every option a distinct name.");
      return;
    }
    const normalizedCurrency = currency.trim().toUpperCase();
    if (type === "money") {
      try {
        if (!/^[A-Z]{3}$/.test(normalizedCurrency))
          throw new Error("Invalid currency");
        const validatedCurrency = new Intl.NumberFormat("en", {
          style: "currency",
          currency: normalizedCurrency,
        }).resolvedOptions().currency;
        if (validatedCurrency !== normalizedCurrency)
          throw new Error("Invalid currency");
      } catch {
        setError("Use a three-letter currency code, such as USD or EUR.");
        return;
      }
    }
    const input: CustomFieldDraft = {
      name: name.trim(),
      type,
      icon,
      currency: type === "money" ? normalizedCurrency : null,
      options: type === "select" ? trimmedOptions : [],
      showOnCreate,
    };
    try {
      if (field)
        await update.mutateAsync({
          fieldId: field.id,
          input: {
            name: input.name,
            icon,
            options: input.options,
            showOnCreate,
          },
        });
      else await create.mutateAsync(input);
      onClose();
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "The field could not be saved.",
      );
    }
  };
  const saveLabel = field ? "Save field" : "Create field";
  return (
    <Dialog
      onOpenChange={(open) => {
        if (!open && !isPending) onClose();
      }}
      open
    >
      <Dialog.Content
        aria-busy={isPending}
        className="flex max-h-[calc(100dvh-15vw-1rem)] flex-col md:max-h-[calc(100dvh-10vw-1rem)]"
        hideClose={isPending}
        onCloseAutoFocus={(event) => {
          if (!onReturnFocus) return;
          event.preventDefault();
          onReturnFocus();
        }}
        onEscapeKeyDown={(event) => {
          if (isPending) event.preventDefault();
        }}
        onInteractOutside={(event) => {
          if (isPending) event.preventDefault();
        }}
        onOpenAutoFocus={(event) => {
          event.preventDefault();
          nameInputRef.current?.focus({ preventScroll: true });
        }}
      >
        <Dialog.Header className="border-border shrink-0 border-b px-6 py-5">
          <Dialog.Title className="pr-8 text-lg">
            {field ? "Edit field" : "Create field"}
          </Dialog.Title>
          <Dialog.Description className="mt-2 px-0 text-base leading-6">
            Set a team property for tasks and reports.
          </Dialog.Description>
        </Dialog.Header>
        <form className="flex min-h-0 flex-1 flex-col" onSubmit={submit}>
          <Dialog.Body className="max-h-none min-h-0 flex-1 scroll-py-5 space-y-5 py-5">
            <Box>
              <label className="mb-2 block" htmlFor={`${controlsId}-name`}>
                Field name
              </label>
              <Flex align="center" gap={2}>
                <CustomFieldIconPicker
                  disabled={isPending}
                  onChange={setIcon}
                  type={type}
                  value={icon}
                />
                <Box className="min-w-0 flex-1">
                  <Input
                    className="h-[var(--settings-control-height)] px-3 leading-normal"
                    disabled={isPending}
                    id={`${controlsId}-name`}
                    maxLength={100}
                    onChange={(event) => {
                      setName(event.target.value);
                    }}
                    placeholder="e.g. Customer, Impact, Launch date"
                    ref={nameInputRef}
                    required
                    value={name}
                  />
                </Box>
              </Flex>
            </Box>
            <Flex align="center" className="gap-4" justify="between">
              <Box className="min-w-0 flex-1">
                <Text id={`${controlsId}-type`}>Field type</Text>
                <Text className="mt-1 leading-6" color="muted">
                  {FIELD_TYPE_DESCRIPTIONS[type]}
                </Text>
              </Box>
              <Select
                disabled={Boolean(field) || isPending}
                onValueChange={(value) => {
                  setType(value as CustomFieldType);
                }}
                value={type}
              >
                <Select.Trigger
                  aria-labelledby={`${controlsId}-type`}
                  className="h-[var(--settings-control-height)] w-40 shrink-0 text-base"
                >
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  {FIELD_TYPES.map((value) => (
                    <Select.Option
                      className="text-base"
                      key={value}
                      value={value}
                    >
                      {CUSTOM_FIELD_TYPE_LABELS[value]}
                    </Select.Option>
                  ))}
                </Select.Content>
              </Select>
            </Flex>
            {type === "select" ? (
              <Box>
                <Text className="mb-2">Options</Text>
                <Box className="space-y-2">
                  {options.map((option) => (
                    <Flex align="center" gap={2} key={option.id}>
                      <Box className="min-w-0 flex-1">
                        <Input
                          aria-label="Option name"
                          className="h-[var(--settings-control-height)] px-3 leading-normal"
                          disabled={isPending}
                          maxLength={100}
                          onChange={(event) => {
                            setOptions((current) =>
                              current.map((entry) =>
                                entry.id === option.id
                                  ? { ...entry, name: event.target.value }
                                  : entry,
                              ),
                            );
                          }}
                          placeholder="Option name"
                          value={option.name}
                        />
                      </Box>
                      <Button
                        aria-label={`Remove ${option.name || "option"}`}
                        className="h-[var(--settings-control-height)] shrink-0 px-2"
                        color="tertiary"
                        disabled={isPending}
                        onClick={() => {
                          setOptions((current) =>
                            current.filter((entry) => entry.id !== option.id),
                          );
                        }}
                        size="sm"
                        type="button"
                        variant="naked"
                      >
                        Remove
                      </Button>
                    </Flex>
                  ))}
                </Box>
                <Button
                  className="mt-3 h-[var(--settings-control-height)]"
                  color="tertiary"
                  disabled={isPending || options.length >= 100}
                  leftIcon={<PlusIcon />}
                  onClick={() => {
                    setOptions((current) => [
                      ...current,
                      { id: crypto.randomUUID(), name: "" },
                    ]);
                  }}
                  size="sm"
                  type="button"
                  variant="outline"
                >
                  Add option
                </Button>
                {field ? (
                  <Text className="mt-2 leading-6" color="muted">
                    Removing an option keeps its existing values for history.
                  </Text>
                ) : null}
              </Box>
            ) : null}
            {type === "money" ? (
              <Flex align="center" className="gap-4" justify="between">
                <Box className="min-w-0 flex-1">
                  <label className="block" htmlFor={`${controlsId}-currency`}>
                    Currency
                  </label>
                  <Text className="mt-1" color="muted">
                    {field
                      ? "Currency is fixed after creation."
                      : "Use a three-letter code, such as USD or EUR."}
                  </Text>
                </Box>
                <Box className="w-24 shrink-0">
                  <Input
                    className="h-[var(--settings-control-height)] px-3 leading-normal"
                    disabled={Boolean(field) || isPending}
                    id={`${controlsId}-currency`}
                    maxLength={3}
                    onChange={(event) => {
                      setCurrency(event.target.value.toUpperCase());
                    }}
                    placeholder="USD"
                    required
                    value={currency}
                  />
                </Box>
              </Flex>
            ) : null}
            <Flex
              align="center"
              className="border-border gap-4 border-t pt-5"
              justify="between"
            >
              <Box className="min-w-0">
                <label
                  className="block"
                  htmlFor={`${controlsId}-show-on-create`}
                >
                  Show when creating work
                </label>
                <Text className="mt-1" color="muted">
                  Include this field in the create form by default.
                </Text>
              </Box>
              <Switch
                checked={showOnCreate}
                className="shrink-0"
                disabled={isPending}
                id={`${controlsId}-show-on-create`}
                onCheckedChange={setShowOnCreate}
              />
            </Flex>
            {error ? (
              <Text color="danger" role="alert">
                {error}
              </Text>
            ) : null}
          </Dialog.Body>
          <Dialog.Footer className="shrink-0 gap-2 py-4" justify="end">
            <Button
              color="tertiary"
              disabled={isPending}
              onClick={onClose}
              type="button"
              variant="outline"
            >
              Cancel
            </Button>
            <Button
              disabled={isPending}
              loading={isPending}
              loadingText="Saving…"
              type="submit"
            >
              {saveLabel}
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};

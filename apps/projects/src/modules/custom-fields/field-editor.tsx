import type { FormEvent } from "react";
import { useState } from "react";
import { Box, Button, Dialog, Flex, Input, Select, Switch, Text } from "ui";
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

export const CustomFieldEditor = ({
  teamId,
  field,
  onClose,
}: {
  teamId: string;
  field: CustomField | null;
  onClose: () => void;
}) => {
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
        className="mt-0 flex max-h-[calc(100dvh-2rem)] flex-col md:mt-0"
        hideClose={isPending}
        onEscapeKeyDown={(event) => {
          if (isPending) event.preventDefault();
        }}
        onInteractOutside={(event) => {
          if (isPending) event.preventDefault();
        }}
        overlayClassName="items-center py-4"
      >
        <Dialog.Header className="shrink-0 px-6 py-4">
          <Dialog.Title className="text-lg">
            {field ? "Edit field" : "Create field"}
          </Dialog.Title>
          <Dialog.Description className="mt-1 px-0 text-base leading-6">
            Set a team property for tasks and reports.
          </Dialog.Description>
        </Dialog.Header>
        <form className="flex min-h-0 flex-1 flex-col" onSubmit={submit}>
          <Dialog.Body className="max-h-none min-h-0 flex-1 space-y-4">
            <Box className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Flex align="start" className="min-w-0" gap={3}>
                <Box className="w-10 shrink-0">
                  <CustomFieldIconPicker
                    disabled={isPending}
                    onChange={setIcon}
                    type={type}
                    value={icon}
                  />
                </Box>
                <Box className="min-w-0 flex-1">
                  <Input
                    autoFocus
                    className="h-10"
                    disabled={isPending}
                    label="Field name"
                    labelClassName="mb-2 font-medium"
                    maxLength={100}
                    onChange={(event) => {
                      setName(event.target.value);
                    }}
                    placeholder="e.g. Budget"
                    required
                    value={name}
                  />
                </Box>
              </Flex>
              <Box className="min-w-0">
                <Text className="mb-2" fontWeight="medium">
                  Field type
                </Text>
                <Select
                  disabled={Boolean(field) || isPending}
                  onValueChange={(value) => {
                    setType(value as CustomFieldType);
                  }}
                  value={type}
                >
                  <Select.Trigger
                    aria-label="Field type"
                    className="h-10 w-full text-base"
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
                {field ? (
                  <Text className="mt-2 leading-6" color="muted">
                    Type is fixed after creation.
                  </Text>
                ) : null}
              </Box>
            </Box>
            {type === "select" ? (
              <Box>
                <Text className="mb-2" fontWeight="medium">
                  Options
                </Text>
                <Box className="space-y-2">
                  {options.map((option) => (
                    <Flex align="center" gap={2} key={option.id}>
                      <Input
                        aria-label="Option name"
                        className="h-10"
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
                      <Button
                        aria-label={`Remove ${option.name || "option"}`}
                        className="h-10"
                        color="tertiary"
                        disabled={isPending}
                        onClick={() => {
                          setOptions((current) =>
                            current.filter((entry) => entry.id !== option.id),
                          );
                        }}
                        type="button"
                        variant="naked"
                      >
                        Remove
                      </Button>
                    </Flex>
                  ))}
                </Box>
                <Button
                  className="mt-3 h-10"
                  color="tertiary"
                  disabled={isPending || options.length >= 100}
                  onClick={() => {
                    setOptions((current) => [
                      ...current,
                      { id: crypto.randomUUID(), name: "" },
                    ]);
                  }}
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
            <Box
              className={
                type === "money"
                  ? "grid grid-cols-1 items-end gap-4 sm:grid-cols-[auto_1fr]"
                  : undefined
              }
            >
              {type === "money" ? (
                <Box className="w-24">
                  <Input
                    className="h-10"
                    disabled={Boolean(field) || isPending}
                    label="Currency"
                    labelClassName="mb-2 font-medium"
                    maxLength={3}
                    onChange={(event) => {
                      setCurrency(event.target.value.toUpperCase());
                    }}
                    placeholder="USD"
                    required
                    value={currency}
                  />
                </Box>
              ) : null}
              <Flex align="center" className="min-h-10 gap-4" justify="between">
                <Text fontWeight="medium">Show when creating work</Text>
                <Switch
                  aria-label="Show when creating work"
                  checked={showOnCreate}
                  disabled={isPending}
                  onCheckedChange={setShowOnCreate}
                />
              </Flex>
            </Box>
            {error ? (
              <Text color="danger" role="alert">
                {error}
              </Text>
            ) : null}
          </Dialog.Body>
          <Dialog.Footer className="shrink-0 gap-3 py-3" justify="end">
            <Button
              className="h-10"
              color="tertiary"
              disabled={isPending}
              onClick={onClose}
              type="button"
              variant="outline"
            >
              Cancel
            </Button>
            <Button className="h-10" disabled={isPending} type="submit">
              {isPending ? "Saving…" : saveLabel}
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};

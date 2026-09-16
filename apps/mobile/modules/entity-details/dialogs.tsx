import type { QueryKey } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { useRef, useState, useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  Modal,
  KeyboardAvoidingView,
  Platform,
  TextInput,
  FlatList,
  Pressable,
} from "react-native";
import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context";
import { Button, Text, Row, IconButton } from "@/components/ui";
import { QueryState } from "@/components/ui/query-state";
import { useTheme } from "@/hooks/theme";
import { themeColors } from "@/constants/colors";

function DialogShell({
  title,
  onClose,
  busy,
  children,
}: {
  title: string;
  onClose: () => void;
  busy: boolean;
  children: ReactNode;
}) {
  const { resolvedTheme } = useTheme();
  return (
    <Modal
      visible
      presentationStyle="pageSheet"
      animationType="slide"
      onRequestClose={() => {
        if (!busy) onClose();
      }}
    >
      <SafeAreaProvider>
        <SafeAreaView
          style={{
            flex: 1,
            backgroundColor: themeColors[resolvedTheme].background,
          }}
        >
          <KeyboardAvoidingView
            behavior={Platform.OS === "ios" ? "padding" : undefined}
            style={{ flex: 1, padding: 20 }}
          >
            <Row align="center" className="mb-4 gap-3">
              <IconButton
                label="Close"
                icon="close"
                disabled={busy}
                onPress={onClose}
              />
              <Text fontWeight="bold" style={{ flex: 1 }}>
                {title}
              </Text>
            </Row>
            {children}
          </KeyboardAvoidingView>
        </SafeAreaView>
      </SafeAreaProvider>
    </Modal>
  );
}
export function TextActionDialog({
  title,
  description,
  placeholder,
  onSubmit,
  onClose,
}: {
  title: string;
  description: string;
  placeholder: string;
  onSubmit: (text: string) => Promise<void>;
  onClose: () => void;
}) {
  const { resolvedTheme } = useTheme();
  const [text, setText] = useState("");
  const [busy, setBusy] = useState(false);
  const pending = useRef(false);
  const [error, setError] = useState<string>();
  return (
    <DialogShell title={title} busy={busy} onClose={onClose}>
      <Text color="muted">{description}</Text>
      <TextInput
        accessibilityLabel={placeholder}
        multiline
        autoFocus
        editable={!busy}
        maxLength={1000}
        value={text}
        onChangeText={setText}
        placeholder={placeholder}
        placeholderTextColor={themeColors[resolvedTheme].textMuted}
        style={{
          color: themeColors[resolvedTheme].foreground,
          fontSize: 17,
          minHeight: 150,
          marginVertical: 20,
          textAlignVertical: "top",
        }}
      />
      {error ? <Text color="danger">{error}</Text> : null}
      <Button
        loading={busy}
        onPress={() => {
          if (pending.current) return;
          pending.current = true;
          setBusy(true);
          setError(undefined);
          void onSubmit(text.trim())
            .then(onClose)
            .catch((cause: unknown) =>
              setError(
                cause instanceof Error ? cause.message : "Please try again.",
              ),
            )
            .finally(() => {
              pending.current = false;
              setBusy(false);
            });
        }}
      >
        {title}
      </Button>
    </DialogShell>
  );
}
export function SearchActionDialog({
  title,
  description,
  queryKey,
  search,
  onSelect,
  onClose,
}: {
  title: string;
  description: string;
  queryKey: QueryKey;
  search: (
    query: string,
    signal: AbortSignal,
  ) => Promise<{ id: string; title: string }[]>;
  onSelect: (id: string) => Promise<void>;
  onClose: () => void;
}) {
  const { resolvedTheme } = useTheme();
  const [text, setText] = useState("");
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState<string>();
  const [busy, setBusy] = useState(false);
  const pending = useRef(false);
  const [error, setError] = useState<string>();
  useEffect(() => {
    const timer = setTimeout(() => setQuery(text.trim()), 250);
    return () => clearTimeout(timer);
  }, [text]);
  const results = useQuery({
    queryKey: [...queryKey, query],
    queryFn: ({ signal }) => search(query, signal),
  });
  return (
    <DialogShell title={title} onClose={onClose} busy={busy}>
      <Text color="muted">{description}</Text>
      <TextInput
        accessibilityLabel="Search"
        autoFocus
        value={text}
        editable={!busy}
        onChangeText={(value) => {
          setText(value);
          setSelected(undefined);
        }}
        placeholder="Search…"
        placeholderTextColor={themeColors[resolvedTheme].textMuted}
        style={{
          color: themeColors[resolvedTheme].foreground,
          fontSize: 17,
          paddingVertical: 16,
        }}
      />
      {results.isPending || results.error ? (
        <QueryState
          loading={results.isPending}
          title={results.error ? "Could not search" : "Searching"}
          message={results.error?.message}
          onRetry={() => void results.refetch()}
        />
      ) : (
        <FlatList
          style={{ flex: 1 }}
          data={results.data}
          keyExtractor={(item) => item.id}
          renderItem={({ item }) => (
            <Pressable
              accessibilityRole="radio"
              accessibilityState={{ selected: selected === item.id }}
              disabled={busy}
              onPress={() => setSelected(item.id)}
              style={{
                paddingVertical: 14,
                paddingHorizontal: 10,
                borderRadius: 12,
                backgroundColor:
                  selected === item.id
                    ? themeColors[resolvedTheme].stateSelected
                    : "transparent",
              }}
            >
              <Text>{item.title}</Text>
            </Pressable>
          )}
          ListEmptyComponent={<Text color="muted">No matching items.</Text>}
        />
      )}
      {error ? <Text color="danger">{error}</Text> : null}
      <Button
        loading={busy}
        disabled={!selected || busy || text.trim() !== query}
        onPress={() => {
          if (!selected || pending.current) return;
          pending.current = true;
          setBusy(true);
          setError(undefined);
          void onSelect(selected)
            .then(onClose)
            .catch((cause: unknown) =>
              setError(
                cause instanceof Error ? cause.message : "Please try again.",
              ),
            )
            .finally(() => {
              pending.current = false;
              setBusy(false);
            });
        }}
      >
        {title}
      </Button>
    </DialogShell>
  );
}

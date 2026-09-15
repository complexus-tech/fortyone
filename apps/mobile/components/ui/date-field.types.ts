export type DateFieldProps = {
  value: string | null | undefined;
  label: string;
  onChange: (value: Date | null) => void | Promise<void>;
  disabled?: boolean;
};

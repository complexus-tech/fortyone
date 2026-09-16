export type TabOption = {
  value: string;
  label: string;
};

export type TabsControlProps = {
  options: readonly TabOption[];
  value: string;
  onValueChange: (value: string) => void;
  labelSize?: 14 | 15;
  accessibilityLabel?: string;
};

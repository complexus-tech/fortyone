export type NotificationControlsProps = {
  enabled: boolean;
  enabling: boolean;
  testing: boolean;
  unavailable: boolean;
  onEnabledChange: (enabled: boolean) => void;
  onTest: () => void;
};

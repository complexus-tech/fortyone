import type { TrackingEvent } from "../constants/events";

export type Properties = Record<string, unknown> | undefined;

export type Analytics = {
  track: (eventName: TrackingEvent, properties?: Properties) => void;
  startSessionRecording: () => void;
  stopSessionRecording: () => void;
  identify: (
    userId?: string,
    properties?: Properties,
    propertiesToSetOnce?: Properties,
  ) => void;
  logout: (reset_device_id?: boolean | undefined) => void;
};

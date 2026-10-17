let context: AudioContext | null = null;

export const canPlayChatSound = () => context?.state === "running";

/** A user gesture unlocks audio; an incoming event never prompts for permission. */
export const enableChatSound = async () => {
  if (
    typeof window === "undefined" ||
    typeof window.AudioContext !== "function"
  )
    return false;
  context ??= new AudioContext();
  if (context.state === "suspended") await context.resume();
  return context.state === "running";
};

export const playChatSound = () => {
  if (!context || context.state !== "running") return false;
  const now = context.currentTime;
  const gain = context.createGain();
  gain.connect(context.destination);
  gain.gain.setValueAtTime(0, now);
  gain.gain.linearRampToValueAtTime(0.07, now + 0.015);
  gain.gain.exponentialRampToValueAtTime(0.001, now + 0.24);
  const oscillator = context.createOscillator();
  oscillator.type = "sine";
  oscillator.frequency.setValueAtTime(880, now);
  oscillator.frequency.setValueAtTime(660, now + 0.1);
  oscillator.connect(gain);
  oscillator.onended = () => {
    oscillator.disconnect();
    gain.disconnect();
  };
  oscillator.start(now);
  oscillator.stop(now + 0.25);
  return true;
};

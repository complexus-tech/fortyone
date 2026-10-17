export type AudioWaveformData = {
  peaks: readonly number[];
  duration: number;
};

type WaveformSource = {
  id: string;
  url: string;
  durationSeconds?: number;
};

type Subscriber = {
  signal: AbortSignal;
  onAbort: () => void;
  resolve: (data: AudioWaveformData | null) => void;
};

type WaveformJob = {
  key: string;
  source: WaveformSource;
  generation: number;
  controller: AbortController;
  subscribers: Set<Subscriber>;
};

const MAX_BYTES = 20 * 1024 * 1024;
const MAX_DURATION_SECONDS = 5 * 60;
const PEAK_COUNT = 48;
const CACHE_LIMIT = 64;
const FAILURE_TTL_MS = 60_000;
const DOWNLOAD_TIMEOUT_MS = 15_000;
const DECODE_SAMPLE_RATE = 8_000;

const readAudioBytes = async (url: string, signal: AbortSignal) => {
  const response = await fetch(url, { signal, credentials: "same-origin" });
  const contentLength = Number(response.headers.get("content-length"));
  if (!response.ok || !response.body || contentLength > MAX_BYTES) {
    await response.body?.cancel();
    return null;
  }

  const reader = response.body.getReader();
  const chunks: Uint8Array[] = [];
  let length = 0;
  const cancel = () => {
    // Cancellation is best effort: a failed stream is already unusable.
    void reader.cancel().catch(() => undefined);
  };
  signal.addEventListener("abort", cancel, { once: true });

  try {
    while (!signal.aborted) {
      // eslint-disable-next-line no-await-in-loop -- Read sequentially to enforce the byte budget before accepting another chunk.
      const chunk = await reader.read();
      if (chunk.done) break;
      length += chunk.value.byteLength;
      if (length > MAX_BYTES) {
        // eslint-disable-next-line no-await-in-loop -- Stop the stream before releasing its reader.
        await reader.cancel();
        return null;
      }
      chunks.push(chunk.value);
    }
    if (signal.aborted || length === 0) return null;

    const bytes = new Uint8Array(length);
    let offset = 0;
    for (const chunk of chunks) {
      bytes.set(chunk, offset);
      offset += chunk.byteLength;
    }
    return bytes.buffer;
  } finally {
    signal.removeEventListener("abort", cancel);
    reader.releaseLock();
  }
};

const extractWaveform = (buffer: AudioBuffer): AudioWaveformData | null => {
  if (
    !Number.isFinite(buffer.duration) ||
    buffer.duration <= 0 ||
    buffer.duration > MAX_DURATION_SECONDS ||
    buffer.numberOfChannels > 8 ||
    buffer.numberOfChannels < 1 ||
    buffer.length === 0
  ) {
    return null;
  }

  const peaks = Array.from(
    { length: Math.min(PEAK_COUNT, buffer.length) },
    () => 0,
  );
  for (let channel = 0; channel < buffer.numberOfChannels; channel++) {
    const samples = buffer.getChannelData(channel);
    for (let bin = 0; bin < peaks.length; bin++) {
      const start = Math.floor((bin * samples.length) / peaks.length);
      const end = Math.floor(((bin + 1) * samples.length) / peaks.length);
      let peak = peaks[bin];
      for (let index = start; index < end; index++) {
        peak = Math.max(peak, Math.abs(samples[index]));
      }
      peaks[bin] = Number.isFinite(peak) ? peak : 0;
    }
  }
  const largest = Math.max(...peaks);
  return {
    peaks: largest > 0 ? peaks.map((peak) => peak / largest) : peaks,
    duration: buffer.duration,
  };
};

/** Conversation-scoped cache. Encoded bytes and decoded PCM never enter it. */
export const createAudioWaveformStore = () => {
  const cache = new Map<
    string,
    { data: AudioWaveformData | null; expiresAt: number }
  >();
  const jobs = new Map<string, WaveformJob>();
  const queue: WaveformJob[] = [];
  let generation = 0;
  let running = false;

  const settle = (job: WaveformJob, data: AudioWaveformData | null) => {
    for (const subscriber of job.subscribers) {
      subscriber.signal.removeEventListener("abort", subscriber.onAbort);
      subscriber.resolve(data);
    }
    job.subscribers.clear();
    if (jobs.get(job.key) === job) jobs.delete(job.key);
  };

  const remember = (key: string, data: AudioWaveformData | null) => {
    cache.delete(key);
    cache.set(key, {
      data,
      expiresAt: data ? Infinity : Date.now() + FAILURE_TTL_MS,
    });
    if (cache.size > CACHE_LIMIT) {
      const oldest = cache.keys().next().value;
      if (oldest !== undefined) cache.delete(oldest);
    }
  };

  const decode = async (
    job: WaveformJob,
  ): Promise<AudioWaveformData | null> => {
    if (
      typeof OfflineAudioContext === "undefined" ||
      (job.source.durationSeconds ?? 0) > MAX_DURATION_SECONDS
    ) {
      return null;
    }

    const timeout = setTimeout(() => {
      job.controller.abort();
    }, DOWNLOAD_TIMEOUT_MS);
    let bytes: ArrayBuffer | null;
    try {
      bytes = await readAudioBytes(job.source.url, job.controller.signal);
    } finally {
      clearTimeout(timeout);
    }
    if (!bytes || job.controller.signal.aborted) return null;

    // Offline contexts do not open an audio device or expose close(). The tiny
    // render length is unused; decodeAudioData resamples to this lower rate.
    const decoder = new OfflineAudioContext(1, 1, DECODE_SAMPLE_RATE);
    const buffer = await decoder.decodeAudioData(bytes);
    // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition -- Cancellation can change while the browser decodes asynchronously.
    if (job.controller.signal.aborted) return null;
    return extractWaveform(buffer);
  };

  const drain = async () => {
    if (running) return;
    running = true;
    try {
      let job = queue.shift();
      while (job) {
        if (job.subscribers.size > 0 && job.generation === generation) {
          let data: AudioWaveformData | null = null;
          try {
            // eslint-disable-next-line no-await-in-loop -- Keep at most one fetch/decode active for the conversation.
            data = await decode(job);
          } catch {
            // An optional visual must not affect playback or transcription.
          }
          if (job.generation === generation && job.subscribers.size > 0) {
            remember(job.key, data);
            settle(job, data);
          }
        }
        job = queue.shift();
      }
    } finally {
      running = false;
    }
  };

  return {
    load(source: WaveformSource, signal: AbortSignal) {
      if (signal.aborted) return Promise.resolve(null);
      const key = JSON.stringify([source.id, source.url]);
      const cached = cache.get(key);
      if (cached && cached.expiresAt > Date.now()) {
        cache.delete(key);
        cache.set(key, cached);
        return Promise.resolve(cached.data);
      }
      if (cached) cache.delete(key);

      let job = jobs.get(key);
      if (!job) {
        job = {
          key,
          source,
          generation,
          controller: new AbortController(),
          subscribers: new Set(),
        };
        jobs.set(key, job);
        queue.push(job);
      }
      const subscribedJob = job;
      const result = new Promise<AudioWaveformData | null>((resolve) => {
        const subscriber: Subscriber = {
          signal,
          resolve,
          onAbort: () => {
            signal.removeEventListener("abort", subscriber.onAbort);
            subscribedJob.subscribers.delete(subscriber);
            resolve(null);
            if (subscribedJob.subscribers.size === 0) {
              subscribedJob.controller.abort();
              if (jobs.get(key) === subscribedJob) jobs.delete(key);
            }
          },
        };
        subscribedJob.subscribers.add(subscriber);
        signal.addEventListener("abort", subscriber.onAbort, { once: true });
      });
      void drain();
      return result;
    },
    destroy() {
      generation++;
      cache.clear();
      queue.length = 0;
      for (const job of jobs.values()) {
        job.controller.abort();
        settle(job, null);
      }
      jobs.clear();
      // decodeAudioData cannot be aborted. Keep its semaphore until it settles,
      // discard its result, then allow newly mounted consumers to use the store.
    },
  };
};

export type AudioWaveformStore = ReturnType<typeof createAudioWaveformStore>;

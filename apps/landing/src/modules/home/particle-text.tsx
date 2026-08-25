"use client";

import type { CSSProperties, ReactNode } from "react";
import { useEffect, useRef } from "react";
import { cn } from "lib";
import styles from "./particle-text.module.css";

type Particle = {
  driftAmplitude: number;
  driftPhase: number;
  driftSpeed: number;
  entranceOffset: number;
  originX: number;
  originY: number;
  radius: number;
  targetX: number;
  targetY: number;
  velocityX: number;
  velocityY: number;
  x: number;
  y: number;
};

type ParticleTextProps = {
  children: string;
  className?: string;
  entranceDelay?: number;
  offsetX?: number;
  offsetY?: number;
  style?: CSSProperties;
  tone?: "danger" | "primary" | "success";
};

type ParticleVectorProps = Omit<ParticleTextProps, "children"> & {
  children: ReactNode;
  sourceKey: string;
};

type ParticleEffectProps = Omit<ParticleTextProps, "children"> & {
  children: ReactNode;
  sourceKey: string;
  sourceText?: string;
  sourceType: "text" | "vector";
};

const CANVAS_PADDING = 40;
const COMPACT_VIEWPORT_MEDIA_QUERY = "(max-width: 47.999rem)";
const DRIFT_BLEND_DURATION = 280;
const DRIFT_AMPLITUDE_MAX = 1.45;
const DRIFT_AMPLITUDE_MIN = 0.65;
const DRIFT_SPEED_MAX = 0.001;
const DRIFT_SPEED_MIN = 0.00055;
const LIGHT_MODE_RADIUS_SCALE = 1.12;
const MAX_PIXEL_RATIO = 2;
const MIN_PARTICLE_RADIUS = 0.9;
const PARTICLE_ALPHA_THRESHOLD = 96;
const POINTER_RADIUS_MULTIPLIER = 0.82;
const PARTICLE_ASSEMBLY_DURATION = 880;
const PARTICLE_ASSEMBLY_STAGGER = 140;

const clampProgress = (value: number) => Math.min(Math.max(value, 0), 1);

const sampleCubicBezier = (time: number, point1: number, point2: number) => {
  const inverseTime = 1 - time;

  return (
    3 * inverseTime * inverseTime * time * point1 +
    3 * inverseTime * time * time * point2 +
    time * time * time
  );
};

const sampleCubicBezierDerivative = (
  time: number,
  point1: number,
  point2: number,
) => {
  const inverseTime = 1 - time;

  return (
    3 * inverseTime * inverseTime * point1 +
    6 * inverseTime * time * (point2 - point1) +
    3 * time * time * (1 - point2)
  );
};

const easeParticleAssembly = (progress: number) => {
  const clampedProgress = clampProgress(progress);
  let time = clampedProgress;

  for (let iteration = 0; iteration < 5; iteration += 1) {
    const error = sampleCubicBezier(time, 0.23, 0.32) - clampedProgress;
    const derivative = sampleCubicBezierDerivative(time, 0.23, 0.32);

    if (Math.abs(derivative) < 0.0001) break;
    time = clampProgress(time - error / derivative);
  }

  return sampleCubicBezier(time, 1, 1);
};

const getParticleRadius = (x: number, y: number, sampleStep: number) => {
  const variation = ((x * 17 + y * 29) % 11) / 10;
  return Math.max(MIN_PARTICLE_RADIUS, sampleStep * (0.22 + variation * 0.12));
};

const ParticleEffect = ({
  children,
  className,
  entranceDelay = 0,
  offsetX = 0,
  offsetY = 0,
  sourceKey,
  sourceText,
  sourceType,
  style,
  tone = "primary",
}: ParticleEffectProps) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const rootRef = useRef<HTMLSpanElement>(null);
  const sourceRef = useRef<HTMLSpanElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    const root = rootRef.current;
    const source = sourceRef.current;

    if (!canvas || !root || !source) return;

    const context = canvas.getContext("2d");

    if (!context) return;

    const compactViewportQuery = window.matchMedia(
      COMPACT_VIEWPORT_MEDIA_QUERY,
    );
    const reducedMotionQuery = window.matchMedia(
      "(prefers-reduced-motion: reduce)",
    );
    const finePointerQuery = window.matchMedia(
      "(hover: hover) and (pointer: fine)",
    );
    let animationFrame: number | undefined;
    let canBuildParticles = false;
    let color = "currentColor";
    let entranceEndsAt = 0;
    let entranceStartsAt = 0;
    let entranceIsPending = false;
    let hasPlayedEntrance = false;
    let isDisposed = false;
    let isVisible = false;
    let lastHeight = 0;
    let lastWidth = 0;
    let particles: Particle[] = [];
    let pointerX = Number.POSITIVE_INFINITY;
    let pointerY = Number.POSITIVE_INFINITY;
    let pointerRadius = 48;

    const drawParticles = (interactive: boolean, elapsedTime = 0) => {
      context.clearRect(0, 0, canvas.width, canvas.height);
      context.fillStyle = color;
      context.globalAlpha = 1;
      context.beginPath();

      particles.forEach((particle) => {
        if (interactive) {
          const entranceIsActive = elapsedTime < entranceEndsAt;

          if (entranceIsActive) {
            const entranceProgress =
              (elapsedTime - entranceStartsAt - particle.entranceOffset) /
              PARTICLE_ASSEMBLY_DURATION;
            const easedProgress = easeParticleAssembly(entranceProgress);

            particle.x =
              particle.originX +
              (particle.targetX - particle.originX) * easedProgress;
            particle.y =
              particle.originY +
              (particle.targetY - particle.originY) * easedProgress;
            particle.velocityX = 0;
            particle.velocityY = 0;
          } else {
            const driftBlend = clampProgress(
              (elapsedTime - entranceEndsAt) / DRIFT_BLEND_DURATION,
            );
            const driftX =
              Math.sin(
                elapsedTime * particle.driftSpeed + particle.driftPhase,
              ) *
              particle.driftAmplitude *
              driftBlend;
            const driftY =
              Math.cos(
                elapsedTime * particle.driftSpeed * 0.82 +
                  particle.driftPhase * 1.37,
              ) *
              particle.driftAmplitude *
              driftBlend;

            if (finePointerQuery.matches) {
              const deltaX = particle.x - pointerX;
              const deltaY = particle.y - pointerY;
              const distance = Math.hypot(deltaX, deltaY);

              if (distance < pointerRadius) {
                const safeDistance = Math.max(distance, 0.1);
                const force = (1 - safeDistance / pointerRadius) * 2.1;
                particle.velocityX += (deltaX / safeDistance) * force;
                particle.velocityY += (deltaY / safeDistance) * force;
              }
            }

            particle.velocityX +=
              (particle.targetX + driftX - particle.x) * 0.075;
            particle.velocityY +=
              (particle.targetY + driftY - particle.y) * 0.075;
            particle.velocityX *= 0.84;
            particle.velocityY *= 0.84;
            particle.x += particle.velocityX;
            particle.y += particle.velocityY;
          }
        } else {
          particle.x = particle.targetX;
          particle.y = particle.targetY;
        }

        context.moveTo(particle.x + particle.radius, particle.y);
        context.arc(particle.x, particle.y, particle.radius, 0, Math.PI * 2);
      });

      context.fill();
    };

    const animate = (elapsedTime: number) => {
      drawParticles(!reducedMotionQuery.matches, elapsedTime);
      animationFrame = window.requestAnimationFrame(animate);
    };

    const stopAnimation = () => {
      if (animationFrame === undefined) return;
      window.cancelAnimationFrame(animationFrame);
      animationFrame = undefined;
    };

    const startAnimation = () => {
      if (
        animationFrame !== undefined ||
        compactViewportQuery.matches ||
        !isVisible ||
        particles.length === 0 ||
        document.visibilityState === "hidden"
      ) {
        return;
      }

      if (reducedMotionQuery.matches) {
        drawParticles(false);
        return;
      }

      if (entranceIsPending) {
        entranceStartsAt = performance.now() + entranceDelay;
        entranceEndsAt =
          entranceStartsAt +
          PARTICLE_ASSEMBLY_DURATION +
          PARTICLE_ASSEMBLY_STAGGER +
          18;
        entranceIsPending = false;
        hasPlayedEntrance = true;

        particles.forEach((particle) => {
          particle.x = particle.originX;
          particle.y = particle.originY;
          particle.velocityX = 0;
          particle.velocityY = 0;
        });
      }

      animationFrame = window.requestAnimationFrame(animate);
    };

    const resetParticles = () => {
      stopAnimation();
      particles = [];
      entranceIsPending = false;
      lastHeight = 0;
      lastWidth = 0;
      canvas.height = 0;
      canvas.width = 0;
      delete root.dataset.ready;
    };

    const buildParticles = (force = false) => {
      if (!canBuildParticles) return;
      if (compactViewportQuery.matches) {
        resetParticles();
        return;
      }

      const bounds = source.getBoundingClientRect();
      const computedStyle = window.getComputedStyle(source);
      const rootStyle = window.getComputedStyle(root);
      const height = Math.ceil(bounds.height);
      const width = Math.ceil(bounds.width);

      if (height === 0 || width === 0) return;
      if (!force && height === lastHeight && width === lastWidth) return;

      lastHeight = height;
      lastWidth = width;

      const pixelRatio = Math.min(
        window.devicePixelRatio || 1,
        MAX_PIXEL_RATIO,
      );
      const canvasHeight = height + CANVAS_PADDING * 2;
      const canvasWidth = width + CANVAS_PADDING * 2;

      canvas.height = Math.ceil(canvasHeight * pixelRatio);
      canvas.width = Math.ceil(canvasWidth * pixelRatio);
      canvas.style.height = `${canvasHeight}px`;
      canvas.style.left = `${-CANVAS_PADDING}px`;
      canvas.style.top = `${-CANVAS_PADDING}px`;
      canvas.style.width = `${canvasWidth}px`;
      context.setTransform(pixelRatio, 0, 0, pixelRatio, 0, 0);
      context.imageSmoothingEnabled = true;

      const samplingCanvas = document.createElement("canvas");
      const samplingContext = samplingCanvas.getContext("2d", {
        willReadFrequently: true,
      });

      if (!samplingContext) return;

      samplingCanvas.height = canvasHeight;
      samplingCanvas.width = canvasWidth;
      samplingContext.fillStyle = "#000";

      if (sourceType === "text" && sourceText !== undefined) {
        samplingContext.font = [
          computedStyle.fontStyle,
          computedStyle.fontVariant,
          computedStyle.fontWeight,
          computedStyle.fontSize,
          computedStyle.fontFamily,
        ].join(" ");
        samplingContext.textBaseline = "alphabetic";

        const metrics = samplingContext.measureText(sourceText);
        const glyphHeight =
          metrics.actualBoundingBoxAscent + metrics.actualBoundingBoxDescent;
        const baseline =
          CANVAS_PADDING +
          (height - glyphHeight) / 2 +
          metrics.actualBoundingBoxAscent;

        samplingContext.fillText(sourceText, CANVAS_PADDING, baseline);
      } else {
        const svg = source.querySelector("svg");
        const viewBox = svg?.viewBox.baseVal;

        if (!svg || !viewBox || viewBox.width === 0 || viewBox.height === 0) {
          return;
        }

        const scale = Math.min(width / viewBox.width, height / viewBox.height);
        const renderedWidth = viewBox.width * scale;
        const renderedHeight = viewBox.height * scale;

        samplingContext.save();
        samplingContext.translate(
          CANVAS_PADDING + (width - renderedWidth) / 2,
          CANVAS_PADDING + (height - renderedHeight) / 2,
        );
        samplingContext.scale(scale, scale);
        samplingContext.translate(-viewBox.x, -viewBox.y);

        svg.querySelectorAll("path").forEach((path) => {
          const pathData = path.getAttribute("d");

          if (pathData) samplingContext.fill(new Path2D(pathData));
        });
        samplingContext.restore();
      }

      const pixels = samplingContext.getImageData(
        0,
        0,
        canvasWidth,
        canvasHeight,
      ).data;
      const sourceSize =
        sourceType === "text"
          ? Number.parseFloat(computedStyle.fontSize)
          : height;
      const radiusScale = document.documentElement.classList.contains("dark")
        ? 1
        : LIGHT_MODE_RADIUS_SCALE;
      const sampleStep = Math.max(2, Math.round(sourceSize / 26));
      const nextParticles: Particle[] = [];
      const shouldAnimateEntrance =
        !hasPlayedEntrance && !reducedMotionQuery.matches;

      for (
        let y = Math.floor(sampleStep / 2);
        y < canvasHeight;
        y += sampleStep
      ) {
        for (
          let x = Math.floor(sampleStep / 2);
          x < canvasWidth;
          x += sampleStep
        ) {
          const alpha = pixels[(y * canvasWidth + x) * 4 + 3];

          if (alpha < PARTICLE_ALPHA_THRESHOLD) continue;

          const driftVariation = ((x * 31 + y * 19) % 101) / 100;
          const targetX = x + offsetX;
          const targetY = y + offsetY;
          const horizontalProgress = clampProgress(
            (x - CANVAS_PADDING) / width,
          );
          const flowPhase = ((x * 11 + y * 17) % 360) * (Math.PI / 180);
          const flowDistance = sourceSize * (0.3 + driftVariation * 0.12);
          const originX =
            targetX - flowDistance + Math.cos(flowPhase) * sourceSize * 0.08;
          const originY =
            targetY +
            sourceSize * 0.12 +
            Math.sin(flowPhase) * sourceSize * 0.18;

          nextParticles.push({
            driftAmplitude:
              DRIFT_AMPLITUDE_MIN +
              driftVariation * (DRIFT_AMPLITUDE_MAX - DRIFT_AMPLITUDE_MIN),
            driftPhase: ((x * 13 + y * 7) % 360) * (Math.PI / 180),
            driftSpeed:
              DRIFT_SPEED_MIN +
              driftVariation * (DRIFT_SPEED_MAX - DRIFT_SPEED_MIN),
            entranceOffset:
              horizontalProgress * PARTICLE_ASSEMBLY_STAGGER +
              driftVariation * 18,
            originX: shouldAnimateEntrance ? originX : targetX,
            originY: shouldAnimateEntrance ? originY : targetY,
            radius: getParticleRadius(x, y, sampleStep) * radiusScale,
            targetX,
            targetY,
            velocityX: 0,
            velocityY: 0,
            x: shouldAnimateEntrance ? originX : targetX,
            y: shouldAnimateEntrance ? originY : targetY,
          });
        }
      }

      particles = nextParticles;
      entranceStartsAt = 0;
      entranceEndsAt = 0;
      entranceIsPending = shouldAnimateEntrance;

      if (!shouldAnimateEntrance) hasPlayedEntrance = true;

      pointerRadius = sourceSize * POINTER_RADIUS_MULTIPLIER;
      color = rootStyle.color;
      root.dataset.ready = "true";

      if (!shouldAnimateEntrance || isVisible) {
        drawParticles(false);
      }

      startAnimation();
    };

    const handlePointerMove = (event: PointerEvent) => {
      const bounds = root.getBoundingClientRect();
      pointerX = event.clientX - bounds.left + CANVAS_PADDING;
      pointerY = event.clientY - bounds.top + CANVAS_PADDING;
    };

    const handlePointerLeave = () => {
      pointerX = Number.POSITIVE_INFINITY;
      pointerY = Number.POSITIVE_INFINITY;
    };

    const handleVisibilityChange = () => {
      if (document.visibilityState === "hidden") {
        stopAnimation();
      } else {
        startAnimation();
      }
    };

    const handleReducedMotionChange = () => {
      buildParticles(true);
    };
    const handleCompactViewportChange = () => {
      buildParticles(true);
    };

    const resizeObserver = new ResizeObserver(() => {
      buildParticles();
    });
    const themeObserver = new MutationObserver(() => {
      buildParticles(true);
    });
    const intersectionObserver = new IntersectionObserver(([entry]) => {
      isVisible = entry?.isIntersecting ?? false;

      if (isVisible) {
        startAnimation();
      } else {
        stopAnimation();
      }
    });

    themeObserver.observe(document.documentElement, {
      attributeFilter: ["class"],
      attributes: true,
    });
    intersectionObserver.observe(root);
    root.addEventListener("pointermove", handlePointerMove);
    root.addEventListener("pointerleave", handlePointerLeave);
    document.addEventListener("visibilitychange", handleVisibilityChange);
    compactViewportQuery.addEventListener(
      "change",
      handleCompactViewportChange,
    );
    reducedMotionQuery.addEventListener("change", handleReducedMotionChange);

    void document.fonts.ready.then(() => {
      if (isDisposed) return;

      canBuildParticles = true;
      buildParticles(true);
      resizeObserver.observe(source);
    });

    return () => {
      isDisposed = true;
      stopAnimation();
      resizeObserver.disconnect();
      themeObserver.disconnect();
      intersectionObserver.disconnect();
      root.removeEventListener("pointermove", handlePointerMove);
      root.removeEventListener("pointerleave", handlePointerLeave);
      document.removeEventListener("visibilitychange", handleVisibilityChange);
      compactViewportQuery.removeEventListener(
        "change",
        handleCompactViewportChange,
      );
      reducedMotionQuery.removeEventListener(
        "change",
        handleReducedMotionChange,
      );
    };
  }, [entranceDelay, offsetX, offsetY, sourceKey, sourceText, sourceType]);

  return (
    <span
      className={cn(styles.root, className)}
      data-tone={tone}
      ref={rootRef}
      style={style}
    >
      <span
        className={cn(styles.source, {
          [styles.vectorSource]: sourceType === "vector",
        })}
        ref={sourceRef}
        style={{ left: offsetX, top: offsetY }}
      >
        {children}
      </span>
      <canvas aria-hidden="true" className={styles.canvas} ref={canvasRef} />
    </span>
  );
};

export const ParticleText = ({ children, ...props }: ParticleTextProps) => (
  <ParticleEffect
    {...props}
    sourceKey={children}
    sourceText={children}
    sourceType="text"
  >
    {children}
  </ParticleEffect>
);

export const ParticleVector = ({
  children,
  sourceKey,
  ...props
}: ParticleVectorProps) => (
  <ParticleEffect {...props} sourceKey={sourceKey} sourceType="vector">
    {children}
  </ParticleEffect>
);

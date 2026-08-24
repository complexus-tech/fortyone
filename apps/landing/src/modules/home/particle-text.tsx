"use client";

import type { CSSProperties } from "react";
import { useEffect, useRef } from "react";
import { cn } from "lib";
import styles from "./particle-text.module.css";

type Particle = {
  driftAmplitude: number;
  driftPhase: number;
  driftSpeed: number;
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
  offsetX?: number;
  offsetY?: number;
  style?: CSSProperties;
  tone?: "danger" | "primary" | "success";
};

const CANVAS_PADDING = 10;
const DRIFT_AMPLITUDE_MAX = 1.1;
const DRIFT_AMPLITUDE_MIN = 0.45;
const DRIFT_SPEED_MAX = 0.001;
const DRIFT_SPEED_MIN = 0.00055;
const LIGHT_MODE_RADIUS_SCALE = 1.12;
const MAX_PIXEL_RATIO = 2;
const MIN_PARTICLE_RADIUS = 0.9;
const PARTICLE_ALPHA_THRESHOLD = 96;
const POINTER_RADIUS_MULTIPLIER = 0.82;

const getParticleRadius = (x: number, y: number, sampleStep: number) => {
  const variation = ((x * 17 + y * 29) % 11) / 10;
  return Math.max(MIN_PARTICLE_RADIUS, sampleStep * (0.22 + variation * 0.12));
};

export const ParticleText = ({
  children,
  className,
  offsetX = 0,
  offsetY = 0,
  style,
  tone = "primary",
}: ParticleTextProps) => {
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

    const reducedMotionQuery = window.matchMedia(
      "(prefers-reduced-motion: reduce)",
    );
    let animationFrame: number | undefined;
    let color = "currentColor";
    let isDisposed = false;
    let isVisible = true;
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
          const driftX =
            Math.sin(elapsedTime * particle.driftSpeed + particle.driftPhase) *
            particle.driftAmplitude;
          const driftY =
            Math.cos(
              elapsedTime * particle.driftSpeed * 0.82 +
                particle.driftPhase * 1.37,
            ) * particle.driftAmplitude;
          const deltaX = particle.x - pointerX;
          const deltaY = particle.y - pointerY;
          const distance = Math.hypot(deltaX, deltaY);

          if (distance < pointerRadius) {
            const safeDistance = Math.max(distance, 0.1);
            const force = (1 - safeDistance / pointerRadius) * 2.1;
            particle.velocityX += (deltaX / safeDistance) * force;
            particle.velocityY += (deltaY / safeDistance) * force;
          }

          particle.velocityX +=
            (particle.targetX + driftX - particle.x) * 0.075;
          particle.velocityY +=
            (particle.targetY + driftY - particle.y) * 0.075;
          particle.velocityX *= 0.84;
          particle.velocityY *= 0.84;
          particle.x += particle.velocityX;
          particle.y += particle.velocityY;
        } else {
          particle.x = particle.targetX;
          particle.y = particle.targetY;
        }

        context.moveTo(particle.x + particle.radius, particle.y);
        context.arc(particle.x, particle.y, particle.radius, 0, Math.PI * 2);
      });

      context.fill();
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
        !isVisible ||
        document.visibilityState === "hidden"
      ) {
        return;
      }

      if (reducedMotionQuery.matches) {
        drawParticles(false);
        return;
      }

      animationFrame = window.requestAnimationFrame(animate);
    };

    const buildParticles = () => {
      const bounds = source.getBoundingClientRect();
      const computedStyle = window.getComputedStyle(source);
      const rootStyle = window.getComputedStyle(root);
      const height = Math.ceil(bounds.height);
      const width = Math.ceil(bounds.width);

      if (height === 0 || width === 0) return;

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
      samplingContext.font = [
        computedStyle.fontStyle,
        computedStyle.fontVariant,
        computedStyle.fontWeight,
        computedStyle.fontSize,
        computedStyle.fontFamily,
      ].join(" ");
      samplingContext.fillStyle = "#000";
      samplingContext.textBaseline = "alphabetic";

      const metrics = samplingContext.measureText(children);
      const glyphHeight =
        metrics.actualBoundingBoxAscent + metrics.actualBoundingBoxDescent;
      const baseline =
        CANVAS_PADDING +
        (height - glyphHeight) / 2 +
        metrics.actualBoundingBoxAscent;

      samplingContext.fillText(children, CANVAS_PADDING, baseline);

      const pixels = samplingContext.getImageData(
        0,
        0,
        canvasWidth,
        canvasHeight,
      ).data;
      const fontSize = Number.parseFloat(computedStyle.fontSize);
      const radiusScale = document.documentElement.classList.contains("dark")
        ? 1
        : LIGHT_MODE_RADIUS_SCALE;
      const sampleStep = Math.max(2, Math.round(fontSize / 26));
      const nextParticles: Particle[] = [];

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

          nextParticles.push({
            driftAmplitude:
              DRIFT_AMPLITUDE_MIN +
              driftVariation * (DRIFT_AMPLITUDE_MAX - DRIFT_AMPLITUDE_MIN),
            driftPhase: ((x * 13 + y * 7) % 360) * (Math.PI / 180),
            driftSpeed:
              DRIFT_SPEED_MIN +
              driftVariation * (DRIFT_SPEED_MAX - DRIFT_SPEED_MIN),
            radius: getParticleRadius(x, y, sampleStep) * radiusScale,
            targetX: x + offsetX,
            targetY: y + offsetY,
            velocityX: 0,
            velocityY: 0,
            x: x + offsetX,
            y: y + offsetY,
          });
        }
      }

      particles = nextParticles;
      pointerRadius = fontSize * POINTER_RADIUS_MULTIPLIER;
      color = rootStyle.color;
      root.dataset.ready = "true";
      drawParticles(false);
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

    const resizeObserver = new ResizeObserver(buildParticles);
    const themeObserver = new MutationObserver(buildParticles);
    const intersectionObserver = new IntersectionObserver(([entry]) => {
      isVisible = entry?.isIntersecting ?? false;

      if (isVisible) {
        startAnimation();
      } else {
        stopAnimation();
      }
    });

    resizeObserver.observe(source);
    themeObserver.observe(document.documentElement, {
      attributeFilter: ["class"],
      attributes: true,
    });
    intersectionObserver.observe(root);
    root.addEventListener("pointermove", handlePointerMove);
    root.addEventListener("pointerleave", handlePointerLeave);
    document.addEventListener("visibilitychange", handleVisibilityChange);
    reducedMotionQuery.addEventListener("change", buildParticles);

    void document.fonts.ready.then(() => {
      if (!isDisposed) buildParticles();
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
      reducedMotionQuery.removeEventListener("change", buildParticles);
    };
  }, [children, offsetX, offsetY]);

  return (
    <span
      className={cn(styles.root, className)}
      data-tone={tone}
      ref={rootRef}
      style={style}
    >
      <span
        className={styles.source}
        ref={sourceRef}
        style={{ left: offsetX, top: offsetY }}
      >
        {children}
      </span>
      <canvas aria-hidden="true" className={styles.canvas} ref={canvasRef} />
    </span>
  );
};

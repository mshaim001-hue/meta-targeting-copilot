import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatReach(lower?: number | null, upper?: number | null) {
  if (lower == null && upper == null) return "Estimate unavailable"
  const fmt = (n: number) =>
    n >= 1_000_000 ? `${(n / 1_000_000).toFixed(1)}M` : n >= 1_000 ? `${Math.round(n / 1000)}k` : String(n)
  if (lower != null && upper != null) return `${fmt(lower)} – ${fmt(upper)}`
  return fmt(lower ?? upper ?? 0)
}

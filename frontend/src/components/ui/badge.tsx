import { cn } from "@/lib/utils"

export function Badge({ className, ...props }: React.ComponentProps<"span">) {
  return (
    <span
      className={cn("inline-flex items-center rounded-full border border-rule px-2 py-0.5 text-[11px] font-medium tracking-wide", className)}
      {...props}
    />
  )
}

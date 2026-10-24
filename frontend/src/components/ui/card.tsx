import { cn } from "@/lib/utils"

export function Card({ className, ...props }: React.ComponentProps<"div">) {
  return <div className={cn("rounded-xl border border-rule bg-white/55 p-5 shadow-[0_1px_0_rgba(28,25,20,0.04)]", className)} {...props} />
}

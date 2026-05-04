import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const treeVariants = cva(
  "font-mono whitespace-pre leading-[1.0] tracking-tighter select-none",
  {
    variants: {
      size: {
        sm: "text-[6px]",
        md: "text-[9px]",
        lg: "text-[12px]",
        xl: "text-[16px]",
      },
    },
    defaultVariants: {
      size: "md",
    },
  },
)

interface TreeWordmarkProps extends VariantProps<typeof treeVariants> {
  className?: string
}

export function TreeWordmark({ size, className }: TreeWordmarkProps) {
  return (
    <div className={cn(treeVariants({ size }), className)}>
      {"████████╗██████╗ ███████╗███████╗"}
      <br />
      {"╚══██╔══╝██╔══██╗██╔════╝██╔════╝"}
      <br />
      {"   ██║   ██████╔╝█████╗  █████╗  "}
      <br />
      {"   ██║   ██╔══██╗██╔══╝  ██╔══╝  "}
      <br />
      {"   ██║   ██║  ██║███████╗███████╗"}
      <br />
      {"   ╚═╝   ╚═╝  ╚═╝╚══════╝╚══════╝"}
    </div>
  )
}

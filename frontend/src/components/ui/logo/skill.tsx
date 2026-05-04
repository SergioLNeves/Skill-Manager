import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const skillVariants = cva(
  "font-mono whitespace-pre leading-[1.1] tracking-tighter select-none",
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
      size: "lg",
    },
  },
)

interface SkillWordmarkProps extends VariantProps<typeof skillVariants> {
  className?: string
}

export function SkillWordmark({ size, className }: SkillWordmarkProps) {
  return (
    <div className={cn(skillVariants({ size }), className)}>
      {"███████╗██╗  ██╗██╗██╗      ██╗     "}
      <br />
      {"██╔════╝██║ ██╔╝██║██║      ██║     "}
      <br />
      {"███████╗█████╔╝ ██║██║      ██║     "}
      <br />
      {"╚════██║██╔═██╗ ██║██║      ██║     "}
      <br />
      {"███████║██║  ██╗██║███████╗███████╗"}
      <br />
      {"╚══════╝╚═╝  ╚═╝╚═╝╚══════╝╚══════╝"}
    </div>
  )
}

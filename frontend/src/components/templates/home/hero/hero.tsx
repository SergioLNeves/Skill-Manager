import { ArrowRight, BookOpen, Network } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Skill, Tree } from "@/components/ui/logo"

export default function Hero() {
  return (
    <section className="flex flex-col">
      <div className="flex flex-col md:flex-row items-center justify-center gap-4">
        <div className="m-2 flex flex-col gap-2">
          <Skill className="md:ml-8" />
          <Tree />
        </div>

        <div className="flex flex-col gap-4 max-w-lg">
          <h1 className="text-xl md:text-4xl font-normal leading-[1.2] tracking-tighter text-center md:text-start">
            Gerencie suas skills, acompanhe seu progresso e organize seu
            conhecimento técnico{" "}
            <span className="text-muted-foreground">do seu jeito</span>.
          </h1>
        </div>
      </div>

      <div className="flex flex-col md:flex-row justify-center mt-9 gap-3">
        <Button size="lg">
          Explorar skills <ArrowRight className="size-4" />
        </Button>
        <Button size="lg" variant="outline">
          <BookOpen className="size-4" /> Todas as skills
        </Button>
        <Button size="lg" variant="outline">
          <Network className="size-4" /> Visão em grafo
        </Button>
      </div>
    </section>
  )
}

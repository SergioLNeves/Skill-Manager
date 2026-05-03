import { Card, CardTitle, CardHeader, CardContent, CardFooter } from "@/components/ui/card"

interface MetricsProps {
  skills: number
  categories: string[]
  tags: string[]
}

export default function Metrics({ skills, categories = [], tags = [] }: MetricsProps) {
  return (
    <section className="grid md:grid-cols-3 bg-card justify-center items-center border border-primary">
      <Card id="Skills">
        <CardHeader className="uppercase font-thin text-muted-foreground text-xs font-mono tracking-[0.2em]">
          Skills
        </CardHeader>
        <CardContent>
          <CardTitle className="text-4xl leading line-clamp-1">
            {skills ?? 0}
          </CardTitle>
        </CardContent>
        <CardFooter className="flex flex-wrap gap-1 mt-2 font-thin text-muted-foreground">
          Na sua árvore
        </CardFooter>
      </Card>

      <Card id="Categories" className="border-y md:border-x">
        <CardHeader className="uppercase font-thin text-muted-foreground text-xs font-mono tracking-[0.2em]">
          Categorias
        </CardHeader>
        <CardContent className="flex flex-col gap-2">
          <CardTitle className="text-4xl line-clamp-1">
            {categories.length}
          </CardTitle>
        </CardContent>
        <CardFooter className="flex flex-wrap gap-1 mt-2 font-thin text-muted-foreground">
          Navegue por tópico
        </CardFooter>
      </Card>

      <Card id="Tags">
        <CardHeader className="uppercase font-thin text-muted-foreground text-xs font-mono tracking-[0.2em]">
          Tags
        </CardHeader>
        <CardContent className="flex flex-col gap-2">
          <CardTitle className="text-4xl line-clamp-1">{tags.length}</CardTitle>
        </CardContent>
        <CardFooter className="flex flex-wrap gap-1 mt-2 font-thin text-muted-foreground">
          Filtre e conecte
        </CardFooter>
      </Card>
    </section>
  )
}

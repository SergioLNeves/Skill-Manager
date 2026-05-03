import { createFileRoute } from '@tanstack/react-router'
import Hero from '@/components/templates/home/hero/hero'
import Install from '@/components/templates/home/install/install'
import Metrics from '@/components/templates/home/metrics/metrics'
import RecentSkills from '@/components/templates/home/recent-skills/recent-skills'
import { metricsMock, recentSkillsMock } from '@/types/metrics'

export const Route = createFileRoute('/')({
  component: Home,
})

function Home() {
  return (
    <div className="flex flex-col container justify-center items-center min-w-screen max-w-md px-4 py-10">
      <div className="flex flex-col max-w-2xl w-full gap-16">
        <Hero />
        <Install />
        <Metrics {...metricsMock} />
        <RecentSkills items={recentSkillsMock} />
      </div>
    </div>
  )
}

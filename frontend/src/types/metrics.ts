import type { SkillItem } from "@/components/templates/home/recent-skills/recent-skills"

export const metricsMock = {
  skills: 24,
  categories: ["Frontend", "Backend", "DevOps", "Mobile", "Design", "Data"],
  tags: ["react", "typescript", "go", "docker", "kubernetes", "sql", "rest", "graphql", "ci/cd", "testing"],
}

export const recentSkillsMock: SkillItem[] = [
  { name: "React Query", tags: ["react", "async", "cache"], category: "Frontend" },
  { name: "Go Routines", tags: ["go", "concurrency"], category: "Backend" },
  { name: "Docker Compose", tags: ["docker", "devops"], category: "DevOps" },
  { name: "Tailwind CSS", tags: ["css", "utility", "design"], category: "Frontend" },
  { name: "PostgreSQL", tags: ["sql", "database"], category: "Backend" },
  { name: "GitHub Actions", tags: ["ci/cd", "automation"], category: "DevOps" },
  { name: "Expo Router", tags: ["react-native", "mobile", "routing"], category: "Mobile" },
]

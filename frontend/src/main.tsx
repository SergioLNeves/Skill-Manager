import React from 'react'
import { createRoot } from 'react-dom/client'
import { createRouter, RouterProvider } from '@tanstack/react-router'
import { routeTree } from './routeTree.gen'
import './index.css'

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

const router = createRouter({ routeTree })
const container = document.getElementById('root')

createRoot(container!).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>
)

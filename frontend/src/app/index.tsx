import { useState } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Sidebar } from '../widgets/sidebar'
import { DashboardPage } from '../pages/dashboard'
import { ProjectsPage } from '../pages/projects'
import { PaymentsPage } from '../pages/payments'
import '../app/styles/index.css'

export function App() {
  const [menuOpen, setMenuOpen] = useState(false)

  return (
    <BrowserRouter>
      <Sidebar isOpen={menuOpen} onOpen={() => setMenuOpen(true)} onClose={() => setMenuOpen(false)} />
      <main className="main">
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/projects" element={<ProjectsPage />} />
          <Route path="/payments" element={<PaymentsPage />} />
        </Routes>
      </main>
    </BrowserRouter>
  )
}

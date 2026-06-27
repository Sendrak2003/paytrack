import { NavLink } from 'react-router-dom'
import { BarChart3, Home, FolderOpen, CreditCard, X, Menu } from 'lucide-react'

interface Props {
  isOpen: boolean
  onOpen: () => void
  onClose: () => void
}

export function Sidebar({ isOpen, onOpen, onClose }: Props) {
  return (
    <>
      {/* Mobile top bar */}
      <div className="mobile-topbar">
        <button className="burger" onClick={onOpen} aria-label="Меню">
          <Menu size={20} />
        </button>
        <span className="mobile-logo"><BarChart3 size={18} /> PayTrack</span>
      </div>

      {/* Backdrop */}
      {isOpen && <div className="sidebar-backdrop" onClick={onClose} />}

      <aside className={`sidebar ${isOpen ? 'open' : ''}`}>
        <div className="sidebar-logo">
          <BarChart3 size={20} /> PayTrack
          <button className="sidebar-close" onClick={onClose} aria-label="Закрыть"><X size={16} /></button>
        </div>
        <nav className="sidebar-nav" onClick={onClose}>
          <NavLink to="/" end className={({ isActive }) => isActive ? 'active' : ''}>
            <span className="nav-icon"><Home size={16} /></span> Дашборд
          </NavLink>
          <NavLink to="/projects" className={({ isActive }) => isActive ? 'active' : ''}>
            <span className="nav-icon"><FolderOpen size={16} /></span> Проекты
          </NavLink>
          <NavLink to="/payments" className={({ isActive }) => isActive ? 'active' : ''}>
            <span className="nav-icon"><CreditCard size={16} /></span> Оплаты
          </NavLink>
        </nav>
      </aside>
    </>
  )
}

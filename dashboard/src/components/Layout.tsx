import { NavLink, Outlet } from 'react-router-dom'

const nav = [
  { to: '/',            label: 'Events' },
  { to: '/deliveries',  label: 'Deliveries' },
  { to: '/endpoints',   label: 'Endpoints' },
]

export function Layout() {
  return (
    <div className="min-h-screen flex flex-col">
      <header className="border-b border-gray-800 bg-gray-950 px-6 py-3 flex items-center gap-8 shrink-0">
        <span className="text-sm font-semibold tracking-tight text-white">
          <span className="text-emerald-400">⬡</span> EventHook
        </span>
        <nav className="flex gap-1">
          {nav.map(({ to, label }) => (
            <NavLink
              key={to}
              to={to}
              end={to === '/'}
              className={({ isActive }) =>
                `px-3 py-1.5 rounded text-sm transition-colors ${
                  isActive
                    ? 'bg-gray-800 text-white'
                    : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800/50'
                }`
              }
            >
              {label}
            </NavLink>
          ))}
        </nav>
      </header>
      <main className="flex-1 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}

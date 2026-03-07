// Fixed-width sidebar with always-visible navigation labels
import { NavLink } from 'react-router';
import { ScrollText, Layers, Server, type LucideIcon } from 'lucide-react';

interface NavItemProps {
  to: string;
  icon: LucideIcon;
  label: string;
}

function NavItem({ to, icon: Icon, label }: NavItemProps) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        `nav-item ${isActive ? 'active' : ''}`
      }
    >
      <span className="nav-icon">
        <Icon size={20} />
      </span>
      <span className="nav-label">{label}</span>
    </NavLink>
  );
}

export function Sidebar() {
  return (
    <aside className="sidebar">
      <div className="logo">
        <span className="logo-icon">
          <img src="/logo.png" alt="Mock Server" width={160} height={160} style={{ borderRadius: 16 }} />
        </span>
        <span className="logo-text">Mock Server</span>
      </div>
      <nav className="flex-1">
        <NavItem to="/logs" icon={ScrollText} label="Logs" />
        <NavItem to="/mocks" icon={Layers} label="Mocks" />
        <NavItem to="/hosts" icon={Server} label="Hosts" />
      </nav>
    </aside>
  );
}


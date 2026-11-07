import { NavLink, Route, Routes } from 'react-router-dom'
import { navigation, paths, sections } from './routes.js'

function HomePage() {
  return (
    <section className="page-card hero">
      <span className="eyebrow">Dobrodošli</span>
      <h1>Prostor za boljši delovni dan.</h1>
      <p>
        DeskBooker bo omogočal pregled razpoložljivih delovnih mest,
        sejnih sob, parkirišč in opreme.
      </p>
      <div className="quick-links">
        {sections.map((section) => (
          <NavLink key={section.path} className="quick-link" to={section.path}>
            <span className="quick-link-icon" aria-hidden="true">{section.icon}</span>
            <span>
              <strong>{section.label}</strong>
              <small>{section.description}</small>
            </span>
            <span aria-hidden="true">›</span>
          </NavLink>
        ))}
      </div>
    </section>
  )
}

function ComingSoonPage({ title, description }) {
  return (
    <section className="page-card" aria-labelledby="page-title">
      <span className="eyebrow">V pripravi</span>
      <h1 id="page-title">{title}</h1>
      <p>{description}</p>
      <p className="notice">
        Navigacija že deluje. Podatke, obrazce in rezervacije bomo povezali v naslednjih nalogah.
      </p>
    </section>
  )
}

function NotFoundPage() {
  return (
    <section className="page-card" aria-labelledby="page-title">
      <span className="eyebrow">404</span>
      <h1 id="page-title">Stran ne obstaja</h1>
      <p>Preveri spletni naslov ali se vrni na začetno stran.</p>
      <NavLink className="button-link" to={paths.home}>Na začetno stran</NavLink>
    </section>
  )
}

function Navigation() {
  return (
    <nav className="site-nav" aria-label="Glavna navigacija">
      {navigation.map(({ path, label, end }) => (
        <NavLink
          key={path}
          to={path}
          end={end}
          className={({ isActive }) => `nav-link${isActive ? ' active' : ''}`}
        >
          {label}
        </NavLink>
      ))}
    </nav>
  )
}

export default function App() {
  return (
    <div className="app-shell">
      <header className="site-header">
        <NavLink className="brand" to={paths.home} aria-label="DeskBooker – začetna stran">
          <span className="brand-icon" aria-hidden="true">▣</span>
          <span>DeskBooker</span>
        </NavLink>
        <Navigation />
      </header>

      <main className="main-content" id="main-content">
        <Routes>
          <Route path={paths.home} element={<HomePage />} />
          {sections.map(({ path, label, description }) => (
            <Route
              key={path}
              path={path}
              element={<ComingSoonPage title={label} description={description} />}
            />
          ))}
          <Route
            path={paths.reservations}
            element={
              <ComingSoonPage
                title="Moje rezervacije"
                description="Tukaj boš spremljal in upravljal svoje rezervacije."
              />
            }
          />
          <Route
            path={paths.login}
            element={
              <ComingSoonPage
                title="Prijava"
                description="Prijava uporabnikov bo dodana v nalogi za avtentikacijo."
              />
            }
          />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </main>

      <footer className="site-footer">DeskBooker · Razvojna različica</footer>
    </div>
  )
}

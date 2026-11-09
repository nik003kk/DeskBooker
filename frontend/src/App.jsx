import { useState } from 'react'
import { NavLink, Route, Routes } from 'react-router-dom'
import { Button, ButtonLink, Card, Field, SelectField, StatusBadge } from './components/ui.jsx'
import { navigation, paths, sections } from './routes.js'
import RegisterPage from './components/RegisterPage.jsx'
import LoginPage from './components/LoginPage.jsx'
import { useAuth } from './AuthContext.jsx'
import ProtectedRoute from './components/ProtectedRoute.jsx'
import ProfilePage from './components/ProfilePage.jsx'
import CreateResourcePage from './components/CreateResourcePage.jsx'
import AdminResourcesPage from './components/AdminResourcesPage.jsx'
import EditResourcePage from './components/EditResourcePage.jsx'
import ManagerRoute from './components/ManagerRoute.jsx'
import ResourceSearchPage from './components/ResourceSearchPage.jsx'
import BookingConfirmPage from './components/BookingConfirmPage.jsx'
import BookingSuccessPage from './components/BookingSuccessPage.jsx'
import CheckInPage from './components/CheckInPage.jsx'
import AnalyticsPage from './components/AnalyticsPage.jsx'
import { canManageResources } from './resource-permissions.js'

const resourceSamples = {
  [paths.desks]: [
    { id: 'D-101', title: 'Miza D-101', detail: '1. nadstropje · Cona A', badge: 'Monitor' },
    { id: 'D-102', title: 'Miza D-102', detail: '1. nadstropje · Cona A', badge: 'Standardna' },
    { id: 'D-201', title: 'Miza D-201', detail: '2. nadstropje · Cona B', badge: 'Monitor' },
  ],
  [paths.rooms]: [
    { id: 'S-1', title: 'Triglav', detail: '1. nadstropje · 8 oseb', badge: 'Sejna soba' },
    { id: 'S-2', title: 'Pohorje', detail: '2. nadstropje · 4 osebe', badge: 'Sejna soba' },
    { id: 'S-3', title: 'Bled', detail: '2. nadstropje · 12 oseb', badge: 'Sejna soba' },
  ],
  [paths.parking]: [
    { id: 'P-01', title: 'Parkirišče P-01', detail: 'Pokrito parkirišče', badge: 'Pokrito' },
    { id: 'P-02', title: 'Parkirišče P-02', detail: 'Pokrito · Električno polnjenje', badge: 'EV' },
    { id: 'P-03', title: 'Parkirišče P-03', detail: 'Zunanje parkirišče', badge: 'Zunanje' },
  ],
  [paths.equipment]: [
    { id: 'O-1', title: 'Monitor 27″', detail: 'Zunanja oprema delovnega mesta', badge: 'Monitor' },
    { id: 'O-2', title: 'USB-C priklopna postaja', detail: 'Priklop dodatnih naprav', badge: 'Dodatek' },
    { id: 'O-3', title: 'Projektor', detail: 'Za predstavitve in sestanke', badge: 'AV oprema' },
  ],
}

function HomePage() {
  return (
    <div className="page-stack">
      <Card className="hero">
        <div className="hero-copy">
          <StatusBadge tone="blue">Dobrodošli v DeskBooker</StatusBadge>
          <h1>Prostor za tvoj najboljši delovni dan.</h1>
          <p>Delovna mesta, sejne sobe, parkirišča in oprema — vse na enem mestu.</p>
          <div className="hero-actions">
            <ButtonLink to={paths.desks}>Razišči delovna mesta <span aria-hidden="true">→</span></ButtonLink>
            <ButtonLink to={paths.reservations} variant="secondary">Moje rezervacije</ButtonLink>
          </div>
        </div>
        <div className="hero-graphic" aria-hidden="true">
          <span className="hero-symbol">▦</span>
          <span className="hero-graphic-text">Tvoja pisarna. Tvoja izbira.</span>
          <span className="hero-mini-card">DeskBooker · Workspace</span>
        </div>
      </Card>
      <section className="section-block" aria-labelledby="resources-title">
        <div className="section-heading">
          <div>
            <span className="eyebrow">Razišči možnosti</span>
            <h2 id="resources-title">Kaj želiš rezervirati?</h2>
          </div>
          <StatusBadge>Predogled aplikacije</StatusBadge>
        </div>
        <div className="resource-grid">
          {sections.map((section) => (
            <Card as="article" className="resource-card" key={section.path}>
              <span className="resource-icon" aria-hidden="true">{section.icon}</span>
              <h3>{section.label}</h3>
              <p>{section.description}</p>
              <NavLink className="card-link" to={section.path}>
                Odpri stran <span aria-hidden="true">→</span>
              </NavLink>
            </Card>
          ))}
        </div>
      </section>
    </div>
  )
}

function SearchForm({ onSearch, onReset }) {
  const [query, setQuery] = useState('')
  const [floor, setFloor] = useState('all')

  function handleSubmit(event) {
    event.preventDefault()
    onSearch({ query: query.trim(), floor })
  }

  function handleReset() {
    setQuery('')
    setFloor('all')
    onReset()
  }

  return (
    <Card className="filter-card" aria-label="Iskalni filter">
      <form onSubmit={handleSubmit} className="filter-form" role="search">
        <Field
          label="Iskanje po imenu"
          type="search"
          placeholder="Vnesi ime ali oznako ..."
          value={query}
          onChange={(event) => setQuery(event.target.value)}
        />
        <SelectField
          label="Nadstropje"
          value={floor}
          onChange={(event) => setFloor(event.target.value)}
          options={[
            { value: 'all', label: 'Vsa nadstropja' },
            { value: '1', label: '1. nadstropje' },
            { value: '2', label: '2. nadstropje' },
          ]}
        />
        <div className="filter-actions">
          <Button type="submit">Prikaži</Button>
          <Button variant="ghost" onClick={handleReset}>Ponastavi</Button>
        </div>
      </form>
    </Card>
  )
}

function ResourcePage({ path, title, description }) {
  const [filter, setFilter] = useState({ query: '', floor: 'all' })
  const [hasSearched, setHasSearched] = useState(false)
  const samples = resourceSamples[path] || []
  const filtered = samples.filter((item) => {
    const matchesName = `${item.title} ${item.id} ${item.detail}`.toLowerCase().includes(filter.query.toLowerCase())
    const matchesFloor = filter.floor === 'all' || item.detail.includes(`${filter.floor}. nadstropje`)
    return matchesName && matchesFloor
  })

  return (
    <div className="page-stack">
      <div className="page-intro">
        <div>
          <span className="eyebrow">Rezervacijski viri</span>
          <h1>{title}</h1>
          <p>{description}</p>
        </div>
        <StatusBadge tone="blue">Demo prikaz</StatusBadge>
      </div>
      <SearchForm
        onSearch={(value) => { setFilter(value); setHasSearched(true) }}
        onReset={() => { setFilter({ query: '', floor: 'all' }); setHasSearched(false) }}
      />
      <p className="demo-notice" role="note">
        To so primeri za preverjanje postavitve in obrazcev. Podatki niso povezani z bazo in rezervacije še niso mogoče.
      </p>
      <div className="section-heading">
        <h2>Predogled virov</h2>
        <span className="result-count" aria-live="polite">{filtered.length} {filtered.length === 1 ? 'rezultat' : 'rezultatov'}</span>
      </div>
      {filtered.length > 0 ? (
        <div className="resource-grid">
          {filtered.map((item) => (
            <Card as="article" className="listing-card" key={item.id}>
              <div className="listing-top"><span className="listing-icon" aria-hidden="true">▤</span><StatusBadge>{item.badge}</StatusBadge></div>
              <h3>{item.title}</h3>
              <p>{item.detail}</p>
              <Button variant="secondary" disabled title="Rezervacije bodo omogočene v naslednjih nalogah">Rezerviraj (kmalu)</Button>
            </Card>
          ))}
        </div>
      ) : (
        <Card className="empty-state" role="status">
          <span aria-hidden="true">⌕</span>
          <h3>Ni zadetkov</h3>
          <p>Za ta iskalni pogoj ni predstavitvenih virov. Poskusi drugačen izraz ali ponastavi filter.</p>
          <Button variant="secondary" onClick={() => { setFilter({ query: '', floor: 'all' }); setHasSearched(false) }}>Počisti filter</Button>
        </Card>
      )}
      {hasSearched && <span className="sr-only" role="status">Prikazani so filtrirani primeri.</span>}
    </div>
  )
}

function ComingSoonPage({ title, description }) {
  return (
    <Card className="info-page">
      <StatusBadge tone="blue">V pripravi</StatusBadge>
      <h1>{title}</h1>
      <p>{description}</p>
      <p className="demo-notice">Ta funkcionalnost bo dodana v naslednjih nalogah. Za zdaj je na voljo le osnovni uporabniški vmesnik.</p>
      <ButtonLink to={paths.home} variant="secondary">Nazaj na začetno stran</ButtonLink>
    </Card>
  )
}

function NotFoundPage() {
  return (
    <Card className="info-page">
      <span className="error-code">404</span>
      <h1>Stran ne obstaja</h1>
      <p>Preveri spletni naslov ali se vrni na začetno stran.</p>
      <ButtonLink to={paths.home}>Na začetno stran</ButtonLink>
    </Card>
  )
}

function Navigation() {
  const { isAuthenticated, user, logout } = useAuth()
  return (
    <nav className="site-nav" aria-label="Glavna navigacija">
      {navigation.filter(({path}) => (path !== paths.profile || isAuthenticated) && (![paths.createResource,paths.adminResources,paths.analytics].includes(path) || canManageResources(user))).map(({ path, label, end }) => (
        <NavLink
          key={path}
          to={path}
          end={end}
          className={({ isActive }) => `nav-link${isActive ? ' active' : ''}`}
        >
          {label}
        </NavLink>
      ))}
      {isAuthenticated && <button type="button" className="nav-link auth-logout" onClick={logout}>Odjava ({user?.first_name || user?.email || "uporabnik"})</button>}
    </nav>
  )
}

export default function App() {
  return (
    <div className="app-shell">
      <a className="skip-link" href="#main-content">Preskoči na vsebino</a>
      <header className="site-header">
        <div className="header-inner">
          <NavLink className="brand" to={paths.home} aria-label="DeskBooker – začetna stran">
            <span className="brand-icon" aria-hidden="true">▣</span>
            <span>DeskBooker</span>
          </NavLink>
          <Navigation />
        </div>
      </header>

      <main className="main-content" id="main-content">
        <Routes>
          <Route path={paths.home} element={<HomePage />} />
          {sections.map(({ path, label, description }) => (
            <Route
              key={path}
              path={path}
              element={<ResourcePage path={path} title={label} description={description} />}
            />
          ))}
          <Route path={paths.editResource} element={<ManagerRoute><EditResourcePage /></ManagerRoute>} />
          <Route path={paths.adminResources} element={<ManagerRoute><AdminResourcesPage /></ManagerRoute>} />
          <Route path={paths.createResource} element={<ManagerRoute><CreateResourcePage /></ManagerRoute>} />
          <Route path={paths.profile} element={<ProtectedRoute><ProfilePage /></ProtectedRoute>} />
          <Route path={paths.analytics} element={<ManagerRoute><AnalyticsPage /></ManagerRoute>} />
          <Route path={paths.search} element={<ProtectedRoute><ResourceSearchPage /></ProtectedRoute>} />
          <Route path={paths.checkIn} element={<ProtectedRoute><CheckInPage /></ProtectedRoute>} />
          <Route path={paths.bookingSuccess} element={<ProtectedRoute><BookingSuccessPage /></ProtectedRoute>} />
          <Route path={paths.bookingConfirm} element={<ProtectedRoute><BookingConfirmPage /></ProtectedRoute>} />
          <Route path={paths.reservations} element={<ProtectedRoute><ComingSoonPage title="Moje rezervacije" description="Tukaj boš spremljal in upravljal svoje rezervacije." /></ProtectedRoute>} />
          <Route path={paths.register} element={<RegisterPage />} />
          <Route path={paths.login} element={<LoginPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </main>

      <footer className="site-footer"><span>© DeskBooker · Razvojna različica</span><span>Ustvarjeno za hibridno pisarno</span></footer>
    </div>
  )
}


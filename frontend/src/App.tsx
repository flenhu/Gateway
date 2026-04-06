import { BrowserRouter, NavLink, Navigate, Route, Routes } from 'react-router-dom'
import { CompareView } from './pages/CompareView'
import './App.css'

function App() {
  return (
    <BrowserRouter>
      <div className="app-shell">
        <header className="app-header">
          <div className="app-header__brand">
            <p className="app-header__eyebrow">LLM Gateway</p>
            <h1>Compare models or let the gateway choose.</h1>
            <p className="app-header__lede">
              Phase 3 starts here: the frontend is now routed and ready for
              mock-driven API integration.
            </p>
          </div>

          <nav className="app-nav" aria-label="Primary">
            <NavLink
              end
              to="/"
              className={({ isActive }) =>
                isActive ? 'app-nav__link app-nav__link--active' : 'app-nav__link'
              }
            >
              Best Pick
            </NavLink>
            <NavLink
              to="/compare"
              className={({ isActive }) =>
                isActive ? 'app-nav__link app-nav__link--active' : 'app-nav__link'
              }
            >
              Compare
            </NavLink>
          </nav>
        </header>

        <main className="app-main">
          <Routes>
            <Route path="/" element={<BestPickPlaceholder />} />
            <Route path="/compare" element={<CompareView />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  )
}

function BestPickPlaceholder() {
  return (
    <section className="mode-panel">
      <div className="mode-panel__copy">
        <p className="mode-panel__label">Default Route</p>
        <h2>Best Pick</h2>
        <p>
          This route will send one request and let the gateway choose the best
          available provider.
        </p>
      </div>

      <div className="mode-panel__details">
        <InfoTile
          title="Input Flow"
          body="A shared prompt input will live here in the next step."
        />
        <InfoTile
          title="Response Shape"
          body="One response card will render the chosen provider, latency, cost, and token usage."
        />
      </div>
    </section>
  )
}

type InfoTileProps = {
  title: string
  body: string
}

function InfoTile({ title, body }: InfoTileProps) {
  return (
    <article className="info-tile">
      <h3>{title}</h3>
      <p>{body}</p>
    </article>
  )
}

export default App

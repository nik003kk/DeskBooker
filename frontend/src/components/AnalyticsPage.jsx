import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../AuthContext.jsx'
import { Button, Card } from './ui.jsx'
import { defaultAnalyticsPeriod, formatPercent, loadAnalytics } from '../analytics-api.js'
import { OccupancyChart, StatusChart } from './AnalyticsCharts.jsx'
import AnalyticsFilters from './AnalyticsFilters.jsx'
import { previousAnalyticsPeriod, percentagePointChange, reservationChange } from '../analytics-periods.js'
import { paths } from '../routes.js'

export default function AnalyticsPage() {
  const { token, logout } = useAuth()
  const [refresh, setRefresh] = useState(0)
  const [previous, setPrevious] = useState(null)
  const [period, setPeriod] = useState(() => defaultAnalyticsPeriod())
  const [report, setReport] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  useEffect(() => {
    const controller = new AbortController()
    setLoading(true); setError(''); setReport(null); setPrevious(null)
    const options = { token, logout, signal: controller.signal, apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api' }
    Promise.all([loadAnalytics(period, options), loadAnalytics(previousAnalyticsPeriod(period), options)])
      .then(([data, earlier]) => { if (!controller.signal.aborted) { setReport(data); setPrevious(earlier) } })
      .catch(err => { if (!controller.signal.aborted) setError(err.status === 403 ? 'Za analitiko potrebuješ aktivno vlogo vodje ali skrbnika.' : 'Analitike ni mogoče naložiti. Poskusi ponovno.') })
      .finally(() => { if (!controller.signal.aborted) setLoading(false) })
    return () => controller.abort()
  }, [period, token, logout, refresh])
  return <div className="page-stack analytics-page">
    <div className="section-heading"><div><span className="eyebrow">Pregled pisarne</span><h1>Analitika zasedenosti</h1><p>{period.from} – {period.to} · Europe/Ljubljana</p></div></div>
    <AnalyticsFilters period={period} loading={loading} onChange={setPeriod} onRefresh={() => setRefresh(value => value + 1)} />
    {loading && <Card aria-busy="true"><p role="status">Nalagam analitiko ...</p></Card>}
    {error && <Card><p role="alert">{error}</p><Button onClick={() => setRefresh(value => value + 1)}>Poskusi znova</Button></Card>}
    {report && <>
      <p className="field-hint">Primerjava s predhodnim enako dolgim obdobjem: {previous.from} – {previous.to}</p>
      <div className="analytics-kpis">
        <KPI title="Skupna zasedenost" value={formatPercent(report.occupancy_rate)} hint="Rezervirani čas aktivnih virov, pon–pet 09–17" comparison={percentagePointChange(report.occupancy_rate, previous.occupancy_rate)} />
        <KPI title="Vse rezervacije" value={report.total_reservations} hint="Po datumu začetka v izbranem obdobju" comparison={reservationChange(report.total_reservations, previous.total_reservations)} />
        <KPI title="Aktivne rezervacije" value={report.active_reservations} hint="Potrjene ali s prisotnostjo, ki se še niso končale" />
        <KPI title="No-show" value={formatPercent(report.no_show.rate)} hint={`${report.no_show.expired} poteklih od ${report.no_show.eligible} znanih izidov`} comparison={percentagePointChange(report.no_show.rate, previous.no_show.rate)} />
      </div>
      {report.total_reservations === 0 && <Card><h2>V tem obdobju še ni rezervacij</h2><p>Ko bodo ustvarjene, se bodo tukaj prikazali zbirni podatki.</p></Card>}
      <div className="analytics-charts"><OccupancyChart points={report.weekdays} /><StatusChart counts={report.status_counts} total={report.total_reservations} /></div>
      <Card><h2>Kako beremo podatke</h2><p>Zasedenost meri rezervirani čas trenutno aktivnih virov med 09:00 in 17:00 od ponedeljka do petka. Zgodovina aktivacij ni na voljo. No-show primerja potekle rezervacije s poteklimi, prisotnimi in zaključenimi; prihodnje rezervacije in preklici so iz tega deleža izključeni.</p><p className="field-hint">Zadnje osveževanje: {new Intl.DateTimeFormat('sl-SI', { timeZone: 'Europe/Ljubljana', dateStyle: 'short', timeStyle: 'medium' }).format(new Date(report.generated_at))}</p></Card>
    </>}
    <Card><div className="section-heading"><h2>Upravljanje virov</h2><Link className="ui-button ui-button--primary" to={paths.createResource}>Dodaj vir</Link></div><p>Uredi podatke in opremo ter upravljaj aktivnost virov.</p><Link to={paths.adminResources}>Odpri seznam virov →</Link></Card>
  </div>
}
function KPI({ title, value, hint, comparison }) { return <Card className="analytics-kpi"><h2>{title}</h2><strong className="analytics-kpi-value">{value}</strong><p>{hint}</p>{comparison && <small className="analytics-comparison">{comparison}</small>}</Card> }

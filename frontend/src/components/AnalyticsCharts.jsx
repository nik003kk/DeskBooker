import { useId } from 'react'
import { analyticsStatuses, formatPercent, statusLabels, weekdayLabels } from '../analytics-api.js'
import { Card } from './ui.jsx'

export function OccupancyChart({ points }) {
  const title = useId()
  return <Card className="analytics-chart-card"><h2>Zasedenost po dnevih v tednu</h2><p className="field-hint">Tehtani delež rezerviranega časa · pon–pet 09–17</p>
    {points.every(point => point.occupancy_rate === null) && <p>V tem obdobju ni razpoložljive kapacitete.</p>}
    <svg className="analytics-bars" viewBox="0 0 600 250" role="img" aria-labelledby={title}>
      <title id={title}>Zasedenost od ponedeljka do petka v odstotkih</title>
      {[0, 25, 50, 75, 100].map(value => <g key={value}><line x1="48" x2="585" y1={195 - value * 1.6} y2={195 - value * 1.6} stroke="#e1e8f2" strokeDasharray="4 4" /><text x="38" y={200 - value * 1.6} textAnchor="end" className="analytics-axis">{value}%</text></g>)}
      {points.map((point, index) => {
        const height = (point.occupancy_rate ?? 0) * 1.6
        const x = 74 + index * 103
        return <g key={point.weekday}>
          <rect x={x} y={195 - height} width="58" height={height} rx="7" fill="#4282f5"><title>{weekdayLabels[index]}: {formatPercent(point.occupancy_rate)}</title></rect>
          <text x={x + 29} y={Math.max(22, 183 - height)} textAnchor="middle" className="analytics-bar-value">{formatPercent(point.occupancy_rate)}</text>
          <text x={x + 29} y="226" textAnchor="middle" className="analytics-axis">{weekdayLabels[index]}</text>
        </g>
      })}
    </svg>
    <details><summary>Prikaži podatke grafa</summary><div className="analytics-table-scroll"><table><caption className="sr-only">Zasedenost po dnevih v tednu</caption><thead><tr><th>Dan</th><th>Rezervirane minute</th><th>Razpoložljive minute</th><th>Zasedenost</th></tr></thead><tbody>{points.map(point => <tr key={point.weekday}><th scope="row">{weekdayLabels[point.weekday - 1]}</th><td>{Math.round(point.booked_minutes * 100) / 100}</td><td>{Math.round(point.capacity_minutes * 100) / 100}</td><td>{formatPercent(point.occupancy_rate)}</td></tr>)}</tbody></table></div></details>
  </Card>
}
const colors = ['#c3ccdd', '#8eb9ff', '#3269e8', '#2d9d76', '#edb65b', '#df7080']
export function StatusChart({ counts, total }) {
  const title = useId()
  let offset = 0
  return <Card className="analytics-chart-card"><h2>Status rezervacij</h2><p className="field-hint">Vse rezervacije po začetku v izbranem obdobju</p>
    <div className="analytics-status-layout">
      <svg viewBox="0 0 240 240" className="analytics-donut" role="img" aria-labelledby={title}>
        <title id={title}>Porazdelitev statusov: skupaj {total} rezervacij</title>
        <circle cx="120" cy="120" r="82" fill="none" stroke="#edf1f7" strokeWidth="32" />
        {analyticsStatuses.map((status, index) => {
          const share = total ? counts[status] / total * 100 : 0
          const start = offset; offset += share
          return share > 0 && <circle key={status} cx="120" cy="120" r="82" fill="none" stroke={colors[index]} strokeWidth="32" pathLength="100" strokeDasharray={`${share} ${100 - share}`} strokeDashoffset={-start} transform="rotate(-90 120 120)"><title>{statusLabels[status]}: {counts[status]} ({formatPercent(share)})</title></circle>
        })}
        <text x="120" y="119" textAnchor="middle" className="analytics-donut-total">{total}</text><text x="120" y="141" textAnchor="middle" className="analytics-axis">rezervacij</text>
      </svg>
      <ul className="analytics-legend">{analyticsStatuses.map((status, index) => <li key={status}><span className="analytics-legend-label"><span className="analytics-color" style={{ background: colors[index] }} aria-hidden="true" />{statusLabels[status]}</span><strong>{counts[status]}</strong><span>{formatPercent(total ? counts[status] / total * 100 : null)}</span></li>)}</ul>
    </div>
    {total === 0 && <p>V tem obdobju ni rezervacij.</p>}
  </Card>
}

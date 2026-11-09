import { useEffect, useRef, useState } from 'react'
import { Button } from './ui.jsx'
import { submitCheckIn } from '../check-in-api.js'

export default function CheckInModal({ reservation, options, onClose, onSuccess }) {
  const dialog = useRef(null)
  const inputs = useRef([])
  const pending = useRef(false)
  const alive = useRef(true)
  const [digits, setDigits] = useState(['', '', '', ''])
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  useEffect(() => {
    alive.current = true
    const previous = document.activeElement
    dialog.current.showModal()
    inputs.current[0]?.focus()
    return () => { alive.current = false; dialog.current?.close(); previous?.focus() }
  }, [])

  useEffect(() => {
    if (error && !busy) inputs.current[0]?.focus()
  }, [error, busy])

  function change(index, value) {
    if (!/^\d?$/.test(value)) return
    setDigits(current => current.map((digit, i) => i === index ? value : digit))
    setError('')
    if (value && index < 3) inputs.current[index + 1]?.focus()
  }
  function paste(event, index) {
    event.preventDefault()
    const value = event.clipboardData.getData('text').trim()
    if (!/^\d{1,4}$/.test(value)) return
    const chars = value.slice(0, 4 - index).split('')
    setDigits(current => current.map((digit, i) => i >= index && i < index + chars.length ? chars[i - index] : digit))
    setError('')
    inputs.current[Math.min(3, index + chars.length)]?.focus()
  }
  async function submit(event) {
    event.preventDefault()
    if (pending.current || !/^\d{4}$/.test(digits.join(''))) return
    pending.current = true
    setBusy(true); setError('')
    try {
      const saved = await submitCheckIn(reservation.id, digits.join(''), options)
      if (alive.current) onSuccess(saved)
    } catch (err) {
      if (alive.current) {
        setError(err instanceof TypeError ? 'Povezava s strežnikom ni uspela. Poskusi znova.' : err.message)
        setDigits(['', '', '', ''])
        inputs.current[0]?.focus()
      }
    } finally {
      pending.current = false
      if (alive.current) setBusy(false)
    }
  }
  return (
    <dialog ref={dialog} className="check-in-dialog" aria-labelledby="check-in-title"
      aria-describedby="check-in-hint" onCancel={event => { if (pending.current) event.preventDefault(); else onClose() }}>
      <form onSubmit={submit}>
        <span className="eyebrow">Potrditev prihoda</span>
        <h2 id="check-in-title">Check-in s PIN kodo</h2>
        <p id="check-in-hint">Vnesi štirimestno kodo, ki si jo prejel ob rezervaciji.</p>
        <fieldset disabled={busy} className="check-in-pin-fields">
          <legend className="sr-only">Štirimestna PIN koda</legend>
          {digits.map((digit, index) => <input key={index} ref={element => { inputs.current[index] = element }}
            aria-label={'PIN, številka ' + (index + 1)} type="text" inputMode="numeric" pattern="[0-9]"
            maxLength={1} required autoComplete="off" value={digit} onFocus={event => event.target.select()}
            onChange={event => change(index, event.target.value)} onPaste={event => paste(event, index)}
            onKeyDown={event => {
              if (event.key === 'Backspace' && !digit && index > 0) inputs.current[index - 1]?.focus()
              if (event.key === 'ArrowLeft' && index > 0) inputs.current[index - 1]?.focus()
              if (event.key === 'ArrowRight' && index < 3) inputs.current[index + 1]?.focus()
            }} />)}
        </fieldset>
        {error && <p role="alert" className="auth-feedback auth-feedback--error">{error}</p>}
        <div className="check-in-actions">
          <Button type="submit" disabled={busy || !/^\d{4}$/.test(digits.join(''))}>{busy ? 'Potrjujem ...' : 'Potrdi prisotnost'}</Button>
          <Button variant="secondary" disabled={busy} onClick={onClose}>Zapri</Button>
        </div>
      </form>
    </dialog>
  )
}

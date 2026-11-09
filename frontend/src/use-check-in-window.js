import { useEffect, useState } from 'react'
import { checkInWindow } from './check-in-window.js'

export function useCheckInWindow(reservation) {
  const [now, setNow] = useState(Date.now)
  useEffect(() => {
    const update = () => setNow(Date.now())
    update()
    const timer = window.setInterval(update, 1000)
    document.addEventListener('visibilitychange', update)
    return () => { window.clearInterval(timer); document.removeEventListener('visibilitychange', update) }
  }, [reservation?.id, reservation?.starts_at, reservation?.ends_at, reservation?.status])
  return checkInWindow(reservation, now)
}

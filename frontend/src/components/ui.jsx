import { useId } from 'react'
import { Link } from 'react-router-dom'

export function Button({ children, variant = 'primary', type = 'button', className = '', ...props }) {
  return (
    <button type={type} className={`ui-button ui-button--${variant} ${className}`.trim()} {...props}>
      {children}
    </button>
  )
}

export function ButtonLink({ children, to, variant = 'primary', className = '' }) {
  return <Link className={`ui-button ui-button--${variant} ${className}`.trim()} to={to}>{children}</Link>
}

export function Card({ children, className = '', as: Component = 'section', ...props }) {
  return <Component className={`ui-card ${className}`.trim()} {...props}>{children}</Component>
}

export function Field({ label, hint, id, className = '', ...inputProps }) {
  const generatedId = useId()
  const inputId = id || generatedId
  const hintId = `${inputId}-hint`
  return (
    <div className={`form-field ${className}`.trim()}>
      <label htmlFor={inputId}>{label}</label>
      <input id={inputId} aria-describedby={hint ? hintId : undefined} {...inputProps} />
      {hint && <small id={hintId} className="field-hint">{hint}</small>}
    </div>
  )
}

export function SelectField({ label, options, id, className = '', ...selectProps }) {
  const generatedId = useId()
  return (
    <div className={`form-field ${className}`.trim()}>
      <label htmlFor={id || generatedId}>{label}</label>
      <select id={id || generatedId} {...selectProps}>
        {options.map(({ value, label: optionLabel }) => (
          <option key={value} value={value}>{optionLabel}</option>
        ))}
      </select>
    </div>
  )
}

export function StatusBadge({ children, tone = 'neutral' }) {
  return <span className={`status-badge status-badge--${tone}`}>{children}</span>
}

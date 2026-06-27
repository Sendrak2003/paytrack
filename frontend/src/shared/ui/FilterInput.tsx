import { Search } from 'lucide-react'
import type { InputHTMLAttributes } from 'react'

interface Props extends InputHTMLAttributes<HTMLInputElement> {
  withIcon?: boolean
}

export function FilterInput({ withIcon, className = '', ...rest }: Props) {
  if (withIcon) {
    return (
      <div className="filter-input-wrapper">
        <Search size={14} className="filter-input-icon" />
        <input className={`filter-input filter-input-with-icon ${className}`} {...rest} />
      </div>
    )
  }
  return <input className={`filter-input ${className}`} {...rest} />
}

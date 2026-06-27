import type { ButtonHTMLAttributes, ReactNode } from 'react'

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'ghost' | 'success'
  size?: 'sm' | 'md'
  children: ReactNode
}

export function Button({ variant = 'ghost', size = 'md', children, className = '', ...rest }: Props) {
  const cls = `btn btn-${variant} ${size === 'sm' ? 'btn-sm' : ''} ${className}`.trim()
  return <button className={cls} {...rest}>{children}</button>
}

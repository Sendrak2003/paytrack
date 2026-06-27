/**
 * Shared API client — базовые fetch-обёртки.
 * Не знает о бизнес-сущностях (чистая архитектура).
 */
const API_ROOT = import.meta.env.VITE_API_URL ?? ''
const BASE = `${API_ROOT}/api/v1`

export async function get<T>(url: string): Promise<T> {
  const res = await fetch(BASE + url)
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}

export async function put<T>(url: string, body: unknown): Promise<T> {
  const res = await fetch(BASE + url, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}

export async function post<T>(url: string, body: unknown): Promise<T> {
  const res = await fetch(BASE + url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}

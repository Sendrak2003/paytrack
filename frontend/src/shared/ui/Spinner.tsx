export function Spinner({ text = 'Загрузка...' }: { text?: string }) {
  return (
    <div className="loading">
      <div className="spinner" />
      <div>{text}</div>
    </div>
  )
}

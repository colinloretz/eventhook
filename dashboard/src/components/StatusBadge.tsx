const colors: Record<string, string> = {
  delivered: 'bg-emerald-500/20 text-emerald-400 ring-emerald-500/30',
  success:   'bg-emerald-500/20 text-emerald-400 ring-emerald-500/30',
  pending:   'bg-gray-500/20 text-gray-400 ring-gray-500/30',
  retrying:  'bg-yellow-500/20 text-yellow-400 ring-yellow-500/30',
  failed:    'bg-red-500/20 text-red-400 ring-red-500/30',
  failure:   'bg-red-500/20 text-red-400 ring-red-500/30',
  timeout:   'bg-orange-500/20 text-orange-400 ring-orange-500/30',
  ignored:   'bg-gray-600/20 text-gray-500 ring-gray-600/30',
}

export function StatusBadge({ status }: { status: string }) {
  const cls = colors[status] ?? 'bg-gray-500/20 text-gray-400 ring-gray-500/30'
  return (
    <span className={`inline-flex items-center rounded px-2 py-0.5 text-xs font-medium ring-1 ring-inset ${cls}`}>
      {status}
    </span>
  )
}

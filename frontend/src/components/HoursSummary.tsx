import type { Horas } from '../types'

// Destaque de horas mensais por integrante — visão rápida do gestor
// sobre a distribuição de carga do mês.
export default function HoursSummary({
  horas,
  mesRotulo,
}: {
  horas: Horas[]
  mesRotulo: string
}) {
  return (
    <div className="rounded-lg bg-white p-4 shadow dark:bg-slate-800">
      <div className="mb-2 text-sm font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">
        Horas de plantão — {mesRotulo}
        <span className="ml-2 font-normal normal-case tracking-normal text-slate-400 dark:text-slate-500">
          (somente os dias deste mês)
        </span>
      </div>
      <div className="flex flex-wrap gap-3">
        {horas.length === 0 && (
          <span className="text-slate-400">
            Nenhum plantão neste mês. Configure o rodízio na aba Rodízio.
          </span>
        )}
        {horas.map((h) => (
          <div
            key={h.membro_id}
            className="flex items-center gap-3 rounded-lg border border-slate-200 px-4 py-2 dark:border-slate-700"
            style={{ borderLeft: `6px solid ${h.cor}` }}
          >
            <span className="font-medium">{h.nome}</span>
            <span className="text-2xl font-bold" style={{ color: h.cor }}>
              {h.total_formatado}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}

import type { Dia } from '../types'
import { rotuloTipo } from './ExceptionModal'

const DIAS_SEMANA = ['Dom', 'Seg', 'Ter', 'Qua', 'Qui', 'Sex', 'Sáb']

export default function CalendarGrid({
  ano,
  mes,
  dias,
  onSelecionarDia,
}: {
  ano: number
  mes: number
  dias: Dia[]
  onSelecionarDia?: (dia: Dia) => void
}) {
  const podeEditar = !!onSelecionarDia
  const primeiroDia = new Date(ano, mes - 1, 1).getDay()
  const hoje = new Date()
  const hojeISO = `${hoje.getFullYear()}-${String(hoje.getMonth() + 1).padStart(2, '0')}-${String(hoje.getDate()).padStart(2, '0')}`

  return (
    <div className="overflow-x-auto">
      <div className="grid min-w-[900px] grid-cols-7 gap-1.5">
        {DIAS_SEMANA.map((d) => (
          <div
            key={d}
            className="px-2 py-1 text-center text-sm font-semibold uppercase text-slate-500 dark:text-slate-400"
          >
            {d}
          </div>
        ))}
        {Array.from({ length: primeiroDia }).map((_, i) => (
          <div key={`vazio-${i}`} />
        ))}
        {dias.map((dia) => {
          const numero = Number(dia.data.slice(8, 10))
          const ehHoje = dia.data === hojeISO
          const fimDeSemana = [0, 6].includes(
            new Date(`${dia.data}T12:00:00`).getDay(),
          )
          return (
            <button
              key={dia.data}
              onClick={() => onSelecionarDia?.(dia)}
              className={`min-h-28 rounded-lg border p-1.5 text-left align-top transition ${
                podeEditar ? 'hover:ring-2 hover:ring-blue-400' : 'cursor-default'
              } ${
                ehHoje
                  ? 'border-blue-500 bg-blue-50 dark:bg-blue-950'
                  : fimDeSemana
                    ? 'border-slate-200 bg-slate-50 dark:border-slate-700 dark:bg-slate-800/60'
                    : 'border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800'
              }`}
            >
              <div className="mb-1 flex items-center justify-between">
                <span
                  className={`text-sm font-bold ${ehHoje ? 'text-blue-600 dark:text-blue-400' : 'text-slate-600 dark:text-slate-300'}`}
                >
                  {numero}
                </span>
                {dia.descoberto && (
                  <span title="Há horário descoberto neste dia">⚠️</span>
                )}
              </div>
              <div className="space-y-1">
                {dia.ferias.map((f) => (
                  <div
                    key={`f-${f.membro_id}`}
                    className="truncate rounded bg-cyan-100 px-1.5 py-0.5 text-xs font-medium text-cyan-700 dark:bg-cyan-950 dark:text-cyan-300"
                    title={`${f.nome} está de férias neste dia`}
                  >
                    🏖️ Férias: {f.nome}
                  </div>
                ))}
                {dia.sem_plantao && (
                  <div
                    className="truncate rounded bg-purple-100 px-1.5 py-0.5 text-xs font-medium text-purple-700 dark:bg-purple-950 dark:text-purple-300"
                    title={`Dia sem plantão${dia.motivo ? `: ${dia.motivo}` : ''}`}
                  >
                    🎉 {dia.motivo || 'Sem plantão'}
                  </div>
                )}
                {dia.slots.map((s, i) => (
                  <div
                    key={i}
                    className="truncate rounded px-1.5 py-0.5 text-xs font-medium text-white"
                    style={{ backgroundColor: s.cor }}
                    title={`${s.nome} — ${s.inicio} às ${s.fim}${s.origem !== 'rodizio' ? ` (${s.origem})` : ''}`}
                  >
                    {s.nome} {s.inicio}–{s.fim}
                    {s.origem !== 'rodizio' && ' ✱'}
                  </div>
                ))}
                {dia.pendentes.map((p) => (
                  <div
                    key={`p-${p.id}`}
                    className="truncate rounded border border-dashed border-amber-500 bg-amber-50 px-1.5 py-0.5 text-xs font-medium text-amber-700 dark:bg-amber-950 dark:text-amber-300"
                    title={`Pendente de confirmação: ${rotuloTipo(p.tipo)}`}
                  >
                    ⏳ {rotuloTipo(p.tipo)}
                    {p.tipo !== 'feriado' && p.tipo !== 'edicao_dia' && ` ${p.inicio}–${p.fim}`}
                  </div>
                ))}
              </div>
            </button>
          )
        })}
      </div>
      <div className="mt-2 flex gap-4 text-xs text-slate-500 dark:text-slate-400">
        <span>✱ alteração confirmada (substituição/troca/individual/edição)</span>
        <span>⏳ pendente de confirmação</span>
        <span>⚠️ horário descoberto</span>
        <span>🎉 dia sem plantão</span>
        <span>🏖️ férias</span>
      </div>
    </div>
  )
}

import { useCallback, useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Calendario, Dia, Membro } from '../types'
import CalendarGrid from '../components/CalendarGrid'
import HoursSummary from '../components/HoursSummary'
import ExceptionModal from '../components/ExceptionModal'

const MESES = [
  'Janeiro', 'Fevereiro', 'Março', 'Abril', 'Maio', 'Junho',
  'Julho', 'Agosto', 'Setembro', 'Outubro', 'Novembro', 'Dezembro',
]

export default function CalendarioPage({ logado }: { logado: boolean }) {
  const hoje = new Date()
  const [ano, setAno] = useState(hoje.getFullYear())
  const [mes, setMes] = useState(hoje.getMonth() + 1)
  const [cal, setCal] = useState<Calendario | null>(null)
  const [membros, setMembros] = useState<Membro[]>([])
  const [erro, setErro] = useState('')
  const [diaSelecionado, setDiaSelecionado] = useState<Dia | null>(null)

  const carregar = useCallback(() => {
    api.calendario(ano, mes).then(setCal).catch((e) => setErro(e.message))
    if (logado) {
      api.listarMembros().then(setMembros).catch(() => {})
    }
  }, [ano, mes, logado])

  useEffect(carregar, [carregar])

  const navegar = (delta: number) => {
    let m = mes + delta
    let a = ano
    if (m < 1) { m = 12; a-- }
    if (m > 12) { m = 1; a++ }
    setMes(m)
    setAno(a)
  }

  return (
    <div className="space-y-4">
      {erro && (
        <div className="rounded bg-red-100 px-4 py-2 text-red-800 dark:bg-red-950 dark:text-red-200">{erro}</div>
      )}

      {cal && <HoursSummary horas={cal.horas} mesRotulo={`${MESES[mes - 1]} ${ano}`} />}

      <div className="flex items-center justify-between">
        <button
          onClick={() => navegar(-1)}
          className="rounded bg-white px-4 py-2 shadow hover:bg-slate-50 dark:bg-slate-800 dark:hover:bg-slate-700"
        >
          ← {''}Mês anterior
        </button>
        <h2 className="text-xl font-bold">
          {MESES[mes - 1]} de {ano}
        </h2>
        <button
          onClick={() => navegar(1)}
          className="rounded bg-white px-4 py-2 shadow hover:bg-slate-50 dark:bg-slate-800 dark:hover:bg-slate-700"
        >
          Próximo mês →
        </button>
      </div>

      {cal && (
        <CalendarGrid
          ano={ano}
          mes={mes}
          dias={cal.dias}
          onSelecionarDia={logado ? setDiaSelecionado : undefined}
        />
      )}

      <p className="text-sm text-slate-500 dark:text-slate-400">
        {logado
          ? 'Clique em um dia para registrar substituição, troca, plantão individual, ausência, feriado ou editar o dia. Alterações só entram em vigor após confirmação.'
          : 'Visualização somente leitura. Faça login como administrador (no topo da página) para editar a escala.'}
      </p>

      {logado && diaSelecionado && (
        <ExceptionModal
          dia={diaSelecionado}
          membros={membros}
          onFechar={() => setDiaSelecionado(null)}
          onAlterado={() => {
            setDiaSelecionado(null)
            carregar()
          }}
        />
      )}
    </div>
  )
}

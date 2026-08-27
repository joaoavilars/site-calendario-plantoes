import { useCallback, useEffect, useState } from 'react'
import { api } from '../api/client'
import type { AuditEntry } from '../types'

const ENTIDADES: Record<string, string> = {
  member: 'Integrante',
  rule: 'Rodízio',
  exception: 'Agendamento',
}

const ACOES: Record<string, string> = {
  create: 'criado',
  update: 'atualizado',
  deactivate: 'desativado',
  replace: 'substituído',
  end: 'encerrado',
  confirmado: 'confirmado',
  rejeitado: 'rejeitado',
  cancelado: 'cancelado',
}

export default function HistoricoPage() {
  const [entradas, setEntradas] = useState<AuditEntry[]>([])
  const [entidade, setEntidade] = useState('')
  const [de, setDe] = useState('')
  const [ate, setAte] = useState('')
  const [erro, setErro] = useState('')
  const [detalhe, setDetalhe] = useState<AuditEntry | null>(null)

  const carregar = useCallback(() => {
    api
      .historico({ entidade, de, ate })
      .then(setEntradas)
      .catch((e) => setErro(e.message))
  }, [entidade, de, ate])

  useEffect(carregar, [carregar])

  return (
    <div className="space-y-4">
      {erro && <div className="rounded bg-red-100 px-4 py-2 text-red-800 dark:bg-red-950 dark:text-red-200">{erro}</div>}

      <div className="rounded-lg bg-white p-4 shadow dark:bg-slate-800">
        <h2 className="mb-3 font-bold">Histórico de alterações</h2>
        <div className="mb-4 flex flex-wrap items-end gap-3">
          <div>
            <label className="mb-1 block text-sm font-medium">Entidade</label>
            <select
              value={entidade}
              onChange={(e) => setEntidade(e.target.value)}
              className="rounded border border-slate-300 bg-white px-2 py-1.5 dark:border-slate-600 dark:bg-slate-700"
            >
              <option value="">Todas</option>
              <option value="member">Integrantes</option>
              <option value="rule">Rodízio</option>
              <option value="exception">Agendamentos</option>
            </select>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">De</label>
            <input type="date" value={de} onChange={(e) => setDe(e.target.value)}
              className="rounded border border-slate-300 bg-white px-2 py-1.5 dark:border-slate-600 dark:bg-slate-700" />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Até</label>
            <input type="date" value={ate} onChange={(e) => setAte(e.target.value)}
              className="rounded border border-slate-300 bg-white px-2 py-1.5 dark:border-slate-600 dark:bg-slate-700" />
          </div>
        </div>

        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b text-slate-500 dark:border-slate-700 dark:text-slate-400">
              <th className="py-2">Quando (UTC)</th>
              <th>O quê</th>
              <th>Ação</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {entradas.map((e) => (
              <tr key={e.id} className="border-b last:border-0 dark:border-slate-700">
                <td className="py-2">{e.criado_em}</td>
                <td>{ENTIDADES[e.entidade] ?? e.entidade} #{e.entidade_id}</td>
                <td>{ACOES[e.acao] ?? e.acao}</td>
                <td className="text-right">
                  <button
                    onClick={() => setDetalhe(detalhe?.id === e.id ? null : e)}
                    className="text-blue-600 hover:underline dark:text-blue-400"
                  >
                    {detalhe?.id === e.id ? 'ocultar' : 'detalhes'}
                  </button>
                </td>
              </tr>
            ))}
            {entradas.length === 0 && (
              <tr><td colSpan={4} className="py-4 text-center text-slate-400">Nenhum registro.</td></tr>
            )}
          </tbody>
        </table>
        {detalhe && (
          <pre className="mt-3 overflow-x-auto rounded bg-slate-800 p-3 text-xs text-slate-100 dark:bg-slate-950">
            {JSON.stringify(detalhe.dados, null, 2)}
          </pre>
        )}
      </div>
    </div>
  )
}

import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Membro } from '../types'

const CORES = ['#22c55e', '#eab308', '#3b82f6', '#ef4444', '#a855f7', '#f97316', '#14b8a6', '#ec4899']

export default function IntegrantesPage() {
  const [membros, setMembros] = useState<Membro[]>([])
  const [nome, setNome] = useState('')
  const [cor, setCor] = useState(CORES[0])
  const [mencao, setMencao] = useState('')
  const [erro, setErro] = useState('')
  const [msg, setMsg] = useState('')

  const carregar = () => {
    api.listarMembros(true).then(setMembros).catch((e) => setErro(e.message))
  }
  useEffect(carregar, [])

  const criar = async () => {
    setErro('')
    try {
      await api.criarMembro({ nome, cor, mencao_telegram: mencao })
      setNome('')
      setMencao('')
      carregar()
    } catch (e) {
      setErro((e as Error).message)
    }
  }

  const alternarAtivo = async (m: Membro) => {
    setErro('')
    try {
      if (m.ativo) await api.desativarMembro(m.id)
      else await api.atualizarMembro(m.id, { ativo: true })
      carregar()
    } catch (e) {
      setErro((e as Error).message)
    }
  }

  const testarAlerta = async () => {
    setErro('')
    setMsg('')
    try {
      await api.testarAlerta()
      setMsg('Mensagem de teste enviada ao grupo do Telegram!')
    } catch (e) {
      setErro((e as Error).message)
    }
  }

  return (
    <div className="space-y-6">
      {erro && <div className="rounded bg-red-100 px-4 py-2 text-red-800 dark:bg-red-950 dark:text-red-200">{erro}</div>}
      {msg && <div className="rounded bg-green-100 px-4 py-2 text-green-800 dark:bg-green-950 dark:text-green-200">{msg}</div>}

      <div className="rounded-lg bg-white p-4 shadow dark:bg-slate-800">
        <h2 className="mb-3 font-bold">Novo integrante</h2>
        <div className="flex flex-wrap items-end gap-3">
          <div>
            <label className="mb-1 block text-sm font-medium">Nome</label>
            <input
              value={nome}
              onChange={(e) => setNome(e.target.value)}
              className="rounded border border-slate-300 bg-white px-2 py-1.5 dark:border-slate-600 dark:bg-slate-700"
              placeholder="ex.: Maria"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Cor no calendário</label>
            <div className="flex gap-1">
              {CORES.map((c) => (
                <button
                  key={c}
                  onClick={() => setCor(c)}
                  className={`h-8 w-8 rounded-full border-2 ${cor === c ? 'border-slate-800 dark:border-white' : 'border-transparent'}`}
                  style={{ backgroundColor: c }}
                />
              ))}
            </div>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Menção no Telegram (opcional)</label>
            <input
              value={mencao}
              onChange={(e) => setMencao(e.target.value)}
              className="rounded border border-slate-300 bg-white px-2 py-1.5 dark:border-slate-600 dark:bg-slate-700"
              placeholder="@usuario"
            />
          </div>
          <button
            onClick={criar}
            disabled={!nome.trim()}
            className="rounded bg-blue-600 px-4 py-2 font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            Cadastrar
          </button>
        </div>
        <p className="mt-2 text-sm text-slate-500 dark:text-slate-400">
          Após cadastrar, configure os dias e horários de plantão na aba Rodízio.
        </p>
      </div>

      <div className="rounded-lg bg-white p-4 shadow dark:bg-slate-800">
        <h2 className="mb-3 font-bold">Integrantes</h2>
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b text-slate-500 dark:border-slate-700 dark:text-slate-400">
              <th className="py-2">Nome</th>
              <th>Telegram</th>
              <th>Situação</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {membros.map((m) => (
              <tr key={m.id} className="border-b last:border-0 dark:border-slate-700">
                <td className="py-2">
                  <span
                    className="mr-2 inline-block h-3 w-3 rounded-full"
                    style={{ backgroundColor: m.cor }}
                  />
                  {m.nome}
                </td>
                <td>{m.mencao_telegram || '—'}</td>
                <td>{m.ativo ? 'Ativo' : 'Inativo'}</td>
                <td className="text-right">
                  <button
                    onClick={() => alternarAtivo(m)}
                    className={`rounded px-3 py-1 text-white ${m.ativo ? 'bg-red-500 hover:bg-red-600' : 'bg-green-600 hover:bg-green-700'}`}
                  >
                    {m.ativo ? 'Desativar' : 'Reativar'}
                  </button>
                </td>
              </tr>
            ))}
            {membros.length === 0 && (
              <tr><td colSpan={4} className="py-4 text-center text-slate-400">Nenhum integrante cadastrado.</td></tr>
            )}
          </tbody>
        </table>
        <p className="mt-2 text-sm text-slate-500 dark:text-slate-400">
          Desativar um integrante encerra o rodízio dele a partir de hoje; o histórico é preservado.
        </p>
      </div>

      <div className="rounded-lg bg-white p-4 shadow dark:bg-slate-800">
        <h2 className="mb-2 font-bold">Alertas do Telegram</h2>
        <button
          onClick={testarAlerta}
          className="rounded bg-slate-700 px-4 py-2 font-medium text-white hover:bg-slate-800"
        >
          Enviar mensagem de teste ao grupo
        </button>
      </div>
    </div>
  )
}

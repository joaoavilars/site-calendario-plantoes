import { useCallback, useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Excecao, Membro, Preview, Regra } from '../types'

const DIAS = ['Domingo', 'Segunda', 'Terça', 'Quarta', 'Quinta', 'Sexta', 'Sábado']

// Faixas rápidas: uma linha vira várias regras (uma por dia) ao salvar
const FAIXAS: Record<string, { rotulo: string; dias: number[] }> = {
  '1-5': { rotulo: 'Seg a Sex', dias: [1, 2, 3, 4, 5] },
  '1-6': { rotulo: 'Seg a Sáb', dias: [1, 2, 3, 4, 5, 6] },
  '0-6': { rotulo: 'Todos os dias', dias: [0, 1, 2, 3, 4, 5, 6] },
  wknd: { rotulo: 'Sáb e Dom', dias: [6, 0] },
}

function expandeDias(valor: string): number[] {
  return FAIXAS[valor]?.dias ?? [Number(valor)]
}

interface LinhaRegra {
  chave: number
  dias: string // '0'..'6' ou chave de FAIXAS
  intervalo_semanas: number
  inicio: string
  fim: string
  vigente_de: string // âncora da alternância quando intervalo > 1
}

let seq = 1

// Agrupa as regras individuais do banco de volta em faixas quando os
// horários/intervalo coincidem (ex.: 5 regras seg–sex viram uma linha)
function agrupaRegras(regras: Regra[]): LinhaRegra[] {
  const grupos = new Map<string, { regra: Regra; dias: Set<number> }>()
  for (const r of regras) {
    const ancora = r.intervalo_semanas > 1 ? r.vigente_de : ''
    const key = `${r.inicio}|${r.fim}|${r.intervalo_semanas}|${ancora}`
    if (!grupos.has(key)) grupos.set(key, { regra: r, dias: new Set() })
    grupos.get(key)!.dias.add(r.dia_semana)
  }
  const linhas: LinhaRegra[] = []
  for (const { regra, dias } of grupos.values()) {
    const setKey = [...dias].sort((a, b) => a - b).join(',')
    const faixa = Object.entries(FAIXAS).find(
      ([, f]) => [...f.dias].sort((a, b) => a - b).join(',') === setKey,
    )
    const base = {
      intervalo_semanas: regra.intervalo_semanas,
      inicio: regra.inicio,
      fim: regra.fim,
      vigente_de: regra.intervalo_semanas > 1 ? regra.vigente_de : '',
    }
    if (faixa) {
      linhas.push({ chave: seq++, dias: faixa[0], ...base })
    } else {
      for (const d of [...dias].sort((a, b) => a - b)) {
        linhas.push({ chave: seq++, dias: String(d), ...base })
      }
    }
  }
  return linhas
}

const inputCls =
  'rounded border border-slate-300 bg-white px-2 py-1 dark:border-slate-600 dark:bg-slate-700'

export default function RodizioPage() {
  const [membros, setMembros] = useState<Membro[]>([])
  const [membroId, setMembroId] = useState<number | ''>('')
  const [linhas, setLinhas] = useState<LinhaRegra[]>([])
  const [erro, setErro] = useState('')
  const [msg, setMsg] = useState('')
  // férias
  const [feriasDe, setFeriasDe] = useState('')
  const [feriasAte, setFeriasAte] = useState('')
  const [feriasSub, setFeriasSub] = useState<number | ''>('')
  const [feriasObs, setFeriasObs] = useState('')
  const [feriasLista, setFeriasLista] = useState<Excecao[]>([])
  const [feriasPreview, setFeriasPreview] = useState<{ excecao: Excecao; preview: Preview } | null>(null)

  useEffect(() => {
    api.listarMembros().then(setMembros).catch((e) => setErro(e.message))
  }, [])

  const carregarRegras = useCallback((id: number) => {
    api
      .listarRegras(id)
      .then((regras) => setLinhas(agrupaRegras(regras)))
      .catch((e) => setErro(e.message))
  }, [])

  const carregarFerias = useCallback((id: number) => {
    api
      .listarExcecoes({ tipo: 'ferias', membro: id })
      .then((lista) =>
        setFeriasLista(lista.filter((f) => f.status === 'pendente' || f.status === 'confirmado')),
      )
      .catch((e) => setErro(e.message))
  }, [])

  const selecionar = (id: number | '') => {
    setMembroId(id)
    setMsg('')
    setErro('')
    setFeriasPreview(null)
    if (id !== '') {
      carregarRegras(id)
      carregarFerias(id)
    } else {
      setLinhas([])
      setFeriasLista([])
    }
  }

  const agendarFerias = async () => {
    if (membroId === '') return
    setErro('')
    setMsg('')
    try {
      const resp = await api.criarExcecao({
        data: feriasDe,
        tipo: 'ferias',
        membro_original_id: membroId,
        membro_substituto_id: feriasSub === '' ? null : feriasSub,
        data_fim: feriasAte,
        observacao: feriasObs,
      })
      setFeriasPreview(resp)
    } catch (e) {
      setErro((e as Error).message)
    }
  }

  const decidirFerias = async (acao: 'confirmar' | 'descartar') => {
    if (!feriasPreview || membroId === '') return
    setErro('')
    try {
      if (acao === 'confirmar') {
        await api.confirmarExcecao(feriasPreview.excecao.id)
        setMsg('Férias confirmadas! O calendário já reflete o período.')
      } else {
        await api.rejeitarExcecao(feriasPreview.excecao.id)
      }
      setFeriasPreview(null)
      setFeriasDe('')
      setFeriasAte('')
      setFeriasSub('')
      setFeriasObs('')
      carregarFerias(membroId)
    } catch (e) {
      setErro((e as Error).message)
    }
  }

  const cancelarFerias = async (id: number) => {
    if (membroId === '') return
    setErro('')
    try {
      await api.cancelarExcecao(id)
      setMsg('Período de férias cancelado; os plantões do rodízio voltam a valer.')
      carregarFerias(membroId)
    } catch (e) {
      setErro((e as Error).message)
    }
  }

  const adicionar = () => {
    setLinhas([
      ...linhas,
      { chave: seq++, dias: '1-5', intervalo_semanas: 1, inicio: '08:00', fim: '12:00', vigente_de: '' },
    ])
  }

  const atualizar = (chave: number, campo: keyof LinhaRegra, valor: string | number) => {
    setLinhas(linhas.map((l) => (l.chave === chave ? { ...l, [campo]: valor } : l)))
  }

  const remover = (chave: number) => {
    setLinhas(linhas.filter((l) => l.chave !== chave))
  }

  const salvar = async () => {
    if (membroId === '') return
    setErro('')
    setMsg('')
    try {
      await api.salvarRegras(
        membroId,
        linhas.flatMap((l) =>
          expandeDias(l.dias).map((dia) => ({
            dia_semana: dia,
            intervalo_semanas: l.intervalo_semanas,
            inicio: l.inicio,
            fim: l.fim,
            vigente_de: l.intervalo_semanas > 1 && l.vigente_de ? l.vigente_de : undefined,
          })),
        ),
      )
      setMsg('Rodízio salvo! O calendário já reflete as novas regras a partir de hoje.')
      carregarRegras(membroId)
    } catch (e) {
      setErro((e as Error).message)
    }
  }

  return (
    <div className="space-y-4">
      {erro && <div className="rounded bg-red-100 px-4 py-2 text-red-800 dark:bg-red-950 dark:text-red-200">{erro}</div>}
      {msg && <div className="rounded bg-green-100 px-4 py-2 text-green-800 dark:bg-green-950 dark:text-green-200">{msg}</div>}

      <div className="rounded-lg bg-white p-4 shadow dark:bg-slate-800">
        <h2 className="mb-3 font-bold">Rodízio de plantões</h2>
        <label className="mb-1 block text-sm font-medium">Integrante</label>
        <select
          value={membroId}
          onChange={(e) => selecionar(e.target.value ? Number(e.target.value) : '')}
          className={`${inputCls} py-1.5`}
        >
          <option value="">Selecione…</option>
          {membros.map((m) => (
            <option key={m.id} value={m.id}>{m.nome}</option>
          ))}
        </select>

        {membroId !== '' && (
          <div className="mt-4 space-y-2">
            {linhas.map((l) => (
              <div key={l.chave} className="flex flex-wrap items-center gap-2 rounded border border-slate-200 p-2 dark:border-slate-700">
                <select
                  value={l.dias}
                  onChange={(e) => atualizar(l.chave, 'dias', e.target.value)}
                  className={inputCls}
                >
                  <optgroup label="Faixas">
                    {Object.entries(FAIXAS).map(([valor, f]) => (
                      <option key={valor} value={valor}>{f.rotulo}</option>
                    ))}
                  </optgroup>
                  <optgroup label="Dia específico">
                    {DIAS.map((d, i) => (
                      <option key={i} value={String(i)}>{d}</option>
                    ))}
                  </optgroup>
                </select>
                <span className="text-sm">das</span>
                <input
                  type="time"
                  value={l.inicio}
                  onChange={(e) => atualizar(l.chave, 'inicio', e.target.value)}
                  className={inputCls}
                />
                <span className="text-sm">às</span>
                <input
                  type="time"
                  value={l.fim}
                  onChange={(e) => atualizar(l.chave, 'fim', e.target.value)}
                  className={inputCls}
                />
                <select
                  value={l.intervalo_semanas}
                  onChange={(e) => atualizar(l.chave, 'intervalo_semanas', Number(e.target.value))}
                  className={inputCls}
                >
                  <option value={1}>toda semana</option>
                  <option value={2}>a cada 2 semanas</option>
                  <option value={3}>a cada 3 semanas</option>
                  <option value={4}>a cada 4 semanas</option>
                </select>
                {l.intervalo_semanas > 1 && (
                  <>
                    <span className="text-sm">começando em</span>
                    <input
                      type="date"
                      value={l.vigente_de}
                      onChange={(e) => atualizar(l.chave, 'vigente_de', e.target.value)}
                      className={inputCls}
                      title="Primeira data desta alternância (ex.: primeiro sábado do integrante)"
                    />
                  </>
                )}
                <button
                  onClick={() => remover(l.chave)}
                  className="ml-auto rounded bg-red-500 px-2 py-1 text-sm text-white hover:bg-red-600"
                >
                  Remover
                </button>
              </div>
            ))}

            <div className="flex gap-2 pt-2">
              <button
                onClick={adicionar}
                className="rounded border border-slate-300 px-4 py-2 hover:bg-slate-50 dark:border-slate-600 dark:hover:bg-slate-700"
              >
                + Adicionar plantão
              </button>
              <button
                onClick={salvar}
                className="rounded bg-blue-600 px-4 py-2 font-medium text-white hover:bg-blue-700"
              >
                Salvar rodízio
              </button>
            </div>
            <p className="text-sm text-slate-500 dark:text-slate-400">
              Escolha uma faixa (ex.: Seg a Sex) para criar o plantão em todos
              os dias de uma vez, ou um dia específico. Salvar substitui o
              rodízio vigente deste integrante a partir de hoje; os meses
              anteriores continuam como eram (histórico preservado). Use
              &quot;a cada 2 semanas&quot; para escalas alternadas, como sábado
              sim / sábado não, informando a primeira data da alternância.
            </p>
          </div>
        )}
      </div>

      {membroId !== '' && (
        <div className="rounded-lg bg-white p-4 shadow dark:bg-slate-800">
          <h2 className="mb-3 font-bold">🏖️ Férias e afastamentos</h2>

          {!feriasPreview ? (
            <div className="flex flex-wrap items-end gap-3">
              <div>
                <label className="mb-1 block text-sm font-medium">De</label>
                <input type="date" value={feriasDe} onChange={(e) => setFeriasDe(e.target.value)} className={inputCls} />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">Até</label>
                <input type="date" value={feriasAte} onChange={(e) => setFeriasAte(e.target.value)} className={inputCls} />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">Substituto (opcional)</label>
                <select
                  value={feriasSub}
                  onChange={(e) => setFeriasSub(e.target.value ? Number(e.target.value) : '')}
                  className={inputCls}
                >
                  <option value="">Sem substituto</option>
                  {membros
                    .filter((m) => m.id !== membroId)
                    .map((m) => (
                      <option key={m.id} value={m.id}>{m.nome}</option>
                    ))}
                </select>
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">Observação</label>
                <input
                  type="text"
                  value={feriasObs}
                  onChange={(e) => setFeriasObs(e.target.value)}
                  placeholder="ex.: férias anuais"
                  className={inputCls}
                />
              </div>
              <button
                onClick={agendarFerias}
                disabled={!feriasDe || !feriasAte}
                className="rounded bg-blue-600 px-4 py-2 font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Agendar férias
              </button>
            </div>
          ) : (
            <div className="space-y-3">
              <div className="rounded border border-blue-200 bg-blue-50 px-4 py-2 text-sm text-blue-800 dark:border-blue-800 dark:bg-blue-950 dark:text-blue-200">
                Prévia: {feriasPreview.preview.dias.length} dia(s) com plantão
                afetado no período. Nada será alterado até você confirmar.
              </div>
              <div className="rounded border border-slate-200 p-3 dark:border-slate-700">
                <div className="mb-2 text-xs font-semibold uppercase text-slate-500 dark:text-slate-400">
                  Horas no mês de início (antes → depois)
                </div>
                {feriasPreview.preview.horas_depois.map((h) => {
                  const antes = feriasPreview.preview.horas_antes.find((a) => a.membro_id === h.membro_id)
                  const mudou = antes?.total_minutos !== h.total_minutos
                  return (
                    <div key={h.membro_id} className={`text-sm ${mudou ? 'font-semibold' : ''}`}>
                      {h.nome}: {antes?.total_formatado ?? '0h00'} → {h.total_formatado}
                    </div>
                  )
                })}
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => decidirFerias('descartar')}
                  className="rounded px-4 py-2 text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-700"
                >
                  Descartar
                </button>
                <button
                  onClick={() => decidirFerias('confirmar')}
                  className="rounded bg-green-600 px-4 py-2 font-medium text-white hover:bg-green-700"
                >
                  Confirmar férias
                </button>
              </div>
            </div>
          )}

          {feriasLista.length > 0 && (
            <div className="mt-4">
              <div className="mb-1 text-xs font-semibold uppercase text-slate-500 dark:text-slate-400">
                Períodos agendados
              </div>
              {feriasLista.map((f) => (
                <div
                  key={f.id}
                  className="flex flex-wrap items-center justify-between gap-2 border-b py-2 text-sm last:border-0 dark:border-slate-700"
                >
                  <span>
                    {dataBR(f.data)} a {f.data_fim ? dataBR(f.data_fim) : '?'} ·{' '}
                    {f.membro_substituto_id
                      ? `substituto: ${membros.find((m) => m.id === f.membro_substituto_id)?.nome ?? '?'}`
                      : 'sem substituto'}
                    {f.observacao && ` · ${f.observacao}`} ·{' '}
                    <span className={f.status === 'confirmado' ? 'text-green-600 dark:text-green-400' : 'text-amber-600 dark:text-amber-400'}>
                      {f.status}
                    </span>
                  </span>
                  <button
                    onClick={() => cancelarFerias(f.id)}
                    className="rounded bg-red-500 px-2 py-1 text-white hover:bg-red-600"
                  >
                    Cancelar
                  </button>
                </div>
              ))}
            </div>
          )}
          <p className="mt-3 text-sm text-slate-500 dark:text-slate-400">
            Durante as férias os plantões do integrante saem do calendário
            (com substituto, ele os assume automaticamente e as horas contam
            para ele). O rodízio continua correndo por baixo: no retorno, a
            escala — inclusive alternâncias de sábado — volta exatamente de
            onde parou, sem precisar mexer nas regras.
          </p>
        </div>
      )}
    </div>
  )
}

function dataBR(iso: string) {
  return `${iso.slice(8, 10)}/${iso.slice(5, 7)}/${iso.slice(0, 4)}`
}

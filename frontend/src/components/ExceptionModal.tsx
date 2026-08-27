import { useState } from 'react'
import { api } from '../api/client'
import type { Dia, Excecao, Membro, Preview, Slot, TipoExcecao } from '../types'

const TIPOS: Array<{ id: TipoExcecao; rotulo: string; descricao: string }> = [
  { id: 'substituicao', rotulo: 'Substituição', descricao: 'Outro integrante assume o plantão neste dia' },
  { id: 'troca', rotulo: 'Troca', descricao: 'Dois integrantes trocam plantões entre duas datas' },
  { id: 'extra', rotulo: 'Plantão individual', descricao: 'Inclui um plantão avulso para um integrante neste dia' },
  { id: 'ausencia', rotulo: 'Ausência', descricao: 'Integrante não fará o plantão (fica descoberto)' },
  { id: 'feriado', rotulo: 'Sem plantão (feriado)', descricao: 'Cancela toda a escala automática deste dia' },
  { id: 'edicao_dia', rotulo: 'Editar o dia', descricao: 'Redefine manualmente toda a escala deste dia' },
]

const ROTULOS_EXTRAS: Record<string, string> = { ferias: 'Férias' }

export const rotuloTipo = (t: string) =>
  TIPOS.find((x) => x.id === t)?.rotulo ?? ROTULOS_EXTRAS[t] ?? t

function dataBR(iso: string) {
  return `${iso.slice(8, 10)}/${iso.slice(5, 7)}/${iso.slice(0, 4)}`
}

function ListaSlots({ slots }: { slots: Slot[] }) {
  if (slots.length === 0)
    return <span className="text-sm italic text-slate-400">sem plantões</span>
  return (
    <div className="space-y-1">
      {slots.map((s, i) => (
        <div key={i} className="flex items-center gap-2 text-sm">
          <span
            className="inline-block h-3 w-3 rounded-full"
            style={{ backgroundColor: s.cor }}
          />
          {s.nome} — {s.inicio} às {s.fim}
        </div>
      ))}
    </div>
  )
}

interface LinhaSlot {
  chave: number
  membro_id: number | ''
  inicio: string
  fim: string
}

let seqSlot = 1

const inputCls =
  'rounded border border-slate-300 bg-white px-2 py-1.5 dark:border-slate-600 dark:bg-slate-700'

export default function ExceptionModal({
  dia,
  membros,
  onFechar,
  onAlterado,
}: {
  dia: Dia
  membros: Membro[]
  onFechar: () => void
  onAlterado: () => void
}) {
  const [tipo, setTipo] = useState<TipoExcecao>('substituicao')
  const [original, setOriginal] = useState<number | ''>('')
  const [substituto, setSubstituto] = useState<number | ''>('')
  const [dataTroca, setDataTroca] = useState('')
  const [inicio, setInicio] = useState(dia.slots[0]?.inicio ?? '08:00')
  const [fim, setFim] = useState(dia.slots[0]?.fim ?? '12:00')
  const [obs, setObs] = useState('')
  // editor do dia: começa com a escala atual do dia
  const [linhas, setLinhas] = useState<LinhaSlot[]>(() =>
    dia.slots.map((s) => ({ chave: seqSlot++, membro_id: s.membro_id, inicio: s.inicio, fim: s.fim })),
  )
  const [erro, setErro] = useState('')
  const [salvando, setSalvando] = useState(false)
  const [pendente, setPendente] = useState<{ excecao: Excecao; preview: Preview } | null>(null)

  const precisaOriginal = tipo === 'substituicao' || tipo === 'troca' || tipo === 'ausencia'
  const precisaSubstituto = tipo === 'substituicao' || tipo === 'troca' || tipo === 'extra'
  const mostraHorarios = tipo !== 'feriado' && tipo !== 'edicao_dia'

  const criar = async () => {
    setErro('')
    setSalvando(true)
    try {
      const resp = await api.criarExcecao({
        data: dia.data,
        tipo,
        membro_original_id: precisaOriginal && original !== '' ? original : null,
        membro_substituto_id: precisaSubstituto && substituto !== '' ? substituto : null,
        data_troca: tipo === 'troca' && dataTroca ? dataTroca : null,
        inicio: mostraHorarios ? inicio : undefined,
        fim: mostraHorarios ? fim : undefined,
        slots:
          tipo === 'edicao_dia'
            ? linhas
                .filter((l) => l.membro_id !== '')
                .map((l) => ({ membro_id: l.membro_id as number, inicio: l.inicio, fim: l.fim }))
            : undefined,
        observacao: obs,
      })
      setPendente(resp)
    } catch (e) {
      setErro((e as Error).message)
    } finally {
      setSalvando(false)
    }
  }

  const decidir = async (acao: 'confirmar' | 'descartar', id: number) => {
    setErro('')
    setSalvando(true)
    try {
      if (acao === 'confirmar') await api.confirmarExcecao(id)
      else await api.rejeitarExcecao(id)
      onAlterado()
    } catch (e) {
      setErro((e as Error).message)
      setSalvando(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-lg bg-white p-6 shadow-xl dark:bg-slate-800">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-lg font-bold">
            Agendamento — {dataBR(dia.data)}
            {dia.sem_plantao && (
              <span className="ml-2 rounded bg-purple-100 px-2 py-0.5 text-sm font-medium text-purple-700 dark:bg-purple-950 dark:text-purple-300">
                🎉 dia sem plantão{dia.motivo ? `: ${dia.motivo}` : ''}
              </span>
            )}
          </h3>
          <button onClick={onFechar} className="text-2xl leading-none text-slate-400 hover:text-slate-600 dark:hover:text-slate-200">
            ×
          </button>
        </div>

        {erro && (
          <div className="mb-4 rounded bg-red-100 px-4 py-2 text-red-800 dark:bg-red-950 dark:text-red-200">{erro}</div>
        )}

        {dia.pendentes.length > 0 && !pendente && (
          <div className="mb-4 rounded border border-amber-300 bg-amber-50 p-3 dark:border-amber-700 dark:bg-amber-950">
            <div className="mb-2 font-semibold text-amber-800 dark:text-amber-200">
              Pendentes de confirmação neste dia
            </div>
            {dia.pendentes.map((p) => (
              <div key={p.id} className="mb-2 flex items-center justify-between gap-2 text-sm">
                <span>
                  {rotuloTipo(p.tipo)}
                  {p.tipo !== 'feriado' && p.tipo !== 'edicao_dia' && ` — ${p.inicio} às ${p.fim}`}
                  {p.observacao && ` (${p.observacao})`}
                </span>
                <span className="flex gap-2">
                  <button
                    disabled={salvando}
                    onClick={() => decidir('confirmar', p.id)}
                    className="rounded bg-green-600 px-2 py-1 text-white hover:bg-green-700"
                  >
                    Confirmar
                  </button>
                  <button
                    disabled={salvando}
                    onClick={() => decidir('descartar', p.id)}
                    className="rounded bg-slate-400 px-2 py-1 text-white hover:bg-slate-500"
                  >
                    Rejeitar
                  </button>
                </span>
              </div>
            ))}
          </div>
        )}

        {!pendente ? (
          <div className="space-y-4">
            <div>
              <label className="mb-1 block text-sm font-medium">Tipo</label>
              <div className="grid grid-cols-2 gap-2">
                {TIPOS.map((t) => (
                  <button
                    key={t.id}
                    onClick={() => setTipo(t.id)}
                    className={`rounded border p-2 text-left text-sm transition ${
                      tipo === t.id
                        ? 'border-blue-500 bg-blue-50 dark:bg-blue-950'
                        : 'border-slate-200 hover:border-slate-300 dark:border-slate-700 dark:hover:border-slate-500'
                    }`}
                  >
                    <div className="font-semibold">{t.rotulo}</div>
                    <div className="text-xs text-slate-500 dark:text-slate-400">{t.descricao}</div>
                  </button>
                ))}
              </div>
            </div>

            {tipo === 'feriado' && (
              <p className="rounded bg-purple-50 px-4 py-2 text-sm text-purple-800 dark:bg-purple-950 dark:text-purple-200">
                Todos os plantões automáticos deste dia serão cancelados e as
                horas do mês recalculadas. Plantões individuais adicionados
                depois continuam valendo.
              </p>
            )}

            {tipo === 'edicao_dia' && (
              <div>
                <label className="mb-1 block text-sm font-medium">
                  Escala deste dia (substitui totalmente a automática)
                </label>
                <div className="space-y-2">
                  {linhas.map((l) => (
                    <div key={l.chave} className="flex flex-wrap items-center gap-2 rounded border border-slate-200 p-2 dark:border-slate-700">
                      <select
                        value={l.membro_id}
                        onChange={(e) =>
                          setLinhas(linhas.map((x) => x.chave === l.chave
                            ? { ...x, membro_id: e.target.value ? Number(e.target.value) : '' }
                            : x))
                        }
                        className={inputCls}
                      >
                        <option value="">Selecione…</option>
                        {membros.map((m) => (
                          <option key={m.id} value={m.id}>{m.nome}</option>
                        ))}
                      </select>
                      <span className="text-sm">das</span>
                      <input
                        type="time"
                        value={l.inicio}
                        onChange={(e) =>
                          setLinhas(linhas.map((x) => x.chave === l.chave ? { ...x, inicio: e.target.value } : x))
                        }
                        className={inputCls}
                      />
                      <span className="text-sm">às</span>
                      <input
                        type="time"
                        value={l.fim}
                        onChange={(e) =>
                          setLinhas(linhas.map((x) => x.chave === l.chave ? { ...x, fim: e.target.value } : x))
                        }
                        className={inputCls}
                      />
                      <button
                        onClick={() => setLinhas(linhas.filter((x) => x.chave !== l.chave))}
                        className="ml-auto rounded bg-red-500 px-2 py-1 text-sm text-white hover:bg-red-600"
                      >
                        Remover
                      </button>
                    </div>
                  ))}
                  <button
                    onClick={() =>
                      setLinhas([...linhas, { chave: seqSlot++, membro_id: '', inicio: '08:00', fim: '12:00' }])
                    }
                    className="rounded border border-slate-300 px-3 py-1.5 text-sm hover:bg-slate-50 dark:border-slate-600 dark:hover:bg-slate-700"
                  >
                    + Adicionar plantão
                  </button>
                </div>
                <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">
                  Ex.: transformar uma quarta-feira em escala de sábado. Para
                  deixar o dia sem nenhum plantão, use &quot;Sem plantão (feriado)&quot;.
                </p>
              </div>
            )}

            {(precisaOriginal || precisaSubstituto) && (
              <div className="grid grid-cols-2 gap-4">
                {precisaOriginal && (
                  <div>
                    <label className="mb-1 block text-sm font-medium">
                      {tipo === 'ausencia' ? 'Quem estará ausente' : 'Quem sai'}
                    </label>
                    <select
                      value={original}
                      onChange={(e) => setOriginal(e.target.value ? Number(e.target.value) : '')}
                      className={`w-full ${inputCls}`}
                    >
                      <option value="">Selecione…</option>
                      {membros.map((m) => (
                        <option key={m.id} value={m.id}>{m.nome}</option>
                      ))}
                    </select>
                  </div>
                )}
                {precisaSubstituto && (
                  <div>
                    <label className="mb-1 block text-sm font-medium">
                      {tipo === 'extra' ? 'Quem fará o plantão' : 'Quem entra'}
                    </label>
                    <select
                      value={substituto}
                      onChange={(e) => setSubstituto(e.target.value ? Number(e.target.value) : '')}
                      className={`w-full ${inputCls}`}
                    >
                      <option value="">Selecione…</option>
                      {membros.map((m) => (
                        <option key={m.id} value={m.id}>{m.nome}</option>
                      ))}
                    </select>
                  </div>
                )}
              </div>
            )}

            {tipo === 'troca' && (
              <div>
                <label className="mb-1 block text-sm font-medium">
                  Data em que quem entra será retribuído (assume os plantões do outro)
                </label>
                <input
                  type="date"
                  value={dataTroca}
                  onChange={(e) => setDataTroca(e.target.value)}
                  className={inputCls}
                />
              </div>
            )}

            {mostraHorarios && (
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="mb-1 block text-sm font-medium">Início</label>
                  <input
                    type="time"
                    value={inicio}
                    onChange={(e) => setInicio(e.target.value)}
                    className={`w-full ${inputCls}`}
                  />
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium">Fim</label>
                  <input
                    type="time"
                    value={fim}
                    onChange={(e) => setFim(e.target.value)}
                    className={`w-full ${inputCls}`}
                  />
                </div>
              </div>
            )}

            <div>
              <label className="mb-1 block text-sm font-medium">
                {tipo === 'feriado' ? 'Motivo (ex.: Natal, Ano Novo)' : 'Observação'}
              </label>
              <input
                type="text"
                value={obs}
                onChange={(e) => setObs(e.target.value)}
                placeholder={tipo === 'feriado' ? 'ex.: Natal' : 'ex.: consulta médica'}
                className={`w-full ${inputCls}`}
              />
            </div>

            <div className="flex justify-end gap-2">
              <button onClick={onFechar} className="rounded px-4 py-2 text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-700">
                Cancelar
              </button>
              <button
                onClick={criar}
                disabled={salvando}
                className="rounded bg-blue-600 px-4 py-2 font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                {salvando ? 'Gerando prévia…' : 'Ver prévia da alteração'}
              </button>
            </div>
          </div>
        ) : (
          <div className="space-y-4">
            <div className="rounded border border-blue-200 bg-blue-50 px-4 py-2 text-sm text-blue-800 dark:border-blue-800 dark:bg-blue-950 dark:text-blue-200">
              Confira a prévia abaixo. Nada será alterado até você confirmar.
            </div>
            {pendente.preview.dias.map((d) => (
              <div key={d.data} className="rounded border border-slate-200 p-3 dark:border-slate-700">
                <div className="mb-2 font-semibold">{dataBR(d.data)}</div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <div className="mb-1 text-xs font-semibold uppercase text-slate-500 dark:text-slate-400">Antes</div>
                    <ListaSlots slots={d.antes} />
                  </div>
                  <div>
                    <div className="mb-1 text-xs font-semibold uppercase text-slate-500 dark:text-slate-400">Depois</div>
                    <ListaSlots slots={d.depois} />
                  </div>
                </div>
              </div>
            ))}
            <div className="rounded border border-slate-200 p-3 dark:border-slate-700">
              <div className="mb-2 text-xs font-semibold uppercase text-slate-500 dark:text-slate-400">
                Horas no mês (antes → depois)
              </div>
              {pendente.preview.horas_depois.map((h) => {
                const antes = pendente.preview.horas_antes.find(
                  (a) => a.membro_id === h.membro_id,
                )
                const mudou = antes?.total_minutos !== h.total_minutos
                return (
                  <div key={h.membro_id} className={`text-sm ${mudou ? 'font-semibold' : ''}`}>
                    {h.nome}: {antes?.total_formatado ?? '0h00'} → {h.total_formatado}
                  </div>
                )
              })}
            </div>
            <div className="flex justify-end gap-2">
              <button
                disabled={salvando}
                onClick={() => decidir('descartar', pendente.excecao.id)}
                className="rounded px-4 py-2 text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-700"
              >
                Descartar
              </button>
              <button
                disabled={salvando}
                onClick={() => decidir('confirmar', pendente.excecao.id)}
                className="rounded bg-green-600 px-4 py-2 font-medium text-white hover:bg-green-700 disabled:opacity-50"
              >
                Confirmar alteração
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

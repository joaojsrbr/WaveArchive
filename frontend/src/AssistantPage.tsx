import { useEffect, useState, type FormEvent } from "react";
import { assistantChatStream, deleteAIConversation, getSettings, listAIConversations, listBuilds, listCharacters, listTeams, syncCharacterGuides, testAIProvider } from "./lib/backend";
import type { AIConversation, AIProviderStatus, AppSettings, Build, Character, Team } from "./types";

export function AssistantPage({ onError }: { onError: (message: string) => void }) {
  const [conversations, setConversations] = useState<AIConversation[]>([]);
  const [current, setCurrent] = useState<AIConversation>();
  const [builds, setBuilds] = useState<Build[]>([]);
  const [teams, setTeams] = useState<Team[]>([]);
  const [characters,setCharacters]=useState<Character[]>([]);
  const [context, setContext] = useState("general:0");
  const [question, setQuestion] = useState("");
  const [endpoint, setEndpoint] = useState(() => localStorage.getItem("wavearchive:ollama-endpoint") || "http://127.0.0.1:11434");
  const [model, setModel] = useState(() => localStorage.getItem("wavearchive:ollama-model") || "qwen2.5:7b");
  const [provider,setProvider]=useState<AppSettings["aiProvider"]>("ollama");
  const [mode,setMode]=useState<AppSettings["aiMode"]>("strict");
  const [apiKey,setAPIKey]=useState("");
  const [sending, setSending] = useState(false);
  const [liveText,setLiveText]=useState("");
  const [providerStatus,setProviderStatus]=useState<AIProviderStatus>();

  async function load() {
    try {
      const [nextConversations, nextBuilds, nextTeams, settings,nextCharacters] = await Promise.all([listAIConversations(), listBuilds(), listTeams(), getSettings(),listCharacters({query:"",element:0,rarity:0,ownedOnly:false,favorites:false,sort:"name"})]);
      setConversations(nextConversations); setBuilds(nextBuilds); setTeams(nextTeams);
      setCharacters(nextCharacters);
      setProvider(settings.aiProvider); setMode(settings.aiMode); setEndpoint(settings.aiEndpoint); setModel(settings.aiModel);
      if (current) setCurrent(nextConversations.find((item) => item.id === current.id));
      onError("");
    } catch (cause) { onError(messageFrom(cause)); }
  }
  useEffect(() => { void load(); const off=window.runtime?.EventsOn("ai:chunk",(payload)=>setLiveText(value=>value+String(payload)));return ()=>{off?.()} }, []);
  async function send(event: FormEvent) {
    event.preventDefault(); if (!question.trim()) return;
    const [contextType, rawID] = context.split(":");
    setSending(true);
    try {
      localStorage.setItem("wavearchive:ollama-endpoint", endpoint); localStorage.setItem("wavearchive:ollama-model", model);
      setLiveText("");
      const conversation = await assistantChatStream({
        conversationId: current?.id ?? 0,
        contextType: current?.contextType ?? contextType,
        contextId: current?.contextId ?? (Number(rawID) || undefined),
        question, endpoint, model, provider, apiKey, mode
      });
      setCurrent(conversation); setLiveText(""); setQuestion(""); await load();
    } catch (cause) { onError(messageFrom(cause)); }
    finally { setSending(false); }
  }
  async function testProvider(){
    try{setProviderStatus(await testAIProvider({provider,endpoint,model,apiKey,mode,context:"status",dataJson:"{}"}));onError("")}
    catch(cause){onError(messageFrom(cause))}
  }
  async function syncGuide(){
    const [kind,id]=context.split(":");if(kind!=="character"||!Number(id))return;
    try{const guides=await syncCharacterGuides(Number(id),"en");setProviderStatus({provider,online:true,models:providerStatus?.models??[],message:`${guides.length} guias oficiais armazenados`});onError("")}
    catch(cause){onError(messageFrom(cause))}
  }
  async function remove(id: number) {
    try { await deleteAIConversation(id); if (current?.id === id) setCurrent(undefined); await load(); }
    catch (cause) { onError(messageFrom(cause)); }
  }
  return <div className="assistantPage">
    <div className="pageIntro"><div><span className="eyebrow">{provider.toUpperCase()} · CONTEXTO CONTROLADO</span><h1>ASSISTENTE IA</h1><p>Pergunte sobre builds e equipes usando dados calculados pelo WaveArchive.</p></div></div>
    <div className="assistantLayout">
      <aside className="conversationSidebar"><button className="primaryButton" onClick={() => { setCurrent(undefined); setContext("general:0"); }}>＋ Nova conversa</button><div>{conversations.map((conversation) => <article className={current?.id === conversation.id ? "active" : ""} key={conversation.id} onClick={() => { setCurrent(conversation); setContext(`${conversation.contextType}:${conversation.contextId ?? 0}`); }}><span>{conversation.contextType.toUpperCase()}</span><strong>{conversation.title}</strong><small>{conversation.model}</small><button onClick={(event) => { event.stopPropagation(); void remove(conversation.id); }}>×</button></article>)}</div></aside>
      <section className="chatPanel">
        <header><div><span>✦</span><div><small>{current ? `${current.contextType} #${current.contextId ?? "geral"}` : `NOVA ANÁLISE · ${mode}`}</small><h2>{current?.title ?? "Como posso analisar seus dados?"}</h2>{providerStatus&&<em className="providerStatus">{providerStatus.online?"●":"○"} {providerStatus.message}</em>}</div></div><div><label>Endpoint<input value={endpoint} onChange={(event) => setEndpoint(event.target.value)} /></label><label>Modelo<input list="ai-models" value={model} onChange={(event) => setModel(event.target.value)} /></label><datalist id="ai-models">{providerStatus?.models.map(item=><option value={item} key={item}/>)}</datalist>{provider==="gemini"&&<label>Chave da sessão<input type="password" value={apiKey} onChange={e=>setAPIKey(e.target.value)} autoComplete="off"/></label>}<button onClick={()=>void testProvider()}>Testar e listar modelos</button></div></header>
        <div className="chatMessages">{!current?.messages.length&&!liveText ? <div className="assistantEmpty"><span>✦</span><h3>Análise fundamentada</h3><p>Escolha personagem, build ou equipe. O RAG busca somente no catálogo local e nos guias oficiais sincronizados.</p></div> : <>{current?.messages.map((message) => <article className={message.role} key={message.id}><small>{message.role === "user" ? "VOCÊ" : `${current.model} · ${current.provider.toUpperCase()}`}</small><p>{message.content}</p>{message.role==="assistant"&&<em className="aiSource">Fonte: {current.contextType} #{current.contextId??"geral"} · SQLite + engine local</em>}</article>)}{liveText&&<article className="assistant streaming"><small>GERANDO · STREAMING</small><p>{liveText}<i className="streamCursor">▋</i></p></article>}</>}</div>
        <form className="chatComposer" onSubmit={send}>{!current && <div className="contextPicker"><select value={context} onChange={(event) => setContext(event.target.value)}><option value="general:0">Contexto geral</option><optgroup label="Personagens">{characters.map(character=><option value={`character:${character.id}`} key={`c${character.id}`}>{character.name}</option>)}</optgroup><optgroup label="Builds">{builds.map((build) => <option value={`build:${build.id}`} key={`b${build.id}`}>{build.name}</option>)}</optgroup><optgroup label="Equipes">{teams.map((team) => <option value={`team:${team.id}`} key={`t${team.id}`}>{team.name}</option>)}</optgroup></select>{context.startsWith("character:")&&<button type="button" onClick={()=>void syncGuide()}>Sincronizar guias oficiais</button>}</div>}<textarea rows={3} value={question} onChange={(event) => setQuestion(event.target.value)} placeholder="Ex.: qual é o maior gargalo desta build e qual troca teria mais impacto?" /><button className="primaryButton" disabled={sending || !question.trim()}>{sending ? "Analisando…" : "Enviar"}</button><small>{provider==="gemini"?"Gemini é remoto. A chave não é persistida.":"O provedor local precisa estar em execução; o contexto não sai do computador."}</small></form>
      </section>
    </div>
  </div>;
}
function messageFrom(cause: unknown) { return cause instanceof Error ? cause.message : String(cause); }

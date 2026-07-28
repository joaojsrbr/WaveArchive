#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
wuwa_scraper.py
================

Raspa dados de personagens (e suas armas de assinatura) de Wuthering Waves
usando a API estática do Nanoka (https://static.nanoka.cc) e gera, para
cada personagem escolhido, um arquivo .md pronto para ser colado em uma IA,
já contendo um pré-prompt pedindo um resumo de kit/build.

COMO USAR
---------
A arma de assinatura de cada personagem é detectada AUTOMATICAMENTE (o
próprio JSON do personagem informa, no campo "recommend.weapon", qual é
a arma feita especificamente para ele). Você não precisa descobrir o ID
da arma na mão — mas pode informar um "weapon_id" manualmente se quiser
forçar uma arma diferente.

JEITO MAIS FÁCIL — modo interativo:
    python wuwa_scraper.py --pick

    Isso mostra a lista de todos os personagens com seus IDs, você digita
    os IDs que quiser (separados por vírgula) e, se quiser, pode informar
    o ID de uma arma específica (Enter para deixar o script detectar a
    arma de assinatura sozinho). Os .md já saem prontos na pasta ./output.

MODO COMPARAÇÃO — compare 2 ou mais personagens lado a lado:
    python wuwa_scraper.py --compare

    Mostra a lista de personagens, você escolhe 2+ IDs e o script gera um
    único arquivo .md com os dados de todos lado a lado, mais um pré-prompt
    pedindo à IA que compare os kits, diga quem é melhor em cada papel,
    quem escala mais com sequências e qual vale a pena puxar primeiro.

    Também aceita IDs direto na linha de comando:
        python wuwa_scraper.py --compare 1607 1610 1109

JEITO MANUAL — editando o script:
1) Edite a lista CHARACTERS_TO_SCRAPE lá embaixo com os personagens que
   você quer. Basta o "character_id" — deixe "weapon_id": None para a
   arma de assinatura ser detectada automaticamente.

   Para descobrir o ID de um personagem, rode:
       python wuwa_scraper.py --list

2) Rode:
       python wuwa_scraper.py

3) Os arquivos .md vão parar na pasta ./output, um por personagem.
   É só abrir o .md e colar o conteúdo inteiro numa IA (ChatGPT, Claude, etc).

REQUISITOS
----------
    pip install requests

A versão do jogo usada nas URLs (ex: "3.5.3") fica na constante GAME_VERSION.
Se quiser sempre pegar a versão mais nova automaticamente, deixe
AUTO_DETECT_VERSION = True (o script consulta o manifest.json do Nanoka).
"""

import json
import os
import re
import sys
import time
import unicodedata
from pathlib import Path
from typing import Optional

try:
    import requests
except ImportError:
    print("Este script precisa da lib 'requests'. Instale com:")
    print("    pip install requests")
    sys.exit(1)

try:
    from PIL import Image
    import io
    HAS_PILLOW = True
except ImportError:
    HAS_PILLOW = False


# =============================================================================
# CONFIGURAÇÃO — EDITE AQUI
# =============================================================================

# Se True, o script consulta o manifest.json e usa a versão "latest" do jogo.
# Se False, usa o valor fixo em GAME_VERSION.
AUTO_DETECT_VERSION = True
GAME_VERSION = "3.5.3"  # usado só se AUTO_DETECT_VERSION = False

# Idioma dos textos (en, pt, ko, ja, zh ... depende do que o Nanoka publica;
# "en" é o mais completo/estável)
LANGUAGE = "en"

# Pasta onde os .md finais serão salvos
OUTPUT_DIR = "output"

# Lista de personagens que você quer raspar (usada quando você roda só
# "python wuwa_scraper.py", sem flags). "weapon_id" é opcional (a arma de
# assinatura/signature weapon do personagem) — deixe None se não souber.
#
# Por padrão fica vazia: é mais fácil usar "python wuwa_scraper.py --pick"
# para escolher os personagens direto pelo terminal. Mas se preferir,
# preencha aqui manualmente, por exemplo:
#
# CHARACTERS_TO_SCRAPE = [
#     {"character_id": 1607, "weapon_id": 21050056, "name_hint": "Cantarella"},
#     {"character_id": 1610, "weapon_id": None, "name_hint": "Yangyang Xuanling"},
# ]
CHARACTERS_TO_SCRAPE = []

BASE_URL = "https://static.nanoka.cc"
REQUEST_DELAY_SECONDS = 0.3  # gentileza com o servidor entre requisições
TIMEOUT = 20


# =============================================================================
# FUNÇÕES DE REDE
# =============================================================================

def get_session() -> requests.Session:
    s = requests.Session()
    s.headers.update({
        "User-Agent": "Mozilla/5.0 (compatible; wuwa-scraper/1.0; +https://ww.nanoka.cc)"
    })
    return s


def fetch_json(session: requests.Session, url: str):
    """Busca um JSON, retorna None em caso de erro (404 etc)."""
    try:
        resp = session.get(url, timeout=TIMEOUT)
        if resp.status_code != 200:
            return None
        return resp.json()
    except (requests.RequestException, json.JSONDecodeError) as e:
        print(f"  [aviso] falha ao buscar {url}: {e}")
        return None


def detect_game_version(session: requests.Session) -> str:
    """Pega a versão 'latest' do WuWa a partir do manifest.json do Nanoka."""
    manifest = fetch_json(session, f"{BASE_URL}/manifest.json")
    if manifest and "ww" in manifest and "latest" in manifest["ww"]:
        version = manifest["ww"]["latest"]
        # versões tipo "3.5.3" às vezes vêm com sufixo "+hash"; tiramos isso
        version = version.split("+")[0]
        print(f"Versão detectada automaticamente: {version}")
        return version
    print(f"[aviso] não consegui detectar a versão, usando fallback {GAME_VERSION}")
    return GAME_VERSION


def fetch_character_index(session: requests.Session, version: str) -> dict:
    """Índice geral: id -> dados resumidos de cada personagem."""
    url = f"{BASE_URL}/ww/{version}/character.json"
    data = fetch_json(session, url)
    return data or {}


def fetch_character_detail(session: requests.Session, version: str, char_id: int) -> dict:
    url = f"{BASE_URL}/ww/{version}/{LANGUAGE}/character/{char_id}.json"
    return fetch_json(session, url)


def fetch_weapon_detail(session: requests.Session, version: str, weapon_id: int) -> dict:
    url = f"{BASE_URL}/ww/{version}/{LANGUAGE}/weapon/{weapon_id}.json"
    return fetch_json(session, url)


# =============================================================================
# IMAGENS — DOWNLOAD E CONVERSÃO
# =============================================================================

ASSETS_BASE_URL = "https://static.nanoka.cc/assets/ww"


def build_image_url(raw_path: str) -> str:
    """
    Converte path Unity do JSON para URL real de imagem .webp.

    Ex: "/Game/Aki/UI/UIResources/Common/Image/IconRoleHead256/T_IconRoleHead256_70_UI.T_IconRoleHead256_70_UI"
    →   "https://static.nanoka.cc/assets/ww/UIResources/Common/Image/IconRoleHead256/T_IconRoleHead256_70_UI.webp"
    """
    prefix = "/Game/Aki/UI/"
    if raw_path.startswith(prefix):
        relative = raw_path[len(prefix):]
    else:
        relative = raw_path.lstrip("/")

    # O path vem duplicado: "Folder/Filename.Filename" — pega só o nome antes do ponto
    parts = relative.rsplit("/", 1)
    if len(parts) == 2:
        folder, filename_dup = parts
        filename = filename_dup.split(".")[0]
        return f"{ASSETS_BASE_URL}/{folder}/{filename}.webp"

    filename = relative.split(".")[0]
    return f"{ASSETS_BASE_URL}/{filename}.webp"


def download_and_convert_image(session: requests.Session, url: str, dest_path: str) -> bool:
    """Baixa imagem .webp e converte para .png. Retorna True se sucesso."""
    if not HAS_PILLOW:
        print("  [erro] Pillow não instalado. Rode: pip install Pillow")
        return False
    try:
        resp = session.get(url, timeout=TIMEOUT)
        if resp.status_code != 200:
            print(f"  [aviso] HTTP {resp.status_code} ao baixar {url}")
            return False
        img = Image.open(io.BytesIO(resp.content))
        img.save(dest_path, "PNG")
        return True
    except Exception as e:
        print(f"  [aviso] falha ao baixar/converter {url}: {e}")
        return False


def download_character_images(
    session: requests.Session,
    index_entry: dict,
    char_slug: str,
    out_dir: Path,
) -> dict:
    """Baixa icon + background do personagem como PNG. Retorna dict com filenames."""
    files = {}

    icon_path = index_entry.get("icon", "")
    bg_path = index_entry.get("background", "")

    if icon_path:
        url = build_image_url(icon_path)
        fname = f"{char_slug}_icon.png"
        dest = out_dir / fname
        print(f"   -> Baixando icon: {url}")
        if download_and_convert_image(session, url, str(dest)):
            files["icon"] = fname
            print(f"      [ok] {fname}")
        time.sleep(REQUEST_DELAY_SECONDS)

    if bg_path:
        url = build_image_url(bg_path)
        fname = f"{char_slug}_bg.png"
        dest = out_dir / fname
        print(f"   -> Baixando background: {url}")
        if download_and_convert_image(session, url, str(dest)):
            files["background"] = fname
            print(f"      [ok] {fname}")
        time.sleep(REQUEST_DELAY_SECONDS)

    return files


def download_weapon_image(
    session: requests.Session,
    weapon_data: dict,
    char_slug: str,
    out_dir: Path,
) -> Optional[str]:
    """Baixa icon da arma como PNG. Retorna filename ou None."""
    icon_path = weapon_data.get("icon", "")
    if not icon_path:
        return None

    url = build_image_url(icon_path)
    fname = f"{char_slug}_weapon.png"
    dest = out_dir / fname
    print(f"   -> Baixando icon da arma: {url}")
    if download_and_convert_image(session, url, str(dest)):
        print(f"      [ok] {fname}")
        time.sleep(REQUEST_DELAY_SECONDS)
        return fname
    time.sleep(REQUEST_DELAY_SECONDS)
    return None


# =============================================================================
# LIMPEZA DE TEXTO
# =============================================================================

# Os textos do jogo vêm com tags tipo <te href=850094>Fisalia</te>,
# <color=Dark>...</color>, <size=40>...</size> etc. Isso é formatação
# de rich-text do Unity e não interessa pra um resumo em markdown.

TAG_RE = re.compile(r"<[^>]+>")


def clean_text(text):
    """Remove tags de rich-text (<te>, <color>, <size>...) e normaliza espaços."""
    if not text:
        return ""
    if not isinstance(text, str):
        return text
    text = TAG_RE.sub("", text)
    text = text.replace("\\n", "\n")
    # remove espaços duplicados nas linhas, mas preserva quebras de parágrafo
    lines = [ln.strip() for ln in text.split("\n")]
    return "\n".join(lines).strip()


# As descrições de skills/sequências usam placeholders {0}, {1}, ... cujos
# valores reais vêm em uma lista separada ("param" ou "simple_param"). Sem
# substituir isso, a IA recebe textos cheios de buracos tipo "por {0}s".
PLACEHOLDER_RE = re.compile(r"\{(\d+)\}")

# Tags tipo {Cus:Ipt,Touch=Tap PC=Press Gamepad=Press} são variações de
# texto por plataforma/gênero (ex: {Male=he;Female=she}). Pegamos sempre a
# primeira opção disponível, que já dá um texto legível.
CUS_TAG_RE = re.compile(r"\{Cus:[^,]+,([^}]+)\}")
GENDER_TAG_RE = re.compile(r"\{(?:Male|Female)=([^;}]+)(?:;[^}]+)?\}")
SAPTAG_RE = re.compile(r"<SapTag=\d+>(.*?)</SapTag>")


def apply_params(text: str, params) -> str:
    """Substitui {0}, {1}... pelos valores em params, e simplifica tags
    condicionais de plataforma/gênero pegando a primeira opção."""
    if not text:
        return ""

    def replace_index(match):
        idx = int(match.group(1))
        if params and 0 <= idx < len(params):
            value = params[idx]
            # alguns params são listas aninhadas (ex: [["40.00%", ...]])
            if isinstance(value, list):
                value = value[0] if value else ""
            return str(value)
        return match.group(0)  # mantém {N} se não houver valor (melhor que apagar)

    text = PLACEHOLDER_RE.sub(replace_index, text)

    # {Cus:Ipt,Touch=Tap PC=Press Gamepad=Press} -> pega a primeira opção (Tap)
    def replace_cus(match):
        options = match.group(1).strip().split()
        if options:
            first = options[0]
            return first.split("=")[-1] if "=" in first else first
        return ""

    text = CUS_TAG_RE.sub(replace_cus, text)
    # {Male=he;Female=she} -> "he"
    text = GENDER_TAG_RE.sub(lambda m: m.group(1), text)
    # <SapTag=1>texto</SapTag> -> texto
    text = SAPTAG_RE.sub(lambda m: m.group(1), text)

    return text


def slugify(name: str) -> str:
    """Transforma um nome em algo seguro para nome de arquivo."""
    if not name:
        return "unknown"
    name = unicodedata.normalize("NFKD", name).encode("ascii", "ignore").decode("ascii")
    name = re.sub(r"[^\w\s-]", "", name).strip().lower()
    name = re.sub(r"[\s_-]+", "_", name)
    return name or "unknown"


# =============================================================================
# MAPAS DE CÓDIGOS (element / weapon type) — conforme observado nos dados
# =============================================================================

ELEMENT_MAP = {
    1: "Glacio",
    2: "Fusion",
    3: "Electro",
    4: "Aero",
    5: "Spectro",
    6: "Havoc",
}

WEAPON_TYPE_MAP = {
    1: "Broadblade",
    2: "Sword",
    3: "Pistols",
    4: "Gauntlets",
    5: "Rectifier",
}

ELEMENT_PALETTE = {
    1: {  # Glacio
        "background": "deep navy/void (#0a0e1a)",
        "accent": "icy cyan (#7fdbff)",
        "text": "frost white (#f0f8ff)",
        "secondary": "pale blue (#b8d4e3)",
        "motif": "ice crystals, frost patterns, snowflakes",
        "mood": "cold, serene, crystalline",
    },
    2: {  # Fusion
        "background": "charcoal/dark ember (#1a0a0a)",
        "accent": "molten orange (#ff6b35)",
        "text": "warm gold (#ffd700)",
        "secondary": "ember red (#c0392b)",
        "motif": "flames, embers, heat distortion",
        "mood": "fiery, passionate, intense",
    },
    3: {  # Electro
        "background": "deep indigo (#0d0221)",
        "accent": "electric violet (#b388ff)",
        "text": "lightning white (#e8e0ff)",
        "secondary": "plasma blue (#6c5ce7)",
        "motif": "sparks, electric arcs, lightning bolts",
        "mood": "electric, dynamic, crackling energy",
    },
    4: {  # Aero
        "background": "midnight teal (#0a1a1a)",
        "accent": "mint/sky blue (#64ffda)",
        "text": "cloud white (#f5f5f5)",
        "secondary": "soft teal (#26a69a)",
        "motif": "wind currents, feathers, flowing ribbons",
        "mood": "breezy, free, ethereal",
    },
    5: {  # Spectro
        "background": "dark bronze (#1a1408)",
        "accent": "golden amber (#ffab00)",
        "text": "warm ivory (#fff8e1)",
        "secondary": "honey (#f9a825)",
        "motif": "light rays, prisms, golden particles",
        "mood": "radiant, warm, luminous",
    },
    6: {  # Havoc
        "background": "midnight black (#0a0008)",
        "accent": "dark magenta/violet (#9c27b0)",
        "text": "pale crimson (#ffcdd2)",
        "secondary": "deep plum (#6a1b9a)",
        "motif": "feathers, shadows, dark petals, void particles",
        "mood": "dark, mysterious, dangerous elegance",
    },
}


def describe_element(code):
    return ELEMENT_MAP.get(code, f"Desconhecido ({code})")


def describe_weapon_type(code):
    return WEAPON_TYPE_MAP.get(code, f"Desconhecido ({code})")


# =============================================================================
# EXTRAÇÃO / FORMATAÇÃO DOS DADOS DO PERSONAGEM
# =============================================================================

def extract_signature_weapon_id(char_data: dict) -> Optional[int]:
    """
    Descobre o ID da arma de assinatura (signature weapon) do personagem
    automaticamente, sem precisar que o usuário informe na mão.

    No JSON do personagem existe um campo "recommend": {"weapon": [...]}
    com uma lista de armas recomendadas para aquele personagem — o
    primeiro ID da lista é a arma de assinatura (a feita especificamente
    para ele). Também existe, em alguns dumps, uma lista solta "weapon"
    fora de "recommend" com o mesmo propósito — tentamos os dois, nessa
    ordem de confiança.
    """
    recommend = char_data.get("recommend")
    if isinstance(recommend, dict):
        weapons = recommend.get("weapon")
        if isinstance(weapons, list) and weapons:
            first = weapons[0]
            if isinstance(first, int):
                return first
            if isinstance(first, str) and first.isdigit():
                return int(first)

    # Fallback: às vezes a lista de armas recomendadas aparece solta,
    # como uma lista (não como o int de "tipo de arma", que também usa
    # a chave "weapon" no nível raiz do personagem).
    loose_weapon = char_data.get("weapon")
    if isinstance(loose_weapon, list) and loose_weapon:
        first = loose_weapon[0]
        if isinstance(first, int):
            return first
        if isinstance(first, str) and first.isdigit():
            return int(first)

    return None


def extract_chains(char_data: dict) -> list:
    """
    Extrai as Sequências/Cadeias (constelações) do personagem, com nome
    e descrição limpa (já com os valores reais no lugar de {0}, {1}...).
    É informação central para entender como o kit escala com cópias
    extras (S1~S6).
    """
    chains = []
    raw = char_data.get("chains") or {}
    for chain_id, chain in sorted(
        raw.items(), key=lambda kv: int(kv[0]) if str(kv[0]).isdigit() else 0
    ):
        name = chain.get("name", "")
        raw_desc = apply_params(chain.get("desc", ""), chain.get("param"))
        desc = clean_text(raw_desc)
        if not name and not desc:
            continue
        chains.append({"id": chain_id, "name": name, "desc": desc})
    return chains


def extract_skills(char_data: dict) -> list:
    """
    Extrai as skills de skill_trees, pegando nome + descrição com os
    placeholders {0}, {1}... já substituídos pelos valores reais (via
    "simple_param"/"param"), sem os arrays de dano por nível (são dados
    demais e não ajudam a IA a resumir o "estilo de jogo" do personagem).
    """
    skills = []
    skill_trees = char_data.get("skill_trees", {})
    for _, node in skill_trees.items():
        skill = node.get("skill")
        if not skill or not skill.get("name"):
            continue
        desc_raw = skill.get("simple_desc") or skill.get("desc", "")
        params = skill.get("simple_param") or skill.get("param")
        desc = clean_text(apply_params(desc_raw, params))
        skills.append({
            "type": skill.get("type", ""),
            "name": skill.get("name", ""),
            "desc": desc,
        })
    return skills


def extract_tags(char_data: dict) -> list:
    tags = []
    for _, tag in (char_data.get("tag") or {}).items():
        name = tag.get("name")
        desc = tag.get("desc")
        if name:
            tags.append(f"{name}" + (f" — {desc}" if desc else ""))
    return tags


def extract_forte_summary(char_data: dict) -> list:
    """
    Pega a explicação em prosa do mecanismo central do kit (campo
    "forte.desc_list"). É um resumo curto, em texto corrido, de como o
    "gimmick" do personagem funciona (ex: sistema de Trance/Shiver da
    Cantarella) — ótimo contexto de alto nível para a IA antes de entrar
    nos detalhes de cada skill.
    """
    forte = char_data.get("forte") or {}
    desc_list = forte.get("desc_list") or []
    return [clean_text(d) for d in desc_list if d]


def extract_skin_names(char_data: dict) -> list:
    """Nomes das skins/trajes do personagem (informação leve, opcional)."""
    names = []
    for _, skin in (char_data.get("skin") or {}).items():
        name = skin.get("name") or skin.get("title_name")
        title = skin.get("title_name")
        if name:
            label = name if not title or title == name else f"{name} ({title})"
            names.append(label)
    return names


def extract_goods(char_data: dict) -> list:
    """Itens/relíquias pessoais (lore), tipo o objeto especial do personagem."""
    goods = []
    for g in char_data.get("goods", []) or []:
        goods.append({
            "title": g.get("title", ""),
            "content": clean_text(g.get("content", "")),
        })
    return goods


def extract_stories(char_data: dict) -> list:
    stories = []
    for s in char_data.get("stories", []) or []:
        stories.append({
            "title": s.get("title", ""),
            "content": clean_text(s.get("content", "")),
        })
    return stories


def fetch_guide_recommendations(session: requests.Session, role_id: str) -> dict:
    """Busca atributos, echo e time recomendados na API oficial de guias."""
    list_url = f"https://guide-server.aki-game.net/introduction/list?roleGbId={role_id}"
    headers = {"User-Agent": "Mozilla/5.0", "x-language": "en"}
    try:
        r = session.get(list_url, headers=headers)
        data = r.json()
        items = data.get("data", [])
        if not items:
            return {"attributes": [], "echo": "", "teammates": []}

        for guide in items:
            guide_id = guide.get("id")
            if not guide_id:
                continue

            info_url = f"https://guide-server.aki-game.net/introduction/info?roleGbId={role_id}&id={guide_id}"
            ir = session.get(info_url, headers=headers)
            idata = ir.json().get("data")
            if not idata:
                continue

            # Attributes
            results = []
            role_attr = idata.get("roleAttribute", {})
            attr_items = role_attr.get("items", [])
            if attr_items:
                for attr in attr_items:
                    texts = attr.get("texts", [])
                    name = ""
                    for t in texts:
                        if t.get("language") == "en":
                            name = t.get("name")
                            break
                    if not name and texts:
                        name = texts[0].get("name")

                    amt = attr.get("recommendAmount")
                    if name:
                        if amt:
                            results.append(f"{name} ({amt})")
                        else:
                            results.append(name)

            # Echo
            echo_name = ""
            echo_main = idata.get("echo", {}).get("main", {})
            echo_texts = echo_main.get("echoProps", {}).get("texts", [])
            if echo_texts:
                for t in echo_texts:
                    if t.get("language") == "en":
                        echo_name = t.get("name")
                        break
                if not echo_name and echo_texts:
                    echo_name = echo_texts[0].get("name")

            # Teammates
            teammates = []
            tm_items = idata.get("teammate", {}).get("items", [])
            if tm_items:
                for item in tm_items:
                    tm_main = item.get("main", {})
                    tm_texts = tm_main.get("texts", [])
                    if tm_texts:
                        tm_name = ""
                        for t in tm_texts:
                            if t.get("language") == "en":
                                tm_name = t.get("name")
                                break
                        if not tm_name and tm_texts:
                            tm_name = tm_texts[0].get("name")

                        tm_pic = tm_main.get("cardPictureUrl", "")
                        if tm_name:
                            if tm_pic:
                                teammates.append(f"{tm_name} ({tm_pic})")
                            else:
                                teammates.append(tm_name)

            # Skills Priority
            skills_rec = []
            add_point = idata.get("roleSkill", {}).get("addPointTarget", [])
            for s in add_point:
                s_texts = s.get("texts", [])
                s_name = ""
                for t in s_texts:
                    if t.get("language") == "en":
                        s_name = t.get("name")
                        break
                if not s_name and s_texts:
                    s_name = s_texts[0].get("name")

                s_level = s.get("recommendLevel")
                if s_name and s_level:
                    skills_rec.append(f"{s_name} (Level {s_level})")

            return {
                "attributes": results,
                "echo": echo_name,
                "teammates": teammates,
                "skills": skills_rec
            }
    except Exception as e:
        print(f"   [aviso] Falha ao buscar guia para {role_id}: {e}")
    return {"attributes": [], "echo": "", "teammates": [], "skills": []}


# =============================================================================
# MONTAGEM DO MARKDOWN
# =============================================================================

PRE_PROMPT_TEMPLATE = """\
# PROMPT PARA IA — NÃO APAGUE ESTA SEÇÃO

Você é um especialista em Wuthering Waves. Abaixo estão os dados brutos
(extraídos diretamente do datamine do jogo) do personagem **{char_name}**
{weapon_clause}.

Com base SOMENTE nas informações fornecidas abaixo, faça um resumo claro e
objetivo contendo:

1. **Visão geral do personagem** (papel no time: DPS, sub-DPS, suporte,
   healer, etc — use as tags fornecidas como referência).
2. **Resumo do kit/skills**: explique de forma simples o que cada
   habilidade faz e como elas se conectam (combo, rotação básica).
3. **Sequências (chains)**: resuma o que cada sequência (S1 a S6) agrega
   ao kit, e diga se alguma é especialmente importante (ex: libera um
   comportamento novo) versus as que só aumentam dano.
4. **Pontos fortes e fracos** do kit, na sua avaliação técnica.
5. **Sinergias**: que tipo de personagem/elemento combina bem no time.
6. {weapon_instruction}
7. Não invente números de dano específicos que não estejam no texto —
   foque na lógica do kit, não nas porcentagens.
8. Pode usar lore/contexto do personagem apenas como toque final, não como
   foco principal.
9. Pesquise na internet o basico do Wuthering Waves
10. Quero o basico da lore se possivel (opcional)

Seja direto, use bullet points quando fizer sentido, e escreva como se
fosse para alguém que já manja do jogo mas quer entender rápido se vale a
pena puxar (pull) esse personagem.

# Regras

- Não invente informações.
- Não invente porcentagens.
- Não faça cálculos de dano.
- Não use termos vagos como "muito forte" sem justificativa.
- Baseie todas as conclusões nas informações fornecidas.
- Priorize análise sobre descrição.
- Escreva de forma parecida com uma análise de theorycrafting profissional.

## Análise das Sequências

Para cada sequência (S1–S6):

### S1
- O que adiciona ao kit.
- Impacto real.

### S2
- O que adiciona ao kit.
- Impacto real.

(... repetir até S6)

Ao final, faça um resumo:

### Sequências Mais Importantes
Classifique:

- Alta prioridade
- Média prioridade
- Baixa prioridade

Explique quais realmente mudam a gameplay e quais são apenas aumento de dano.

---

## Arma Assinatura

Analise a arma separadamente.

Explique:

- O que ela fortalece.
- Quais partes do kit recebem mais benefício.
- Se parece essencial ou apenas otimizada.
- Quanto sua passiva conversa com as mecânicas do personagem.

---

## Veredito

Conclua em até 10 linhas:

- Qual é a proposta do personagem.
- Para quem ele parece ser destinado.
- O quão completo o kit parece.
- Se as mecânicas possuem boa sinergia interna.


---

# DADOS BRUTOS

"""


def build_markdown(char_data: dict, weapon_data: Optional[dict], version: str, guide_data: dict = None) -> str:
    char_name = char_data.get("name") or char_data.get("nick_name") or "Desconhecido"
    nickname = char_data.get("nick_name", "")
    rarity = char_data.get("rarity", "?")
    weapon_type = describe_weapon_type(char_data.get("weapon"))
    element = describe_element(char_data.get("element"))
    desc = clean_text(char_data.get("desc", ""))

    info = char_data.get("chara_info", {}) or {}
    tags = extract_tags(char_data)
    forte_summary = extract_forte_summary(char_data)
    skills = extract_skills(char_data)
    chains = extract_chains(char_data)
    skin_names = extract_skin_names(char_data)
    goods = extract_goods(char_data)
    stories = extract_stories(char_data)

    weapon_clause = "e sua arma de assinatura" if weapon_data else "(sem arma de assinatura informada)"
    weapon_instruction = (
        "Comente também como a arma de assinatura reforça (ou não) o kit do personagem."
        if weapon_data else
        "Ignore este item, pois nenhuma arma foi fornecida."
    )

    md = []
    md.append(PRE_PROMPT_TEMPLATE.format(
        char_name=char_name,
        weapon_clause=weapon_clause,
        weapon_instruction=weapon_instruction,
    ))

    # ---- Ficha geral ----
    md.append(f"## {char_name}" + (f' "{nickname}"' if nickname and nickname != char_name else ""))
    md.append("")
    md.append(f"- **Raridade:** {rarity} estrelas")
    md.append(f"- **Elemento:** {element}")
    md.append(f"- **Tipo de arma:** {weapon_type}")
    if info.get("birth"):
        md.append(f"- **Aniversário:** {info.get('birth')}")
    if info.get("sex"):
        md.append(f"- **Sexo:** {info.get('sex')}")
    if info.get("country"):
        md.append(f"- **Região/País:** {info.get('country')}")
    if info.get("influence"):
        md.append(f"- **Facção/Influência:** {info.get('influence')}")
    if skin_names:
        md.append(f"- **Skin(s) disponível(is):** {', '.join(skin_names)}")
    md.append("")

    if desc:
        md.append("### Descrição")
        md.append(desc)
        md.append("")

    if tags:
        md.append("### Tags de papel no time")
        for t in tags:
            md.append(f"- {t}")
        md.append("")

    if guide_data:
        attrs = guide_data.get("attributes", [])
        if attrs:
            md.append("### Atributos Recomendados (Main/Substats)")
            for attr in attrs:
                md.append(f"- {attr}")
            md.append("")

        echo = guide_data.get("echo")
        if echo:
            md.append("### Echo Recomendado")
            md.append(f"- {echo}")
            md.append("")

        team = guide_data.get("teammates", [])
        if team:
            md.append("### Time Recomendado")
            for t in team:
                md.append(f"- {t}")
            md.append("")

        skills_rec = guide_data.get("skills", [])
        if skills_rec:
            md.append("### Prioridade de Skills")
            for sr in skills_rec:
                md.append(f"- {sr}")
            md.append("")

    # ---- Talento (Forte / habilidade especial — versão LORE/narrativa) ----
    if info.get("talent_name"):
        md.append(f"### Forte (lore): {info.get('talent_name')}")
        if info.get("talent_doc"):
            md.append(clean_text(info.get("talent_doc")))
        md.append("")

    # ---- Mecânica do Forte (explicação técnica de como o gimmick funciona) ----
    if forte_summary:
        md.append("### Como o Forte funciona (mecânica)")
        md.append("")
        for paragraph in forte_summary:
            md.append(paragraph)
            md.append("")

    # ---- Skills / kit ----
    if skills:
        md.append("### Skills / Kit")
        for s in skills:
            md.append(f"**{s['name']}** ({s['type']})")
            md.append("")
            md.append(s["desc"])
            md.append("")

    # ---- Chains / Sequências (constelações) ----
    if chains:
        md.append("### Sequências (Chains / Constelações)")
        md.append("")
        for c in chains:
            md.append(f"**Sequência {c['id']}: {c['name']}**")
            md.append("")
            md.append(c["desc"])
            md.append("")

    # ---- Arma de assinatura ----
    if weapon_data:
        md.append("---")
        md.append("")
        md.append("## Arma de Assinatura")
        md.append("")
        w_name = weapon_data.get("name", "Desconhecida")
        w_rarity = weapon_data.get("rarity", "?")
        w_desc = clean_text(weapon_data.get("desc", ""))
        md.append(f"- **Nome:** {w_name}")
        md.append(f"- **Raridade:** {w_rarity} estrelas")
        if weapon_data.get("weapon_type") is not None:
            md.append(f"- **Tipo:** {describe_weapon_type(weapon_data.get('weapon_type'))}")
        md.append("")
        if w_desc:
            md.append("### Descrição / Lore da arma")
            md.append(w_desc)
            md.append("")

        # efeito/skill da arma (a chave exata pode variar; tentamos as mais comuns)
        weapon_skill = (
            weapon_data.get("skill")
            or weapon_data.get("effect")
            or weapon_data.get("skill_desc")
        )
        if weapon_skill:
            md.append("### Efeito (Passiva da arma)")
            if isinstance(weapon_skill, dict):
                w_skill_name = weapon_skill.get("name", "")
                w_skill_desc = clean_text(
                    weapon_skill.get("simple_desc") or weapon_skill.get("desc", "")
                )
                if w_skill_name:
                    md.append(f"**{w_skill_name}**")
                md.append(w_skill_desc)
            else:
                md.append(clean_text(str(weapon_skill)))
            md.append("")

    # ---- Lore complementar (opcional, ajuda a IA a dar um toque de personalidade) ----
    if goods:
        md.append("---")
        md.append("")
        md.append("## Itens pessoais (lore)")
        for g in goods:
            md.append(f"**{g['title']}**")
            md.append("")
            md.append(g["content"])
            md.append("")

    if stories:
        md.append("---")
        md.append("")
        md.append("## Histórias / Stories (lore extra, opcional)")
        for s in stories:
            md.append(f"### {s['title']}")
            md.append(s["content"])
            md.append("")

    md.append("---")
    md.append(f"*Dados extraídos de static.nanoka.cc — Wuthering Waves v{version} — idioma: {LANGUAGE}*")

    return "\n".join(md)


def build_raw_markdown(char_data: dict, weapon_data: Optional[dict], version: str, guide_data: dict = None) -> str:
    """Gera .md somente com dados brutos, sem nenhum prompt/instrução."""
    char_name = char_data.get("name") or char_data.get("nick_name") or "Desconhecido"
    nickname = char_data.get("nick_name", "")
    rarity = char_data.get("rarity", "?")
    weapon_type = describe_weapon_type(char_data.get("weapon"))
    element = describe_element(char_data.get("element"))
    desc = clean_text(char_data.get("desc", ""))

    info = char_data.get("chara_info", {}) or {}
    tags = extract_tags(char_data)
    forte_summary = extract_forte_summary(char_data)
    skills = extract_skills(char_data)
    chains = extract_chains(char_data)
    skin_names = extract_skin_names(char_data)
    goods = extract_goods(char_data)
    stories = extract_stories(char_data)

    md = []
    md.append(f"# {char_name}" + (f' \"{nickname}\"' if nickname and nickname != char_name else ""))
    md.append("")
    md.append(f"- **Raridade:** {rarity} estrelas")
    md.append(f"- **Elemento:** {element}")
    md.append(f"- **Tipo de arma:** {weapon_type}")
    if info.get("birth"):
        md.append(f"- **Aniversário:** {info.get('birth')}")
    if info.get("sex"):
        md.append(f"- **Sexo:** {info.get('sex')}")
    if info.get("country"):
        md.append(f"- **Região/País:** {info.get('country')}")
    if info.get("influence"):
        md.append(f"- **Facção/Influência:** {info.get('influence')}")
    if skin_names:
        md.append(f"- **Skin(s) disponível(is):** {', '.join(skin_names)}")
    md.append("")

    if desc:
        md.append("## Descrição")
        md.append(desc)
        md.append("")

    if tags:
        md.append("## Tags de papel no time")
        for t in tags:
            md.append(f"- {t}")
        md.append("")

    if guide_data:
        attrs = guide_data.get("attributes", [])
        if attrs:
            md.append("## Atributos Recomendados")
            for attr in attrs:
                md.append(f"- {attr}")
            md.append("")

        echo = guide_data.get("echo")
        if echo:
            md.append("## Echo Recomendado")
            md.append(f"- {echo}")
            md.append("")

        team = guide_data.get("teammates", [])
        if team:
            md.append("## Time Recomendado")
            for t in team:
                md.append(f"- {t}")
            md.append("")

        skills_rec = guide_data.get("skills", [])
        if skills_rec:
            md.append("## Prioridade de Skills")
            for sr in skills_rec:
                md.append(f"- {sr}")
            md.append("")

    if info.get("talent_name"):
        md.append(f"## Forte (lore): {info.get('talent_name')}")
        if info.get("talent_doc"):
            md.append(clean_text(info.get("talent_doc")))
        md.append("")

    if forte_summary:
        md.append("## Como o Forte funciona (mecânica)")
        md.append("")
        for paragraph in forte_summary:
            md.append(paragraph)
            md.append("")

    if skills:
        md.append("## Skills / Kit")
        for s in skills:
            md.append(f"**{s['name']}** ({s['type']})")
            md.append("")
            md.append(s["desc"])
            md.append("")

    if chains:
        md.append("## Sequências (Chains / Constelações)")
        md.append("")
        for c in chains:
            md.append(f"**Sequência {c['id']}: {c['name']}**")
            md.append("")
            md.append(c["desc"])
            md.append("")

    if weapon_data:
        md.append("---")
        md.append("")
        md.append("## Arma de Assinatura")
        md.append("")
        w_name = weapon_data.get("name", "Desconhecida")
        w_rarity = weapon_data.get("rarity", "?")
        w_desc = clean_text(weapon_data.get("desc", ""))
        md.append(f"- **Nome:** {w_name}")
        md.append(f"- **Raridade:** {w_rarity} estrelas")
        if weapon_data.get("weapon_type") is not None:
            md.append(f"- **Tipo:** {describe_weapon_type(weapon_data.get('weapon_type'))}")
        md.append("")
        if w_desc:
            md.append("### Descrição / Lore da arma")
            md.append(w_desc)
            md.append("")
        weapon_skill = (
            weapon_data.get("skill")
            or weapon_data.get("effect")
            or weapon_data.get("skill_desc")
        )
        if weapon_skill:
            md.append("### Efeito (Passiva da arma)")
            if isinstance(weapon_skill, dict):
                w_skill_name = weapon_skill.get("name", "")
                w_skill_desc = clean_text(
                    weapon_skill.get("simple_desc") or weapon_skill.get("desc", "")
                )
                if w_skill_name:
                    md.append(f"**{w_skill_name}**")
                md.append(w_skill_desc)
            else:
                md.append(clean_text(str(weapon_skill)))
            md.append("")

    if goods:
        md.append("---")
        md.append("")
        md.append("## Itens pessoais (lore)")
        for g in goods:
            md.append(f"**{g['title']}**")
            md.append("")
            md.append(g["content"])
            md.append("")

    if stories:
        md.append("---")
        md.append("")
        md.append("## Histórias / Stories (lore extra)")
        for s in stories:
            md.append(f"### {s['title']}")
            md.append(s["content"])
            md.append("")

    md.append("---")
    md.append(f"*Dados extraídos de static.nanoka.cc — Wuthering Waves v{version} — idioma: {LANGUAGE}*")

    return "\n".join(md)


def build_card_markdown(
    char_data: dict,
    weapon_data: Optional[dict],
    version: str,
    image_files: dict,
    guide_data: dict = None,
) -> str:
    """
    Gera .md autocontido: prompt Awwwards + dados brutos.
    Pronto para colar no Gemini/ChatGPT junto com a imagem de referência.
    """
    char_name = char_data.get("name") or char_data.get("nick_name") or "Desconhecido"
    nickname = char_data.get("nick_name", "")
    element_code = char_data.get("element", 0)
    element_name = describe_element(element_code)
    weapon_type = describe_weapon_type(char_data.get("weapon"))
    rarity = char_data.get("rarity", "?")

    tags = extract_tags(char_data)
    tags_str = ", ".join([t.split(" — ")[0] for t in tags]) if tags else "N/A"

    attrs_str = ", ".join(guide_data.get("attributes", [])) if guide_data else ""
    echo_str = guide_data.get("echo", "") if guide_data else ""
    team_str = ", ".join(guide_data.get("teammates", [])) if guide_data else ""
    skills_str = ", ".join(guide_data.get("skills", [])) if guide_data else ""

    desc = clean_text(char_data.get("desc", ""))
    desc_short = desc.split(".")[0].strip() if desc else ""
    if len(desc_short.split()) > 15:
        desc_short = " ".join(desc_short.split()[:15]) + "..."

    palette = ELEMENT_PALETTE.get(element_code, ELEMENT_PALETTE[6])

    bg_file = image_files.get("background", "N/A")
    icon_file = image_files.get("icon", "N/A")
    weapon_file = image_files.get("weapon", "N/A")

    weapon_name = weapon_data.get("name", "") if weapon_data else ""

    md = []

    # ---- INSTRUÇÕES ----
    md.append(f"# Card Awwwards — {char_name}")
    md.append("")
    md.append("## Como usar")
    md.append("1. Abra o Gemini ou ChatGPT.")
    if weapon_file != "N/A":
        md.append(f"2. **Anexe as imagens de referência** (`{bg_file}` e `{weapon_file}`) — arte full do personagem e ícone da arma.")
    else:
        md.append(f"2. **Anexe a imagem de referência** (`{bg_file}`) — esta é a arte full do personagem.")
    md.append("3. **Cole todo o conteúdo deste arquivo** na mesma mensagem.")
    md.append("4. Envie e aguarde a IA gerar o card.")
    md.append("")
    img_note = f"> Imagens disponíveis na mesma pasta: `{icon_file}` (ícone), `{bg_file}` (arte full)"
    if weapon_file != "N/A":
        img_note += f", `{weapon_file}` (arma)"
    md.append(img_note)
    md.append("")

    # ---- PROMPT ----
    md.append("---")
    md.append("")
    md.append("## Prompt")
    md.append("")
    md.append("```")
    if weapon_file != "N/A":
        md.append(f"You have attachments: reference images (character art and signature weapon) and this markdown file with {char_name}'s data and lore.")
    else:
        md.append(f"You have attachments: a reference image and this markdown file with {char_name}'s data and lore.")
    md.append("")
    md.append("CRITICAL: Use the reference images ONLY for exact visual likeness — face, hairstyle, outfit colors/silhouette, weapon design and any companion creatures.")
    md.append("Do NOT invent, hallucinate, or create any visual elements, strange anatomy, or non-sensical details that are not grounded in the reference images. Stay completely faithful to the provided designs.")
    md.append("Do not copy any UI, text or layout from the reference images.")
    md.append("")
    md.append("Use the DATA section below as the ONLY source for every piece of text that appears on the card. Extract and use:")
    display_name = char_name + (f' \"{nickname}\"' if nickname and nickname != char_name else "")
    md.append(f"- Full name: {display_name}")
    md.append(f"- Element: {element_name}")
    md.append(f"- Weapon type: {weapon_type}")
    md.append(f"- Rarity: {rarity} stars")
    md.append(f"- Team role tags: {tags_str}")
    if attrs_str:
        md.append(f"- Recommended stats: {attrs_str}")
    if echo_str:
        md.append(f"- Recommended echo: {echo_str}")
    if team_str:
        md.append(f"- Recommended team: {team_str}")
    if skills_str:
        md.append(f"- Skill priority: {skills_str}")
    if weapon_name:
        md.append(f"- Signature weapon: {weapon_name}")
    md.append(f'- Short lore line (use this or pick a better one from the stories below): \"{desc_short}\"')
    md.append("Do not invent, translate loosely, or add any stat, number or fact that is not present in this file.")
    md.append("")
    md.append("DESIGN DIRECTION — premium editorial character card, awwwards-level quality, NOT a generic gacha-game gold-bordered card:")
    md.append("- Vertical card, ratio 2:3.")
    md.append("- Full illustrated character art as the dominant visual, matching the reference image's palette.")
    md.append(f"- Color palette: {palette['background']} background, {palette['accent']} accent, {palette['text']} text, {palette['secondary']} secondary — nothing else competes with these.")
    md.append(f"- Visual motif tied to the {element_name} element: {palette['motif']} — used subtly, with restraint, not as clutter.")
    md.append(f"- Overall mood: {palette['mood']}.")
    md.append("- Typography-led composition: pair an elegant italic serif for the name (large, confident) with a clean geometric sans for role tags, and a technical monospace for small labels/eyebrow text — this contrast should feel intentional and refined, not decorative.")
    md.append("- Replace ornate gold RPG borders with a minimal architectural frame: a single thin hairline border in the accent color, with one asymmetric accent (e.g. a corner bracket, a vertical rule, or a small element/rarity mark) rather than a symmetric ornamental frame.")
    md.append("- Generous negative space; do not fill every inch of the card with decoration. Let the art breathe.")
    md.append(f"- Rarity should be shown with quiet confidence (e.g. {rarity} small minimal star marks or a small numeral), not a loud banner.")
    md.append("- No watermark, no fake logos, no placeholder text — every word on the card must come from the data below.")
    md.append("")
    md.append("The final image should look like a museum-quality character showcase card an award-winning design studio would make — refined, confident, a little unexpected — not a template RPG card.")
    md.append("```")
    md.append("")

    # ---- DICA ----
    md.append("### Se o resultado sair genérico")
    md.append('Peça pro modelo, na sequência: **"remove any gold ornamental border, keep the frame minimal and asymmetric, and make the typography the hero of the composition, not decoration."** Isso costuma puxar mais pro lado editorial.')
    md.append("")

    # ---- DADOS BRUTOS ----
    md.append("---")
    md.append("")
    md.append("# DADOS DO PERSONAGEM")
    md.append("")
    raw_content = build_raw_markdown(char_data, weapon_data, version, guide_data)
    md.append(raw_content)

    return "\n".join(md)


# =============================================================================
# MODO --list / --pick / --compare : helpers de exibição e coleta de IDs
# =============================================================================

def _print_character_table(index: dict):
    """Imprime a tabela de personagens (reutilizada em --list, --pick e --compare)."""
    print(f"\n{'ID':<6} {'Nome':<20} {'Apelido':<35} Raridade  Elemento")
    print("-" * 90)
    for char_id, data in sorted(index.items(), key=lambda kv: int(kv[0])):
        name = data.get("en") or "(sem nome / Rover variant)"
        nick = data.get("nickname", "")
        rarity = data.get("rank", "?")
        element = describe_element(data.get("element"))
        print(f"{char_id:<6} {name:<20} {nick:<35} {rarity:<8}  {element}")
    print(f"\nTotal: {len(index)} personagens.")


def list_characters(session: requests.Session, version: str):
    index = fetch_character_index(session, version)
    if not index:
        print("Não consegui buscar o índice de personagens.")
        return
    _print_character_table(index)
    print("\nCopie os IDs que quiser para CHARACTERS_TO_SCRAPE no topo do script.")


def pick_characters_interactively(session: requests.Session, version: str) -> list:
    """
    Mostra a lista de personagens e deixa escolher os IDs digitando ali
    mesmo no terminal, sem precisar editar o script. Retorna uma lista no
    mesmo formato de CHARACTERS_TO_SCRAPE.
    """
    index = fetch_character_index(session, version)
    if not index:
        print("Não consegui buscar o índice de personagens.")
        return []

    _print_character_table(index)

    print(
        "\nDigite os IDs dos personagens que você quer, separados por vírgula"
        "\n(ex: 1607, 1610, 1109). Deixe em branco para cancelar."
    )
    raw = input("IDs: ").strip()
    if not raw:
        print("Nenhum ID informado, cancelando.")
        return []

    chosen_ids = []
    for piece in raw.split(","):
        piece = piece.strip()
        if not piece:
            continue
        if not piece.isdigit():
            print(f"  [aviso] '{piece}' não parece um ID válido, ignorando.")
            continue
        if piece not in index:
            print(f"  [aviso] ID {piece} não encontrado no índice, ignorando.")
            continue
        chosen_ids.append(piece)

    if not chosen_ids:
        print("Nenhum ID válido selecionado.")
        return []

    selections = []
    for char_id in chosen_ids:
        data = index[char_id]
        name = data.get("en") or data.get("nickname") or char_id
        print(f"\nPersonagem: {name} (ID {char_id})")
        weapon_raw = input(
            "  ID da arma de assinatura (Enter para detectar automaticamente): "
        ).strip()
        weapon_id = int(weapon_raw) if weapon_raw.isdigit() else None
        selections.append({
            "character_id": int(char_id),
            "weapon_id": weapon_id,
            "name_hint": name,
        })

    return selections



# =============================================================================
# MODO --compare : compara 2+ personagens lado a lado num único .md
# =============================================================================

COMPARE_PRE_PROMPT_TEMPLATE = """# PROMPT PARA IA — COMPARAÇÃO DE PERSONAGENS — NÃO APAGUE ESTA SEÇÃO

Você é um especialista em Wuthering Waves. Abaixo estão os dados brutos
(extraídos diretamente do datamine do jogo) de {n} personagens para comparação:
{char_names_list}

Com base SOMENTE nas informações fornecidas abaixo, faça uma análise comparativa
cobrindo os seguintes pontos:

1. **Resumo rápido de cada personagem** (papel, elemento, tipo de arma, gimmick central).
2. **Comparação de kits**:
   - Qual tem o kit mais autossuficiente (menos dependente de outros personagens)?
   - Qual tem maior potencial de dano bruto vs utilidade para o time?
   - Qual tem a rotação mais simples / mais complexa?
3. **Sequências (Chains)**:
   - Para cada personagem: quais sequências são realmente transformadoras vs apenas numéricas?
   - Comparando S0 vs S0: quem performa melhor sem investimento em cópias?
   - Se houver budget para puxar sequências: em qual personagem vale mais investir?
4. **Armas de assinatura** (se disponíveis):
   - Qual arma é mais essencial para o seu portador?
   - Alguma delas funciona em mais de um dos personagens comparados?
5. **Sinergia entre eles**: faz sentido usá-los no mesmo time? Por quê (ou por quê não)?
6. **Veredicto — quem puxar primeiro?**
   - Considerando quem está em lacunas de meta, versatilidade e curva de investimento.
   - Responda diretamente, com justificativa técnica baseada nos dados.

# Regras
- Não invente informações ou porcentagens.
- Baseie todas as conclusões nas informações fornecidas.
- Seja direto e use bullet points onde fizer sentido.
- Escreva como uma análise de theorycrafting, não como review casual.

---

# DADOS BRUTOS DOS PERSONAGENS

"""


def _fetch_char_and_weapon(
    session: requests.Session,
    version: str,
    entry: dict,
) -> tuple:
    """Busca dados do personagem e sua arma. Retorna (char_data, weapon_data)."""
    char_id = entry["character_id"]
    weapon_id = entry.get("weapon_id")
    hint = entry.get("name_hint", char_id)

    print(f"-> Personagem {char_id} ({hint})")
    char_data = fetch_character_detail(session, version, char_id)
    time.sleep(REQUEST_DELAY_SECONDS)

    if not char_data:
        print(f"   [erro] não consegui obter dados do personagem {char_id}, pulando.")
        return None, None

    if not weapon_id:
        weapon_id = extract_signature_weapon_id(char_data)
        if weapon_id:
            print(f"   -> Arma de assinatura detectada automaticamente: {weapon_id}")

    weapon_data = None
    if weapon_id:
        print(f"   -> Buscando arma {weapon_id}")
        weapon_data = fetch_weapon_detail(session, version, weapon_id)
        time.sleep(REQUEST_DELAY_SECONDS)
        if not weapon_data:
            print(f"   [aviso] não consegui obter dados da arma {weapon_id}.")

    return char_data, weapon_data


def _build_char_block(char_data: dict, weapon_data, version: str, index: int) -> list:
    """
    Retorna as linhas markdown de UM personagem para uso dentro do arquivo
    de comparação (sem o pre-prompt e sem lore, para não inchaçar demais).
    """
    char_name = char_data.get("name") or char_data.get("nick_name") or "Desconhecido"
    nickname = char_data.get("nick_name", "")
    rarity = char_data.get("rarity", "?")
    weapon_type = describe_weapon_type(char_data.get("weapon"))
    element = describe_element(char_data.get("element"))
    desc = clean_text(char_data.get("desc", ""))

    info = char_data.get("chara_info", {}) or {}
    tags = extract_tags(char_data)
    forte_summary = extract_forte_summary(char_data)
    skills = extract_skills(char_data)
    chains = extract_chains(char_data)

    md = []
    sep = "=" * 70
    md.append(sep)
    md.append(f"## Personagem {index}: {char_name}" + (
        f' "{nickname}"' if nickname and nickname != char_name else ""
    ))
    md.append(sep)
    md.append("")
    md.append(f"- **Raridade:** {rarity} estrelas")
    md.append(f"- **Elemento:** {element}")
    md.append(f"- **Tipo de arma:** {weapon_type}")
    if info.get("country"):
        md.append(f"- **Região/País:** {info.get('country')}")
    if info.get("influence"):
        md.append(f"- **Facção/Influência:** {info.get('influence')}")
    md.append("")

    if desc:
        md.append("### Descrição")
        md.append(desc)
        md.append("")

    if tags:
        md.append("### Tags de papel no time")
        for t in tags:
            md.append(f"- {t}")
        md.append("")

    if info.get("talent_name"):
        md.append(f"### Forte (lore): {info.get('talent_name')}")
        if info.get("talent_doc"):
            md.append(clean_text(info.get("talent_doc")))
        md.append("")

    if forte_summary:
        md.append("### Como o Forte funciona (mecânica)")
        md.append("")
        for paragraph in forte_summary:
            md.append(paragraph)
            md.append("")

    if skills:
        md.append("### Skills / Kit")
        for s in skills:
            md.append(f"**{s['name']}** ({s['type']})")
            md.append("")
            md.append(s["desc"])
            md.append("")

    if chains:
        md.append("### Sequências (Chains / Constelações)")
        md.append("")
        for c in chains:
            md.append(f"**Sequência {c['id']}: {c['name']}**")
            md.append("")
            md.append(c["desc"])
            md.append("")

    if weapon_data:
        md.append("### Arma de Assinatura")
        md.append("")
        w_name = weapon_data.get("name", "Desconhecida")
        w_rarity = weapon_data.get("rarity", "?")
        w_desc = clean_text(weapon_data.get("desc", ""))
        md.append(f"- **Nome:** {w_name}")
        md.append(f"- **Raridade:** {w_rarity} estrelas")
        if weapon_data.get("weapon_type") is not None:
            md.append(f"- **Tipo:** {describe_weapon_type(weapon_data.get('weapon_type'))}")
        md.append("")
        if w_desc:
            md.append("#### Descrição / Lore da arma")
            md.append(w_desc)
            md.append("")
        weapon_skill = (
            weapon_data.get("skill")
            or weapon_data.get("effect")
            or weapon_data.get("skill_desc")
        )
        if weapon_skill:
            md.append("#### Efeito (Passiva da arma)")
            if isinstance(weapon_skill, dict):
                w_skill_name = weapon_skill.get("name", "")
                w_skill_desc = clean_text(
                    weapon_skill.get("simple_desc") or weapon_skill.get("desc", "")
                )
                if w_skill_name:
                    md.append(f"**{w_skill_name}**")
                md.append(w_skill_desc)
            else:
                md.append(clean_text(str(weapon_skill)))
            md.append("")

    return md


def build_compare_markdown(entries_data: list, version: str) -> str:
    """
    Monta o .md de comparação.

    entries_data: lista de (char_data, weapon_data) — pares já buscados.
    """
    names = []
    for char_data, _ in entries_data:
        n = char_data.get("name") or char_data.get("nick_name") or "Desconhecido"
        names.append(f"**{n}**")

    char_names_list = ", ".join(names[:-1]) + f" e {names[-1]}" if len(names) > 1 else names[0]

    md = []
    md.append(COMPARE_PRE_PROMPT_TEMPLATE.format(
        n=len(entries_data),
        char_names_list=char_names_list,
    ))

    for i, (char_data, weapon_data) in enumerate(entries_data, start=1):
        md.extend(_build_char_block(char_data, weapon_data, version, i))
        md.append("")

    md.append("---")
    md.append(
        f"*Dados extraídos de static.nanoka.cc — "
        f"Wuthering Waves v{version} — idioma: {LANGUAGE}*"
    )
    return "\n".join(md)


def pick_ids_for_compare(session: requests.Session, version: str, argv_ids: list) -> list:
    """
    Resolve os IDs para o modo --compare.

    Se argv_ids foi passado (ex: --compare 1607 1610), usa direto.
    Caso contrário, exibe a tabela e pede pro usuário digitar.
    Retorna lista no mesmo formato de CHARACTERS_TO_SCRAPE.
    """
    index = fetch_character_index(session, version)
    if not index:
        print("Não consegui buscar o índice de personagens.")
        return []

    if argv_ids:
        raw_ids = argv_ids
    else:
        _print_character_table(index)
        print(
            "\nDigite os IDs dos personagens que você quer COMPARAR (mínimo 2),"
            "\nseparados por vírgula (ex: 1607, 1610, 1109)."
        )
        raw = input("IDs: ").strip()
        if not raw:
            print("Nenhum ID informado, cancelando.")
            return []
        raw_ids = [p.strip() for p in raw.split(",")]

    chosen_ids = []
    for piece in raw_ids:
        piece = str(piece).strip()
        if not piece:
            continue
        if not piece.isdigit():
            print(f"  [aviso] \'{piece}\' não parece um ID válido, ignorando.")
            continue
        if piece not in index:
            print(f"  [aviso] ID {piece} não encontrado no índice, ignorando.")
            continue
        chosen_ids.append(piece)

    if len(chosen_ids) < 2:
        print("É necessário pelo menos 2 IDs válidos para comparação.")
        return []

    selections = []
    for char_id in chosen_ids:
        data = index[char_id]
        name = data.get("en") or data.get("nickname") or char_id
        selections.append({
            "character_id": int(char_id),
            "weapon_id": None,  # sempre detecta automaticamente no compare
            "name_hint": name,
        })
    return selections


def compare_characters(session: requests.Session, version: str, characters: list):
    """Busca todos os personagens e gera o .md de comparação."""
    out_dir = Path(OUTPUT_DIR)
    out_dir.mkdir(parents=True, exist_ok=True)

    print(f"\nComparando {len(characters)} personagem(ns) — versão {version}...\n")

    entries_data = []
    names_slug = []
    for entry in characters:
        char_data, weapon_data = _fetch_char_and_weapon(session, version, entry)
        if char_data is None:
            continue
        entries_data.append((char_data, weapon_data))
        char_name = char_data.get("name") or str(entry["character_id"])
        names_slug.append(slugify(char_name))

    if len(entries_data) < 2:
        print("[erro] Menos de 2 personagens carregados com sucesso; não é possível comparar.")
        return

    md_content = build_compare_markdown(entries_data, version)

    filename = "compare_" + "_vs_".join(names_slug) + ".md"
    filepath = out_dir / filename
    filepath.write_text(md_content, encoding="utf-8")
    print(f"\n[ok] Comparação salva em {filepath}")
    print("Abra o arquivo e cole o conteúdo inteiro numa IA para a análise comparativa.")


# =============================================================================
# MAIN
# =============================================================================

def _fetch_and_prepare(session, version, entry):
    """Busca dados do personagem, arma e atributos, retorna tuple."""
    char_id = entry["character_id"]
    weapon_id = entry.get("weapon_id")
    hint = entry.get("name_hint", char_id)

    print(f"-> Personagem {char_id} ({hint})")
    char_data = fetch_character_detail(session, version, char_id)
    time.sleep(REQUEST_DELAY_SECONDS)

    if not char_data:
        print(f"   [erro] não consegui obter dados do personagem {char_id}, pulando.")
        return None, None, None

    if not weapon_id:
        weapon_id = extract_signature_weapon_id(char_data)
        if weapon_id:
            print(f"   -> Arma de assinatura detectada automaticamente: {weapon_id}")

    weapon_data = None
    if weapon_id:
        print(f"   -> Buscando arma {weapon_id}")
        weapon_data = fetch_weapon_detail(session, version, weapon_id)
        time.sleep(REQUEST_DELAY_SECONDS)
        if not weapon_data:
            print(f"   [aviso] não consegui obter dados da arma {weapon_id}.")

    guide_data = fetch_guide_recommendations(session, str(char_id))
    if guide_data and (guide_data.get("attributes") or guide_data.get("echo")):
        print("   -> Guia oficial carregado (Atributos, Echo, Time).")

    return char_data, weapon_data, guide_data


def _download_all_images(session, index, char_id, slug, char_dir, weapon_data):
    """Baixa todas as imagens (personagem + arma) e retorna dict de filenames."""
    image_files = {}
    index_entry = index.get(str(char_id), {})
    if index_entry:
        image_files = download_character_images(session, index_entry, slug, char_dir)

    if weapon_data:
        weapon_fname = download_weapon_image(session, weapon_data, slug, char_dir)
        if weapon_fname:
            image_files["weapon"] = weapon_fname

    return image_files


def scrape_characters(session: requests.Session, version: str, characters: list):
    """Modo padrão / --pick: gera análise + raw + card + imagens, tudo em pasta por personagem."""
    out_dir = Path(OUTPUT_DIR)
    out_dir.mkdir(parents=True, exist_ok=True)

    index = fetch_character_index(session, version)

    print(f"\nRaspando {len(characters)} personagem(ns) — versão {version}...\n")

    for entry in characters:
        char_data, weapon_data, guide_data = _fetch_and_prepare(session, version, entry)
        if char_data is None:
            continue

        char_id = entry["character_id"]
        char_name = char_data.get("name") or str(char_id)
        slug = f"{char_id}_{slugify(char_name)}"

        # Cria pasta do personagem
        char_dir = out_dir / slug
        char_dir.mkdir(parents=True, exist_ok=True)

        # Baixa imagens
        image_files = _download_all_images(session, index, char_id, slug, char_dir, weapon_data)

        # .md de análise (com prompt — comportamento existente)
        md_content = build_markdown(char_data, weapon_data, version, guide_data)
        filepath = char_dir / f"{slug}.md"
        filepath.write_text(md_content, encoding="utf-8")
        print(f"   [ok] análise salva em {filepath}")

        # .md bruto (sem prompt)
        raw_content = build_raw_markdown(char_data, weapon_data, version, guide_data)
        raw_path = char_dir / f"{slug}_raw.md"
        raw_path.write_text(raw_content, encoding="utf-8")
        print(f"   [ok] dados brutos salvos em {raw_path}")

        # .md card Awwwards (prompt + dados autocontido)
        card_content = build_card_markdown(char_data, weapon_data, version, image_files, guide_data)
        card_path = char_dir / f"{slug}_card.md"
        card_path.write_text(card_content, encoding="utf-8")
        print(f"   [ok] card Awwwards salvo em {card_path}")
        print()

    print("Concluído! Os arquivos estão na pasta:", out_dir.resolve())


def card_characters(session: requests.Session, version: str, characters: list):
    """Modo --card: baixa imagens + gera _raw.md e _card.md (sem análise)."""
    out_dir = Path(OUTPUT_DIR)
    out_dir.mkdir(parents=True, exist_ok=True)

    index = fetch_character_index(session, version)

    print(f"\nGerando cards para {len(characters)} personagem(ns) — versão {version}...\n")

    for entry in characters:
        char_data, weapon_data, guide_data = _fetch_and_prepare(session, version, entry)
        if char_data is None:
            continue

        char_id = entry["character_id"]
        char_name = char_data.get("name") or str(char_id)
        slug = f"{char_id}_{slugify(char_name)}"

        # Cria pasta do personagem
        char_dir = out_dir / slug
        char_dir.mkdir(parents=True, exist_ok=True)

        # Baixa imagens
        image_files = _download_all_images(session, index, char_id, slug, char_dir, weapon_data)

        # .md bruto (sem prompt)
        raw_content = build_raw_markdown(char_data, weapon_data, version, guide_data)
        raw_path = char_dir / f"{slug}_raw.md"
        raw_path.write_text(raw_content, encoding="utf-8")
        print(f"   [ok] dados brutos salvos em {raw_path}")

        # .md card Awwwards (prompt + dados autocontido)
        card_content = build_card_markdown(char_data, weapon_data, version, image_files, guide_data)
        card_path = char_dir / f"{slug}_card.md"
        card_path.write_text(card_content, encoding="utf-8")
        print(f"   [ok] card Awwwards salvo em {card_path}")
        print()

    print("Concluído! Os arquivos estão na pasta:", out_dir.resolve())


def main():
    session = get_session()

    version = GAME_VERSION
    if AUTO_DETECT_VERSION:
        version = detect_game_version(session)

    if "--list" in sys.argv:
        list_characters(session, version)
        return

    if "--pick" in sys.argv:
        characters = pick_characters_interactively(session, version)
        if not characters:
            return
        scrape_characters(session, version, characters)
        return

    if "--card" in sys.argv:
        characters = pick_characters_interactively(session, version)
        if not characters:
            return
        card_characters(session, version, characters)
        return

    if "--compare" in sys.argv:
        # IDs podem ser passados direto: python wuwa_scraper.py --compare 1607 1610
        idx = sys.argv.index("--compare")
        argv_ids = [a for a in sys.argv[idx + 1:] if not a.startswith("--")]
        characters = pick_ids_for_compare(session, version, argv_ids)
        if not characters:
            return
        compare_characters(session, version, characters)
        return

    if not CHARACTERS_TO_SCRAPE:
        print(
            "Nenhum personagem configurado em CHARACTERS_TO_SCRAPE.\n"
            "Use 'python wuwa_scraper.py --pick' para escolher interativamente,\n"
            "    'python wuwa_scraper.py --card' para gerar cards Awwwards,\n"
            "ou edite a lista CHARACTERS_TO_SCRAPE no topo do script."
        )
        return

    scrape_characters(session, version, CHARACTERS_TO_SCRAPE)


if __name__ == "__main__":
    main()
